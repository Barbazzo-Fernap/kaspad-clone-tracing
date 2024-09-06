package transactionprocessor

import (
	"github.com/bugnanetwork/bugnad/domain/consensus/model"
)

type transactionProcessor struct {
	databaseContext model.DBReader

	krc721Store model.KRC721Store
	bvmStore    model.BVMStore
}

// New instantiates a new TransactionProcessor
func New(
	databaseContext model.DBReader,
	krc721Store model.KRC721Store,
	bvmStore model.BVMStore,
) model.TransactionProcessor {

	return &transactionProcessor{
		databaseContext: databaseContext,
		krc721Store:     krc721Store,
		bvmStore:        bvmStore,
	}
}
