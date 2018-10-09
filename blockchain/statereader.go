package blockchain

import (
	"math/big"
	"errors"

	"github.com/elastos/Elastos.ELA.Utility/common"
	"github.com/elastos/Elastos.ELA.Utility/crypto"
	"github.com/elastos/Elastos.ELA/blockchain"

	"github.com/elastos/Elastos.ELA.SideChain/vm"
	"github.com/elastos/Elastos.ELA.SideChain/contract"
	"github.com/elastos/Elastos.ELA.SideChain/core"
	"github.com/elastos/Elastos.ELA.SideChain/vm/types"
	"github.com/elastos/Elastos.ELA.SideChain/smartcontract/states"
	"github.com/elastos/Elastos.ELA.SideChain/log"
)

type StateReader struct {
	serviceMap map[string]func(engine *vm.ExecutionEngine) bool
}

func NewStateReader() *StateReader {
	var stateReader StateReader
	stateReader.serviceMap = make(map[string]func(*vm.ExecutionEngine) bool, 0)

	stateReader.Register("Neo.Runtime.GetTrigger", stateReader.RuntimeGetTrigger)
	stateReader.Register("Neo.Runtime.CheckWitness", stateReader.RuntimeCheckWitness)
	stateReader.Register("Neo.Runtime.Notify", stateReader.RuntimeNotify)
	stateReader.Register("Neo.Runtime.Log", stateReader.RuntimeLog)
	stateReader.Register("Neo.Runtime.GetTime", stateReader.RuntimeGetTime)

	stateReader.Register("Neo.Blockchain.GetHeight", stateReader.BlockChainGetHeight)
	stateReader.Register("Neo.Blockchain.GetHeader", stateReader.BlockChainGetHeader)
	stateReader.Register("Neo.Blockchain.GetBlock", stateReader.BlockChainGetBlock)
	stateReader.Register("Neo.Blockchain.GetTransaction", stateReader.BlockChainGetTransaction)
	stateReader.Register("Neo.Blockchain.GetAccount", stateReader.BlockChainGetAccount)
	stateReader.Register("Neo.Blockchain.GetAsset", stateReader.BlockChainGetAsset)

	stateReader.Register("Neo.Header.GetHash", stateReader.HeaderGetHash)
	stateReader.Register("Neo.Header.GetVersion", stateReader.HeaderGetVersion)
	stateReader.Register("Neo.Header.GetPrevHash", stateReader.HeaderGetPrevHash)
	stateReader.Register("Neo.Header.GetMerkleRoot", stateReader.HeaderGetMerkleRoot)
	stateReader.Register("Neo.Header.GetTimestamp", stateReader.HeaderGetTimestamp)
	stateReader.Register("Neo.Header.GetConsensusData", stateReader.HeaderGetConsensusData)

	stateReader.Register("Neo.Block.GetTransactionCount", stateReader.BlockgetTransactionCount)
	stateReader.Register("Neo.Block.GetTransactions", stateReader.BlockGetTransactions)
	stateReader.Register("Neo.Block.GetTransaction", stateReader.BlockGetTransaction)

	stateReader.Register("Neo.Transaction.GetHash", stateReader.TransactionGetHash)
	stateReader.Register("Neo.Transaction.GetType", stateReader.TransactionGetType)
	stateReader.Register("Neo.Transaction.GetAttributes", stateReader.TransactionGetAttributes)
	stateReader.Register("Neo.Transaction.GetInputs", stateReader.TransactionGetInputs)
	stateReader.Register("Neo.Transaction.GetOutputs", stateReader.TransactionGetOutputs)
	stateReader.Register("Neo.Transaction.GetReferences", stateReader.TransactionGetReferences)

	stateReader.Register("Neo.Attribute.GetUsage", stateReader.AttributeGetUsage)
	stateReader.Register("Neo.Attribute.GetData", stateReader.AttributeGetData)

	stateReader.Register("Neo.Input.GetHash", stateReader.InputGetHash)
	stateReader.Register("Neo.Input.GetIndex", stateReader.InputGetIndex)

	stateReader.Register("Neo.Output.GetAssetId", stateReader.OutputGetAssetId)
	stateReader.Register("Neo.Output.GetValue", stateReader.OutputGetValue)
	stateReader.Register("Neo.Output.GetScriptHash", stateReader.OutputGetCodeHash)

	stateReader.Register("Neo.Account.GetScriptHash", stateReader.AccountGetCodeHash)
	stateReader.Register("Neo.Account.GetBalance", stateReader.AccountGetBalance)

	stateReader.Register("Neo.Asset.GetAssetId", stateReader.AssetGetAssetId)
	stateReader.Register("Neo.Asset.GetAssetType", stateReader.AssetGetAssetType)
	stateReader.Register("Neo.Asset.GetAmount", stateReader.AssetGetAmount)
	stateReader.Register("Neo.Asset.GetAvailable", stateReader.AssetGetAvailable)
	stateReader.Register("Neo.Asset.GetPrecision", stateReader.AssetGetPrecision)
	stateReader.Register("Neo.Asset.GetOwner", stateReader.AssetGetOwner)
	stateReader.Register("Neo.Asset.GetAdmin", stateReader.AssetGetAdmin)
	stateReader.Register("Neo.Asset.GetIssuer", stateReader.AssetGetIssuer)

	stateReader.Register("Neo.Contract.GetScript", stateReader.ContractGetCode)

	stateReader.Register("Neo.Storage.GetContext", stateReader.StorageGetContext)

	return &stateReader
}

