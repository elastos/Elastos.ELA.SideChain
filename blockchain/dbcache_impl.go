package blockchain

import (
	"bytes"
	"math/big"

	"github.com/elastos/Elastos.ELA.SideChain/smartcontract/storage"
	"github.com/elastos/Elastos.ELA.SideChain/smartcontract/states"
	. "github.com/elastos/Elastos.ELA.SideChain/store"

	"github.com/elastos/Elastos.ELA.Utility/common"
)

type DBCache struct {
	RWSet *storage.RWSet
	db    IChainStore
}

func NewDBCache(db IChainStore) *DBCache {
	return &DBCache{
		RWSet: storage.NewRWSet(),
		db:    db,
	}
}

func (cache *DBCache) Commit() {
	rwSet := cache.RWSet.WriteSet
	for k, v := range rwSet {
		key := make([]byte, 0)
		key = append([]byte{byte(v.Prefix)}, []byte(k)...)
		if v.IsDeleted {
			cache.db.BatchDelete(key)
		} else {
			b := new(bytes.Buffer)
			v.Item.Serialize(b)
			value := make([]byte, 0)
			value = append(value, b.Bytes()...)
			cache.db.BatchPut(key, value)
		}
	}
}

func (cache *DBCache) TryGetInternal(prefix DataEntryPrefix, key string) (states.IStateValueInterface, error) {
	k := make([]byte, 0)
	k = append([]byte{byte(prefix)}, []byte(key)...)
	value, err := cache.db.Get(k)
	if err != nil {
		return nil, err
	}
	return states.GetStateValue(prefix, value)
}

func (cache *DBCache) TryDelete(prefix DataEntryPrefix, key string) bool {
	k := make([]byte, 0)
	k = append([]byte{byte(prefix)}, []byte(key)...)
	err := cache.db.Delete(k)
	return err == nil
}

func (cache *DBCache) GetOrAdd(prefix DataEntryPrefix, key string, value states.IStateValueInterface) (states.IStateValueInterface, error) {
	if v, ok := cache.RWSet.WriteSet[key]; ok {
		if v.IsDeleted {
			v.Item = value
			v.IsDeleted = false
		}
	} else {
		item, err := cache.TryGetInternal(prefix, key)
		if err != nil && err.Error() != ("leveldb: not found") {
			return nil, err
		}
		write := &storage.Write{
			Prefix:    prefix,
			Key:       key,
			Item:      item,
			IsDeleted: false,
		}
		if write.Item == nil {
			write.Item = value
		}
		cache.RWSet.WriteSet[key] = write
	}
	return cache.RWSet.WriteSet[key].Item, nil
}

func (cache *DBCache) TryGet(prefix DataEntryPrefix, key string) (states.IStateValueInterface, error) {
	if v, ok := cache.RWSet.WriteSet[key]; ok {
		return v.Item, nil
	} else {
		return cache.TryGetInternal(prefix, key)
	}
}

func (cache *DBCache) GetWriteSet() *storage.RWSet {
	return cache.RWSet
}

func (cache *DBCache) GetBalance(hash common.Uint168) *big.Int {
	return big.NewInt(100)
}

func (cache *DBCache) GetCodeSize(hash common.Uint168) int {
	return 0
}

func (cache *DBCache) AddBalance(hash common.Uint168, int2 *big.Int) {

}

func (cache *DBCache) GetChainStoreDb() IChainStore {
	return cache.db
}

func (cache *DBCache) Suicide(codeHash common.Uint168) bool {
	skey := storage.KeyToStr(&codeHash)
	cache.RWSet.Delete(skey)
	return true;
}
