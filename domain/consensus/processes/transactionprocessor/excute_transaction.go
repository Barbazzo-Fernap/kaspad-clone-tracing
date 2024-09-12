package transactionprocessor

import (
	"fmt"
	"math/big"
	"time"

	"github.com/bugnanetwork/bugnad/domain/bvm/state"
	"github.com/bugnanetwork/bugnad/domain/bvm/vm"
	"github.com/bugnanetwork/bugnad/domain/consensus/model"
	"github.com/bugnanetwork/bugnad/domain/consensus/model/externalapi"
	"github.com/bugnanetwork/bugnad/domain/consensus/utils/consensushashing"
	"github.com/bugnanetwork/bugnad/domain/consensus/utils/constants"
	"github.com/bugnanetwork/bugnad/domain/consensus/utils/txscript"
	"github.com/bugnanetwork/bugnad/util"
)

type TxPayloadExcutor struct {
	Action string `json:"action"`
	Args   []any  `json:"args"`
}

func (t *transactionProcessor) Excute(
	stagingArea *model.StagingArea,
	povBlockHash *externalapi.DomainHash,
	blockDaaScore uint64,
	tx *externalapi.DomainTransaction,
) error {
	s := stagingArea
	if povBlockHash.String() == "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff" {
		s = model.NewStagingArea()
	}

	for i, input := range tx.Inputs {
		if err := t.excuteTXInput(tx, blockDaaScore, s, input, i); err != nil {
			if err.Error() == "invalid script" {
				continue
			}
			return err
		}
	}

	return nil
}