func (s *StateReader) Register(methodName string, handler func(engine *vm.ExecutionEngine) bool) bool {
	if _, ok := s.serviceMap[methodName]; ok {
		return false
	}
	s.serviceMap[methodName] = handler
	return true;
}

func (s *StateReader) GetServiceMap() map[string]func(*vm.ExecutionEngine) bool {
	return s.serviceMap
}

func (s *StateReader) RuntimeGetTrigger(e *vm.ExecutionEngine) bool {
	vm.PushData(e, int(e.GetTrigger()))
	return true
}

func (s *StateReader) RuntimeNotify(e *vm.ExecutionEngine) bool {
	item := vm.PopStackItem(e)
	s.NotifyInfo(item)
	return true
}

func (s *StateReader) NotifyInfo(item types.StackItem)  {
	switch item.(type) {
	case *types.Boolean:
		log.Info("notifyInfo",item.GetBoolean())
	case *types.Integer:
		log.Info("notifyInfo",item.GetBigInteger())
	case *types.ByteArray:
		log.Info("notifyInfo",common.BytesToHexString(item.GetByteArray()))
	case *types.GeneralInterface:
		interop := item.GetInterface()
		log.Info("notifyInfo",interop)
	case *types.Array:
		items := item.GetArray();
		for i := 0; i < len(items); i++ {
			s.NotifyInfo(items[i])
		}
	}
}

func (s *StateReader) RuntimeLog(e *vm.ExecutionEngine) bool {
	data := vm.PopByteArray(e)
	log.Info("log hex", common.BytesToHexString(data));
	log.Info("RuntimeLog msg:", string(data))
	return true
}

func (s *StateReader) RuntimeGetTime(e *vm.ExecutionEngine) bool {
	var height uint32 = 0
	if DefaultLedger != nil {
		height = DefaultLedger.Blockchain.BlockHeight
		hash, err := DefaultLedger.Store.GetBlockHash(height)
		if err != nil {
			block, gerr := GetGenesisBlock()
			if gerr != nil {
				return false
			}
			hash = block.Hash()
		}
		header, err := DefaultLedger.Store.GetHeader(hash)
		vm.PushData(e, header.Timestamp)
	}
	return true;
}

