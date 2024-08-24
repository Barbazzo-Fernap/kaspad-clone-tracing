package serialization

import (
	"fmt"

	"github.com/bugnanetwork/bugnad/domain/consensus/model"
	"github.com/bugnanetwork/bugnad/domain/consensus/utils/krc721"
)

// KRC721ToDBKRC721 converts KRC721 to DbKRC721
func KRC721CollectionToDBKRC721Collection(krc721 model.KRC721Collection) *DbKRC721Collection {
	return &DbKRC721Collection{
		Id:          ScriptPublicKeyToDBScriptPublicKey(krc721.ID()),
		Owner:       ScriptPublicKeyToDBScriptPublicKey(krc721.Owner()),
		Name:        krc721.Name(),
		Symbol:      krc721.Symbol(),
		MaxSupply:   krc721.MaxSupply(),
		TotalSupply: krc721.TotalSupply(),
		BaseURI:     krc721.BaseURI(),
	}
}

// DBKRC721ToKRC721 converts DbKRC721 to KRC721
func DBKRC721CollectionToKRC721Collection(dbKRC721 *DbKRC721Collection) (model.KRC721Collection, error) {
	id, err := DBScriptPublicKeyToScriptPublicKey(dbKRC721.Id)
	if err != nil {
		return nil, fmt.Errorf("Error in DBKRC721ToKRC721: %s", err)
	}

	owner, err := DBScriptPublicKeyToScriptPublicKey(dbKRC721.Owner)
	if err != nil {
		return nil, fmt.Errorf("Error in DBKRC721ToKRC721: %s", err)
	}

	return krc721.NewKRC721Collection(
		id,
		owner,
		dbKRC721.Name,
		dbKRC721.Symbol,
		dbKRC721.MaxSupply,
		dbKRC721.TotalSupply,
		dbKRC721.BaseURI,
	)
}
