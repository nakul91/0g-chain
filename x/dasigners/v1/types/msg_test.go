package types_test

import (
	"math/big"
	"testing"

	"github.com/0glabs/0g-chain/crypto/bn254util"
	"github.com/0glabs/0g-chain/x/dasigners/v1/testutil"
	"github.com/0glabs/0g-chain/x/dasigners/v1/types"
	"github.com/consensys/gnark-crypto/ecc/bn254"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/suite"
)

type MsgTestSuite struct {
	testutil.Suite
}

func (suite *MsgTestSuite) Test_MsgRegisterSigner() {
	sk := big.NewInt(1)
	pkG1 := new(bn254.G1Affine).ScalarMultiplication(bn254util.GetG1Generator(), sk)
	pkG2 := new(bn254.G2Affine).ScalarMultiplication(bn254util.GetG2Generator(), sk)
	hash := types.PubkeyRegistrationHash(common.HexToAddress("0x9685C4EB29309820CDC62663CC6CC82F3D42E964"), big.NewInt(8888))
	signature := new(bn254.G1Affine).ScalarMultiplication(hash, sk)
	msg := &types.MsgRegisterSigner{
		Signer: &types.Signer{
			Account:  "9685C4EB29309820CDC62663CC6CC82F3D42E964",
			Socket:   "0.0.0.0:1234",
			PubkeyG1: bn254util.SerializeG1(pkG1),
			PubkeyG2: bn254util.SerializeG2(pkG2),
		},
		Signature: bn254util.SerializeG1(signature),
	}
	suite.Assert().EqualValues(len(msg.GetSigners()), 1)
	suite.Assert().EqualValues(msg.GetSigners()[0].String(), "0g1j6zuf6efxzvzpnwxye3ucmxg9u7596ty686hna")
	suite.Assert().NoError(msg.ValidateBasic())
}

func (suite *MsgTestSuite) Test_MsgUpdateSocket() {
	msg := &types.MsgUpdateSocket{
		Account: "9685C4EB29309820CDC62663CC6CC82F3D42E964",
		Socket:  "0.0.0.0:1234",
	}
	suite.Assert().EqualValues(len(msg.GetSigners()), 1)
	suite.Assert().EqualValues(msg.GetSigners()[0].String(), "0g1j6zuf6efxzvzpnwxye3ucmxg9u7596ty686hna")
	suite.Assert().NoError(msg.ValidateBasic())
}

func (suite *MsgTestSuite) Test_MsgRegisterNextEpoch() {
	hash := types.EpochRegistrationHash(common.HexToAddress("0x9685C4EB29309820CDC62663CC6CC82F3D42E964"), 1, big.NewInt(8888))
	signature := new(bn254.G1Affine).ScalarMultiplication(hash, big.NewInt(1))
	msg := &types.MsgRegisterNextEpoch{
		Account:   "9685C4EB29309820CDC62663CC6CC82F3D42E964",
		Signature: bn254util.SerializeG1(signature),
	}
	suite.Assert().EqualValues(len(msg.GetSigners()), 1)
	suite.Assert().EqualValues(msg.GetSigners()[0].String(), "0g1j6zuf6efxzvzpnwxye3ucmxg9u7596ty686hna")
	suite.Assert().NoError(msg.ValidateBasic())
}

func TestSuite(t *testing.T) {
	suite.Run(t, new(MsgTestSuite))
}
