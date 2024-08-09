package keeper_test

import (
	"encoding/hex"
	"math/big"
	"strings"
	"testing"

	"github.com/0glabs/0g-chain/crypto/bn254util"
	"github.com/0glabs/0g-chain/x/dasigners/v1"
	"github.com/0glabs/0g-chain/x/dasigners/v1/keeper"
	"github.com/0glabs/0g-chain/x/dasigners/v1/testutil"
	"github.com/0glabs/0g-chain/x/dasigners/v1/types"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/consensys/gnark-crypto/ecc/bn254"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/suite"
)

const (
	signer1 = "9685C4EB29309820CDC62663CC6CC82F3D42E964"
	signer2 = "9685C4EB29309820CDC62663CC6CC82F3D42E965"
)

type KeeperTestSuite struct {
	testutil.Suite
}

func (suite *KeeperTestSuite) testRegisterSignerInvalidSignature() {
	sk := big.NewInt(1)
	pkG1 := new(bn254.G1Affine).ScalarMultiplication(bn254util.GetG1Generator(), sk)
	pkG2 := new(bn254.G2Affine).ScalarMultiplication(bn254util.GetG2Generator(), sk)
	hash := types.PubkeyRegistrationHash(common.HexToAddress(signer1), big.NewInt(8888))
	signature := new(bn254.G1Affine).ScalarMultiplication(hash, big.NewInt(2))
	msg := &types.MsgRegisterSigner{
		Signer: &types.Signer{
			Account:  signer1,
			Socket:   "0.0.0.0:1234",
			PubkeyG1: bn254util.SerializeG1(pkG1),
			PubkeyG2: bn254util.SerializeG2(pkG2),
		},
		Signature: bn254util.SerializeG1(signature),
	}
	_, err := suite.Keeper.RegisterSigner(sdk.WrapSDKContext(suite.Ctx), msg)
	suite.Assert().ErrorIs(err, types.ErrInvalidSignature)
}

func (suite *KeeperTestSuite) testRegisterSignerSuccess() *types.Signer { // resgister signer
	sk := big.NewInt(1)
	pkG1 := new(bn254.G1Affine).ScalarMultiplication(bn254util.GetG1Generator(), sk)
	pkG2 := new(bn254.G2Affine).ScalarMultiplication(bn254util.GetG2Generator(), sk)
	hash := types.PubkeyRegistrationHash(common.HexToAddress(signer1), big.NewInt(8888))
	signature := new(bn254.G1Affine).ScalarMultiplication(hash, sk)
	signer := &types.Signer{
		Account:  signer1,
		Socket:   "0.0.0.0:1234",
		PubkeyG1: bn254util.SerializeG1(pkG1),
		PubkeyG2: bn254util.SerializeG2(pkG2),
	}
	msg := &types.MsgRegisterSigner{
		Signer:    signer,
		Signature: bn254util.SerializeG1(signature),
	}
	oldEventNum := len(suite.Ctx.EventManager().Events())
	_, err := suite.Keeper.RegisterSigner(sdk.WrapSDKContext(suite.Ctx), msg)
	suite.Assert().NoError(err)
	events := suite.Ctx.EventManager().Events()
	suite.Assert().EqualValues(len(events), oldEventNum+1)
	suite.Assert().EqualValues(events[len(events)-1], sdk.NewEvent(
		types.EventTypeUpdateSigner,
		sdk.NewAttribute(types.AttributeKeySigner, signer.Account),
		sdk.NewAttribute(types.AttributeKeySocket, signer.Socket),
		sdk.NewAttribute(types.AttributeKeyPublicKeyG1, hex.EncodeToString(signer.PubkeyG1)),
		sdk.NewAttribute(types.AttributeKeyPublicKeyG2, hex.EncodeToString(signer.PubkeyG2)),
	))
	return signer
}

func (suite *KeeperTestSuite) testQuerySigner(signer *types.Signer) {
	response, err := suite.Keeper.Signer(sdk.WrapSDKContext(suite.Ctx), &types.QuerySignerRequest{
		Accounts: []string{signer1},
	})
	suite.Assert().NoError(err)
	suite.Assert().EqualValues(len(response.Signer), 1)
	suite.Assert().EqualValues(response.Signer[0], signer)
	_, err = suite.Keeper.Signer(sdk.WrapSDKContext(suite.Ctx), &types.QuerySignerRequest{
		Accounts: []string{signer1, signer2},
	})
	suite.Assert().ErrorIs(err, types.ErrSignerNotFound)
}

