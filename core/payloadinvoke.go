package core

import (
	"io"

	"github.com/elastos/Elastos.ELA.Utility/common"
)

type PayloadInvoke struct {
	CodeHash    common.Uint168
	Code        []byte
	ProgramHash common.Uint168
}

func (ic *PayloadInvoke) Data(version byte) []byte{
	return []byte{0}
}

func (ic *PayloadInvoke) Serialize(w io.Writer, version byte) error {
	err := ic.CodeHash.Serialize(w)
	if err != nil {
		return err;
	}
	err = common.WriteVarBytes(w, ic.Code)
	if err != nil {
		return err
	}
	err = ic.ProgramHash.Serialize(w)
	if err != nil {
		return err
	}
	return nil
}

func (ic *PayloadInvoke) Deserialize(r io.Reader, version byte) error {
	codeHash := common.Uint168{}
	err := codeHash.Deserialize(r)
	if err != nil {
		return err
	}
	ic.CodeHash = codeHash

	ic.Code, err = common.ReadVarBytes(r)
	if err != nil {
		return err
	}

	programHash := common.Uint168{}
	err = programHash.Deserialize(r)
	if err != nil {
		return err
	}
	ic.ProgramHash = programHash;

	return nil
}