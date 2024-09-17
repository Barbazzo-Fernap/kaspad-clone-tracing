package main

import (
	"github.com/bugnanetwork/bugnad/domain/consensus/utils/txscript"
	"github.com/bugnanetwork/bugnad/util"
)

type SmartContractInput interface {
	Address() util.Address
	RedeemScript(signature []byte) ([]byte, error)
}

type DeploySmartContractinput struct {
	addr           util.Address
	secretContract []byte
	contractCode   []byte
}

func NewDeploySmartContractinput(
	pubkhThem []byte,
	contractCode []byte,
	prefix util.Bech32Prefix,
) (*DeploySmartContractinput, error) {
	builder := txscript.NewScriptBuilder()

	// Verify their signature is being used to redeem the output
	builder.AddData(pubkhThem)
	builder.AddOp(txscript.OpCheckSig)

	for i := 0; i < len(contractCode); i += txscript.MaxScriptElementSize {
		builder.AddOp(txscript.OpDrop)
	}

	builder.AddOp(txscript.Op2Drop)
	secretContract, err := builder.Script()
	if err != nil {
		return nil, err
	}

	contractP2SHaddress, err := util.NewAddressScriptHash(secretContract, prefix)
	if err != nil {
		return nil, err
	}

	return &DeploySmartContractinput{
		addr:           contractP2SHaddress,
		secretContract: secretContract,
		contractCode:   contractCode,
	}, nil
}

func (d *DeploySmartContractinput) Address() util.Address {
	return d.addr
}

func (d *DeploySmartContractinput) RedeemScript(signature []byte) ([]byte, error) {
	builder := txscript.NewScriptBuilder()

	builder.AddData([]byte("bugna_script"))
	builder.AddData([]byte("deploy"))

	for i := 0; i < len(d.contractCode); i += txscript.MaxScriptElementSize {
		start := i
		end := i + txscript.MaxScriptElementSize
		if end > len(d.contractCode) {
			end = len(d.contractCode)
		}

		d := d.contractCode[start:end]
		builder.AddData(d)
	}

	builder.AddData(signature)
	builder.AddData(d.secretContract)
	return builder.Script()

}
