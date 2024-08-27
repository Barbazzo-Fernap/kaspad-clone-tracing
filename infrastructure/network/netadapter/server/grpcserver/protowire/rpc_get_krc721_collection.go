package protowire

import (
	"github.com/bugnanetwork/bugnad/app/appmessage"
	"github.com/pkg/errors"
)

func (x *BugnadMessage_GetKRC721CollectionRequest) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "BugnadMessage_GetKRC721CollectionRequest is nil")
	}
	return x.GetKRC721CollectionRequest.toAppMessage()
}

func (x *BugnadMessage_GetKRC721CollectionRequest) fromAppMessage(message *appmessage.GetKRC721CollectionRequestMessage) error {
	x.GetKRC721CollectionRequest = &GetKRC721CollectionRequestMessage{
		Address: message.Address,
	}
	return nil
}

func (x *GetKRC721CollectionRequestMessage) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "GetKRC721CollectionRequestMessage is nil")
	}
	return &appmessage.GetKRC721CollectionRequestMessage{
		Address: x.Address,
	}, nil
}

func (x *BugnadMessage_GetKRC721CollectionResponse) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "BugnadMessage_GetKRC721CollectionResponse is nil")
	}
	return x.GetKRC721CollectionResponse.toAppMessage()
}

func (x *BugnadMessage_GetKRC721CollectionResponse) fromAppMessage(message *appmessage.GetKRC721CollectionResponseMessage) error {
	var err *RPCError
	if message.Error != nil {
		err = &RPCError{Message: message.Error.Message}
	}

	x.GetKRC721CollectionResponse = &GetKRC721CollectionResponseMessage{
		Id:          message.Collection.ID,
		Owner:       message.Collection.Owner,
		Name:        message.Collection.Name,
		Symbol:      message.Collection.Symbol,
		MaxSupply:   message.Collection.MaxSupply,
		TotalSupply: message.Collection.TotalSupply,
		BaseURI:     message.Collection.BaseURI,
		Error:       err,
	}
	return nil
}

func (x *GetKRC721CollectionResponseMessage) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "GetKRC721CollectionResponseMessage is nil")
	}
	rpcErr, err := x.Error.toAppMessage()
	// Error is an optional field
	if err != nil && !errors.Is(err, errorNil) {
		return nil, err
	}
	return &appmessage.GetKRC721CollectionResponseMessage{
		Collection: &appmessage.RPCKRC721Collection{
			ID:          x.Id,
			Owner:       x.Owner,
			Name:        x.Name,
			Symbol:      x.Symbol,
			MaxSupply:   x.MaxSupply,
			TotalSupply: x.TotalSupply,
			BaseURI:     x.BaseURI,
		},
		Error: rpcErr,
	}, nil
}
