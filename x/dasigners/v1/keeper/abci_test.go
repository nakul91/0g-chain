package keeper_test

import (
	"testing"

	"github.com/0glabs/0g-chain/x/dasigners/v1/keeper"
	"github.com/0glabs/0g-chain/x/dasigners/v1/testutil"
	"github.com/0glabs/0g-chain/x/dasigners/v1/types"
	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/suite"
)

type AbciTestSuite struct {
	testutil.Suite
}

func (suite *AbciTestSuite) TestBeginBlock_NotContinuous() {
	// suite.App.InitializeFromGenesisStates()
	// dasigners.InitGenesis(suite.Ctx, suite.Keeper, *types.DefaultGenesisState())
	params := suite.Keeper.GetParams(suite.Ctx)
	suite.Assert().EqualValues(params, types.DefaultGenesisState().Params)

	epoch, err := suite.Keeper.GetEpochNumber(suite.Ctx)
	suite.Require().NoError(err)
	suite.Assert().EqualValues(epoch, 0)

	suite.Assert().NotPanics(func() {
		suite.Keeper.BeginBlock(suite.Ctx.WithBlockHeight(int64(params.EpochBlocks*10)), abci.RequestBeginBlock{})
	})
	epoch, err = suite.Keeper.GetEpochNumber(suite.Ctx)
	suite.Require().NoError(err)
	suite.Assert().EqualValues(epoch, 1)
}

