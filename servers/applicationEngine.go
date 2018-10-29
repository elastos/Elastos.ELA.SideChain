package servers

import (
	"github.com/elastos/Elastos.ELA.SideChain/core"
	"github.com/elastos/Elastos.ELA.SideChain/blockchain"
	"github.com/elastos/Elastos.ELA.SideChain/vm"
	"github.com/elastos/Elastos.ELA.SideChain/vm/interfaces"
)
var Store  blockchain.IChainStore
var Table interfaces.IScriptTable
func RunScript(script []byte) *vm.ExecutionEngine {
	container := core.Transaction{Inputs:[]*core.Input{}, Outputs:[]*core.Output{}}
	dbCache := blockchain.NewDBCache(Store)
	stateMachine := blockchain.NewStateMachine(dbCache, dbCache)
	e := vm.NewExecutionEngine(
		&container,
		new(vm.CryptoECDsa),
		vm.MAXSTEPS,
		Table,
		stateMachine,
		99999 * 100000000,
		vm.Application,
	)
	e.LoadScript(script, false)
	e.Execute()
	return e
}