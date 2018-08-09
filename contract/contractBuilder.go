package contract

import (
	"github.com/elastos/Elastos.ELA.Utility/crypto"
	"errors"
	"github.com/elastos/Elastos.ELA.SideChain/vm"
	"github.com/elastos/Elastos.ELA.SideChain/common"
)

//create a Single Singature contract for owner
func CreateSignatureContract(ownerPubKey *crypto.PublicKey) (*Contract, error) {
	msg := "[Contract],CreateSignatureContract failed."
	temp, err := ownerPubKey.EncodePoint(true)
	if err != nil {
		return nil, errors.New(msg)
	}
	signatureReedScript, err := CreateSignatureRedeemScript(ownerPubKey)
	if err != nil {
		return nil, errors.New(msg)
	}
	hash, err := common.ToCodeHash(temp)
	if err != nil {
		return nil, errors.New(msg)
	}
	signatureReedScriptToCodeHash, err := common.ToCodeHash(signatureReedScript)
	if err != nil {
		return nil, errors.New(msg)
	}
	return &Contract{
		Code: signatureReedScript,
		Parameters: []ContractParameterType{Signature},
		ProgramHash: signatureReedScriptToCodeHash,
		OwnerPubkeyHash: hash,
	}, nil
}

func CreateSignatureRedeemScript(pubkey *crypto.PublicKey) ([]byte, error) {
	temp, err := pubkey.EncodePoint(true)
	if err != nil {
		return nil, errors.New("[Contract],CreateSignatureRedeemScript failed.")
	}
	sb := NewProgramBuilder()
	sb.PushData(temp)
	sb.AddOp(vm.CHECKSIG)
	return sb.ToArray(), nil
}

