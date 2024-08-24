package model

import "github.com/bugnanetwork/bugnad/domain/consensus/model/externalapi"

// KRC721Collection represents a nft collection
type KRC721Collection interface {
	ID() *externalapi.ScriptPublicKey
	Owner() *externalapi.ScriptPublicKey
	Name() string
	Symbol() string
	BaseURI() string
	TokenURI(tokenID uint64) string
	TotalSupply() uint64
	MaxSupply() uint64
	Clone() KRC721Collection

	// update
	MintNextTokenID() (uint64, error)
}