func (s *StateReader) CheckWitnessHash(engine *vm.ExecutionEngine, programHash common.Uint168) (bool, error) {
	if engine.GetDataContainer() == nil {
		return false, errors.New("CheckWitnessHash getDataContainer is null")
	}

	tx := engine.GetDataContainer().(*core.Transaction)
	hashForVerify, err := GetTxProgramHashes(tx)
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
	h, err := crypto.ToProgramHash(c)
	if err != nil {
		return false, err
	}
	return s.CheckWitnessHash(engine, *h)
}

func (s *StateReader) RuntimeCheckWitness(e *vm.ExecutionEngine) bool {
	data := vm.PopByteArray(e)
	var (
		result bool
		err    error
	)
	if len(data) == 21 {
		program, err := common.Uint168FromBytes(data)
		if err != nil {
			return false
		}
		result, err = s.CheckWitnessHash(e, *program)
	} else if len(data) == 20 {
		temp := []byte{33}
		data = append(temp, data...)
		program, err := common.Uint168FromBytes(data)
		if err != nil {
			return false
		}
		result, err = s.CheckWitnessHash(e, *program)
	} else if len(data) == 33 {
		publickKey, err := crypto.DecodePoint(data)
		if err != nil {
			return false
		}
		result, err = s.CheckWitnessPublicKey(e, publickKey)
	} else {
		return false
	}
	if err != nil {
		return false
	}
	vm.PushData(e, result)
	return true
}

func (s *StateReader) BlockChainGetHeight(e *vm.ExecutionEngine) bool {
	var i uint32 = 0
	if DefaultLedger != nil {
		i = DefaultLedger.Store.GetHeight();
	}
	vm.PushData(e, i)
	return true
}

func (s *StateReader) BlockChainGetHeader(e *vm.ExecutionEngine) bool {
	var (
		header *core.Header
		err    error
	)
	data := vm.PopByteArray(e)
	l := len(data)

	if l <= 5 {
		b := new(big.Int)
		height := b.SetBytes(common.BytesReverse(data)).Int64()
		if DefaultLedger != nil {
			hash, err := DefaultLedger.Store.GetBlockHash((uint32(height)))
			if err != nil {
				return false
			}
			header, err = DefaultLedger.Store.GetHeader(hash)
		}
	} else if l == 32 {
		hash, _ := common.Uint256FromBytes(data)
		if DefaultLedger != nil {
			header, err = DefaultLedger.Store.GetHeader(*hash)
		}
	} else {
		return false
	}
	if err != nil {
		return false
	}
	vm.PushData(e, header)
	return true
}

func (s *StateReader) BlockChainGetBlock(e *vm.ExecutionEngine) bool {
	data := vm.PopByteArray(e)
	var (
		block *core.Block
		err   error
	)
	l := len(data)
	if l <= 5 {
		b := new(big.Int)
		height := uint32(b.SetBytes(common.BytesReverse(data)).Int64())
		if DefaultLedger != nil {
			hash, err := DefaultLedger.Store.GetBlockHash(height)
			if err != nil {
				return false
			}
			block, err = DefaultLedger.Store.GetBlock(hash)
		}
	} else if l == 32 {
		hash, err := common.Uint256FromBytes(data)
		if err != nil {
			return false
		}
		if DefaultLedger != nil {
			block, err = DefaultLedger.Store.GetBlock(*hash)
		}
	} else {
		return false
	}
	if err != nil {
		return false
	}
	vm.PushData(e, block)
	return true
}

func (s *StateReader) BlockChainGetTransaction(e *vm.ExecutionEngine) bool {
	d := vm.PopByteArray(e)
	hash, err := common.Uint256FromBytes(d)
	if err != nil {
		return false
	}
	tx, _, err := DefaultLedger.Store.GetTransaction(*hash)
	if err != nil {
		return false
	}
	vm.PushData(e, tx)
	return true
}

