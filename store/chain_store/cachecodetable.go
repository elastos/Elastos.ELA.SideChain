package chain_store

import (
	"github.com/elastos/Elastos.ELA.SideChain/store"
	"github.com/elastos/Elastos.ELA.SideChain/smartcontract/states"
	"github.com/elastos/Elastos.ELA.SideChain/blockchain"
	"github.com/elastos/Elastos.ELA.SideChain/vm/interfaces"
	"github.com/elastos/Elastos.ELA.SideChain/core"
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

func (table *CacheCodeTable) GetTxReference(tx *interfaces.IDataContainer) (map[interfaces.IIntPut]interfaces.IOutput, error) {
	txn := (*tx).(*core.Transaction)
	store := table.dbCache.GetChainStoreDb()
	reference, err := store.GetTxReference(txn)
	if err != nil {
		return nil, err
	}
	var data = map[interfaces.IIntPut]interfaces.IOutput{}
	for input, output := range reference {
		data[input] = output
	}
	return data, nil
}