package appmessage

type RPCKRC721Token struct {
	ID    uint64
	Owner string
	URI   string
}

// GetKRC721TokenRequestMessage is an appmessage corresponding to
// its respective RPC message
type GetKRC721TokenRequestMessage struct {
	baseMessage
	CollectionAddress string
	TokenID           uint64
}

// Command returns the protocol command string for the message
func (msg *GetKRC721TokenRequestMessage) Command() MessageCommand {
	return CmdGetKRC721TokenRequestMessage
}

// NewGetKRC721TokenRequestMessage returns a instance of the message
func NewGetKRC721TokenRequestMessage(collectionAddress string, tokenID uint64) *GetKRC721TokenRequestMessage {
	return &GetKRC721TokenRequestMessage{
		baseMessage:       baseMessage{},
		CollectionAddress: collectionAddress,
		TokenID:           tokenID,
	}
}

// GetKRC721TokenResponseMessage is an appmessage corresponding to
// its respective RPC message
type GetKRC721TokenResponseMessage struct {
	baseMessage
	Token *RPCKRC721Token

	Error *RPCError
}

// Command returns the protocol command string for the message
func (msg *GetKRC721TokenResponseMessage) Command() MessageCommand {
	return CmdGetKRC721TokenResponseMessage
}

// NewGetKRC721TokenResponseMessage returns a instance of the message
func NewGetKRC721TokenResponseMessage() *GetKRC721TokenResponseMessage {
	return &GetKRC721TokenResponseMessage{}
}
