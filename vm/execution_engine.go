package vm

import (
	"io"
	_ "math/big"
	_ "sort"

	"github.com/elastos/Elastos.ELA.SideChain/vm/interfaces"
	"github.com/elastos/Elastos.ELA.SideChain/vm/utils"
	"github.com/elastos/Elastos.ELA.Utility/common"
)

const MAXSTEPS int = 1200

func NewExecutionEngine(container interfaces.IDataContainer, crypto interfaces.ICrypto, maxSteps int,
						table interfaces.IScriptTable, service IGeneralService, gas common.Fixed64) *ExecutionEngine {
	var engine ExecutionEngine

	engine.crypto = crypto
	engine.table = table

	engine.dataContainer = container
	engine.invocationStack = utils.NewRandAccessStack()
	engine.opCount = 0

	engine.evaluationStack = utils.NewRandAccessStack()
	engine.altStack = utils.NewRandAccessStack()
	engine.state = BREAK

	engine.context = nil
	engine.opCode = 0

	engine.maxSteps = maxSteps

	engine.service = NewGeneralService()
	if service != nil {
		engine.service.MergeMap(service.GetServiceMap())
	}


	engine.gas = gas.IntValue()
	return &engine
}

type ExecutionEngine struct {
	crypto  interfaces.ICrypto
	table   interfaces.IScriptTable
	service *GeneralService

	dataContainer   interfaces.IDataContainer
	invocationStack *utils.RandomAccessStack
	opCount         int

	maxSteps int

	evaluationStack *utils.RandomAccessStack
	altStack        *utils.RandomAccessStack
	state           VMState

	context *ExecutionContext

	//current opcode
	opCode OpCode
	gas    int64
}

func (e *ExecutionEngine) GetDataContainer() interfaces.IDataContainer {
	return e.dataContainer
}

func (e *ExecutionEngine) GetState() VMState {
	return e.state
}

func (e *ExecutionEngine) GetEvaluationStack() *utils.RandomAccessStack {
	return e.evaluationStack
}

func (e *ExecutionEngine) GetExecuteResult() bool {
	return AssertStackItem(e.evaluationStack.Pop()).GetBoolean()
}

func (e *ExecutionEngine) ExecutingScript() []byte {
	context := AssertExecutionContext(e.invocationStack.Peek(0))
	if context != nil {
		return context.Script
	}
	return nil
}

func (e *ExecutionEngine) CallingScript() []byte {
	if e.invocationStack.Count() > 1 {
		context := AssertExecutionContext(e.invocationStack.Peek(1))
		if context != nil {
			return context.Script
		}
		return nil
	}
	return nil
}


func (e *ExecutionEngine) Create(caller common.Uint168, code []byte) ([]byte, error) {
	return code, nil
}

func (e *ExecutionEngine) Call(caller common.Uint168, codeHash common.Uint168, input []byte) ([]byte, error) {
	e.LoadScript(input, false)
	e.Execute()
	return nil, nil
}

func (e *ExecutionEngine) Hash160(script []byte) []byte {
	return e.crypto.Hash168(script)
}

func (e *ExecutionEngine) EntryScript() []byte {
	context := AssertExecutionContext(e.invocationStack.Peek(e.invocationStack.Count() - 1))
	if context != nil {
		return context.Script
	}
	return nil
}

func (e *ExecutionEngine) LoadScript(script []byte, pushOnly bool) {
	e.invocationStack.Push(NewExecutionContext(script, pushOnly, nil))
}

func (e *ExecutionEngine) Execute() {
	e.state = e.state & (^BREAK)
	for {
		if e.state == FAULT || e.state == HALT || e.state == BREAK {
			break
		}
		err := e.StepInto()
		if err != nil {
			break;
		}
	}
}

func (e *ExecutionEngine) StepInto() error {
	if e.invocationStack.Count() == 0 {
		e.state = VMState(e.state | HALT)
	}
	if e.state&HALT == HALT || e.state&FAULT == FAULT {
		return nil
	}

	context := AssertExecutionContext(e.invocationStack.Peek(0))

	var opCode OpCode

	if context.GetInstructionPointer() >= len(context.Script) {
		opCode = RET
	} else {
		op, err :=  context.OpReader.ReadByte()
		if err == io.EOF && opCode == 0 {
			e.state = FAULT
			return err
		}
		opCode = OpCode(op)
	}

	e.opCount++
	state, err := e.ExecuteOp(OpCode(opCode), context)
	switch state {
	case VMState(HALT):
		e.state = VMState(e.state | HALT)
		return err
	case VMState(FAULT):
		e.state = VMState(e.state | FAULT)
		return err
	}
	return nil
}

func (e *ExecutionEngine) ExecuteOp(opCode OpCode, context *ExecutionContext) (VMState, error) {
	if opCode > PUSH16 && opCode != RET && context.PushOnly {
		return FAULT, nil
	}
	if opCode > PUSH16 && e.opCount > e.maxSteps {
		return FAULT, nil
	}
	if opCode >= PUSHBYTES1 && opCode <= PUSHBYTES75 {
		err := pushData(e, context.OpReader.ReadBytes(int(opCode)))
		if err != nil {
			return FAULT, err
		}
		return NONE, nil
	}
	e.opCode = opCode
	e.context = context
	opExec := OpExecList[opCode]
	if opExec.Exec == nil {
		return FAULT, nil
	}
	state, err := opExec.Exec(e)
	if err != nil {
		return state, err
	}
	return NONE, nil
}

func (e *ExecutionEngine) StepOut() {
	e.state = e.state & (^BREAK)
	c := e.invocationStack.Count()
	for {
		if e.state == FAULT || e.state == HALT || e.state == BREAK || e.invocationStack.Count() >= c {
			break
		}
		e.StepInto()
	}
}

func (e *ExecutionEngine) StepOver() {
	if e.state == FAULT || e.state == HALT {
		return
	}
	e.state = e.state & (^BREAK)
	c := e.invocationStack.Count()
	for {
		if e.state == FAULT || e.state == HALT || e.state == BREAK || e.invocationStack.Count() > c {
			break
		}
		e.StepInto()
	}
}

func (e *ExecutionEngine) AddBreakPoint(position uint) {
	//b := e.context.BreakPoints
	//b = append(b, position)
}

func (e *ExecutionEngine) RemoveBreakPoint(position uint) bool {
	//if e.invocationStack.Count() == 0 { return false }
	//b := e.context.BreakPoints
	return true
}
