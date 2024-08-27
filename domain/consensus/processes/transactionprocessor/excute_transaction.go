package transactionprocessor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/bugnanetwork/bugnad/domain/consensus/model"
	"github.com/bugnanetwork/bugnad/domain/consensus/model/externalapi"
	"github.com/bugnanetwork/bugnad/domain/consensus/utils/krc721"
	"github.com/bugnanetwork/bugnad/domain/consensus/utils/txscript"
	"github.com/bugnanetwork/bugnad/util"
)

type TxPayloadExcutor struct {
	Type   string   `json:"type"`
	Action string   `json:"action"`
	Args   []string `json:"args"`
}

func (t *transactionProcessor) Excute(
	stagingArea *model.StagingArea,
	tx *externalapi.DomainTransaction,
	povBlockHash *externalapi.DomainHash,
) error {
	for _, input := range tx.Inputs {
		if err := t.excuteTXInput(stagingArea, input); err != nil {
			if err.Error() == "invalid script" {
				continue
			}
			return err
		}
	}

	return nil
}

func (t *transactionProcessor) excuteTXInput(stagingArea *model.StagingArea, input *externalapi.DomainTransactionInput) error {
	if !txscript.IsPayToScriptHash(input.UTXOEntry.ScriptPublicKey()) {
		return fmt.Errorf("invalid script")
	}

	if !bytes.Contains(input.SignatureScript, []byte("bugna_script")) {
		return fmt.Errorf("invalid script")
	}

	datas, err := txscript.PushedData(input.SignatureScript)
	if err != nil {
		return fmt.Errorf("err txscript.PushedData: %w", err)
	}

	datas, err = txscript.PushedData(datas[1])
	if err != nil {
		return fmt.Errorf("err txscript.PushedData: %w", err)
	}

	if bytes.Compare(datas[2], []byte("bugna_script")) != 0 {
		return fmt.Errorf("invalid script")
	}

	payload := &TxPayloadExcutor{}
	err = json.Unmarshal(datas[5], payload)
	if err != nil {
		return fmt.Errorf("err json.Unmarshal: %w", err)
	}

	operatorAddr, _ := util.NewAddressPublicKey(datas[0], util.Bech32PrefixBugna)
	operator, err := txscript.PayToAddrScript(operatorAddr)
	if err != nil {
		return fmt.Errorf("err txscript.PayToAddrScript: %w", err)
	}

	switch payload.Type {
	case "krc721":
		return t.excuteKRC721(
			stagingArea,
			input.UTXOEntry.ScriptPublicKey(),
			operator,
			payload)
	default:
		return fmt.Errorf("invalid type: %s", payload.Type)
	}
}

func (t *transactionProcessor) excuteKRC721(
	stagingArea *model.StagingArea,
	inputAddress *externalapi.ScriptPublicKey,
	owner *externalapi.ScriptPublicKey,
	payload *TxPayloadExcutor,
) error {
	switch payload.Action {
	case "deploy":
		if len(payload.Args) != 4 {
			return fmt.Errorf("invalid args, action deploy")
		}

		name := payload.Args[0]
		symbol := payload.Args[1]
		maxSupply, err := strconv.ParseUint(payload.Args[2], 10, 64)
		if err != nil {
			return fmt.Errorf("err strconv.ParseUint: %w", err)
		}
		baseURI := payload.Args[3]

		c, err := krc721.NewKRC721Collection(
			inputAddress,
			owner,
			name,
			symbol,
			maxSupply,
			0,
			baseURI,
		)
		if err != nil {
			return fmt.Errorf("err krc721.NewKRC721Collection: %w", err)
		}

		t.krc721Store.Deploy(
			stagingArea,
			c,
		)
	case "mint":
		if len(payload.Args) != 1 {
			return fmt.Errorf("invalid args, action mint")
		}

		addr, _ := util.DecodeAddress(payload.Args[0], util.Bech32PrefixBugna)
		collectionAddr, _ := txscript.PayToAddrScript(addr)

		err := t.krc721Store.Mint(
			t.databaseContext,
			stagingArea,
			model.ScriptPublicKeyString(collectionAddr.String()),
			model.ScriptPublicKeyString(owner.String()),
		)
		if err != nil {
			return fmt.Errorf("err krc721Store.Mint: %w", err)
		}
	default:
		return fmt.Errorf("invalid action: %s", payload.Action)
	}
	return nil
}
