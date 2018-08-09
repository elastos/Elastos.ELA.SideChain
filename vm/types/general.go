package types

import (
	"math/big"
	"github.com/elastos/Elastos.ELA.SideChain/vm/interfaces"
)

type GeneralInterface struct {
	object interfaces.IGeneralInterface
}

func NewGeneralInterface(value interfaces.IGeneralInterface) *GeneralInterface {
	var ii GeneralInterface
	ii.object = value
	return &ii
}

func (ii *GeneralInterface) Equals(other GeneralInterface) bool {
	return ii.object == other.object
}

func (ii *GeneralInterface) GetBigInteger() big.Int {
	return big.Int{}
}

func (ii *GeneralInterface) GetBoolean() bool {
	if ii.object == nil {
		return false
	}
	return true
}

func (ii *GeneralInterface) GetByteArray() []byte {
	return ii.object.Bytes()
}

func (ii *GeneralInterface) GetInterface() interfaces.IGeneralInterface {
	return ii.object
}

func (ii *GeneralInterface) GetArray() []interfaces.IGeneralInterface {
	return []interfaces.IGeneralInterface{}
}
