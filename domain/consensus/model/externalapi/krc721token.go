package externalapi

type KRC721Collection interface {
	CollectionID() *ScriptPublicKey
	Owner() *ScriptPublicKey // The public key script for the output.
}

type KRC721Token interface {
	TokenID() uint64
	Owner() *ScriptPublicKey // The public key script for the output.
}
