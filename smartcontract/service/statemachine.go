package service

import (
	"github.com/elastos/Elastos.ELA.Utility/common"
	"github.com/elastos/Elastos.ELA.SideChain/smartcontract/storage"
)

type StateMachine struct {
	*StateReader
	CloneCache *storage.CloneCache
}

func NewStateMachine(dbCache storage.DBCache, innerCache storage.DBCache) *StateMachine {
	var stateMachine StateMachine
	stateMachine.CloneCache = storage.NewCloneDBCache(innerCache, dbCache)
	stateMachine.StateReader = NewStateReader()
	//todo add Register method

	return &stateMachine
}

func contains(programHashes []common.Uint168, programHash common.Uint168) bool {
	for _, v := range programHashes {
		if v == programHash {
			return true
		}
	}
	return false
}
