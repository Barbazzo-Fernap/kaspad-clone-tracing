package serialization

import (
	"github.com/bugnanetwork/bugnad/domain/bvm/state"
	"github.com/bugnanetwork/bugnad/domain/consensus/model/externalapi"
)

// KRC721ToDBKRC721 converts KRC721 to DbKRC721
func AccountToDBAccount(account *state.Account) *DbAccount {
	codeHash, _ := externalapi.NewDomainHashFromByteSlice(account.CodeHash)
	return &DbAccount{
		Nonce:    account.Nonce,
		CodeHash: DomainHashToDbHash(codeHash),
	}
}

// DBKRC721ToKRC721 converts DbKRC721 to KRC721
func DBAccountToAccount(dbAccount *DbAccount) (*state.Account, error) {
	return &state.Account{
		Nonce:    dbAccount.Nonce,
		CodeHash: dbAccount.CodeHash.GetHash(),
	}, nil
}
