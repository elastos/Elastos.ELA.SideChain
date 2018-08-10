package service

import (
	"errors"
	"math/big"

	"github.com/elastos/Elastos.ELA.Utility/common"
	"github.com/elastos/Elastos.ELA.Utility/crypto"

	"github.com/elastos/Elastos.ELA.SideChain/vm"
	"github.com/elastos/Elastos.ELA.SideChain/contract"
	common2 "github.com/elastos/Elastos.ELA.SideChain/common"
	"github.com/elastos/Elastos.ELA.SideChain/blockchain"
	"github.com/elastos/Elastos.ELA.SideChain/core"
	"github.com/elastos/Elastos.ELA.SideChain/vm/types"
	"github.com/elastos/Elastos.ELA.SideChain/smartcontract/states"
)

type StateReader struct {
	serviceMap map[string]func(engine *vm.ExecutionEngine) (bool, error)
}

func NewStateReader() *StateReader {
	var stateReader StateReader
	stateReader.serviceMap = make(map[string]func(*vm.ExecutionEngine) (bool, error), 0)
	return &stateReader
}

func (s *StateReader) Register(methodName string, handler func(engine *vm.ExecutionEngine) (bool, error)) bool {
	if _, ok := s.serviceMap[methodName]; ok {
		return false
	}
	s.serviceMap[methodName] = handler
	return true;
}

func (s *StateReader) GetServiceMap() map[string]func(*vm.ExecutionEngine) (bool, error) {
	return s.serviceMap
}

func (s *StateReader) RuntimeGetTrigger(e *vm.ExecutionEngine) (bool, error) {
	return true, nil
}

func (s *StateReader) RuntimeNotify(e *vm.ExecutionEngine) (bool, error) {
	vm.PopStackItem(e)
	return true, nil
}

func (s *StateReader) RuntimeLog(e *vm.ExecutionEngine) (bool, error) {
	return true, nil
}

func (s *StateReader) CheckWitnessHash(engine *vm.ExecutionEngine, programHash common.Uint168) (bool, error) {
	tx := engine.GetDataContainer().(*core.Transaction)
	hashForVerify, err := blockchain.GetTxProgramHashes(tx)
	if err != nil {
		return false, err
	}
	return contains(hashForVerify, programHash), nil
}

func (s *StateReader) CheckWitnessPublicKey(engine *vm.ExecutionEngine, publicKey *crypto.PublicKey) (bool, error) {
	c, err := contract.CreateSignatureRedeemScript(publicKey)
	if err != nil {
		return false, err
	}
	h, err := common2.ToCodeHash(c)
	if err != nil {
		return false, err
	}
	return s.CheckWitnessHash(engine, h)
}

func (s *StateReader) RuntimeCheckWitness(e *vm.ExecutionEngine) (bool, error) {
	data := vm.PopByteArray(e)
	var (
		result bool
		err    error
	)
	if len(data) == 21 {
		program, err := common.Uint168FromBytes(data)
		if err != nil {
			return false, err
		}
		result, err = s.CheckWitnessHash(e, *program)
	} else if len(data) == 33 {
		publickKey, err := crypto.DecodePoint(data)
		if err != nil {
			return false, err
		}
		result, err = s.CheckWitnessPublicKey(e, publickKey)
	} else {
		return false, errors.New("[RuntimeCheckWitness] data invalid.")
	}
	if err != nil {
		return false, err
	}
	vm.PushData(e, result)
}

func (s *StateReader) BlockChainGetHeight(e *vm.ExecutionEngine) (bool, error) {
	var i uint32 = 0
	if blockchain.DefaultLedger != nil {
		i = blockchain.DefaultLedger.Store.GetHeight();
	}
	vm.PushData(e, i)
	return true, nil
}

func (s *StateReader) BlockChainGetHeader(e *vm.ExecutionEngine) (bool, error) {
	var (
		header *core.Header
		err    error
	)
	data := vm.PopByteArray(e)
	l := len(data)

	if l <= 5 {
		b := new(big.Int)
		height := b.SetBytes(common.BytesReverse(data)).Int64()
		if blockchain.DefaultLedger != nil {
			hash, err := blockchain.DefaultLedger.Store.GetBlockHash((uint32(height)))
			if err != nil {
				return false, err
			}
			header, err = blockchain.DefaultLedger.Store.GetHeader(hash)
		}
	} else if l == 32 {
		hash, _ := common.Uint256FromBytes(data)
		if blockchain.DefaultLedger != nil {
			header, err = blockchain.DefaultLedger.Store.GetHeader(*hash)
		}
	} else {
		return false, errors.New("[BlockChainGetHeader] data invalid.")
	}
	if err != nil {
		return false, err
	}
	vm.PushData(e, header)
	return true, nil
}

