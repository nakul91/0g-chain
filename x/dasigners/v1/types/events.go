package types

// Module event types
const (
	EventTypeUpdateSigner = "update_signer"
	EventTypeUpdateParams = "update_params"

	AttributeKeySigner            = "signer"
	AttributeKeySocket            = "socket"
	AttributeKeyPublicKeyG1       = "pubkey_g1"
	AttributeKeyPublicKeyG2       = "pubkey_g2"
	AttributeKeyBlockHeight       = "block_height"
	AttributeKeyTokensPerVote     = "tokens_per_vote"
	AttributeKeyMaxVotesPerSigner = "max_votes_per_signer"
	AttributeKeyMaxQuorums        = "max_quorums"
	AttributeKeyEpochBlocks       = "epoch_blocks"
	AttributeKeyEncodedSlices     = "encoded_slices"
)