func (s *StateReader) BlockChainGetAccount(e *vm.ExecutionEngine) bool {
	d := vm.PopByteArray(e)
	hash, err := common.Uint168FromBytes(d)
	if err != nil {
		return false
	}

	account, err := DefaultLedger.Store.GetAccount(*hash)
	vm.PushData(e, account)
	return true
}

func (s *StateReader) BlockChainGetAsset(e *vm.ExecutionEngine) bool {
	d := vm.PopByteArray(e)
	hash, err := common.Uint256FromBytes(d)
	if err != nil {
		return false
	}
	asset, err := DefaultLedger.Store.GetAsset(*hash)
	if err != nil {
		return false
	}
	vm.PushData(e, asset)
	return true
}

func (s *StateReader) HeaderGetHash(e *vm.ExecutionEngine) bool {
	d := vm.PopInteropInterface(e)
	if d == nil {
		return false
	}
	hash := d.(*core.Header).Hash()
	vm.PushData(e, hash.Bytes())
	return true
}

func (s *StateReader) HeaderGetVersion(e *vm.ExecutionEngine) bool {
	d := vm.PopInteropInterface(e)
	if d == nil {
		return false
	}
	version := d.(*core.Header).Version
	vm.PushData(e, version)
	return true
}

func (s *StateReader) HeaderGetPrevHash(e *vm.ExecutionEngine) bool {
	d := vm.PopInteropInterface(e)
	if d == nil {
		return false
	}
	preHash := d.(*core.Header).Previous
	vm.PushData(e, preHash.Bytes())
	return true
}

func (s *StateReader) HeaderGetMerkleRoot(e *vm.ExecutionEngine) bool {
	d := vm.PopInteropInterface(e)
	if d == nil {
		return false
	}
	root := d.(*core.Header).MerkleRoot
	vm.PushData(e, root.Bytes())
	return true
}

func (s *StateReader) HeaderGetTimestamp(e *vm.ExecutionEngine) bool {
	d := vm.PopInteropInterface(e)
	if d == nil {
		return false
	}
	timeStamp := d.(*core.Header).Timestamp
	vm.PushData(e, timeStamp)
	return true
}

func (s *StateReader) HeaderGetConsensusData(e *vm.ExecutionEngine) bool {
	d := vm.PopInteropInterface(e)
	if d == nil {
		return false
	}
	consensusData := d.(*core.Header).Nonce
	vm.PushData(e, consensusData)
	return true
}

func (s *StateReader) BlockgetTransactionCount(e *vm.ExecutionEngine) bool {
	d := vm.PopInteropInterface(e)
	if d == nil {
		return false
	}
	transactions := d.(*core.Block).Transactions
	vm.PushData(e, len(transactions))
	return true
}

func (s *StateReader) BlockGetTransactions(e *vm.ExecutionEngine) bool {
	d := vm.PopInteropInterface(e)
	if d == nil {
		return false
	}
	transactions := d.(*core.Block).Transactions
	list := make([]types.GeneralInterface, 0)
	for _, v := range transactions {
		list = append(list, *types.NewGeneralInterface(v))
	}
	vm.PushData(e, list)
	return true
}

func (s *StateReader) BlockGetTransaction(e *vm.ExecutionEngine) bool {
	index := vm.PopInt(e)
	if index < 0 {
		return false
	}
	d := vm.PopInteropInterface(e)
	if d == nil {
		return false
	}
	transactions := d.(*core.Block).Transactions
	if index >= len(transactions) {
		return false
	}
	vm.PushData(e, transactions[index])
	return false
}

func (s *StateReader) TransactionGetHash(e *vm.ExecutionEngine) bool {
	d := vm.PopInteropInterface(e)
	if d == nil {
		return false
	}
	txHash := d.(*core.Transaction).Hash()
	vm.PushData(e, txHash.Bytes())
	return true
}

func (s *StateReader) TransactionGetType(e *vm.ExecutionEngine) bool {
	d := vm.PopInteropInterface(e)
	if d == nil {
		return false
	}
	txType := d.(*core.Transaction).TxType
	vm.PushData(e, int(txType))
	return true
}

