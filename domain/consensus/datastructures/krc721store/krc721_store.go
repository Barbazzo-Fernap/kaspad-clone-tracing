package krc721store

import (
	"bytes"
	"fmt"

	"github.com/bugnanetwork/bugnad/domain/consensus/database/serialization"
	"github.com/bugnanetwork/bugnad/domain/consensus/model"
	"github.com/bugnanetwork/bugnad/util/staging"
	"google.golang.org/protobuf/proto"
)

var krc721CollectionBucketName = []byte("krc721collection")
var krc721TokenSuffix = []byte("krc721token")
var krc721ApprovalSuffix = []byte("kr721approval")

// krc721tore represents a store of KRC721
type krc721Store struct {
	shardID                model.StagingShardID
	krc721CollectionBucket model.DBBucket
	prefixBucket           model.DBBucket
}

// New instantiates a new KRC721Store
func New(prefixBucket model.DBBucket, cacheSize int, preallocate bool) model.KRC721Store {
	return &krc721Store{
		shardID:                staging.GenerateShardingID(),
		krc721CollectionBucket: prefixBucket.Bucket(krc721CollectionBucketName),
		prefixBucket:           prefixBucket,
	}
}

func (k *krc721Store) getCollectionDataBucket(collectionID model.ScriptPublicKeyString, suffix []byte) model.DBBucket {
	return k.krc721CollectionBucket.Bucket(append([]byte(collectionID), suffix...))
}

func (k *krc721Store) Deploy(stagingArea *model.StagingArea, krc721Collection model.KRC721Collection) {
	stagingShard := k.stagingShard(stagingArea)

	stagingShard.toAdd[model.ScriptPublicKeyString(krc721Collection.ID().String())] = krc721Collection.Clone()
}

func (k *krc721Store) IsStaged(stagingArea *model.StagingArea) bool {
	return k.stagingShard(stagingArea).isStaged()
}

func (k *krc721Store) GetCollectionByID(dbContext model.DBReader, stagingArea *model.StagingArea, collectionID model.ScriptPublicKeyString) (model.KRC721Collection, error) {
	stagingShard := k.stagingShard(stagingArea)

	if krc721, ok := stagingShard.toUpdate[collectionID]; ok {
		return krc721.Clone(), nil
	}

	if krc721, ok := stagingShard.toAdd[collectionID]; ok {
		return krc721.Clone(), nil
	}

	krc721Bytes, err := dbContext.Get(k.getCollectionKey(collectionID))
	if err != nil {
		return nil, err
	}

	krc721, err := k.deserializeKRC721(krc721Bytes)
	if err != nil {
		return nil, err
	}

	return krc721.Clone(), nil
}

func (k *krc721Store) Mint(dbContext model.DBReader, stagingArea *model.StagingArea, collectionID model.ScriptPublicKeyString, owner model.ScriptPublicKeyString) error {
	c, err := k.GetCollectionByID(dbContext, stagingArea, collectionID)
	if err != nil {
		return fmt.Errorf("error getting collection: %w", err)
	}

	nextTokenID, err := c.MintNextTokenID()
	if err != nil {
		return fmt.Errorf("error minting next token ID: %w", err)
	}

	stagingShard := k.stagingShard(stagingArea)
	stagingShard.toUpdate[model.ScriptPublicKeyString(c.ID().String())] = c.Clone()
	if _, ok := stagingShard.toAddTokens[model.ScriptPublicKeyString(c.ID().String())]; !ok {
		stagingShard.toAddTokens[model.ScriptPublicKeyString(c.ID().String())] = make(map[uint64]model.ScriptPublicKeyString)
	}
	stagingShard.toAddTokens[model.ScriptPublicKeyString(c.ID().String())][nextTokenID] = owner

	// TODO: emit event?

	return nil
}
func (k *krc721Store) BalanceOf(dbContext model.DBReader, stagingArea *model.StagingArea, collectionID model.ScriptPublicKeyString, owner model.ScriptPublicKeyString) (uint64, error) {
	bucket := k.getCollectionDataBucket(collectionID, krc721TokenSuffix)

	cursor, err := dbContext.Cursor(bucket)
	if err != nil {
		return 0, fmt.Errorf("error creating cursor: %w", err)
	}

	balances := uint64(0)
	for cursor.Next() {
		value, err := cursor.Value()
		if err != nil {
			return 0, fmt.Errorf("error getting value: %w", err)
		}

		if model.ScriptPublicKeyString(value) == owner {
			balances++
		}
	}

	return balances, nil
}
func (k *krc721Store) OwnerOf(dbContext model.DBReader, stagingArea *model.StagingArea, collectionID model.ScriptPublicKeyString, tokenID uint64) (model.ScriptPublicKeyString, error) {
	bucket := k.getCollectionDataBucket(collectionID, krc721TokenSuffix)

	key := bucket.Key([]byte(fmt.Sprintf("%d", tokenID)))
	ownerBytes, err := dbContext.Get(key)
	if err != nil {
		return "", fmt.Errorf("error getting token: %w", err)
	}

	return model.ScriptPublicKeyString(ownerBytes), nil
}

