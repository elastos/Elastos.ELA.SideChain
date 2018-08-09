package storage

import (
	"math/big"

	"github.com/elastos/Elastos.ELA.SideChain/smartcontract/states"
	. "github.com/elastos/Elastos.ELA.SideChain/common"

	"github.com/elastos/Elastos.ELA.Utility/common"
	)

type DBCache interface {
	GetOrAdd(prefix DataEntryPrefix, key string, value states.IStateValueInterface) (states.IStateValueInterface, error)
	TryGet(prefix DataEntryPrefix, key string) (states.IStateValueInterface, error)
	GetWriteSet() *RWSet
	//GetState(codeHash common.Uint168, loc common.Hash) (common.Hash, error)
	//SetState(codeHash common.Uint160, loc, value common.Hash)
	//GetCode(codeHash common.Uint160) ([]byte, error)
	//SetCode(codeHash common.Uint160, code []byte)
	GetBalance(common.Uint168) *big.Int
	GetCodeSize(common.Uint168) int
	AddBalance(common.Uint168, *big.Int)
	Suicide(codeHash common.Uint168) bool
}


type CloneCache struct {
	innerCache DBCache
	dbCache DBCache
}

func NewCloneDBCache(innerCache DBCache, dbCache DBCache) *CloneCache {
	return &CloneCache{
		innerCache: innerCache,
		dbCache: dbCache,
	}
}

func (cloneCache *CloneCache) GetInnerCache() DBCache {
	return cloneCache.innerCache
}

func (cloneCache *CloneCache) Commit()  {
	for _, v := range cloneCache.innerCache.GetWriteSet().WriteSet {
		if v.IsDeleted {
			cloneCache.innerCache.GetWriteSet().Delete(v.Key)
		} else {
			cloneCache.innerCache.GetWriteSet().Add(v.Prefix, v.Key, v.Item)
		}
	}
}

func (cloneCache *CloneCache) TryGet(prefix DataEntryPrefix, key string) (states.IStateValueInterface, error) {
	if v, ok := cloneCache.innerCache.GetWriteSet().WriteSet[key]; ok {
		return v.Item, nil
	} else {
		return cloneCache.dbCache.TryGet(prefix, key)
	}
}