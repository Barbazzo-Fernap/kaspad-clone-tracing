package krc721store

import (
	"fmt"

	"github.com/bugnanetwork/bugnad/domain/consensus/model"
)

type Approval struct {
	Owner    model.ScriptPublicKeyString
	Operator model.ScriptPublicKeyString
}

type krc721tagingShard struct {
	store    *krc721Store
	toUpdate map[model.ScriptPublicKeyString]model.KRC721Collection
	toAdd    map[model.ScriptPublicKeyString]model.KRC721Collection
	toDelete map[model.ScriptPublicKeyString]struct{}

	toAddTokens    map[model.ScriptPublicKeyString]map[uint64]model.ScriptPublicKeyString
	toUpdateTokens map[model.ScriptPublicKeyString]map[uint64]model.ScriptPublicKeyString

	toAddTokenApproval    map[model.ScriptPublicKeyString]map[uint64]*Approval
	toDeleteTokenApproval map[model.ScriptPublicKeyString]map[uint64]*Approval

	toAddApprovalForAll    map[model.ScriptPublicKeyString][]*Approval
	toDeleteApprovalForAll map[model.ScriptPublicKeyString][]*Approval
}

func (k *krc721Store) stagingShard(stagingArea *model.StagingArea) *krc721tagingShard {
	return stagingArea.GetOrCreateShard(k.shardID, func() model.StagingShard {
		return &krc721tagingShard{
			store:                  k,
			toUpdate:               map[model.ScriptPublicKeyString]model.KRC721Collection{},
			toAdd:                  make(map[model.ScriptPublicKeyString]model.KRC721Collection),
			toDelete:               make(map[model.ScriptPublicKeyString]struct{}),
			toAddTokens:            make(map[model.ScriptPublicKeyString]map[uint64]model.ScriptPublicKeyString),
			toUpdateTokens:         make(map[model.ScriptPublicKeyString]map[uint64]model.ScriptPublicKeyString),
			toAddTokenApproval:     make(map[model.ScriptPublicKeyString]map[uint64]*Approval),
			toDeleteTokenApproval:  make(map[model.ScriptPublicKeyString]map[uint64]*Approval),
			toAddApprovalForAll:    make(map[model.ScriptPublicKeyString][]*Approval),
			toDeleteApprovalForAll: map[model.ScriptPublicKeyString][]*Approval{},
		}
	}).(*krc721tagingShard)
}

func (mss *krc721tagingShard) Commit(dbTx model.DBTransaction) error {
	err := mss.commitKRC721Collection(dbTx)
	if err != nil {
		return err
	}

	err = mss.commitKRC721Token(dbTx)
	if err != nil {
		return err
	}

	err = mss.commitKRC721TokenApproval(dbTx)
	if err != nil {
		return err
	}

	return nil
}

func (mss *krc721tagingShard) commitKRC721Token(dbTX model.DBTransaction) error {
	for collectionID, tokens := range mss.toAddTokens {
		for tokenID, owner := range tokens {
			bucket := mss.store.getCollectionDataBucket(collectionID, krc721TokenSuffix)
			key := bucket.Key([]byte(fmt.Sprintf("%d", tokenID)))
			err := dbTX.Put(key, []byte(owner))
			if err != nil {
				return err
			}
		}
	}

	for collectionID, tokens := range mss.toUpdateTokens {
		for tokenID, owner := range tokens {
			bucket := mss.store.getCollectionDataBucket(collectionID, krc721TokenSuffix)
			key := bucket.Key([]byte(fmt.Sprintf("%d", tokenID)))
			err := dbTX.Delete(key)
			if err != nil {
				return err
			}

			err = dbTX.Put(key, []byte(owner))
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (mss *krc721tagingShard) commitKRC721TokenApproval(dbTx model.DBTransaction) error {
	for collectionID, approvals := range mss.toAddTokenApproval {
		for tokenID, approval := range approvals {
			bucket := mss.store.getCollectionDataBucket(collectionID, krc721ApprovalSuffix)
			key := bucket.Key([]byte(fmt.Sprintf("token_%d", tokenID)))
			err := dbTx.Put(key, []byte(approval.Operator))
			if err != nil {
				return err
			}
		}
	}

	for collectionID, approvals := range mss.toDeleteTokenApproval {
		for tokenID, approval := range approvals {
			bucket := mss.store.getCollectionDataBucket(collectionID, krc721ApprovalSuffix)
			key := bucket.Key([]byte(fmt.Sprintf("token_%d", tokenID)))
			err := dbTx.Delete(key)
			if err != nil {
				return err
			}

			err = dbTx.Put(key, []byte(approval.Operator))
		}
	}

	// approval for all
	for collectionID, approvals := range mss.toAddApprovalForAll {
		for _, approval := range approvals {
			bucket := mss.store.getCollectionDataBucket(collectionID, krc721ApprovalSuffix)
			key := bucket.Key([]byte(fmt.Sprintf("%s:%s", approval.Owner, approval.Operator)))
			err := dbTx.Put(key, []byte{1})
			if err != nil {
				return err
			}
		}
	}

	for collectionID, approvals := range mss.toDeleteApprovalForAll {
		for _, approval := range approvals {
			bucket := mss.store.getCollectionDataBucket(collectionID, krc721ApprovalSuffix)
			key := bucket.Key([]byte(fmt.Sprintf("%s:%s", approval.Owner, approval.Operator)))
			err := dbTx.Delete(key)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (mss *krc721tagingShard) commitKRC721Collection(dbTx model.DBTransaction) error {
	for hash := range mss.toDelete {
		err := dbTx.Delete(mss.store.getCollectionKey(hash))
		if err != nil {
			return err
		}
	}

	for hash, krc721Collection := range mss.toAdd {
		krc721Bytes, err := mss.store.serializeKRC721(krc721Collection)
		if err != nil {
			return err
		}
		err = dbTx.Put(mss.store.getCollectionKey(hash), krc721Bytes)
		if err != nil {
			return err
		}
	}

	for hash, krc721Collection := range mss.toUpdate {
		// delete first
		err := dbTx.Delete(mss.store.getCollectionKey(hash))
		if err != nil {
			return err
		}

		krc721Bytes, err := mss.store.serializeKRC721(krc721Collection)
		if err != nil {
			return err
		}

		err = dbTx.Put(mss.store.getCollectionKey(hash), krc721Bytes)
		if err != nil {
			return err
		}
	}

	return nil
}

func (mss *krc721tagingShard) isStaged() bool {
	return len(mss.toAdd) != 0 ||
		len(mss.toDelete) != 0 ||
		len(mss.toUpdate) != 0 ||
		len(mss.toAddTokens) != 0 ||
		len(mss.toUpdateTokens) != 0 ||
		len(mss.toAddTokenApproval) != 0 ||
		len(mss.toDeleteTokenApproval) != 0 ||
		len(mss.toAddApprovalForAll) != 0 ||
		len(mss.toDeleteApprovalForAll) != 0
}
