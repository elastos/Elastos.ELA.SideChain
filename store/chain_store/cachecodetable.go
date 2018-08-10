package chain_store

import (
	"github.com/elastos/Elastos.ELA.SideChain/store"
	"github.com/elastos/Elastos.ELA.SideChain/smartcontract/states"
)

type CacheCodeTable struct {
	dbCache *DBCache
}

func NewCacheCodeTable(dbCache *DBCache) *CacheCodeTable {
	return &CacheCodeTable{dbCache: dbCache}
}

func (table *CacheCodeTable) GetScript(codeHash []byte) ([]byte) {
	value, err := table.dbCache.TryGet(store.ST_Contract, string(codeHash))
	if err != nil {
		return nil
	}
	return value.(*states.ContractState).Code.Code
}
