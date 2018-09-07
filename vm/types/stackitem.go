package types

import (
	"math/big"
	"github.com/elastos/Elastos.ELA.SideChain/vm/interfaces"
)

type StackItem interface {
	Equals(other StackItem) bool
	GetBigInteger() *big.Int
	GetBoolean() bool
	GetByteArray() []byte
	GetInterface() interfaces.IGeneralInterface
	GetArray() []StackItem
	GetMap() map[StackItem]StackItem
}