func (t *transactionProcessor) excuteTXInput(tx *externalapi.DomainTransaction, blockDaaScore uint64, stagingArea *model.StagingArea, input *externalapi.DomainTransactionInput, inputIndex int) error {
	// if !txscript.IsPayToScriptHash(input.UTXOEntry.ScriptPublicKey()) {
	// 	return fmt.Errorf("invalid script")
	// }

	// if !bytes.Contains(input.SignatureScript, []byte("bugna_script")) {
	// 	return fmt.Errorf("invalid script")
	// }

	datas, err := txscript.PushedData(input.SignatureScript)
	if err != nil {
		return fmt.Errorf("err txscript.PushedData: %w", err)
	}

	if len(datas) < 3 {
		return fmt.Errorf("invalid script")
	}

	kind := string(datas[len(datas)-3])
	// payload := &TxPayloadExcutor{}
	// err = json.Unmarshal(datas[InputPayloadJSON], payload)
	// if err != nil {
	// 	return fmt.Errorf("err json.Unmarshal: %w", err)
	// }
	payload := datas[0]

	redeemScript, err := txscript.PushedData(datas[len(datas)-1])
	if err != nil {
		return fmt.Errorf("err txscript.PushedData: %w", err)
	}

	// operatorAddr, _ := util.NewAddressPublicKey(redeemScript[0], util.Bech32PrefixBugna)
	// operator, err := txscript.PayToAddrScript(operatorAddr)
	// if err != nil {
	// 	return fmt.Errorf("err txscript.PayToAddrScript: %w", err)
	// }

	caller := vm.BytesToAddress([]byte(redeemScript[0]))

	stateDB := t.bvmStore.StateDBWrapper(t.databaseContext, stagingArea).(*state.StateDB)
	txID := consensushashing.TransactionID(tx)

	defer func() {
		logs := stateDB.GetLogs(vm.BytesToHash(txID.ByteSlice()))
		for _, log := range logs {
			topics := []externalapi.DomainHash{}
			for _, topic := range log.Topics {
				hash, _ := externalapi.NewDomainHashFromByteSlice(topic.Bytes())
				topics = append(topics, *hash)
			}

			scriptPublicKey, _ := txscript.ScriptHashToScriptPublicKey(log.Address.Bytes())
			tx.Logs = append(tx.Logs, &externalapi.DomainTransactionLog{
				ScriptPublicKey: &externalapi.ScriptPublicKey{
					Script:  scriptPublicKey,
					Version: constants.MaxScriptPublicKeyVersion,
				},
				Topics: topics,
				Data:   log.Data,
				Index:  uint64(log.Index),
			})
		}

		journal := stateDB.DumpJournal()
		tx.Journal = journal

		stateDB.IntermediateRoot(true)
	}()

	context := CreateExecuteContext(blockDaaScore, caller, vm.BytesToHash(txID.ByteSlice()), uint32(inputIndex))
	chainConfig := CreateChainConfig()
	vmConfig := CreateVMDefaultConfig()

	evm := vm.NewEVM(context, stateDB, chainConfig, vmConfig)

	// retrieInput := vm.Hex2Bytes("2e64cec1")
	// ret, _, err := evm.Call(
	// 	vm.AccountRef(caller),
	// 	vm.BytesToAddress(vm.Hex2Bytes("0x0000000000000000000000004CA5C394c676350B641d977F289F096E9469250F")),
	// 	retrieInput,
	// 	evm.GasLimit,
	// 	new(big.Int))
	// if err != nil {
	// 	fmt.Println(err)
	// } else {
	// 	fmt.Println("cccccc -------", big.NewInt(0).SetBytes(ret))
	// }

	switch kind {
	case "deploy":
		nonce := stateDB.GetNonce(caller)
		fmt.Println("------ nonce: ", nonce)
		_, contractAddr, _, err := evm.Create(vm.AccountRef(caller), payload, evm.GasLimit, big.NewInt(0))
		if err != nil {
			return fmt.Errorf("err evm.Create: %w", err)
		}

		addr, _ := util.NewAddressScriptHashFromHash(contractAddr.Bytes(), util.Bech32PrefixBugna)

		// fmt.Println("------ contractAddr: ", contractAddr.Hex())
		// fmt.Printf("------ contractAddr: %x", addr.ScriptAddress())
		fmt.Println("------ contractAddr: ", addr.EncodeAddress())
	case "interact":
		toAddr := vm.BytesToAddress(datas[1])

		retrieInput := vm.Hex2Bytes("2e64cec1")
		ret, _, err := evm.Call(
			vm.AccountRef(caller),
			toAddr,
			retrieInput,
			evm.GasLimit,
			new(big.Int))
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println("------- output", big.NewInt(0).SetBytes(ret))
		}

		fmt.Printf("payload: %x\n", payload)
		fmt.Printf("--------- toAddr: %x\n", toAddr)

		_, _, err = evm.Call(vm.AccountRef(caller), toAddr, payload, evm.GasLimit, big.NewInt(0))
		if err != nil {
			return fmt.Errorf("err evm.Call: %w", err)
		}
	// case "krc721":
	// 	return t.excuteKRC721(
	// 		stagingArea,
	// 		input.UTXOEntry.ScriptPublicKey(),
	// 		operator,
	// 		payload)
	default:
		return fmt.Errorf("invalid type: %s", kind)
	}

	return nil
}