func (s *StateReader) TransactionGetAttributes(e *vm.ExecutionEngine) bool {
	d := vm.PopInteropInterface(e)
	if d == nil {
		return false
	}
	attributes := d.(*core.Transaction).Attributes
	list := make([]types.GeneralInterface, 0)
	for _, v := range attributes {
		list = append(list, *types.NewGeneralInterface(v))
	}
	vm.PushData(e, list)
	return true
}

func (s *StateReader) TransactionGetInputs(e *vm.ExecutionEngine) bool {
	d := vm.PopInteropInterface(e)
	if d == nil {
		return false
	}
	inputs := d.(*core.Transaction).Inputs
	list := make([]types.StackItem, 0)
	for _, v := range inputs {
		list = append(list, types.NewGeneralInterface(v))
	}
	vm.PushData(e, list)
	return true
}

func (s *StateReader) TransactionGetOutputs(e *vm.ExecutionEngine) bool {
	d := vm.PopInteropInterface(e)
	if d == nil {
		return false
	}
	outputs := d.(*core.Transaction).Outputs
	list := make([]types.StackItem, 0)
	for _, v := range outputs {
		list = append(list, types.NewGeneralInterface(v))
	}
	vm.PushData(e, list)
	return true
}

func (s *StateReader) TransactionGetReferences(e *vm.ExecutionEngine) bool {
	d := vm.PopInteropInterface(e)
	if d == nil {
		return false
	}

	references, err := DefaultLedger.Store.GetTxReference(d.(*core.Transaction))
	if err != nil {
		return false
	}
	list := make([]types.StackItem, 0)
	for _, v := range references {
		list = append(list, types.NewGeneralInterface(v))//
	}
	vm.PushData(e, list)
	return true
}

func (s *StateReader) AttributeGetUsage(e *vm.ExecutionEngine) bool {
	d := vm.PopInteropInterface(e)
	if d == nil {
		return false
	}
	attribute := d.(*core.Attribute)
	vm.PushData(e, int(attribute.Usage))
	return true
}

func (s *StateReader) AttributeGetData(e *vm.ExecutionEngine) bool {
	d := vm.PopInteropInterface(e)
	if d == nil {
		return false
	}
	attribute := d.(*core.Attribute)
	vm.PushData(e, attribute.Data)
	return true
}

func (s *StateReader) InputGetHash(e *vm.ExecutionEngine) bool {
	d := vm.PopInteropInterface(e)
	if d == nil {
		return false
	}
	input := d.(*core.Input)
	vm.PushData(e, input.Previous.TxID.Bytes())
	return true
}

func (s *StateReader) InputGetIndex(e *vm.ExecutionEngine) bool {
	d := vm.PopInteropInterface(e)
	if d == nil {
		return false
	}
	input := d.(*core.Input)
	vm.PushData(e, input.Previous.Index)
	return true
}

func (s *StateReader) OutputGetAssetId(e *vm.ExecutionEngine) bool {
	d := vm.PopInteropInterface(e)
	if d == nil {
		return false
	}
	output := d.(*core.Output)
	vm.PushData(e, output.AssetID.Bytes())
	return true
}

func (s *StateReader) OutputGetValue(e *vm.ExecutionEngine) bool {
	d := vm.PopInteropInterface(e)
	if d == nil {
		return false
	}
	output := d.(*core.Output)
	bytes, _ := output.Value.Bytes()
	vm.PushData(e, bytes)
	return true
}

func (s *StateReader) OutputGetCodeHash(e *vm.ExecutionEngine) bool {
	d := vm.PopInteropInterface(e)
	if d == nil {
		return false
	}
	output := d.(*core.Output)
	vm.PushData(e, output.ProgramHash.Bytes())
	return true
}

