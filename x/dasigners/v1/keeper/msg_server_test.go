package keeper_test

import (
	"fmt"
	"testing"

	"github.com/0glabs/0g-chain/x/dasigners/v1/testutil"
	"github.com/0glabs/0g-chain/x/dasigners/v1/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
)

type MsgServerTestSuite struct {
	testutil.Suite
}

func (suite *MsgServerTestSuite) TestChangeParams() {
	govAccAddr := suite.GovKeeper.GetGovernanceAccount(suite.Ctx).GetAddress().String()

	testCases := []struct {
		name      string
		req       *types.MsgChangeParams
		expectErr bool
		errMsg    string
	}{
		{
			name: "invalid signer",
			req: &types.MsgChangeParams{
				Authority: suite.Addresses[0].String(),
				Params: &types.Params{
					TokensPerVote:     10,
					MaxVotesPerSigner: 1024,
					MaxQuorums:        10,
					EpochBlocks:       5760,
					EncodedSlices:     3072,
				},
			},
			expectErr: true,
			errMsg:    "expected gov account as only signer for proposal message",
		},
		{
			name: "success",
			req: &types.MsgChangeParams{
				Authority: govAccAddr,
				Params: &types.Params{
					TokensPerVote:     1,
					MaxVotesPerSigner: 2048,
					MaxQuorums:        1,
					EpochBlocks:       100,
					EncodedSlices:     2048 * 3,
				},
			},
			expectErr: false,
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			oldEventNum := len(suite.Ctx.EventManager().Events())
			_, err := suite.Keeper.ChangeParams(suite.Ctx, tc.req)
			if tc.expectErr {
				suite.Require().Error(err)
				suite.Require().Contains(err.Error(), tc.errMsg)
			} else {
				suite.Require().NoError(err)
				params := suite.Keeper.GetParams(suite.Ctx)
				suite.Require().EqualValues(*tc.req.Params, params)
				suite.Assert().NoError(err)

				events := suite.Ctx.EventManager().Events()
				suite.Assert().EqualValues(len(events), oldEventNum+1)
				suite.Assert().EqualValues(events[len(events)-1], sdk.NewEvent(
					types.EventTypeUpdateParams,
					sdk.NewAttribute(types.AttributeKeyBlockHeight, fmt.Sprint(suite.Ctx.BlockHeader().Height)),
					sdk.NewAttribute(types.AttributeKeyTokensPerVote, fmt.Sprint(params.TokensPerVote)),
					sdk.NewAttribute(types.AttributeKeyMaxQuorums, fmt.Sprint(params.MaxQuorums)),
					sdk.NewAttribute(types.AttributeKeyEpochBlocks, fmt.Sprint(params.EpochBlocks)),
					sdk.NewAttribute(types.AttributeKeyEncodedSlices, fmt.Sprint(params.EncodedSlices)),
				),
				)
			}
		})
	}
}

func TestMsgServerSuite(t *testing.T) {
	suite.Run(t, new(MsgServerTestSuite))
}