//
// func (t *transactionProcessor) excuteKRC721(
// 	stagingArea *model.StagingArea,
// 	inputAddress *externalapi.ScriptPublicKey,
// 	owner *externalapi.ScriptPublicKey,
// 	payload *TxPayloadExcutor,
// ) error {
// 	switch payload.Action {
// 	case "deploy":
// 		if len(payload.Args) != 4 {
// 			return fmt.Errorf("invalid args, action deploy")
// 		}
//
// 		name, ok1 := payload.Args[0].(string)
// 		symbol, ok2 := payload.Args[1].(string)
// 		maxSupply, oke3 := payload.Args[2].(float64)
// 		baseURI, ok4 := payload.Args[3].(string)
// 		if !ok1 || !ok2 || !oke3 || !ok4 {
// 			return fmt.Errorf("invalid args, action deploy, args: [%v, %v, %v, %v]", ok1, ok2, oke3, ok4)
// 		}
//
// 		c, err := krc721.NewKRC721Collection(
// 			inputAddress,
// 			owner,
// 			name,
// 			symbol,
// 			uint64(maxSupply),
// 			0,
// 			baseURI,
// 		)
// 		if err != nil {
// 			return fmt.Errorf("err krc721.NewKRC721Collection: %w", err)
// 		}
//
// 		t.krc721Store.Deploy(
// 			stagingArea,
// 			c,
// 		)
// 	case "mint":
// 		if len(payload.Args) != 1 {
// 			return fmt.Errorf("invalid args, action mint")
// 		}
//
// 		addr, err := util.DecodeAddress(fmt.Sprintf("%v", payload.Args[0]), util.Bech32PrefixBugna)
// 		if err != nil {
// 			return fmt.Errorf("util.DecodeAddress: %w", err)
// 		}
//
// 		collectionAddr, _ := txscript.PayToAddrScript(addr)
//
// 		err = t.krc721Store.Mint(
// 			t.databaseContext,
// 			stagingArea,
// 			model.ScriptPublicKeyString(collectionAddr.String()),
// 			model.ScriptPublicKeyString(owner.String()),
// 		)
// 		if err != nil {
// 			return fmt.Errorf("err krc721Store.Mint: %w", err)
// 		}
// 	case "transfer":
// 		if len(payload.Args) != 3 {
// 			return fmt.Errorf("invalid args, action transfer")
// 		}
//
// 		addr, err := util.DecodeAddress(fmt.Sprintf("%v", payload.Args[0]), util.Bech32PrefixBugna)
// 		if err != nil {
// 			return fmt.Errorf("util.DecodeAddress: %w", err)
// 		}
// 		collectionAddr, _ := txscript.PayToAddrScript(addr)
//
// 		addr, err = util.DecodeAddress(fmt.Sprintf("%v", payload.Args[1]), util.Bech32PrefixBugna)
// 		if err != nil {
// 			return fmt.Errorf("util.DecodeAddress: %w", err)
// 		}
// 		toAddr, _ := txscript.PayToAddrScript(addr)
//
// 		tokenID, ok := payload.Args[2].(float64)
// 		if !ok {
// 			return fmt.Errorf("invalid args, action transfer")
// 		}
//
// 		err = t.krc721Store.TransferFrom(
// 			t.databaseContext,
// 			stagingArea,
// 			model.ScriptPublicKeyString(collectionAddr.String()),
// 			model.ScriptPublicKeyString(owner.String()),
// 			model.ScriptPublicKeyString(owner.String()),
// 			model.ScriptPublicKeyString(toAddr.String()),
// 			uint64(tokenID),
// 		)
// 		if err != nil {
// 			return fmt.Errorf("err krc721Store.TransferFrom: %w", err)
// 		}
// 	default:
// 		return fmt.Errorf("invalid action: %s", payload.Action)
// 	}
// 	return nil
// }

func CreateLogTracer() *vm.StructLogger {
	logConf := vm.LogConfig{
		DisableMemory:  false,
		DisableStack:   false,
		DisableStorage: false,
		Debug:          false,
		Limit:          0,
	}
	return vm.NewStructLogger(&logConf)

}
func CreateChainConfig() *vm.ChainConfig {
	chainCfg := vm.ChainConfig{
		ChainID: big.NewInt(1),
	}
	return &chainCfg
}
func CreateExecuteContext(
	blockDaaScore uint64,
	caller vm.Address,
	txHash vm.Hash,
	txIndex uint32,
) vm.Context {
	context := vm.Context{
		Origin:      caller,
		GasPrice:    new(big.Int),
		TxHash:      txHash,
		TxIndex:     txIndex,
		GasLimit:    vm.MaxUint64,
		BlockNumber: big.NewInt(int64(blockDaaScore)),
		Time:        big.NewInt(time.Now().Unix()),
		Difficulty:  new(big.Int),
	}
	return context
}

func CreateVMDefaultConfig() vm.Config {
	return vm.Config{
		Debug:                   false,
		Tracer:                  CreateLogTracer(),
		NoRecursion:             false,
		EnablePreimageRecording: false,
	}

}
