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
