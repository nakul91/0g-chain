package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	"github.com/0glabs/0g-chain/crypto/bn254util"
	"github.com/0glabs/0g-chain/x/dasigners/v1/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	gov "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/ethereum/go-ethereum/common"
	etherminttypes "github.com/evmos/ethermint/types"
)

var _ types.MsgServer = &Keeper{}

func (k Keeper) ChangeParams(goCtx context.Context, msg *types.MsgChangeParams) (*types.MsgChangeParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	// validate authority
	if k.authority != msg.Authority {
		return nil, errorsmod.Wrapf(gov.ErrInvalidSigner, "expected %s got %s", k.authority, msg.Authority)
	}
	// save params
	k.SetParams(ctx, *msg.Params)
	return &types.MsgChangeParamsResponse{}, nil
}

func (k Keeper) RegisterSigner(goCtx context.Context, msg *types.MsgRegisterSigner) (*types.MsgRegisterSignerResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	// validate sender
	err := k.CheckDelegations(ctx, msg.Signer.Account)
	if err != nil {
		return nil, err
	}
	_, found, err := k.GetSigner(ctx, msg.Signer.Account)
	if err != nil {
		return nil, err
	}
	if found {
		return nil, types.ErrSignerExists
	}
	// validate signature
	chainID, err := etherminttypes.ParseChainID(ctx.ChainID())
	if err != nil {
		return nil, err
	}
	hash := types.PubkeyRegistrationHash(common.HexToAddress(msg.Signer.Account), chainID)
	if !msg.Signer.ValidateSignature(hash, bn254util.DeserializeG1(msg.Signature)) {
		return nil, types.ErrInvalidSignature
	}
	// save signer
	if err := k.SetSigner(ctx, *msg.Signer); err != nil {
		return nil, err
	}
	return &types.MsgRegisterSignerResponse{}, nil
}

func (k Keeper) UpdateSocket(goCtx context.Context, msg *types.MsgUpdateSocket) (*types.MsgUpdateSocketResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	signer, found, err := k.GetSigner(ctx, msg.Account)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, types.ErrSignerNotFound
	}
	signer.Socket = msg.Socket
	if err := k.SetSigner(ctx, signer); err != nil {
		return nil, err
	}
	return &types.MsgUpdateSocketResponse{}, nil
}

func (k Keeper) RegisterNextEpoch(goCtx context.Context, msg *types.MsgRegisterNextEpoch) (*types.MsgRegisterNextEpochResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	// get signer
	err := k.CheckDelegations(ctx, msg.Account)
	if err != nil {
		return nil, err
	}
	signer, found, err := k.GetSigner(ctx, msg.Account)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, types.ErrSignerNotFound
	}
	// validate signature
	epochNumber, err := k.GetEpochNumber(ctx)
	if err != nil {
		return nil, err
	}
	chainID, err := etherminttypes.ParseChainID(ctx.ChainID())
	if err != nil {
		return nil, err
	}
	hash := types.EpochRegistrationHash(common.HexToAddress(msg.Account), epochNumber+1, chainID)
	if !signer.ValidateSignature(hash, bn254util.DeserializeG1(msg.Signature)) {
		return nil, types.ErrInvalidSignature
	}
	// save registration
	k.SetRegistration(ctx, epochNumber+1, msg.Account, msg.Signature)
	return &types.MsgRegisterNextEpochResponse{}, nil
}