func (s *StateReader) BlockChainGetBlock(e *vm.ExecutionEngine) (bool, error) {
	data := vm.PopByteArray(e)
	var (
		block *core.Block
		err   error
	)
	l := len(data)
	if l <= 5 {
		b := new(big.Int)
		height := uint32(b.SetBytes(common.BytesReverse(data)).Int64())
		if blockchain.DefaultLedger != nil {
			hash, err := blockchain.DefaultLedger.Store.GetBlockHash(height)
			if err != nil {
				return false, err
			}
			block, err = blockchain.DefaultLedger.Store.GetBlock(hash)
		}
	} else if l == 32 {
		hash, err := common.Uint256FromBytes(data)
		if err != nil {
			return false, err
		}
		if blockchain.DefaultLedger != nil {
			block, err = blockchain.DefaultLedger.Store.GetBlock(*hash)
		}
	} else {
		return false, errors.New("[BlockChainGetBlock] data invalid.")
	}
	if err != nil {
		return false, err
	}
	vm.PushData(e, block)
	return true, nil
}

func (s *StateReader) BlockChainGetTransaction(e *vm.ExecutionEngine) (bool, error) {
	d := vm.PopByteArray(e)
	hash, err := common.Uint256FromBytes(d)
	if err != nil {
		return false, err
	}
	tx, _, err := blockchain.DefaultLedger.Store.GetTransaction(*hash)
	if err != nil {
		return false, err
	}
	vm.PushData(e, tx)
	return true, nil
}

func (s *StateReader) BlockChainGetAsset(e *vm.ExecutionEngine) (bool, error) {
	d := vm.PopByteArray(e)
	hash, err := common.Uint256FromBytes(d)
	if err != nil {
		return false, err
	}
	asset, err := blockchain.DefaultLedger.Store.GetAsset(*hash)
	if err != nil {
		return false, err
	}
	vm.PushData(e, asset)
	return true, nil
}

func (s *StateReader) HeaderGetHash(e *vm.ExecutionEngine) (bool, error) {
	d := vm.PopInteropInterface(e)
	if d == nil {
		return false, errors.New("Get header error in function headergethash!")
	}
	hash := d.(*core.Header).Hash()
	vm.PushData(e, hash.Bytes())
	return true, nil
}

func (s *StateReader) HeaderGetVersion(e *vm.ExecutionEngine) (bool, error) {
	d := vm.PopInteropInterface(e)
	if d == nil {
		return false, errors.New("Get header error in function headergetversion")
	}
	version := d.(*core.Header).Version
	vm.PushData(e, version)
	return true, nil
}

func (s *StateReader) HeaderGetPrevHash(e *vm.ExecutionEngine) (bool, error) {
	d := vm.PopInteropInterface(e)
	if d == nil {
		return false, errors.New("Get header error in function HeaderGetPrevHash")
	}
	preHash := d.(*core.Header).Previous
	vm.PushData(e, preHash.Bytes())
	return true, nil
}

func (s *StateReader) HeaderGetMerkleRoot(e *vm.ExecutionEngine) (bool, error) {
	d := vm.PopInteropInterface(e)
	if d == nil {
		return false, errors.New("Get header error in function HeaderGetMerkleRoot")
	}
	root := d.(*core.Header).MerkleRoot
	vm.PushData(e, root.Bytes())
	return true, nil
}

func (s *StateReader) HeaderGetConsensusData(e *vm.ExecutionEngine) (bool, error) {
	d := vm.PopInteropInterface(e)
	if d == nil {
		return false, errors.New("Get header error in function HeaderGetConsensusData")
	}
	consensusData := d.(*core.Header).Nonce
	vm.PushData(e, consensusData)
	return true, nil
}

func (s *StateReader) BlockgetTransactionCount(e *vm.ExecutionEngine) (bool, error) {
	d := vm.PopInteropInterface(e)
	if d == nil {
		return false, errors.New("Get block error in function BlockGetTransactionCount")
	}
	transactions := d.(*core.Block).Transactions
	vm.PushData(e, len(transactions))
	return true, nil
}

func (s *StateReader) BlockGetTransactions(e *vm.ExecutionEngine) (bool, error) {
	d := vm.PopInteropInterface(e)
	if d == nil {
		return false, errors.New("Get block data error in function BlockGetTransactions")
	}
	transactions := d.(*core.Block).Transactions
	list := make([]types.GeneralInterface, 0)
	for _, v := range transactions {
		list = append(list, *types.NewGeneralInterface(v))
	}
	vm.PushData(e, list)
	return true, nil
}

