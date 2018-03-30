package payload

import (
	"errors"
	"io"

	"github.com/elastos/Elastos.ELA.SideChain/common/serialization"
)

const RecordPayloadVersion byte = 0x00

type Record struct {
	RecordType string
	RecordData []byte
}

func (a *Record) Data(version byte) []byte {
	//TODO: implement RegisterRecord.Data()
	return []byte{0}
}

// Serialize is the implement of SignableData interface.
func (a *Record) Serialize(w io.Writer, version byte) error {
	err := serialization.WriteVarString(w, a.RecordType)
	if err != nil {
		return errors.New("[RecordDetail], RecordType serialize failed.")
	}
	err = serialization.WriteVarBytes(w, a.RecordData)
	if err != nil {
		return errors.New("[RecordDetail], RecordData serialize failed.")
	}
	return nil
}

// Deserialize is the implement of SignableData interface.
func (a *Record) Deserialize(r io.Reader, version byte) error {
	var err error
	a.RecordType, err = serialization.ReadVarString(r)
	if err != nil {
		return errors.New("[RecordDetail], RecordType deserialize failed.")
	}
	a.RecordData, err = serialization.ReadVarBytes(r)
	if err != nil {
		return errors.New("[RecordDetail], RecordData deserialize failed.")
	}
	return nil
}
