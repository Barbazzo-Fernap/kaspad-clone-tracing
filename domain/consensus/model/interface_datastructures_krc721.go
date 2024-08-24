package model

type ScriptPublicKeyString string

// KRC721Store represents a store of KRC721 collection and token
type KRC721Store interface {
	Store
	Deploy(stagingArea *StagingArea, krc721Collection KRC721Collection)
	IsStaged(stagingArea *StagingArea) bool

	Mint(dbContext DBReader, stagingArea *StagingArea, collectionID, owner ScriptPublicKeyString) error

	BalanceOf(dbContext DBReader, stagingArea *StagingArea, collectionID, owner ScriptPublicKeyString) (uint64, error)
	OwnerOf(dbContext DBReader, stagingArea *StagingArea, collectionID ScriptPublicKeyString, tokenID uint64) (ScriptPublicKeyString, error)

	Approve(stagingArea *StagingArea, collectionID, owner, operator ScriptPublicKeyString, tokenID uint64) error
	GetApproved(dbContext DBReader, stagingArea *StagingArea, collectionID ScriptPublicKeyString, tokenID uint64) (ScriptPublicKeyString, error)

	SetApprovalForAll(stagingArea *StagingArea, collectionID, owner, operator ScriptPublicKeyString, approved bool) error
	IsApprovedForAll(dbContext DBReader, stagingArea *StagingArea, collectionID ScriptPublicKeyString, owner ScriptPublicKeyString, operator ScriptPublicKeyString) bool

	TransferFrom(dbContext DBReader, stagingArea *StagingArea, collectionID ScriptPublicKeyString, operator, from, to ScriptPublicKeyString, tokenID uint64) error

	GetCollectionByID(dbContext DBReader, stagingArea *StagingArea, collectionID ScriptPublicKeyString) (KRC721Collection, error)
}
