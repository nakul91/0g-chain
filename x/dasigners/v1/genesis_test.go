package dasigners_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"

	"github.com/0glabs/0g-chain/app"
	"github.com/0glabs/0g-chain/x/dasigners/v1"
	"github.com/0glabs/0g-chain/x/dasigners/v1/testutil"
	"github.com/0glabs/0g-chain/x/dasigners/v1/types"
)

type GenesisTestSuite struct {
	testutil.Suite
}

func (suite *GenesisTestSuite) TestInitGenesis() {
	// Most genesis validation tests are located in the types directory. The 'invalid' test cases are
	// randomly selected subset of those tests.
	testCases := []struct {
		name       string
		genState   *types.GenesisState
		expectPass bool
	}{
		{
			name:       "normal",
			genState:   types.DefaultGenesisState(),
			expectPass: true,
		},
		{
			name: "normal-more-epochs",
			genState: types.NewGenesisState(types.Params{
				TokensPerVote:     10,
				MaxVotesPerSigner: 1024,
				MaxQuorums:        10,
				EpochBlocks:       5760,
				EncodedSlices:     1,
			}, 0, []*types.Signer{{
				Account:  "0000000000000000000000000000000000000001",
				Socket:   "0.0.0.0:1234",
				PubkeyG1: make([]byte, 64),
				PubkeyG2: make([]byte, 128),
			}}, []*types.Quorums{{
				Quorums: []*types.Quorum{{Signers: []string{"0000000000000000000000000000000000000001"}}},
			}}),
			expectPass: true,
		},
		{
			name: "invalid account format",
			genState: types.NewGenesisState(types.Params{
				TokensPerVote:     10,
				MaxVotesPerSigner: 1024,
				MaxQuorums:        10,
				EpochBlocks:       5760,
				EncodedSlices:     1,
			}, 0, []*types.Signer{{
				Account:  "0x0000000000000000000000000000000000000001",
				Socket:   "0.0.0.0:1234",
				PubkeyG1: make([]byte, 64),
				PubkeyG2: make([]byte, 128),
			}}, []*types.Quorums{{
				Quorums: []*types.Quorum{{Signers: []string{"0x0000000000000000000000000000000000000001"}}},
			}}),
			expectPass: false,
		},
		{
			name: "invalid pubkeyG1 format",
			genState: types.NewGenesisState(types.Params{
				TokensPerVote:     10,
				MaxVotesPerSigner: 1024,
				MaxQuorums:        10,
				EpochBlocks:       5760,
				EncodedSlices:     1,
			}, 0, []*types.Signer{{
				Account:  "0000000000000000000000000000000000000001",
				Socket:   "0.0.0.0:1234",
				PubkeyG1: make([]byte, 63),
				PubkeyG2: make([]byte, 128),
			}}, []*types.Quorums{{
				Quorums: []*types.Quorum{{Signers: []string{"0000000000000000000000000000000000000001"}}},
			}}),
			expectPass: false,
		},
		{
			name: "invalid pubkeyG2 format",
			genState: types.NewGenesisState(types.Params{
				TokensPerVote:     10,
				MaxVotesPerSigner: 1024,
				MaxQuorums:        10,
				EpochBlocks:       5760,
				EncodedSlices:     1,
			}, 0, []*types.Signer{{
				Account:  "0000000000000000000000000000000000000001",
				Socket:   "0.0.0.0:1234",
				PubkeyG1: make([]byte, 64),
				PubkeyG2: make([]byte, 129),
			}}, []*types.Quorums{{
				Quorums: []*types.Quorum{{Signers: []string{"0000000000000000000000000000000000000001"}}},
			}}),
			expectPass: false,
		},
		{
			name: "history missing",
			genState: types.NewGenesisState(types.Params{
				TokensPerVote:     10,
				MaxVotesPerSigner: 1024,
				MaxQuorums:        10,
				EpochBlocks:       5760,
				EncodedSlices:     1,
			}, 0, []*types.Signer{{
				Account:  "0000000000000000000000000000000000000001",
				Socket:   "0.0.0.0:1234",
				PubkeyG1: make([]byte, 64),
				PubkeyG2: make([]byte, 128),
			}}, []*types.Quorums{}),
			expectPass: false,
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// Setup (note: suite.SetupTest is not run before every suite.Run)
			suite.App = app.NewTestApp()
			suite.Keeper = suite.App.GetDASignersKeeper()
			suite.Ctx = suite.App.NewContext(true, tmproto.Header{})

			// Run
			var exportedGenState *types.GenesisState
			run := func() {
				dasigners.InitGenesis(suite.Ctx, suite.Keeper, *tc.genState)
				exportedGenState = dasigners.ExportGenesis(suite.Ctx, suite.Keeper)
			}
			if tc.expectPass {
				suite.Require().NotPanics(run)
			} else {
				suite.Require().Panics(run)
			}

			// Check
			if tc.expectPass {
				expectedJson, err := suite.App.AppCodec().MarshalJSON(tc.genState)
				suite.Require().NoError(err)
				actualJson, err := suite.App.AppCodec().MarshalJSON(exportedGenState)
				suite.Require().NoError(err)
				suite.Equal(expectedJson, actualJson)
			}
		})
	}
}

func TestGenesisTestSuite(t *testing.T) {
	suite.Run(t, new(GenesisTestSuite))
}
