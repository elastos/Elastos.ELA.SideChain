package servers

import (
	"github.com/elastos/Elastos.ELA.SideChain/core"
	"github.com/elastos/Elastos.ELA.SideChain/blockchain"
	"github.com/elastos/Elastos.ELA.SideChain/vm"
	"github.com/elastos/Elastos.ELA.SideChain/vm/interfaces"
	"github.com/elastos/Elastos.ELA.SideChain/events"
	"github.com/elastos/Elastos.ELA.SideChain/vm/types"
	"github.com/elastos/Elastos.ELA.SideChain/log"

	"github.com/elastos/Elastos.ELA.Utility/common"
)
var Store  blockchain.IChainStore
var Table interfaces.IScriptTable
func RunScript(script []byte) *vm.ExecutionEngine {
	container := core.Transaction{Inputs:[]*core.Input{}, Outputs:[]*core.Output{}}
	dbCache := blockchain.NewDBCache(Store)
	stateMachine := blockchain.NewStateMachine(dbCache, dbCache)

	stateMachine.StateReader.NotifyEvent.Subscribe(events.EventRunTimeNotify, onContractNotify)
	stateMachine.StateReader.LogEvent.Subscribe(events.EventRunTimeLog, onContractLog)
	e := vm.NewExecutionEngine(
		&container,
		new(vm.CryptoECDsa),
		vm.MAXSTEPS,
		Table,
		stateMachine,
		99999 * 100000000,
		vm.Application,
		true,
	)
	e.LoadScript(script, false)
	e.Execute()
	return e
}

func onContractNotify(item interface{}) {
	data := item.(types.StackItem)
	notifyInfo(data)
}

func notifyInfo(item types.StackItem)  {
	switch item.(type) {
	case *types.Boolean:
		log.Info("notifyInfo", item.GetBoolean())
	case *types.Integer:
		log.Info("notifyInfo", item.GetBigInteger())
	case *types.ByteArray:
		log.Info("notifyInfo", common.BytesToHexString(item.GetByteArray()))
	case *types.GeneralInterface:
		interop := item.GetInterface()
		log.Info("notifyInfo", string(interop.Bytes()))
	case *types.Array:
		items := item.GetArray();
		for i := 0; i < len(items); i++ {
			notifyInfo(items[i])
		}
	}
}

func onContractLog(item interface{})  {
	data := item.(types.StackItem)
	msg := data.GetByteArray()
	log.Info("onRunTimeLog:", string(msg))
}