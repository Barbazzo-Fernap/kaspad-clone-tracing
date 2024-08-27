package externalapi

type KRC721Collection interface {
	ID() *ScriptPublicKey
	Owner() *ScriptPublicKey // The public key script for the output.
	Name() string
	Symbol() string
	BaseURI() string
	TokenURI(tokenID uint64) string
	TotalSupply() uint64
	MaxSupply() uint64
}

type KRC721Token interface {
	TokenID() uint64
	Owner() *ScriptPublicKey // The public key script for the output.
}
