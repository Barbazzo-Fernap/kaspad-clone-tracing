package main

import (
	"encoding/hex"
	"fmt"
	"time"

	"github.com/bugnanetwork/bugnad/domain/consensus/utils/txscript"
	"github.com/bugnanetwork/bugnad/infrastructure/network/rpcclient"
)

func deploySmartContract(cfg *deployConfig) error {
	myKeyPair, addr, err := parsePrivateKeyInKeyPair(cfg.PrivateKey, cfg.configFlags.ActiveNetParams.Prefix)
	if err != nil {
		return fmt.Errorf("Error parsing private key: %s", err)
	}

	rpcAddress, err := cfg.NetParams().NormalizeRPCServerAddress(cfg.RPCServer)
	if err != nil {
		return fmt.Errorf("Error parsing RPC server address: %s", err)
	}

	//RPC client activation (to communicate with Kaspad)
	client, err := rpcclient.NewRPCClient(rpcAddress)
	if err != nil {
		return fmt.Errorf("Error connecting to the RPC server: %s", err)
	}

	client.SetTimeout(5 * time.Minute)

	resp, err := client.GetBlockDAGInfo()
	if err != nil {
		return fmt.Errorf("Error getting block DAG info: %s", err)
	}

	currentBlockHash := resp.VirtualParentHashes[0]

	contractCode, err := hex.DecodeString(cfg.ContractCode)
	if err != nil {
		return fmt.Errorf("Error decoding contract code: %s", err)
	}

	input, err := NewDeploySmartContractinput(addr.ScriptAddress(), contractCode, cfg.configFlags.ActiveNetParams.Prefix)
	if err != nil {
		return fmt.Errorf("Error creating deploy smart contract input: %s", err)
	}

	transactionID, err := initiateAddressScript(&cfg.configFlags, client, myKeyPair, addr, input)
	if err != nil {
		return fmt.Errorf("Error initiating address script: %s", err)
	}

	log.Info("Waiting for network to deploy the contract....")
	log.Infof("Contract P2SH address: %s", input.Address().String())
	log.Infof("Transaction ID: %s", transactionID)

	// Redeem from P2SH contract
	txID, err := redeemContract(client, transactionID, myKeyPair, addr, input)
	if err != nil {
		return fmt.Errorf("Error redeeming contract: %s", err)
	}

	for {
		select {
		case <-time.After(10 * time.Second):
			return fmt.Errorf("Contract not yet deployed.")
		default:
			blocks, err := client.GetBlocks(currentBlockHash, true, true)
			if err != nil {
				return fmt.Errorf("Error getting blocks: %s", err)
			}

			for _, block := range blocks.Blocks {
				for _, txn := range block.Transactions {
					if txn.VerboseData.TransactionID == txID {
						log.Infof("Contract deployed successfully. BlockHash: %s Transaction ID: %s", block.VerboseData.Hash, txID)
						return nil
					}
				}
			}
		}
	}
}

func secretContract(pubkhThem []byte, contractCode []byte) ([]byte, error) {
	builder := txscript.NewScriptBuilder()

	// Verify their signature is being used to redeem the output
	builder.AddData(pubkhThem)
	builder.AddOp(txscript.OpCheckSig)

	for i := 0; i < len(contractCode); i += txscript.MaxScriptElementSize {
		builder.AddOp(txscript.OpDrop)
	}

	builder.AddOp(txscript.Op2Drop)
	return builder.Script()
}
