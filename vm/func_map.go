package vm

import (
	"github.com/elastos/Elastos.ELA.SideChain/vm/types"
	"errors"
)

func opNewMap(e *ExecutionEngine) (VMState, error) {
	e.GetEvaluationStack().Push(types.NewDictionary())
	return NONE, nil
}

func opKeys(e *ExecutionEngine) (VMState, error) {
	items := PopStackItem(e)
	if _,ok := items.(*types.Dictionary); ok {
		PushData(e, items.(*types.Dictionary).GetKeys())
	} else {
		return FAULT, errors.New("opKeys type error")
	}
	return NONE, nil
}

func opValues(e *ExecutionEngine) (VMState, error) {
	itemArr := PopStackItem(e)
	values := make([]types.StackItem, 0)
	if _, ok := itemArr.(*types.Array); ok {
		items := itemArr.(*types.Array).GetArray()
		values = append(values, items...)
	}else if _,ok := itemArr.(*types.Dictionary); ok {
		items := itemArr.(*types.Dictionary).GetValues()
		values = append(values, items.GetArray()...)
	} else {
		return FAULT, errors.New("opValues type error")
	}
	PushData(e, values)
	return NONE, nil
}