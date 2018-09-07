package core

import (
	"fmt"
	"io"
	"math/big"

	"github.com/elastos/Elastos.ELA.SPV/spvwallet"
	. "github.com/elastos/Elastos.ELA.Utility/common"
)

type Output struct {
	AssetID     Uint256
	Value       Fixed64
	TokenValue  big.Int
	OutputLock  uint32
	ProgramHash Uint168
}

func (o Output) String() string {
	return "Output: {\n\t\t" +
		"AssetID: " + o.AssetID.String() + "\n\t\t" +
		"Value: " + o.Value.String() + "\n\t\t" +
		"OutputLock: " + fmt.Sprint(o.OutputLock) + "\n\t\t" +
		"ProgramHash: " + o.ProgramHash.String() + "\n\t\t" +
		"}"
}

func (o *Output) Serialize(w io.Writer) error {
	err := o.AssetID.Serialize(w)
	if err != nil {
		return err
	}

	if o.AssetID.IsEqual(spvwallet.SystemAssetId) {
		err = o.Value.Serialize(w)
		if err != nil {
			return err
		}
	} else {
		err = WriteVarBytes(w, o.TokenValue.Bytes())
		if err != nil {
			return err
		}
	}

	WriteUint32(w, o.OutputLock)

	err = o.ProgramHash.Serialize(w)
	if err != nil {
		return err
	}

	return nil
}

func (o *Output) Deserialize(r io.Reader) error {
	err := o.AssetID.Deserialize(r)
	if err != nil {
		return err
	}

	if o.AssetID.IsEqual(spvwallet.SystemAssetId) {
		err = o.Value.Deserialize(r)
		if err != nil {
			return err
		}
	} else {
		bytes, err := ReadVarBytes(r)
		if err != nil {
			return err
		}
		o.TokenValue.SetBytes(bytes)
	}

	temp, err := ReadUint32(r)
	if err != nil {
		return err
	}
	o.OutputLock = uint32(temp)

	err = o.ProgramHash.Deserialize(r)
	if err != nil {
		return err
	}

	return nil
}
