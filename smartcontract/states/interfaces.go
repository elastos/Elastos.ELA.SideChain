package states

import (
	"io"
	"github.com/elastos/Elastos.ELA.SideChain/vm/interfaces"

	"bytes"
	"github.com/elastos/Elastos.ELA.SideChain/common"
	)

type IStateValueInterface interface {
	Serialize(w io.Writer) error
	Deserialize(r io.Reader) error
	interfaces.IGeneralInterface
}

type IStateKeyInterface interface {
	Serialize(w io.Writer)  error
	Deserialize(r io.Reader) error
}

var (
	StatesMap = map[common.DataEntryPrefix]IStateValueInterface{
		common.ST_Contract: new(ContractState),
	}
)

func GetStateValue(prefix common.DataEntryPrefix, data[]byte) (IStateValueInterface, error)  {
	r := bytes.NewBuffer(data)
	state := StatesMap[prefix]
	err := state.Deserialize(r)
	if err != nil {
		return nil, err
	}
	return state, nil
}