func (k *krc721Store) Approve(stagingArea *model.StagingArea, collectionID model.ScriptPublicKeyString, owner model.ScriptPublicKeyString, operator model.ScriptPublicKeyString, tokenID uint64) error {
	stagingShard := k.stagingShard(stagingArea)

	if _, ok := stagingShard.toDeleteTokenApproval[collectionID]; !ok {
		stagingShard.toDeleteTokenApproval[collectionID] = make(map[uint64]*Approval)
	}

	stagingShard.toAddTokenApproval[collectionID][tokenID] = &Approval{
		Owner:    owner,
		Operator: operator,
	}

	return nil
}

func (k *krc721Store) GetApproved(dbContext model.DBReader, stagingArea *model.StagingArea, collectionID model.ScriptPublicKeyString, tokenID uint64) (model.ScriptPublicKeyString, error) {
	bucket := k.getCollectionDataBucket(collectionID, krc721ApprovalSuffix)
	key := bucket.Key([]byte(fmt.Sprintf("token_%d", tokenID)))
	operator, err := dbContext.Get(key)
	if err != nil {
		return "", fmt.Errorf("error getting approval: %w", err)
	}

	return model.ScriptPublicKeyString(operator), nil
}

func (k *krc721Store) SetApprovalForAll(stagingArea *model.StagingArea, collectionID model.ScriptPublicKeyString, owner model.ScriptPublicKeyString, operator model.ScriptPublicKeyString, approved bool) error {
	stagingShard := k.stagingShard(stagingArea)

	if approved {
		if _, ok := stagingShard.toAddApprovalForAll[collectionID]; !ok {
			stagingShard.toAddApprovalForAll[collectionID] = []*Approval{}
		}

		stagingShard.toAddApprovalForAll[collectionID] = append(stagingShard.toAddApprovalForAll[collectionID], &Approval{
			Owner:    owner,
			Operator: operator,
		})
	} else {
		if _, ok := stagingShard.toDeleteApprovalForAll[collectionID]; !ok {
			stagingShard.toDeleteApprovalForAll[collectionID] = []*Approval{}
		}

		stagingShard.toDeleteApprovalForAll[collectionID] = append(stagingShard.toDeleteApprovalForAll[collectionID], &Approval{
			Owner:    owner,
			Operator: operator,
		})
	}
	return nil
}

func (k *krc721Store) IsApprovedForAll(dbContext model.DBReader, stagingArea *model.StagingArea, collectionID model.ScriptPublicKeyString, owner model.ScriptPublicKeyString, operator model.ScriptPublicKeyString) bool {
	bucket := k.getCollectionDataBucket(collectionID, krc721ApprovalSuffix)
	key := bucket.Key([]byte(fmt.Sprintf("%s:%s", owner, operator)))
	approvalBytes, err := dbContext.Get(key)
	if err != nil {
		return false
	}

	return bytes.Equal(approvalBytes, []byte{1})
}

func (k *krc721Store) TransferFrom(
	dbContext model.DBReader,
	stagingArea *model.StagingArea,
	collectionID model.ScriptPublicKeyString,
	operator model.ScriptPublicKeyString,
	from model.ScriptPublicKeyString,
	to model.ScriptPublicKeyString,
	tokenID uint64,
) error {
	bucket := k.getCollectionDataBucket(collectionID, krc721TokenSuffix)

	key := bucket.Key([]byte(fmt.Sprintf("%d", tokenID)))
	ownerBytes, err := dbContext.Get(key)
	if err != nil {
		return fmt.Errorf("error getting token: %w", err)
	}

	owner := model.ScriptPublicKeyString(ownerBytes)
	if owner != from {
		return fmt.Errorf("token not owned by from")
	}

	if operator != from {
		if !k.IsApprovedForAll(dbContext, stagingArea, collectionID, owner, operator) {
			approval, err := k.GetApproved(dbContext, stagingArea, collectionID, tokenID)
			if err != nil {
				return fmt.Errorf("error getting approval: %w", err)
			}

			if approval != operator {
				return fmt.Errorf("token not approved for operator")
			}
		}
	}

	stagingShard := k.stagingShard(stagingArea)

	if _, ok := stagingShard.toUpdateTokens[collectionID]; !ok {
		stagingShard.toUpdateTokens[collectionID] = make(map[uint64]model.ScriptPublicKeyString)
	}

	stagingShard.toUpdateTokens[collectionID][tokenID] = to

	return nil
}

func (k *krc721Store) getCollectionKey(key model.ScriptPublicKeyString) model.DBKey {
	return k.krc721CollectionBucket.Key(
		[]byte(key),
	)
}

func (k *krc721Store) getKey(collectionID model.ScriptPublicKeyString, suffix, key []byte) model.DBKey {
	return k.getCollectionDataBucket(collectionID, suffix).Key(key)
}

func (k *krc721Store) serializeKRC721(krc721 model.KRC721Collection) ([]byte, error) {
	return proto.Marshal(serialization.KRC721CollectionToDBKRC721Collection(krc721))
}

func (k *krc721Store) deserializeKRC721(krc721Bytes []byte) (model.KRC721Collection, error) {
	dbKRC721 := &serialization.DbKRC721Collection{}
	err := proto.Unmarshal(krc721Bytes, dbKRC721)
	if err != nil {
		return nil, err
	}

	return serialization.DBKRC721CollectionToKRC721Collection(dbKRC721)
}
