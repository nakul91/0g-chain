package chaincfg

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	AppName   = "surged"
	EnvPrefix = "SURGE"
)

func SetSDKConfig() *sdk.Config {
	config := sdk.GetConfig()
	setBech32Prefixes(config)
	setBip44CoinType(config)
	return config
}
