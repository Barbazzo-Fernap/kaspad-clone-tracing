package rpchandlers

import (
	"github.com/bugnanetwork/bugnad/app/appmessage"
	"github.com/bugnanetwork/bugnad/app/rpc/rpccontext"
	"github.com/bugnanetwork/bugnad/domain/consensus/utils/txscript"
	"github.com/bugnanetwork/bugnad/infrastructure/network/netadapter/router"
	"github.com/bugnanetwork/bugnad/util"
)

// HandleGetKRC721Token handles the respectively named RPC command
func HandleGetKRC721Token(context *rpccontext.Context, _ *router.Router, request appmessage.Message) (appmessage.Message, error) {
	req := request.(*appmessage.GetKRC721TokenRequestMessage)
	addr, err := util.DecodeAddress(req.CollectionAddress, context.Config.NetParams().Prefix)

	response := appmessage.NewGetKRC721TokenResponseMessage()
	if err != nil {
		response.Error = &appmessage.RPCError{
			Message: err.Error(),
		}
		return response, nil
	}

	collectionAddr, err := txscript.PayToAddrScript(addr)
	if err != nil {
		response.Error = &appmessage.RPCError{
			Message: err.Error(),
		}
		return response, nil
	}

	c, err := context.Domain.Consensus().GetKRC721Collection(collectionAddr)
	if err != nil {
		response.Error = &appmessage.RPCError{
			Message: err.Error(),
		}
		return response, nil
	}

	ownerOfToken, err := context.Domain.Consensus().OwnerOfKRC721Token(c.ID(), req.TokenID)
	if err != nil {
		response.Error = &appmessage.RPCError{
			Message: err.Error(),
		}
		return response, nil
	}

	_, ownerAddr, _ := txscript.ExtractScriptPubKeyAddress(ownerOfToken, context.Config.NetParams())

	response.Token = &appmessage.RPCKRC721Token{
		ID:    req.TokenID,
		Owner: ownerAddr.EncodeAddress(),
		URI:   c.TokenURI(req.TokenID),
	}
	return response, nil
}
