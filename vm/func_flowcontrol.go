package vm

import (
	"github.com/elastos/Elastos.ELA.SideChain/vm/errors"
)

func opNop(e *ExecutionEngine) (VMState, error) {
	//time.Sleep(1 * time.Millisecond)
	return NONE, nil
}

func opJmp(e *ExecutionEngine) (VMState, error) {
	offset := int(e.context.OpReader.ReadInt16())
	offset = e.context.GetInstructionPointer() + offset - 3

	if offset < 0 || offset > len(e.context.Script) {
		return FAULT, errors.ErrFault
	}
	fValue := true
	if e.opCode > JMP {
		s := AssertStackItem(e.evaluationStack.Pop())
		fValue = s.GetBoolean()
		if e.opCode == JMPIFNOT {
			fValue = !fValue
		}
	}
	if fValue {
		e.context.SetInstructionPointer(offset)
	}

	return NONE, nil
}

func opCall(e *ExecutionEngine) (VMState, error) {
	e.invocationStack.Push(e.context.Clone())
	e.context.SetInstructionPointer(e.context.GetInstructionPointer() + 2)
	e.opCode = JMP
	e.context = e.invocationStack.Peek(0).(*ExecutionContext)
	opJmp(e)
	return NONE, nil
}

func opRet(e *ExecutionEngine) (VMState, error) {
	e.invocationStack.Pop()
	return NONE, nil
}

func opAppCall(e *ExecutionEngine) (VMState, error) {
	if e.table == nil {
		return FAULT, nil
	}
	script_hash := e.context.OpReader.ReadBytes(21)
	script := e.table.GetScript(script_hash)
	if script == nil {
		return FAULT, nil
	}
	if e.opCode == TAILCALL {
		e.invocationStack.Pop()
	}
	e.LoadScript(script, false)
	return NONE, nil
}

func opSysCall(e *ExecutionEngine) (VMState, error) {
	if e.service == nil {
		return FAULT, nil
	}
	success := e.service.Invoke(e.context.OpReader.ReadVarString(), e)
	if success {
		return NONE, nil
	} else {
		return FAULT, nil
	}
}