func (suite *KeeperTestSuite) testUpdateSocket(signer *types.Signer) {
	signer.Socket = "0.0.0.0:2345"
	msg := &types.MsgUpdateSocket{
		Account: signer2,
		Socket:  signer.Socket,
	}
	_, err := suite.Keeper.UpdateSocket(sdk.WrapSDKContext(suite.Ctx), msg)
	suite.Assert().ErrorIs(err, types.ErrSignerNotFound)
	msg.Account = signer.Account
	oldEventNum := len(suite.Ctx.EventManager().Events())
	_, err = suite.Keeper.UpdateSocket(sdk.WrapSDKContext(suite.Ctx), msg)
	suite.Assert().NoError(err, types.ErrSignerNotFound)
	events := suite.Ctx.EventManager().Events()
	suite.Assert().EqualValues(len(events), oldEventNum+1)
	suite.Assert().EqualValues(events[len(events)-1], sdk.NewEvent(
		types.EventTypeUpdateSigner,
		sdk.NewAttribute(types.AttributeKeySigner, signer.Account),
		sdk.NewAttribute(types.AttributeKeySocket, signer.Socket),
		sdk.NewAttribute(types.AttributeKeyPublicKeyG1, hex.EncodeToString(signer.PubkeyG1)),
		sdk.NewAttribute(types.AttributeKeyPublicKeyG2, hex.EncodeToString(signer.PubkeyG2)),
	))
}

func (suite *KeeperTestSuite) testRegisterEpochInvalidSignature() {
	sk := big.NewInt(2)
	hash := types.EpochRegistrationHash(common.HexToAddress(signer1), 1, big.NewInt(8888))
	signature := new(bn254.G1Affine).ScalarMultiplication(hash, sk)
	msg := &types.MsgRegisterNextEpoch{
		Account:   signer1,
		Signature: bn254util.SerializeG1(signature),
	}
	_, err := suite.Keeper.RegisterNextEpoch(sdk.WrapSDKContext(suite.Ctx), msg)
	suite.Assert().ErrorIs(err, types.ErrInvalidSignature)
}

func (suite *KeeperTestSuite) secondSigner() *types.Signer {
	sk := big.NewInt(11)
	pkG1 := new(bn254.G1Affine).ScalarMultiplication(bn254util.GetG1Generator(), sk)
	pkG2 := new(bn254.G2Affine).ScalarMultiplication(bn254util.GetG2Generator(), sk)
	hash := types.PubkeyRegistrationHash(common.HexToAddress(signer2), big.NewInt(8888))
	signature := new(bn254.G1Affine).ScalarMultiplication(hash, sk)
	signer := &types.Signer{
		Account:  signer2,
		Socket:   "0.0.0.0:1234",
		PubkeyG1: bn254util.SerializeG1(pkG1),
		PubkeyG2: bn254util.SerializeG2(pkG2),
	}
	msg := &types.MsgRegisterSigner{
		Signer:    signer,
		Signature: bn254util.SerializeG1(signature),
	}
	oldEventNum := len(suite.Ctx.EventManager().Events())
	_, err := suite.Keeper.RegisterSigner(sdk.WrapSDKContext(suite.Ctx), msg)
	suite.Assert().NoError(err)
	events := suite.Ctx.EventManager().Events()
	suite.Assert().EqualValues(len(events), oldEventNum+1)
	suite.Assert().EqualValues(events[len(events)-1], sdk.NewEvent(
		types.EventTypeUpdateSigner,
		sdk.NewAttribute(types.AttributeKeySigner, signer.Account),
		sdk.NewAttribute(types.AttributeKeySocket, signer.Socket),
		sdk.NewAttribute(types.AttributeKeyPublicKeyG1, hex.EncodeToString(signer.PubkeyG1)),
		sdk.NewAttribute(types.AttributeKeyPublicKeyG2, hex.EncodeToString(signer.PubkeyG2)),
	))
	// register epoch
	hash = types.EpochRegistrationHash(common.HexToAddress(signer2), 1, big.NewInt(8888))
	signature = new(bn254.G1Affine).ScalarMultiplication(hash, sk)
	msg2 := &types.MsgRegisterNextEpoch{
		Account:   signer2,
		Signature: bn254util.SerializeG1(signature),
	}
	_, err = suite.Keeper.RegisterNextEpoch(sdk.WrapSDKContext(suite.Ctx), msg2)
	suite.Assert().NoError(err, types.ErrSignerNotFound)
	return signer
}

func (suite *KeeperTestSuite) testRegisterEpochSuccess() {
	sk := big.NewInt(1)
	hash := types.EpochRegistrationHash(common.HexToAddress(signer1), 1, big.NewInt(8888))
	signature := new(bn254.G1Affine).ScalarMultiplication(hash, sk)
	msg := &types.MsgRegisterNextEpoch{
		Account:   signer1,
		Signature: bn254util.SerializeG1(signature),
	}
	_, err := suite.Keeper.RegisterNextEpoch(sdk.WrapSDKContext(suite.Ctx), msg)
	suite.Assert().NoError(err, types.ErrSignerNotFound)

}

