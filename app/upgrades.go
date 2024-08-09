package app

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

const (
	UpgradeName_Testnet = "v0.3.1"
)

// RegisterUpgradeHandlers registers the upgrade handlers for the app.
func (app App) RegisterUpgradeHandlers() {
	app.upgradeKeeper.SetUpgradeHandler(
		UpgradeName_Testnet,
		upgradeHandler(app, UpgradeName_Testnet),
	)
}

// upgradeHandler returns an UpgradeHandler for the given upgrade parameters.
func upgradeHandler(
	app App,
	name string,
) upgradetypes.UpgradeHandler {
	return func(
		ctx sdk.Context,
		plan upgradetypes.Plan,
		fromVM module.VersionMap,
	) (module.VersionMap, error) {
		app.Logger().Info(fmt.Sprintf("running %s upgrade handler", name))

		params := app.mintKeeper.GetParams(ctx)
		params.MintDenom = "ua0gi"
		app.mintKeeper.SetParams(ctx, params)

		// run migrations for all modules and return new consensus version map
		return app.mm.RunMigrations(ctx, app.configurator, fromVM)
	}
}
