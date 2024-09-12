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
)

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
	datas, err := txscript.PushedData(input.SignatureScript)
	if err != nil {
		return fmt.Errorf("err txscript.PushedData: %w", err)
	}

	if len(datas) < 3 {
		return fmt.Errorf("invalid script")
	}

	kind := string(datas[len(datas)-3])
	payload := datas[0]

	redeemScript, err := txscript.PushedData(datas[len(datas)-1])
	if err != nil {
		return fmt.Errorf("err txscript.PushedData: %w", err)
	}

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

	switch kind {
	case "deploy":
		_, _, _, err := evm.Create(vm.AccountRef(caller), payload, evm.GasLimit, big.NewInt(0))
		if err != nil {
			return fmt.Errorf("err evm.Create: %w", err)
		}
	case "interact":
		toAddr := vm.BytesToAddress(datas[1])
		_, _, err = evm.Call(vm.AccountRef(caller), toAddr, payload, evm.GasLimit, big.NewInt(0))
		if err != nil {
			return fmt.Errorf("err evm.Call: %w", err)
		}
	default:
		return fmt.Errorf("invalid type: %s", kind)
	}

	return nil
}

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
