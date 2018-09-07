package contract

import (
	"io"
	"fmt"

	"github.com/elastos/Elastos.ELA.Utility/crypto"
	"github.com/elastos/Elastos.ELA.Utility/common"

	"github.com/elastos/Elastos.ELA.SideChain/log"
)

type FunctionCode struct {
	// Contract Code
	Code []byte

	// Contract parameter type list
	ParameterTypes []ContractParameterType

	// Contract return type list
	ReturnType ContractParameterType

	codeHash common.Uint168
}


// method of SerializableData
func (fc *FunctionCode) Serialize(w io.Writer) error {
	err := common.WriteUint8(w, uint8(fc.ReturnType))
	if err != nil {
		return err
	}
	err = common.WriteVarBytes(w, ContractParameterTypeToByte(fc.ParameterTypes))
	if err != nil {
		return err
	}
	err = common.WriteVarBytes(w,fc.Code)
	if err != nil {
		return err
	}
	return nil
}

// method of SerializableData
func (fc *FunctionCode) Deserialize(r io.Reader) error {
	returnType, err := common.ReadUint8(r)
	if err != nil {
		return err
	}
	fc.ReturnType = ContractParameterType(returnType)

	parameterTypes, err := common.ReadVarBytes(r)
	if err != nil {
		return err
	}
	fc.ParameterTypes = ByteToContractParameterType(parameterTypes)

	fc.Code,err = common.ReadVarBytes(r)
	if err != nil {
		return err
	}

	return nil
}

// Get code
func (fc *FunctionCode) GetCode() []byte {
	return fc.Code
}

// Get the list of parameter value
func (fc *FunctionCode) GetParameterTypes() []ContractParameterType {
	return fc.ParameterTypes
}

// Get the list of return value
func (fc *FunctionCode) GetReturnType() ContractParameterType {
	return fc.ReturnType
}

// Get the hash of the smart contract
func (fc *FunctionCode) CodeHash() common.Uint168 {
	zeroHash := common.Uint168{}
	if fc.codeHash == zeroHash {
		hash, err := crypto.ToProgramHash(fc.Code)
		if err != nil {
			log.Debug( fmt.Sprintf("[FunctionCode] ToCodeHash err=%s",err) )
			return *hash
		}
		fc.codeHash = *hash
	}
	return fc.codeHash
}