func (s *StateReader) BlockGetTransaction(e *vm.ExecutionEngine) (bool, error) {
	index := vm.PopInt(e)
	if index < 0 {
		return false, errors.New("index invalid in function BlockGetTransaction")
	}
	d := vm.PopInteropInterface(e)
	if d == nil {
		return false, errors.New("Get transaction data error in function BlockGetTransaction")
	}
	transactions := d.(*core.Block).Transactions
	if index >= len(transactions) {
		return false, errors.New("index over transaction length in function BlockGetTransaction")
	}
	vm.PushData(e, transactions[index])
	return false, nil
}

func (s *StateReader) TransactionGetHash(e *vm.ExecutionEngine) (bool, error) {
	d := vm.PopInteropInterface(e)
	if d == nil {
		return false, errors.New("Get transaction error in function TransactionGetHash")
	}
	txHash := d.(*core.Transaction).Hash()
	vm.PushData(e, txHash.Bytes())
	return true, nil
}

func (s *StateReader) TransactionGetType(e *vm.ExecutionEngine) (bool, error) {
	d := vm.PopInteropInterface(e)
	if d == nil {
		return false, errors.New("Get transaction error in function TransactionGetType")
	}
	txType := d.(*core.Transaction).TxType
	vm.PushData(e, int(txType))
	return true, nil
}

func (s *StateReader) TransactionGetAttributes(e *vm.ExecutionEngine) (bool, error) {
	d := vm.PopInteropInterface(e)
	if d == nil {
		return false, errors.New("Get transaction error in function TransactionGetAttributes")
	}
	attributes := d.(*core.Transaction).Attributes
	list := make([]types.GeneralInterface, 0)
	for _, v := range attributes {
		list = append(list, *types.NewGeneralInterface(v))
	}
	vm.PushData(e, list)
	return true, nil
}

func (s *StateReader) TransactionGetInputs(e *vm.ExecutionEngine) (bool, error) {
	d := vm.PopInteropInterface(e)
	if d == nil {
		return false, errors.New("Get transaction error in function TransactionGetInputs")
	}
	inputs := d.(*core.Transaction).Inputs
	list := make([]types.GeneralInterface, 0)
	for _, v := range inputs {
		list = append(list, *types.NewGeneralInterface(v))
	}
	vm.PushData(e, list)
	return true, nil
}

func (s *StateReader) TransactionGetOutputs(e *vm.ExecutionEngine) (bool, error) {
	d := vm.PopInteropInterface(e)
	if d == nil {
		return false, errors.New("Get transaction error in function TransactionGetOutputs")
	}
	outputs := d.(*core.Transaction).Outputs
	list := make([]types.GeneralInterface, 0)
	for _, v := range outputs {
		list = append(list, *types.NewGeneralInterface(v))
	}
	vm.PushData(e, list)
	return true, nil
}

func (s *StateReader) TransactionGetReferences(e *vm.ExecutionEngine) (bool, error) {
	d := vm.PopInteropInterface(e)
	if d == nil {
		return false, errors.New("Get transaction error in function TransactionGetReferences")
	}

	references, err := blockchain.DefaultLedger.Store.GetTxReference(d.(*core.Transaction))
	if err != nil {
		return false, err
	}
	list := make([]types.GeneralInterface, 0)
	for _, v := range references {
		list = append(list, *types.NewGeneralInterface(v))
	}
	vm.PushData(e, list)
	return true, nil
}

func (s *StateReader) AttributeGetUsage(e *vm.ExecutionEngine) (bool, error) {
	d := vm.PopInteropInterface(e)
	if d == nil {
		return false, errors.New("Get Attribute error in function AttributeGetUsage")
	}
	attribute := d.(*core.Attribute)
	vm.PushData(e, int(attribute.Usage))
	return true, nil
}

func (s *StateReader) AttributeGetData(e *vm.ExecutionEngine) (bool, error) {
	d := vm.PopInteropInterface(e)
	if d == nil {
		return false, errors.New("Get Attribute error in function AttributeGetUsage")
	}
	attribute := d.(*core.Attribute)
	vm.PushData(e, attribute.Data)
	return true, nil
}

func (s *StateReader) InputGetHash(e *vm.ExecutionEngine) (bool, error) {
	d := vm.PopInteropInterface(e)
	if d == nil {
		return false, errors.New("Get UTXOTxInput error in function InputGetHash")
	}
	input := d.(*core.Input)
	vm.PushData(e, input.Previous.TxID.Bytes())
	return true, nil
}

func (s *StateReader) InputGetIndex(e *vm.ExecutionEngine) (bool, error) {
	d := vm.PopInteropInterface(e)
	if d == nil {
		return false, errors.New("Get transaction error in function TransactionGetReferences")
	}
	input := d.(*core.Input)
	vm.PushData(e, input.Previous.Index)
	return true, nil
}

