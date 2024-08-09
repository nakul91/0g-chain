package dasigners_test

import (
	"math/big"
	"strings"
	"testing"

	"github.com/0glabs/0g-chain/crypto/bn254util"
	dasignersprecompile "github.com/0glabs/0g-chain/precompiles/dasigners"
	"github.com/0glabs/0g-chain/precompiles/testutil"
	"github.com/0glabs/0g-chain/x/dasigners/v1"
	"github.com/0glabs/0g-chain/x/dasigners/v1/keeper"
	dasignerskeeper "github.com/0glabs/0g-chain/x/dasigners/v1/keeper"
	"github.com/0glabs/0g-chain/x/dasigners/v1/types"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/consensys/gnark-crypto/ecc/bn254"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
	"github.com/stretchr/testify/suite"

	"cosmossdk.io/math"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/evmos/ethermint/crypto/ethsecp256k1"
)

type DASignersTestSuite struct {
	testutil.PrecompileTestSuite

	abi             abi.ABI
	addr            common.Address
	dasigners       *dasignersprecompile.DASignersPrecompile
	dasignerskeeper dasignerskeeper.Keeper
	signerOne       *testutil.TestSigner
	signerTwo       *testutil.TestSigner
}

func (suite *DASignersTestSuite) AddDelegation(from string, to string, amount math.Int) {
	accAddr, err := sdk.AccAddressFromHexUnsafe(from)
	suite.Require().NoError(err)
	valAddr, err := sdk.ValAddressFromHex(to)
	suite.Require().NoError(err)
	validator, found := suite.StakingKeeper.GetValidator(suite.Ctx, valAddr)
	if !found {
		consPriv, err := ethsecp256k1.GenerateKey()
		suite.Require().NoError(err)
		newValidator, err := stakingtypes.NewValidator(valAddr, consPriv.PubKey(), stakingtypes.Description{})
		suite.Require().NoError(err)
		validator = newValidator
	}
	validator.Tokens = validator.Tokens.Add(amount)
	validator.DelegatorShares = validator.DelegatorShares.Add(amount.ToLegacyDec())
	suite.StakingKeeper.SetValidator(suite.Ctx, validator)
	bonded := suite.dasignerskeeper.GetDelegatorBonded(suite.Ctx, accAddr)
	suite.StakingKeeper.SetDelegation(suite.Ctx, stakingtypes.Delegation{
		DelegatorAddress: accAddr.String(),
		ValidatorAddress: valAddr.String(),
		Shares:           bonded.Add(amount).ToLegacyDec(),
	})
}

func (suite *DASignersTestSuite) SetupTest() {
	suite.PrecompileTestSuite.SetupTest()

	suite.dasignerskeeper = suite.App.GetDASignersKeeper()

	suite.addr = common.HexToAddress(dasignersprecompile.PrecompileAddress)

	precompiles := suite.EvmKeeper.GetPrecompiles()
	precompile, ok := precompiles[suite.addr]
	suite.Assert().EqualValues(ok, true)
	suite.dasigners = precompile.(*dasignersprecompile.DASignersPrecompile)

	suite.signerOne = testutil.GenSigner()
	suite.signerTwo = testutil.GenSigner()
	abi, err := abi.JSON(strings.NewReader(dasignersprecompile.DASignersABI))
	suite.Assert().NoError(err)
	suite.abi = abi
}

func (suite *DASignersTestSuite) runTx(input []byte, signer *testutil.TestSigner, gas uint64) ([]byte, error) {
	contract := vm.NewPrecompile(vm.AccountRef(signer.Addr), vm.AccountRef(suite.addr), big.NewInt(0), gas)
	contract.Input = input

	msgEthereumTx := evmtypes.NewTx(suite.EvmKeeper.ChainID(), 0, &suite.addr, big.NewInt(0), gas, big.NewInt(0), big.NewInt(0), big.NewInt(0), input, nil)
	msgEthereumTx.From = signer.HexAddr
	err := msgEthereumTx.Sign(suite.EthSigner, signer.Signer)
	suite.Assert().NoError(err, "failed to sign Ethereum message")

	proposerAddress := suite.Ctx.BlockHeader().ProposerAddress
	cfg, err := suite.EvmKeeper.EVMConfig(suite.Ctx, proposerAddress, suite.EvmKeeper.ChainID())
	suite.Assert().NoError(err, "failed to instantiate EVM config")

	msg, err := msgEthereumTx.AsMessage(suite.EthSigner, big.NewInt(0))
	suite.Assert().NoError(err, "failed to instantiate Ethereum message")

	evm := suite.EvmKeeper.NewEVM(suite.Ctx, msg, cfg, nil, suite.Statedb)
	precompiles := suite.EvmKeeper.GetPrecompiles()
	evm.WithPrecompiles(precompiles, []common.Address{suite.addr})

	return suite.dasigners.Run(evm, contract, false)
}

