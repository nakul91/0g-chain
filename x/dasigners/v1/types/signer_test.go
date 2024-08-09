package types_test

import (
	"math/big"
	"testing"

	"github.com/0glabs/0g-chain/crypto/bn254util"
	"github.com/0glabs/0g-chain/x/dasigners/v1/types"
	"github.com/consensys/gnark-crypto/ecc/bn254"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

func Test_ValidateSignature(t *testing.T) {
	sk := big.NewInt(1)
	pkG1 := new(bn254.G1Affine).ScalarMultiplication(bn254util.GetG1Generator(), sk)
	pkG2 := new(bn254.G2Affine).ScalarMultiplication(bn254util.GetG2Generator(), sk)
	signer := types.Signer{
		Account:  "9685C4EB29309820CDC62663CC6CC82F3D42E964",
		Socket:   "0.0.0.0:1234",
		PubkeyG1: bn254util.SerializeG1(pkG1),
		PubkeyG2: bn254util.SerializeG2(pkG2),
	}
	assert.NoError(t, signer.Validate())
	hash := types.PubkeyRegistrationHash(common.HexToAddress("0x9685C4EB29309820CDC62663CC6CC82F3D42E964"), big.NewInt(8888))
	signature := new(bn254.G1Affine).ScalarMultiplication(hash, big.NewInt(1))
	assert.Equal(t, signer.ValidateSignature(hash, signature), true)
}
