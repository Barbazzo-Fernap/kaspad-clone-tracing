package appmessage

type RPCKRC721Collection struct {
	ID          string
	Owner       string
	Name        string
	Symbol      string
	MaxSupply   uint64
	TotalSupply uint64
	BaseURI     string
}

// GetKRC721CollectionRequestMessage is an appmessage corresponding to
// its respective RPC message
type GetKRC721CollectionRequestMessage struct {
	baseMessage
	Address string
}

// Command returns the protocol command string for the message
func (msg *GetKRC721CollectionRequestMessage) Command() MessageCommand {
	return CmdGetKRC721CollectionRequestMessage
}

// NewGetKRC721CollectionRequestMessage returns a instance of the message
func NewGetKRC721CollectionRequestMessage(address string) *GetKRC721CollectionRequestMessage {
	return &GetKRC721CollectionRequestMessage{
		Address: address,
	}
}

// GetKRC721CollectionResponseMessage is an appmessage corresponding to
// its respective RPC message
type GetKRC721CollectionResponseMessage struct {
	baseMessage
	Collection *RPCKRC721Collection

	Error *RPCError
}

// Command returns the protocol command string for the message
func (msg *GetKRC721CollectionResponseMessage) Command() MessageCommand {
	return CmdGetKRC721CollectionResponseMessage
}

// NewGetKRC721CollectionResponseMessage returns a instance of the message
func NewGetKRC721CollectionResponseMessage() *GetKRC721CollectionResponseMessage {
	return &GetKRC721CollectionResponseMessage{}
}
