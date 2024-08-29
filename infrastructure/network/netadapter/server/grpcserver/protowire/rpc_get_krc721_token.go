package protowire

import (
	"github.com/bugnanetwork/bugnad/app/appmessage"
	"github.com/pkg/errors"
)

func (x *BugnadMessage_GetKRC721TokenRequest) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "BugnadMessage_GetKRC721TokenRequest is nil")
	}
	return x.GetKRC721TokenRequest.toAppMessage()
}

func (x *BugnadMessage_GetKRC721TokenRequest) fromAppMessage(message *appmessage.GetKRC721TokenRequestMessage) error {
	x.GetKRC721TokenRequest = &GetKRC721TokenRequestMessage{
		CollectionAddress: message.CollectionAddress,
		TokenId:           0,
	}
	return nil
}

func (x *GetKRC721TokenRequestMessage) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "GetKRC721TokenRequestMessage is nil")
	}
	return &appmessage.GetKRC721TokenRequestMessage{
		CollectionAddress: x.CollectionAddress,
		TokenID:           x.TokenId,
	}, nil
}

func (x *BugnadMessage_GetKRC721TokenResponse) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "BugnadMessage_GetKRC721TokenResponse is nil")
	}
	return x.GetKRC721TokenResponse.toAppMessage()
}

func (x *BugnadMessage_GetKRC721TokenResponse) fromAppMessage(message *appmessage.GetKRC721TokenResponseMessage) error {
	var err *RPCError
	if message.Error != nil {
		err = &RPCError{Message: message.Error.Message}
	}

	if message.Token == nil {
		x.GetKRC721TokenResponse = &GetKRC721TokenResponseMessage{
			TokenId: 0,
			Owner:   "",
			Uri:     "",
			Error:   err,
		}
		return nil
	}

	x.GetKRC721TokenResponse = &GetKRC721TokenResponseMessage{
		TokenId: message.Token.ID,
		Owner:   message.Token.Owner,
		Uri:     message.Token.URI,
		Error:   err,
	}
	return nil
}

func (x *GetKRC721TokenResponseMessage) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "GetKRC721TokenResponseMessage is nil")
	}
	rpcErr, err := x.Error.toAppMessage()
	// Error is an optional field
	if err != nil && !errors.Is(err, errorNil) {
		return nil, err
	}
	return &appmessage.GetKRC721TokenResponseMessage{
		Token: &appmessage.RPCKRC721Token{
			ID:    x.TokenId,
			Owner: x.Owner,
			URI:   x.Uri,
		},
		Error: rpcErr,
	}, nil
}
