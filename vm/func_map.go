package vm

import (
	"github.com/elastos/Elastos.ELA.SideChain/vm/types"
	"errors"
)

func opNewMap(e *ExecutionEngine) (VMState, error) {
	e.GetEvaluationStack().Push(types.NewDictionary())
	return NONE, nil
}

func opAppend(e *ExecutionEngine) (VMState, error) {
	newItem := PopStackItem(e)
	itemArr := PopStackItem(e)
	if _, ok := itemArr.(*types.Array); ok {
		items := itemArr.GetArray()
		items = append(items, newItem)
	} else {
		return  FAULT, errors.New("opAppend data error")
	}
	return NONE, nil
}