func (suite *AbciTestSuite) TestBeginBlock_Success() {
	// suite.App.InitializeFromGenesisStates()
	// dasigners.InitGenesis(suite.Ctx, suite.Keeper, *types.DefaultGenesisState())
	suite.Keeper.SetParams(suite.Ctx, types.Params{
		TokensPerVote:     10,
		MaxVotesPerSigner: 200,
		MaxQuorums:        10,
		EpochBlocks:       5760,
		EncodedSlices:     10,
	})
	params := suite.Keeper.GetParams(suite.Ctx)
	epoch, err := suite.Keeper.GetEpochNumber(suite.Ctx)
	suite.Require().NoError(err)
	cnt, err := suite.Keeper.GetQuorumCount(suite.Ctx, epoch)
	suite.Require().NoError(err)
	suite.Assert().EqualValues(cnt, 0)
	// set signer
	suite.Keeper.SetSigner(suite.Ctx, types.Signer{
		Account:  "0000000000000000000000000000000000000001",
		Socket:   "0.0.0.0:1234",
		PubkeyG1: common.LeftPadBytes([]byte{1}, 32),
		PubkeyG2: common.LeftPadBytes([]byte{2}, 64),
	})
	suite.Keeper.SetRegistration(suite.Ctx, epoch+1, "0000000000000000000000000000000000000001", common.LeftPadBytes([]byte{1}, 32))
	suite.Ctx = suite.Ctx.WithBlockHeight(int64(params.EpochBlocks) * int64(epoch+1))
	suite.Keeper.BeginBlock(suite.Ctx, abci.RequestBeginBlock{})
	// check quorums
	lastEpoch := epoch
	epoch, err = suite.Keeper.GetEpochNumber(suite.Ctx)
	suite.Require().NoError(err)
	suite.Assert().EqualValues(epoch, lastEpoch+1)
	cnt, err = suite.Keeper.GetQuorumCount(suite.Ctx, epoch)
	suite.Require().NoError(err)
	suite.Assert().EqualValues(cnt, 0)
	// set delegation, 1 ballot
	suite.AddDelegation("0000000000000000000000000000000000000001", "0000000000000000000000000000000000000001", keeper.BondedConversionRate.Mul(sdk.NewIntFromUint64(params.TokensPerVote)))
	// set signer
	suite.Keeper.SetSigner(suite.Ctx, types.Signer{
		Account:  "0000000000000000000000000000000000000001",
		Socket:   "0.0.0.0:1234",
		PubkeyG1: common.LeftPadBytes([]byte{1}, 32),
		PubkeyG2: common.LeftPadBytes([]byte{2}, 64),
	})
	suite.Keeper.SetRegistration(suite.Ctx, epoch+1, "0000000000000000000000000000000000000001", common.LeftPadBytes([]byte{1}, 32))
	suite.Ctx = suite.Ctx.WithBlockHeight(int64(params.EpochBlocks) * int64(epoch+1))
	suite.Keeper.BeginBlock(suite.Ctx, abci.RequestBeginBlock{})
	// check quorums
	lastEpoch = epoch
	epoch, err = suite.Keeper.GetEpochNumber(suite.Ctx)
	suite.Require().NoError(err)
	suite.Assert().EqualValues(epoch, lastEpoch+1)
	cnt, err = suite.Keeper.GetQuorumCount(suite.Ctx, epoch)
	suite.Require().NoError(err)
	suite.Assert().EqualValues(cnt, 1)
	// set delegation, 10 ballot
	suite.AddDelegation(
		"0000000000000000000000000000000000000001",
		"0000000000000000000000000000000000000001",
		keeper.BondedConversionRate.Mul(sdk.NewIntFromUint64(params.TokensPerVote)).Mul(sdk.NewIntFromUint64(9)),
	)
	suite.Keeper.SetRegistration(suite.Ctx, epoch+1, "0000000000000000000000000000000000000001", common.LeftPadBytes([]byte{1}, 32))
	suite.Ctx = suite.Ctx.WithBlockHeight(int64(params.EpochBlocks) * int64(epoch+1))
	suite.Keeper.BeginBlock(suite.Ctx, abci.RequestBeginBlock{})
	// check quorums
	lastEpoch = epoch
	epoch, err = suite.Keeper.GetEpochNumber(suite.Ctx)
	suite.Require().NoError(err)
	suite.Assert().EqualValues(epoch, lastEpoch+1)
	cnt, err = suite.Keeper.GetQuorumCount(suite.Ctx, epoch)
	suite.Require().NoError(err)
	suite.Assert().EqualValues(cnt, 1)
	// set delegation, 11 ballot
	suite.AddDelegation(
		"0000000000000000000000000000000000000001",
		"0000000000000000000000000000000000000001",
		keeper.BondedConversionRate.Mul(sdk.NewIntFromUint64(params.TokensPerVote)).Mul(sdk.NewIntFromUint64(1)),
	)
	suite.Keeper.SetRegistration(suite.Ctx, epoch+1, "0000000000000000000000000000000000000001", common.LeftPadBytes([]byte{1}, 32))
	suite.Ctx = suite.Ctx.WithBlockHeight(int64(params.EpochBlocks) * int64(epoch+1))
	suite.Keeper.BeginBlock(suite.Ctx, abci.RequestBeginBlock{})
	// check quorums
	lastEpoch = epoch
	epoch, err = suite.Keeper.GetEpochNumber(suite.Ctx)
	suite.Require().NoError(err)
	suite.Assert().EqualValues(epoch, lastEpoch+1)
	cnt, err = suite.Keeper.GetQuorumCount(suite.Ctx, epoch)
	suite.Require().NoError(err)
	suite.Assert().EqualValues(cnt, 2)
	// set delegation, 200 ballot
	suite.AddDelegation(
		"0000000000000000000000000000000000000001",
		"0000000000000000000000000000000000000001",
		keeper.BondedConversionRate.Mul(sdk.NewIntFromUint64(params.TokensPerVote)).Mul(sdk.NewIntFromUint64(200)),
	)
	suite.Keeper.SetRegistration(suite.Ctx, epoch+1, "0000000000000000000000000000000000000001", common.LeftPadBytes([]byte{1}, 32))
	suite.Ctx = suite.Ctx.WithBlockHeight(int64(params.EpochBlocks) * int64(epoch+1))
	suite.Keeper.BeginBlock(suite.Ctx, abci.RequestBeginBlock{})
	// check quorums
	lastEpoch = epoch
	epoch, err = suite.Keeper.GetEpochNumber(suite.Ctx)
	suite.Require().NoError(err)
	suite.Assert().EqualValues(epoch, lastEpoch+1)
	cnt, err = suite.Keeper.GetQuorumCount(suite.Ctx, epoch)
	suite.Require().NoError(err)
	suite.Assert().EqualValues(cnt, 10)
}

func TestAbciSuite(t *testing.T) {
	suite.Run(t, new(AbciTestSuite))
}