func (s *StateReader) OutputGetAssetId(e *vm.ExecutionEngine) (bool, error) {
	d := vm.PopInteropInterface(e)
	if d == nil {
		return false, errors.New("Get TxOutput error in function OutputGetAssetId")
	}
	output := d.(*core.Output)
	vm.PushData(e, output.AssetID.Bytes())
	return true, nil
}

func (s *StateReader) OutputGetValue(e *vm.ExecutionEngine) (bool, error) {
	d := vm.PopInteropInterface(e)
	if d == nil {
		return false, errors.New("Get TxOutput error in function OutputGetValue")
	}
	output := d.(*core.Output)
	bytes, _ := output.Value.Bytes()
	vm.PushData(e, bytes)
	return true, nil
}

func (s *StateReader) AssetGetAssetId(e *vm.ExecutionEngine) (bool, error) {
	d := vm.PopInteropInterface(e)
	if d == nil {
		return false, errors.New("Get AssetState error in function AssetGetAssetId")
	}
	assetState := d.(*states.AssetState)
	vm.PushData(e, assetState.AssetId.Bytes())
	return true, nil
}

func (s *StateReader) AssetGetAssetType(e *vm.ExecutionEngine) (bool, error) {
	d := vm.PopInteropInterface(e)
	if d == nil {
		return false, errors.New("Get AssetState error in function AssetGetAssetType")
	}
	assetState := d.(*states.AssetState)
	vm.PushData(e, int(assetState.AssetType))
	return true, nil
}

func (s *StateReader) AssetGetAmount(e *vm.ExecutionEngine) (bool, error) {
	d := vm.PopInteropInterface(e)
	if d == nil {
		return false, errors.New("Get AssetState error in function AssetGetAmount")
	}
	assetState := d.(*states.AssetState)
	vm.PushData(e, assetState.Amount.IntValue())
	return true, nil
}

func (s *StateReader) AssetGetAvailable(e *vm.ExecutionEngine) (bool, error) {
	d := vm.PopInteropInterface(e)
	if d == nil {
		return false, errors.New("Get AssetState error in function AssetGetAvailable")
	}
	assetState := d.(*states.AssetState)
	vm.PushData(e, assetState.Avaliable.IntValue())
	return true, nil
}

func (s *StateReader) AssetGetPrecidsion(e *vm.ExecutionEngine) (bool, error) {
	d := vm.PopInteropInterface(e)
	if d == nil {
		return false, errors.New("Get AssetState error in function AssetGetPrecision")
	}
	assetState := d.(*states.AssetState)
	vm.PushData(e, int(assetState.Precision))
	return true, nil
}

func (s *StateReader) AssetGetOwner(e *vm.ExecutionEngine) (bool, error) {
	d := vm.PopInteropInterface(e)
	if d == nil {
		return false, errors.New("Get AssetState error in function AssetGetOwner")
	}
	assetState := d.(*states.AssetState)
	owner, err := assetState.Owner.EncodePoint(true)
	if err != nil {
		return false, err
	}
	vm.PushData(e, owner)
	return true, nil
}

func (s *StateReader) AssetGetAdmin(e *vm.ExecutionEngine) (bool, error) {
	d := vm.PopInteropInterface(e)
	if d == nil {
		return false, errors.New("Get AssetState error in function AssetGetAdmin")
	}
	assetState := d.(*states.AssetState)
	vm.PushData(e, assetState.Admin.Bytes())
	return true, nil
}

func (s *StateReader) AssetGetIssuer(e *vm.ExecutionEngine) (bool, error) {
	d := vm.PopInteropInterface(e)
	if d == nil {
		return false, errors.New("Get AssetState error in function AssetGetIssuer")
	}
	assetState := d.(*states.AssetState)
	vm.PushData(e, assetState.Issuer.Bytes())
	return true, nil
}

func (s *StateReader) ContractGetCode(e *vm.ExecutionEngine) (bool, error) {
	d := vm.PopInteropInterface(e)
	if d == nil {
		return false, errors.New("Get ContractState error in function ContractGetCode")
	}
	assetState := d.(*states.ContractState)
	vm.Push(e, assetState.Code.Code)
	return true, nil
}

func (s *StateReader) StorageGetContext(e *vm.ExecutionEngine) (bool, error) {
	codeHash, err := common.Uint168FromBytes(e.Hash160(e.ExecutingScript()))
	if err != nil {
		return false, err
	}
	vm.PushData(e, NewStorageContext(codeHash))
	return true, nil
}
