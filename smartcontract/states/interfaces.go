package states

import (
	"io"
	"bytes"

	"github.com/elastos/Elastos.ELA.SideChain/vm/interfaces"
	. "github.com/elastos/Elastos.ELA.SideChain/store"
)

type IStateValueInterface interface {
	Serialize(w io.Writer) error
	Deserialize(r io.Reader) error
	interfaces.IGeneralInterface
}

type IStateKeyInterface interface {
	Serialize(w io.Writer) error
	Deserialize(r io.Reader) error
}

var (
	StatesMap = map[DataEntryPrefix]IStateValueInterface{
		ST_Contract:   new(ContractState),
		ST_Account:    new(AccountState),
		ST_AssetState: new(AssetState),
	}
)

func GetStateValue(prefix DataEntryPrefix, data []byte) (IStateValueInterface, error) {
	r := bytes.NewBuffer(data)
	state := StatesMap[prefix]
	err := state.Deserialize(r)
	if err != nil {
		return nil, err
	}
	return state, nil
}