func (s *StateReader) AccountGetCodeHash(e *vm.ExecutionEngine) bool {
	d := vm.PopInteropInterface(e)
	if d == nil {
		log.Info("Get AccountState error in function AccountGetCodeHash")
		return false
	}
	accountState := d.(*states.AccountState).ProgramHash
	vm.PushData(e, accountState.Bytes())
	return true
}

func (s *StateReader) AccountGetBalance(e *vm.ExecutionEngine) bool {

	d := vm.PopInteropInterface(e)
	if d == nil {
		log.Info("Get AccountState error in function AccountGetCodeHash")
		return false
	}
	accountState := d.(*states.AccountState)
	if accountState == nil {
		log.Info("Get AccountState error in function AccountGetCodeHash")
		return false
	}

	assetIdByte := vm.PopByteArray(e)
	assetId, err := common.Uint256FromBytes(assetIdByte)
	if err != nil {
		return false
	}

	balance := common.Fixed64(0)
	if v, ok := accountState.Balances[*assetId]; ok {
		balance = v
	}
	vm.PushData(e, balance.IntValue())
	return true
}

func (s *StateReader) AssetGetAssetId(e *vm.ExecutionEngine) bool {
	d := vm.PopInteropInterface(e)
	if d == nil {
		return false
	}
	assetState := d.(*states.AssetState)
	vm.PushData(e, assetState.AssetId.Bytes())
	return true
}

func (s *StateReader) AssetGetAssetType(e *vm.ExecutionEngine) bool {
	d := vm.PopInteropInterface(e)
	if d == nil {
		return false
	}
	assetState := d.(*states.AssetState)
	vm.PushData(e, int(assetState.AssetType))
	return true
}

func (s *StateReader) AssetGetAmount(e *vm.ExecutionEngine) bool {
	d := vm.PopInteropInterface(e)
	if d == nil {
		return false
	}
	assetState := d.(*states.AssetState)
	vm.PushData(e, assetState.Amount.IntValue())
	return true
}

func (s *StateReader) AssetGetAvailable(e *vm.ExecutionEngine) bool {
	d := vm.PopInteropInterface(e)
	if d == nil {
		return false
	}
	assetState := d.(*states.AssetState)
	vm.PushData(e, assetState.Avaliable.IntValue())
	return true
}

func (s *StateReader) AssetGetPrecision(e *vm.ExecutionEngine) bool {
	d := vm.PopInteropInterface(e)
	if d == nil {
		return false
	}
	assetState := d.(*states.AssetState)
	vm.PushData(e, int(assetState.Precision))
	return true
}

func (s *StateReader) AssetGetOwner(e *vm.ExecutionEngine) bool {
	d := vm.PopInteropInterface(e)
	if d == nil {
		return false
	}
	assetState := d.(*states.AssetState)
	owner, err := assetState.Owner.EncodePoint(true)
	if err != nil {
		return false
	}
	vm.PushData(e, owner)
	return true
}

func (s *StateReader) AssetGetAdmin(e *vm.ExecutionEngine) bool {
	d := vm.PopInteropInterface(e)
	if d == nil {
		return false
	}
	assetState := d.(*states.AssetState)
	vm.PushData(e, assetState.Admin.Bytes())
	return true
}

func (s *StateReader) AssetGetIssuer(e *vm.ExecutionEngine) bool {
	d := vm.PopInteropInterface(e)
	if d == nil {
		return false
	}
	assetState := d.(*states.AssetState)
	vm.PushData(e, assetState.Issuer.Bytes())
	return true
}

func (s *StateReader) ContractGetCode(e *vm.ExecutionEngine) bool {
	d := vm.PopInteropInterface(e)
	if d == nil {
		return false
	}
	assetState := d.(*states.ContractState)
	vm.Push(e, assetState.Code.Code)
	return true
}

func (s *StateReader) StorageGetContext(e *vm.ExecutionEngine) bool {
	codeHash, err := common.Uint168FromBytes(e.Hash160(e.ExecutingScript()))
	if err != nil {
		return false
	}
	vm.PushData(e, NewStorageContext(codeHash))
	return true
}
