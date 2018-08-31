package blockchain

import "github.com/elastos/Elastos.ELA.Utility/common"

type StorageContext struct {
	codeHash *common.Uint168
}

func NewStorageContext(codeHash *common.Uint168) *StorageContext {
	var storageContext StorageContext
	storageContext.codeHash = codeHash
	return &storageContext
}

func (sc *StorageContext) Bytes() []byte {
	return sc.codeHash.Bytes()
}