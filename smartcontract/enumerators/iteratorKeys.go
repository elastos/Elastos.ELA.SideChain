package enumerators

import (
	"github.com/elastos/Elastos.ELA.SideChain/store"
)

type IteratorKeys struct {
	iter store.IIterator
}

func NewIteratorKeys(iter store.IIterator) *IteratorKeys {
	var iterKeys IteratorKeys
	iterKeys.iter = iter
	return &iterKeys
}

func (iter *IteratorKeys) Next() bool {
	return iter.iter.Next()
}

func (iter *IteratorKeys) Value() []byte {
	return iter.iter.Key()
}

func (iter *IteratorKeys) Dispose()  {
	iter.iter.Release()
}

func (iter *IteratorKeys) Bytes() []byte {
	return iter.iter.Key()
}