func (suite *KeeperTestSuite) newEpoch(params types.Params) {
	// 1st ballot of signer1: 1d5df5684184f84a8dbd20b158b6478a6e8eb021b1cf81ac281dd4c7af4370ed30231e1a6a1d76bac5f464f10c7e99afa8df3c4643ca447bfc80f248764ab2ac
	// 1st ballot of signer2: 103d29532b47eb7df57049180475d72737f7ab2be4a0f3614aedbb61c8a844a32c76fcbb29b937c56c577121dfd4be8041e2b4acfe2523ae54f0d6f604745b06
	// 2nd ballot of signer2: 93a5bb4c22640a155b18e24c0c584f2bc4bdd94ddb786d86ff3c3816d741e67f
	// sorted ballots: 2-1, 1-1, 2-2
	suite.Ctx = suite.Ctx.WithBlockHeight(int64(params.EpochBlocks) * 1)
	suite.Keeper.BeginBlock(suite.Ctx, abci.RequestBeginBlock{})
}

func (suite *KeeperTestSuite) queryEpochNumber() {
	response, err := suite.Keeper.EpochNumber(sdk.WrapSDKContext(suite.Ctx), &types.QueryEpochNumberRequest{})
	suite.Assert().NoError(err)
	suite.Assert().EqualValues(response.EpochNumber, 1)
}

func (suite *KeeperTestSuite) queryQuorumCount() {
	response, err := suite.Keeper.QuorumCount(sdk.WrapSDKContext(suite.Ctx), &types.QueryQuorumCountRequest{
		EpochNumber: 1,
	})
	suite.Assert().NoError(err)
	suite.Assert().EqualValues(response.QuorumCount, 1)
}

func (suite *KeeperTestSuite) queryEpochQuorum(params types.Params) {
	_, err := suite.Keeper.EpochQuorum(sdk.WrapSDKContext(suite.Ctx), &types.QueryEpochQuorumRequest{
		EpochNumber: 1,
		QuorumId:    1,
	})
	suite.Assert().ErrorIs(err, types.ErrQuorumIdOutOfBound)
	response, err := suite.Keeper.EpochQuorum(sdk.WrapSDKContext(suite.Ctx), &types.QueryEpochQuorumRequest{
		EpochNumber: 1,
		QuorumId:    0,
	})
	suite.Assert().NoError(err)
	quorum := make([]string, 0)
	for i := 0; i < int(params.EncodedSlices); i += 1 {
		if i%3 == 1 {
			quorum = append(quorum, strings.ToLower(signer1))
		} else {
			quorum = append(quorum, strings.ToLower(signer2))
		}
	}
	suite.Assert().EqualValues(response.Quorum.Signers, quorum)
}

func (suite *KeeperTestSuite) queryEpochQuorumRow(params types.Params) {
	_, err := suite.Keeper.EpochQuorumRow(sdk.WrapSDKContext(suite.Ctx), &types.QueryEpochQuorumRowRequest{
		EpochNumber: 1,
		QuorumId:    0,
		RowIndex:    uint32(params.EncodedSlices),
	})
	suite.Assert().ErrorIs(err, types.ErrRowIndexOutOfBound)
	response, err := suite.Keeper.EpochQuorumRow(sdk.WrapSDKContext(suite.Ctx), &types.QueryEpochQuorumRowRequest{
		EpochNumber: 1,
		QuorumId:    0,
		RowIndex:    0,
	})
	suite.Assert().NoError(err)
	suite.Assert().EqualValues(response.Signer, strings.ToLower(signer2))
	response, err = suite.Keeper.EpochQuorumRow(sdk.WrapSDKContext(suite.Ctx), &types.QueryEpochQuorumRowRequest{
		EpochNumber: 1,
		QuorumId:    0,
		RowIndex:    1,
	})
	suite.Assert().NoError(err)
	suite.Assert().EqualValues(response.Signer, strings.ToLower(signer1))
	response, err = suite.Keeper.EpochQuorumRow(sdk.WrapSDKContext(suite.Ctx), &types.QueryEpochQuorumRowRequest{
		EpochNumber: 1,
		QuorumId:    0,
		RowIndex:    2,
	})
	suite.Assert().NoError(err)
	suite.Assert().EqualValues(response.Signer, strings.ToLower(signer2))
}

