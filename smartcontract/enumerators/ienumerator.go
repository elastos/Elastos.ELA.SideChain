package enumerators

import "github.com/elastos/Elastos.ELA.SideChain/vm/types"

type IEnumerator interface {
	Next() bool

	Value() types.StackItem

	Dispose()
}