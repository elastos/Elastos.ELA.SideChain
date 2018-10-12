package blockchain

import "github.com/elastos/Elastos.ELA.Utility/common"

type StorageContext struct {
	codeHash *common.Uint168
	IsReadOnly bool
}

func NewStorageContext(codeHash *common.Uint168) *StorageContext {
	var storageContext StorageContext
	storageContext.codeHash = codeHash
	storageContext.IsReadOnly = false
	return &storageContext
}

func (sc *StorageContext) Bytes() []byte {
	return sc.codeHash.Bytes()
}