func (suite *KeeperTestSuite) queryAggregatePubkeyG1(params types.Params) {
	quorumBitMap := make([]byte, params.EncodedSlices/8-1)
	_, err := suite.Keeper.AggregatePubkeyG1(sdk.WrapSDKContext(suite.Ctx), &types.QueryAggregatePubkeyG1Request{
		EpochNumber:  1,
		QuorumId:     0,
		QuorumBitmap: quorumBitMap,
	})
	suite.Assert().ErrorIs(err, types.ErrQuorumBitmapLengthMismatch)
	quorumBitMap = append(quorumBitMap, byte(0))
	response, err := suite.Keeper.AggregatePubkeyG1(sdk.WrapSDKContext(suite.Ctx), &types.QueryAggregatePubkeyG1Request{
		EpochNumber:  1,
		QuorumId:     0,
		QuorumBitmap: quorumBitMap,
	})
	suite.Assert().NoError(err)
	pkG1 := new(bn254.G1Affine).ScalarMultiplication(bn254util.GetG1Generator(), big.NewInt(0))
	suite.Assert().EqualValues(response.AggregatePubkeyG1, bn254util.SerializeG1(pkG1))
	suite.Assert().EqualValues(response.Total, params.EncodedSlices)
	suite.Assert().EqualValues(response.Hit, 0)

	quorumBitMap[0] = byte(1)
	response, err = suite.Keeper.AggregatePubkeyG1(sdk.WrapSDKContext(suite.Ctx), &types.QueryAggregatePubkeyG1Request{
		EpochNumber:  1,
		QuorumId:     0,
		QuorumBitmap: quorumBitMap,
	})
	suite.Assert().NoError(err)
	pkG1 = new(bn254.G1Affine).ScalarMultiplication(bn254util.GetG1Generator(), big.NewInt(11))
	suite.Assert().EqualValues(response.AggregatePubkeyG1, bn254util.SerializeG1(pkG1))
	suite.Assert().EqualValues(response.Total, params.EncodedSlices)
	suite.Assert().EqualValues(response.Hit, params.EncodedSlices*2/3)

	quorumBitMap[0] = byte(2)
	response, err = suite.Keeper.AggregatePubkeyG1(sdk.WrapSDKContext(suite.Ctx), &types.QueryAggregatePubkeyG1Request{
		EpochNumber:  1,
		QuorumId:     0,
		QuorumBitmap: quorumBitMap,
	})
	suite.Assert().NoError(err)
	pkG1 = new(bn254.G1Affine).ScalarMultiplication(bn254util.GetG1Generator(), big.NewInt(1))
	suite.Assert().EqualValues(response.AggregatePubkeyG1, bn254util.SerializeG1(pkG1))
	suite.Assert().EqualValues(response.Total, params.EncodedSlices)
	suite.Assert().EqualValues(response.Hit, params.EncodedSlices/3)

	quorumBitMap[0] = byte(3)
	response, err = suite.Keeper.AggregatePubkeyG1(sdk.WrapSDKContext(suite.Ctx), &types.QueryAggregatePubkeyG1Request{
		EpochNumber:  1,
		QuorumId:     0,
		QuorumBitmap: quorumBitMap,
	})
	suite.Assert().NoError(err)
	pkG1 = new(bn254.G1Affine).ScalarMultiplication(bn254util.GetG1Generator(), big.NewInt(1+11))
	suite.Assert().EqualValues(response.AggregatePubkeyG1, bn254util.SerializeG1(pkG1))
	suite.Assert().EqualValues(response.Total, params.EncodedSlices)
	suite.Assert().EqualValues(response.Hit, params.EncodedSlices)
}

func (suite *KeeperTestSuite) Test_Keeper() {
	// suite.App.InitializeFromGenesisStates()
	dasigners.InitGenesis(suite.Ctx, suite.Keeper, *types.DefaultGenesisState())
	// add delegation
	params := suite.Keeper.GetParams(suite.Ctx)
	suite.AddDelegation(signer1, signer1, keeper.BondedConversionRate.Mul(sdk.NewIntFromUint64(params.TokensPerVote)))
	suite.AddDelegation(signer2, signer1, keeper.BondedConversionRate.Mul(sdk.NewIntFromUint64(params.TokensPerVote)).Mul(sdk.NewIntFromUint64(2)))
	// test
	suite.testRegisterSignerInvalidSignature()
	signerOne := suite.testRegisterSignerSuccess()
	suite.testQuerySigner(signerOne)
	suite.testUpdateSocket(signerOne)
	suite.testRegisterEpochInvalidSignature()
	suite.testRegisterEpochSuccess()
	suite.secondSigner()
	suite.newEpoch(params)
	suite.queryEpochNumber()
	suite.queryQuorumCount()
	suite.queryEpochQuorum(params)
	suite.queryEpochQuorumRow(params)
	suite.queryAggregatePubkeyG1(params)
}

func TestKeeperSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