func (suite *DASignersTestSuite) registerSigner(testSigner *testutil.TestSigner, sk *big.Int) *types.Signer {
	pkG1 := new(bn254.G1Affine).ScalarMultiplication(bn254util.GetG1Generator(), sk)
	pkG2 := new(bn254.G2Affine).ScalarMultiplication(bn254util.GetG2Generator(), sk)
	hash := types.PubkeyRegistrationHash(testSigner.Addr, big.NewInt(8888))
	signature := new(bn254.G1Affine).ScalarMultiplication(hash, sk)
	signer := &types.Signer{
		Account:  testSigner.HexAddr,
		Socket:   "0.0.0.0:1234",
		PubkeyG1: bn254util.SerializeG1(pkG1),
		PubkeyG2: bn254util.SerializeG2(pkG2),
	}

	input, err := suite.abi.Pack(
		"registerSigner",
		dasignersprecompile.NewIDASignersSignerDetail(signer),
		dasignersprecompile.NewBN254G1Point(bn254util.SerializeG1(signature)),
	)
	suite.Assert().NoError(err)

	oldLogs := suite.Statedb.Logs()
	_, err = suite.runTx(input, testSigner, 10000000)
	suite.Assert().NoError(err)
	logs := suite.Statedb.Logs()
	suite.Assert().EqualValues(len(logs), len(oldLogs)+2)

	_, err = suite.abi.Unpack("SocketUpdated", logs[len(logs)-1].Data)
	suite.Assert().NoError(err)
	_, err = suite.abi.Unpack("NewSigner", logs[len(logs)-2].Data)
	suite.Assert().NoError(err)
	return signer
}

func (suite *DASignersTestSuite) updateSocket(testSigner *testutil.TestSigner, signer *types.Signer) {
	input, err := suite.abi.Pack(
		"updateSocket",
		"0.0.0.0:2345",
	)
	suite.Assert().NoError(err)

	oldLogs := suite.Statedb.Logs()
	_, err = suite.runTx(input, testSigner, 10000000)
	suite.Assert().NoError(err)
	logs := suite.Statedb.Logs()
	suite.Assert().EqualValues(len(logs), len(oldLogs)+1)

	_, err = suite.abi.Unpack("SocketUpdated", logs[len(logs)-1].Data)
	suite.Assert().NoError(err)

	signer.Socket = "0.0.0.0:2345"
}

func (suite *DASignersTestSuite) registerEpoch(testSigner *testutil.TestSigner, sk *big.Int) {
	hash := types.EpochRegistrationHash(common.HexToAddress(testSigner.HexAddr), 1, big.NewInt(8888))
	signature := new(bn254.G1Affine).ScalarMultiplication(hash, sk)

	input, err := suite.abi.Pack(
		"registerNextEpoch",
		dasignersprecompile.NewBN254G1Point(bn254util.SerializeG1(signature)),
	)
	suite.Assert().NoError(err)

	_, err = suite.runTx(input, testSigner, 10000000)
	suite.Assert().NoError(err)
}

func (suite *DASignersTestSuite) queryEpochNumber(testSigner *testutil.TestSigner) {
	input, err := suite.abi.Pack(
		"epochNumber",
	)
	suite.Assert().NoError(err)

	bz, err := suite.runTx(input, testSigner, 10000000)
	suite.Assert().NoError(err)
	out, err := suite.abi.Methods["epochNumber"].Outputs.Unpack(bz)
	suite.Assert().NoError(err)
	suite.Assert().EqualValues(out[0], big.NewInt(1))
}

func (suite *DASignersTestSuite) queryQuorumCount(testSigner *testutil.TestSigner) {
	input, err := suite.abi.Pack(
		"quorumCount",
		big.NewInt(1),
	)
	suite.Assert().NoError(err)

	bz, err := suite.runTx(input, testSigner, 10000000)
	suite.Assert().NoError(err)
	out, err := suite.abi.Methods["quorumCount"].Outputs.Unpack(bz)
	suite.Assert().NoError(err)
	suite.Assert().EqualValues(out[0], big.NewInt(1))
}

