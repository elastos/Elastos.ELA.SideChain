package enumerators

import (
	"github.com/elastos/Elastos.ELA.SideChain/store"
)

type IteratorValues struct {
	iter store.IIterator
}

func NewIteratorValues(iter store.IIterator) *IteratorValues {
	var iterValues IteratorValues
	iterValues.iter = iter
	return &iterValues
}

func (iter *IteratorValues) Next() bool {
	return iter.iter.Next()
}

func (iter *IteratorValues) Value() []byte {
	return iter.iter.Value()
}

func (iter *IteratorValues) Dispose()  {
	iter.iter.Release()
}

func (iter *IteratorValues) Bytes() []byte {
	return iter.iter.Value()
}