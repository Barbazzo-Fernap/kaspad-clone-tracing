package model

import (
	"github.com/bugnanetwork/bugnad/domain/consensus/model/externalapi"
)

// TransactionProcessor processes Transaction
type TransactionProcessor interface {
	Excute(stagingArea *StagingArea, tx *externalapi.DomainTransaction, povBlockHash *externalapi.DomainHash) error
}
