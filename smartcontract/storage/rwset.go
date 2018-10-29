package storage

import (
	"errors"
	"bytes"

	"github.com/elastos/Elastos.ELA.SideChain/smartcontract/states"
	. "github.com/elastos/Elastos.ELA.SideChain/store"
)

type RWSet struct {
	ReadSet  map[string]*Read
	WriteSet map[string]*Write
}

type Write struct {
	Prefix    DataEntryPrefix
	Key       string
	Item      states.IStateValueInterface
	IsDeleted bool
}

type Read struct {
	Key     states.IStateKeyInterface
	Version string
}

func NewRWSet() *RWSet {
	var rwSet = new(RWSet)
	rwSet.WriteSet = make(map[string]*Write, 0)
	rwSet.ReadSet = make(map[string]*Read, 0)
	return rwSet
}

func (rw *RWSet) Add(prefix DataEntryPrefix, key string, value states.IStateValueInterface) error {
	data, ok := rw.WriteSet[key]
	if ok && !data.IsDeleted {
		return errors.New("RWSet is allready added:" + key)
	}
	rw.WriteSet[key] = &Write{
		Prefix:    prefix,
		Key:       key,
		Item:      value,
		IsDeleted: false,
	}
	return nil;
}

func (rw *RWSet) Delete(key string) {
	if _, ok := rw.WriteSet[key]; ok {
		rw.WriteSet[key].Item = nil
		rw.WriteSet[key].IsDeleted = true
	} else {
		rw.WriteSet[key] = &Write{
			Key:       key,
			Item:      nil,
			IsDeleted: true,
		}
	}
}

func KeyToStr(key states.IStateKeyInterface) string {
	k := new(bytes.Buffer)
	key.Serialize(k)
	return k.String()
}