func (suite *DASignersTestSuite) queryGetSigner(testSigner *testutil.TestSigner, answer []*types.Signer) {
	input, err := suite.abi.Pack(
		"getSigner",
		[]common.Address{suite.signerOne.Addr, suite.signerTwo.Addr},
	)
	suite.Assert().NoError(err)

	bz, err := suite.runTx(input, testSigner, 10000000)
	suite.Assert().NoError(err)
	out, err := suite.abi.Methods["getSigner"].Outputs.Unpack(bz)
	suite.Assert().NoError(err)
	res := make([]dasignersprecompile.IDASignersSignerDetail, 0)
	for _, s := range answer {
		res = append(res, dasignersprecompile.NewIDASignersSignerDetail(s))
	}
	suite.Assert().EqualValues(out[0], res)
}

func (suite *DASignersTestSuite) queryIsSigner(testSigner *testutil.TestSigner) {
	input, err := suite.abi.Pack(
		"isSigner",
		suite.signerOne.Addr,
	)
	suite.Assert().NoError(err)

	bz, err := suite.runTx(input, testSigner, 10000000)
	suite.Assert().NoError(err)
	out, err := suite.abi.Methods["isSigner"].Outputs.Unpack(bz)
	suite.Assert().NoError(err)
	suite.Assert().EqualValues(out[0], true)
}

func (suite *DASignersTestSuite) queryRegisteredEpoch(testSigner *testutil.TestSigner, account common.Address, epoch *big.Int) bool {
	input, err := suite.abi.Pack(
		"registeredEpoch",
		account,
		epoch,
	)
	suite.Assert().NoError(err)

	bz, err := suite.runTx(input, testSigner, 10000000)
	suite.Assert().NoError(err)
	out, err := suite.abi.Methods["registeredEpoch"].Outputs.Unpack(bz)
	suite.Assert().NoError(err)
	return out[0].(bool)
}

func (suite *DASignersTestSuite) queryGetQuorum(testSigner *testutil.TestSigner) []common.Address {
	input, err := suite.abi.Pack(
		"getQuorum",
		big.NewInt(1),
		big.NewInt(0),
	)
	suite.Assert().NoError(err)

	bz, err := suite.runTx(input, testSigner, 10000000)
	suite.Assert().NoError(err)
	out, err := suite.abi.Methods["getQuorum"].Outputs.Unpack(bz)
	suite.Assert().NoError(err)
	return out[0].([]common.Address)
}

func (suite *DASignersTestSuite) queryGetQuorumRow(testSigner *testutil.TestSigner, row uint32) common.Address {
	input, err := suite.abi.Pack(
		"getQuorumRow",
		big.NewInt(1),
		big.NewInt(0),
		row,
	)
	suite.Assert().NoError(err)

	bz, err := suite.runTx(input, testSigner, 10000000)
	suite.Assert().NoError(err)
	out, err := suite.abi.Methods["getQuorumRow"].Outputs.Unpack(bz)
	suite.Assert().NoError(err)
	return out[0].(common.Address)
}

func (suite *DASignersTestSuite) queryGetAggPkG1(testSigner *testutil.TestSigner, bitmap []byte) struct {
	AggPkG1 dasignersprecompile.BN254G1Point
	Total   *big.Int
	Hit     *big.Int
} {
	input, err := suite.abi.Pack(
		"getAggPkG1",
		big.NewInt(1),
		big.NewInt(0),
		bitmap,
	)
	suite.Assert().NoError(err)

	bz, err := suite.runTx(input, testSigner, 10000000)
	suite.Assert().NoError(err)
	out, err := suite.abi.Methods["getAggPkG1"].Outputs.Unpack(bz)
	suite.Assert().NoError(err)
	return struct {
		AggPkG1 dasignersprecompile.BN254G1Point
		Total   *big.Int
		Hit     *big.Int
	}{
		AggPkG1: out[0].(dasignersprecompile.BN254G1Point),
		Total:   out[1].(*big.Int),
		Hit:     out[2].(*big.Int),
	}
}

