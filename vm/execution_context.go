package vm

import (
	"github.com/elastos/Elastos.ELA.SideChain/vm/utils"
	"github.com/elastos/Elastos.ELA.SideChain/common"
)

type ExecutionContext struct {
	Script             []byte
	OpReader           *utils.VmReader
	PushOnly           bool
	BreakPoints        []uint
	InstructionPointer int
	CodeHash           []byte
}

func NewExecutionContext(script []byte, pushOnly bool, breakPoints []uint) *ExecutionContext {
	var executionContext ExecutionContext
	executionContext.Script = script
	executionContext.OpReader = utils.NewVmReader(script)
	executionContext.PushOnly = pushOnly
	executionContext.BreakPoints = breakPoints
	executionContext.InstructionPointer = executionContext.OpReader.Position()
	return &executionContext
}

func (ec* ExecutionContext) GetCodeHash() []byte {
	if ec.CodeHash == nil {
		hash, err := common.ToCodeHash(ec.Script)
		if err != nil {
			return nil
		}
		ec.CodeHash = hash.Bytes()
	}
	return ec.CodeHash
}

func (ec *ExecutionContext) NextInstruction() OpCode {
	return OpCode(ec.Script[ec.OpReader.Position()])
}

func (ec *ExecutionContext) Clone() *ExecutionContext {
	return NewExecutionContext(ec.Script, ec.PushOnly, ec.BreakPoints)
}
