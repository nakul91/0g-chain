package types

import (
	"encoding/hex"
	fmt "fmt"

	errorsmod "cosmossdk.io/errors"

	"github.com/0glabs/0g-chain/crypto/bn254util"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _, _, _, _ sdk.Msg = &MsgRegisterSigner{}, &MsgUpdateSocket{}, &MsgRegisterNextEpoch{}, &MsgChangeParams{}

// GetSigners returns the expected signers for a MsgRegisterSigner message.
func (msg *MsgRegisterSigner) GetSigners() []sdk.AccAddress {
	valAddr, err := sdk.ValAddressFromHex(msg.Signer.Account)
	if err != nil {
		panic(err)
	}
	accAddr, err := sdk.AccAddressFromHexUnsafe(hex.EncodeToString(valAddr.Bytes()))
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{accAddr}
}

// ValidateBasic does a sanity check of the provided data
func (msg *MsgRegisterSigner) ValidateBasic() error {
	if err := msg.Signer.Validate(); err != nil {
		return err
	}
	if len(msg.Signature) != bn254util.G1PointSize {
		return fmt.Errorf("invalid signature")
	}
	return nil
}

// GetSignBytes implements the LegacyMsg interface.
func (msg MsgRegisterSigner) GetSignBytes() []byte {
	return sdk.MustSortJSON(AminoCdc.MustMarshalJSON(&msg))
}

// GetSigners returns the expected signers for a MsgUpdateSocket message.
func (msg *MsgUpdateSocket) GetSigners() []sdk.AccAddress {
	valAddr, err := sdk.ValAddressFromHex(msg.Account)
	if err != nil {
		panic(err)
	}
	accAddr, err := sdk.AccAddressFromHexUnsafe(hex.EncodeToString(valAddr.Bytes()))
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{accAddr}
}

// ValidateBasic does a sanity check of the provided data
func (msg *MsgUpdateSocket) ValidateBasic() error {
	if err := ValidateHexAddress(msg.Account); err != nil {
		return err
	}
	return nil
}

// GetSignBytes implements the LegacyMsg interface.
func (msg MsgUpdateSocket) GetSignBytes() []byte {
	return sdk.MustSortJSON(AminoCdc.MustMarshalJSON(&msg))
}

// GetSigners returns the expected signers for a MsgRegisterNextEpoch message.
func (msg *MsgRegisterNextEpoch) GetSigners() []sdk.AccAddress {
	valAddr, err := sdk.ValAddressFromHex(msg.Account)
	if err != nil {
		panic(err)
	}
	accAddr, err := sdk.AccAddressFromHexUnsafe(hex.EncodeToString(valAddr.Bytes()))
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{accAddr}
}

// ValidateBasic does a sanity check of the provided data
func (msg *MsgRegisterNextEpoch) ValidateBasic() error {
	if err := ValidateHexAddress(msg.Account); err != nil {
		return err
	}
	if len(msg.Signature) != bn254util.G1PointSize {
		return fmt.Errorf("invalid signature")
	}
	return nil
}

// GetSignBytes implements the LegacyMsg interface.
func (msg MsgRegisterNextEpoch) GetSignBytes() []byte {
	return sdk.MustSortJSON(AminoCdc.MustMarshalJSON(&msg))
}

// GetSigners returns the expected signers for a MsgSetParams message.
func (msg *MsgChangeParams) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Authority)
	return []sdk.AccAddress{addr}
}

// ValidateBasic does a sanity check of the provided data
func (msg *MsgChangeParams) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return errorsmod.Wrap(err, "authority")
	}

	if err := msg.Params.Validate(); err != nil {
		return errorsmod.Wrap(err, "params")
	}

	return nil
}
