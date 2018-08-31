package chain_store

import (
	"github.com/elastos/Elastos.ELA.SideChain/store"
	"github.com/elastos/Elastos.ELA.SideChain/smartcontract/states"
	"github.com/elastos/Elastos.ELA.SideChain/blockchain"
)

type CacheCodeTable struct {
	dbCache *blockchain.DBCache
}

func NewCacheCodeTable(dbCache *blockchain.DBCache) *CacheCodeTable {
	return &CacheCodeTable{dbCache: dbCache}
}

func (table *CacheCodeTable) GetScript(codeHash []byte) ([]byte) {
	value, err := table.dbCache.TryGet(store.ST_Contract, string(codeHash))
	if err != nil {
		return nil
	}
	return value.(*states.ContractState).Code.Code
}
