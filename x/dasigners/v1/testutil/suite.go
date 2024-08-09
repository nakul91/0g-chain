package testutil

import (
	"strings"

	"cosmossdk.io/math"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/suite"

	"github.com/0glabs/0g-chain/app"
	"github.com/0glabs/0g-chain/chaincfg"
	"github.com/0glabs/0g-chain/x/dasigners/v1/keeper"
	"github.com/0glabs/0g-chain/x/dasigners/v1/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/evmos/ethermint/crypto/ethsecp256k1"
)

// Suite implements a test suite for the module integration tests
type Suite struct {
	suite.Suite

	Keeper        keeper.Keeper
	StakingKeeper *stakingkeeper.Keeper
	App           app.TestApp
	Ctx           sdk.Context
	QueryClient   types.QueryClient
	Addresses     []sdk.AccAddress
}

// SetupTest instantiates a new app, keepers, and sets suite state
func (suite *Suite) SetupTest() {
	chaincfg.SetSDKConfig()
	suite.App = app.NewTestApp()
	suite.App.InitializeFromGenesisStates()
	suite.Keeper = suite.App.GetDASignersKeeper()
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

	// Set query client
	queryHelper := suite.App.NewQueryServerTestHelper(suite.Ctx)
	queryHandler := suite.Keeper
	types.RegisterQueryServer(queryHelper, queryHandler)
	suite.QueryClient = types.NewQueryClient(queryHelper)
}

func (suite *Suite) AddDelegation(from string, to string, amount math.Int) {
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
	bonded := suite.Keeper.GetDelegatorBonded(suite.Ctx, accAddr)
	suite.StakingKeeper.SetDelegation(suite.Ctx, stakingtypes.Delegation{
		DelegatorAddress: accAddr.String(),
		ValidatorAddress: valAddr.String(),
		Shares:           bonded.Add(amount).ToLegacyDec(),
	})
}
