package rpchandlers

import (
	"github.com/bugnanetwork/bugnad/app/appmessage"
	"github.com/bugnanetwork/bugnad/app/rpc/rpccontext"
	"github.com/bugnanetwork/bugnad/domain/consensus/utils/txscript"
	"github.com/bugnanetwork/bugnad/infrastructure/network/netadapter/router"
	"github.com/bugnanetwork/bugnad/util"
)

// HandleGetKRC721Collection handles the respectively named RPC command
func HandleGetKRC721Collection(context *rpccontext.Context, _ *router.Router, request appmessage.Message) (appmessage.Message, error) {
	req := request.(*appmessage.GetKRC721CollectionRequestMessage)
	addr, err := util.DecodeAddress(req.Address, context.Config.NetParams().Prefix)

	response := appmessage.NewGetKRC721CollectionResponseMessage()
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

	_, id, _ := txscript.ExtractScriptPubKeyAddress(c.ID(), context.Config.NetParams())
	_, ownerAddr, _ := txscript.ExtractScriptPubKeyAddress(c.Owner(), context.Config.NetParams())

	response.Collection = &appmessage.RPCKRC721Collection{
		ID:          id.EncodeAddress(),
		Owner:       ownerAddr.EncodeAddress(),
		Name:        c.Name(),
		Symbol:      c.Symbol(),
		MaxSupply:   c.MaxSupply(),
		TotalSupply: c.TotalSupply(),
		BaseURI:     c.BaseURI(),
	}
	return response, nil
}
