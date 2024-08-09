package testutil

import (
	"strings"

	"github.com/0glabs/0g-chain/app"
	"github.com/0glabs/0g-chain/chaincfg"
	dasignersprecompile "github.com/0glabs/0g-chain/precompiles/dasigners"
	"github.com/0glabs/0g-chain/x/bep3/types"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	emtests "github.com/evmos/ethermint/tests"
	evmkeeper "github.com/evmos/ethermint/x/evm/keeper"
	"github.com/evmos/ethermint/x/evm/statedb"
	"github.com/stretchr/testify/suite"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/evmos/ethermint/crypto/ethsecp256k1"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

type PrecompileTestSuite struct {
	suite.Suite

	StakingKeeper *stakingkeeper.Keeper
	App           app.TestApp
	Ctx           sdk.Context
	QueryClient   types.QueryClient
	Addresses     []sdk.AccAddress

	EvmKeeper *evmkeeper.Keeper
	EthSigner ethtypes.Signer
	Statedb   *statedb.StateDB
}

type TestSigner struct {
	Addr    common.Address
	HexAddr string
	PrivKey cryptotypes.PrivKey
	Signer  keyring.Signer
}

func GenSigner() *TestSigner {
	var s TestSigner
	addr, priv := emtests.NewAddrKey()
	s.PrivKey = priv
	s.Addr = addr
	s.HexAddr = dasignersprecompile.ToLowerHexWithoutPrefix(s.Addr)
	s.Signer = emtests.NewSigner(priv)
	return &s
}

func (suite *PrecompileTestSuite) SetupTest() {
	chaincfg.SetSDKConfig()
	suite.App = app.NewTestApp()
	suite.App.InitializeFromGenesisStates()
	suite.StakingKeeper = suite.App.GetStakingKeeper()

	// make block header
	privkey, _ := ethsecp256k1.GenerateKey()
	consAddress := sdk.ConsAddress(privkey.PubKey().Address())
	key, err := privkey.ToECDSA()
	suite.Assert().NoError(err)
	hexAddr := strings.ToLower(crypto.PubkeyToAddress(key.PublicKey).Hex()[2:])
	valAddr, err := sdk.ValAddressFromHex(hexAddr)
	suite.Assert().NoError(err)
	suite.Ctx = suite.App.NewContext(true, tmproto.Header{Height: 1, ChainID: app.TestChainId, ProposerAddress: consAddress})
	newValidator, err := stakingtypes.NewValidator(valAddr, privkey.PubKey(), stakingtypes.Description{})
	suite.Assert().NoError(err)
	err = suite.StakingKeeper.SetValidatorByConsAddr(suite.Ctx, newValidator)
	suite.Assert().NoError(err)
	suite.StakingKeeper.SetValidator(suite.Ctx, newValidator)

	_, accAddresses := app.GeneratePrivKeyAddressPairs(10)
	suite.Addresses = accAddresses

	suite.EvmKeeper = suite.App.GetEvmKeeper()

	suite.EthSigner = ethtypes.LatestSignerForChainID(suite.EvmKeeper.ChainID())

	suite.Statedb = statedb.New(suite.Ctx, suite.EvmKeeper, statedb.NewEmptyTxConfig(common.BytesToHash(suite.Ctx.HeaderHash().Bytes())))
}
