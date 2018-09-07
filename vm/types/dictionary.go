package types

import (
	"github.com/elastos/Elastos.ELA.SideChain/vm/interfaces"
	"math/big"
)

type Dictionary struct {
	dic map[StackItem]StackItem
}

func NewDictionary() *Dictionary {
	var dictionary Dictionary
	dictionary.dic = make(map[StackItem]StackItem)
	return &dictionary;
}


func (dic *Dictionary) GetValue(key StackItem) StackItem {
	for temp := range dic.dic {
		if temp.Equals(key) {
			return dic.dic[temp]
		}
	}
	return nil
}

func (dic *Dictionary) Equals(other StackItem) bool {
	otherMap := other.GetMap()
	for key := range dic.dic {
		if dic.dic[key].Equals(otherMap[key]) == false {
			return false
		}
	}
	return true
}

func (dic *Dictionary) GetMap() map[StackItem]StackItem {
	return dic.dic
}

func (dic *Dictionary) PutStackItem(key, value StackItem)  {
	dic.dic[key] = value
}

func (dic *Dictionary) GetBoolean() bool {
	return false
}

func (dic *Dictionary) GetByteArray() []byte {
	return []byte{}
}

func (dic *Dictionary) GetInterface() interfaces.IGeneralInterface {
	return nil
}

func (dic *Dictionary) GetArray() []StackItem {
	return nil
}

func (dic *Dictionary) GetBigInteger() *big.Int {
	return big.NewInt(0)
}