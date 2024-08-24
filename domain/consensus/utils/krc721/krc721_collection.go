package krc721

import (
	"fmt"

	"github.com/bugnanetwork/bugnad/domain/consensus/model"
	"github.com/bugnanetwork/bugnad/domain/consensus/model/externalapi"
)

type krc721Collection struct {
	id          *externalapi.ScriptPublicKey
	owner       *externalapi.ScriptPublicKey
	name        string
	symbol      string
	maxSupply   uint64
	totalSupply uint64
	baseURI     string
}

// NewKRC721Collection creates a new krc721Collection representing the nft
func NewKRC721Collection(
	id, owner *externalapi.ScriptPublicKey,
	name, symbol string,
	maxSupply uint64,
	totalSupply uint64,
	baseURI string,
) (model.KRC721Collection, error) {
	if name == "" {
		return nil, fmt.Errorf("name cannot be empty")
	}

	if symbol == "" {
		return nil, fmt.Errorf("symbol cannot be empty")
	}

	idClone := externalapi.ScriptPublicKey{Script: make([]byte, len(id.Script)), Version: id.Version}
	copy(idClone.Script, id.Script)

	ownerClone := externalapi.ScriptPublicKey{Script: make([]byte, len(owner.Script)), Version: owner.Version}
	copy(ownerClone.Script, owner.Script)
	return &krc721Collection{
		id:          &idClone,
		owner:       &ownerClone,
		name:        name,
		symbol:      symbol,
		maxSupply:   maxSupply,
		totalSupply: totalSupply,
		baseURI:     baseURI,
	}, nil
}

func (u *krc721Collection) ID() *externalapi.ScriptPublicKey {
	clone := externalapi.ScriptPublicKey{Script: make([]byte, len(u.id.Script)), Version: u.id.Version}
	copy(clone.Script, u.id.Script)
	return &clone

}

func (u *krc721Collection) Owner() *externalapi.ScriptPublicKey {
	clone := externalapi.ScriptPublicKey{Script: make([]byte, len(u.owner.Script)), Version: u.owner.Version}
	copy(clone.Script, u.owner.Script)
	return &clone
}

func (u *krc721Collection) Name() string {
	return u.name
}

func (u *krc721Collection) Symbol() string {
	return u.symbol
}

func (u *krc721Collection) BaseURI() string {
	return u.baseURI
}

func (u *krc721Collection) TokenURI(tokenID uint64) string {
	return fmt.Sprintf("%s/%d", u.baseURI, tokenID)
}

func (u *krc721Collection) TotalSupply() uint64 {
	return u.totalSupply
}

func (u *krc721Collection) MaxSupply() uint64 {
	return u.maxSupply
}

func (u *krc721Collection) MintNextTokenID() (uint64, error) {
	if u.totalSupply >= u.maxSupply {
		return 0, fmt.Errorf("max supply reached")
	}

	u.totalSupply++
	return u.totalSupply, nil
}

func (u *krc721Collection) Clone() model.KRC721Collection {
	return &krc721Collection{
		id:          u.ID(),
		owner:       u.Owner(),
		name:        u.Name(),
		symbol:      u.Symbol(),
		maxSupply:   u.MaxSupply(),
		totalSupply: u.TotalSupply(),
		baseURI:     u.baseURI,
	}
}
