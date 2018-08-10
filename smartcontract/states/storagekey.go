package states

import (
	"io"
	"errors"

	"github.com/elastos/Elastos.ELA.Utility/common"
)

type StorageKey struct {
	CodeHash *common.Uint168
	Key      []byte
}

func NewStorageKey(codeHash *common.Uint168, key []byte) *StorageKey {
	var storage StorageKey
	storage.CodeHash = codeHash
	storage.Key = key
	return &storage
}

func (storageKey *StorageKey) Serialize(w io.Writer) error {
	err := storageKey.CodeHash.Serialize(w)
	if err != nil {
		return err
	}
	err = common.WriteVarBytes(w, storageKey.Key)
	if err != nil {
		return err
	}
	return nil
}

func (storageKey *StorageKey) Deserialize(r io.Reader) error {
	u := new(common.Uint168)
	err := u.Deserialize(r)
	if err != nil {
		return errors.New("StorageKey CodeHash Deserialize fail.")
	}
	storageKey.CodeHash = u
	key, err := common.ReadVarBytes(r)
	if err != nil {
		return errors.New("StorageKey Key Deserialize fail.")
	}
	storageKey.Key = key
	return nil
}