func (suite *DASignersTestSuite) Test_DASigners() {
	// suite.App.InitializeFromGenesisStates()
	dasigners.InitGenesis(suite.Ctx, suite.dasignerskeeper, *types.DefaultGenesisState())
	// add delegation
	params := suite.dasignerskeeper.GetParams(suite.Ctx)
	suite.AddDelegation(suite.signerOne.HexAddr, suite.signerOne.HexAddr, keeper.BondedConversionRate.Mul(sdk.NewIntFromUint64(params.TokensPerVote)))
	suite.AddDelegation(suite.signerTwo.HexAddr, suite.signerOne.HexAddr, keeper.BondedConversionRate.Mul(sdk.NewIntFromUint64(params.TokensPerVote)).Mul(sdk.NewIntFromUint64(2)))
	// tx test
	signer1 := suite.registerSigner(suite.signerOne, big.NewInt(1))
	signer2 := suite.registerSigner(suite.signerTwo, big.NewInt(11))
	suite.updateSocket(suite.signerOne, signer1)
	suite.updateSocket(suite.signerTwo, signer2)
	suite.registerEpoch(suite.signerOne, big.NewInt(1))
	suite.registerEpoch(suite.signerTwo, big.NewInt(11))
	// move to next epoch
	suite.Ctx = suite.Ctx.WithBlockHeight(int64(params.EpochBlocks) * 1)
	suite.dasignerskeeper.BeginBlock(suite.Ctx, abci.RequestBeginBlock{})
	// query test
	suite.queryEpochNumber(suite.signerOne)
	suite.queryQuorumCount(suite.signerOne)
	suite.queryGetSigner(suite.signerOne, []*types.Signer{signer1, signer2})
	suite.queryIsSigner(suite.signerOne)
	suite.Assert().EqualValues(suite.queryRegisteredEpoch(suite.signerOne, suite.signerOne.Addr, big.NewInt(1)), true)
	suite.Assert().EqualValues(suite.queryRegisteredEpoch(suite.signerOne, suite.signerTwo.Addr, big.NewInt(1)), true)
	suite.Assert().EqualValues(suite.queryRegisteredEpoch(suite.signerOne, suite.signerOne.Addr, big.NewInt(2)), false)
	suite.Assert().EqualValues(suite.queryRegisteredEpoch(suite.signerOne, suite.signerTwo.Addr, big.NewInt(0)), false)

	quorum := suite.queryGetQuorum(suite.signerOne)
	suite.Assert().EqualValues(len(quorum), params.EncodedSlices)
	cnt := map[common.Address]int{suite.signerOne.Addr: 0, suite.signerTwo.Addr: 0}
	onePos := len(quorum)
	twoPos := len(quorum)
	for i, v := range quorum {
		suite.Assert().EqualValues(suite.queryGetQuorumRow(suite.signerOne, uint32(i)), v)
		cnt[v] += 1
		if v == suite.signerOne.Addr {
			onePos = min(onePos, i)
		} else {
			twoPos = min(twoPos, i)
		}
	}
	suite.Assert().EqualValues(cnt[suite.signerOne.Addr], len(quorum)/3)
	suite.Assert().EqualValues(cnt[suite.signerTwo.Addr], len(quorum)*2/3)

	bitMap := make([]byte, len(quorum)/8)
	bitMap[onePos/8] |= 1 << (onePos % 8)
	suite.Assert().EqualValues(suite.queryGetAggPkG1(suite.signerOne, bitMap), struct {
		AggPkG1 dasignersprecompile.BN254G1Point
		Total   *big.Int
		Hit     *big.Int
	}{
		AggPkG1: dasignersprecompile.NewBN254G1Point(bn254util.SerializeG1(new(bn254.G1Affine).ScalarMultiplication(bn254util.GetG1Generator(), big.NewInt(1)))),
		Total:   big.NewInt(int64(len(quorum))),
		Hit:     big.NewInt(int64(len(quorum) / 3)),
	})

	bitMap[twoPos/8] |= 1 << (twoPos % 8)
	suite.Assert().EqualValues(suite.queryGetAggPkG1(suite.signerOne, bitMap), struct {
		AggPkG1 dasignersprecompile.BN254G1Point
		Total   *big.Int
		Hit     *big.Int
	}{
		AggPkG1: dasignersprecompile.NewBN254G1Point(bn254util.SerializeG1(new(bn254.G1Affine).ScalarMultiplication(bn254util.GetG1Generator(), big.NewInt(1+11)))),
		Total:   big.NewInt(int64(len(quorum))),
		Hit:     big.NewInt(int64(len(quorum))),
	})

}

func TestKeeperSuite(t *testing.T) {
	suite.Run(t, new(DASignersTestSuite))
}
