package types_test

import (
	"math/big"
	"testing"

	"github.com/0glabs/0g-chain/x/dasigners/v1/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

func Test_PubkeyRegistrationHash(t *testing.T) {
	hash := types.PubkeyRegistrationHash(common.HexToAddress("0x9685C4EB29309820CDC62663CC6CC82F3D42E964"), big.NewInt(8888))
	assert.Equal(t, hash.X.String(), "17347288745752564851578145205408924577042674846071448492673629564958667746090")
	assert.Equal(t, hash.Y.String(), "21456041422468658262738002909407073439935597271458862589356790821116767485654")
}

func Test_EpochRegistrationHash(t *testing.T) {
	hash := types.EpochRegistrationHash(common.HexToAddress("0x9685C4EB29309820CDC62663CC6CC82F3D42E964"), 1, big.NewInt(8888))
	assert.Equal(t, hash.X.String(), "13283083124528531674735853832182424672122091139683454761857829308708073730285")
	assert.Equal(t, hash.Y.String(), "21773064143788270772276852950775943855438706734263253481317981346601766662828")
}
