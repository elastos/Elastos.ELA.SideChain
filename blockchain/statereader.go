package blockchain

import (
	"math/big"
	"errors"
	"io"
	"bytes"

	"github.com/elastos/Elastos.ELA.Utility/common"
	"github.com/elastos/Elastos.ELA.Utility/crypto"

	"github.com/elastos/Elastos.ELA.SideChain/vm"
	"github.com/elastos/Elastos.ELA.SideChain/contract"
	"github.com/elastos/Elastos.ELA.SideChain/core"
	"github.com/elastos/Elastos.ELA.SideChain/vm/types"
	"github.com/elastos/Elastos.ELA.SideChain/smartcontract/states"
	"github.com/elastos/Elastos.ELA.SideChain/log"
	"github.com/elastos/Elastos.ELA.SideChain/store"
	"github.com/elastos/Elastos.ELA.SideChain/smartcontract/enumerators"
	"github.com/elastos/Elastos.ELA.SideChain/events"
)

type StateReader struct {
	serviceMap map[string]func(engine *vm.ExecutionEngine) bool
	NotifyEvent *events.Event
	LogEvent *events.Event
}

func NewStateReader() *StateReader {
	var stateReader StateReader
	stateReader.serviceMap = make(map[string]func(*vm.ExecutionEngine) bool, 0)
	stateReader.NotifyEvent = events.NewEvent()
	stateReader.LogEvent = events.NewEvent()

	stateReader.Register("Neo.Runtime.GetTrigger", stateReader.RuntimeGetTrigger)
	stateReader.Register("Neo.Runtime.CheckWitness", stateReader.RuntimeCheckWitness)
	stateReader.Register("Neo.Runtime.Notify", stateReader.RuntimeNotify)
	stateReader.Register("Neo.Runtime.Log", stateReader.RuntimeLog)
	stateReader.Register("Neo.Runtime.GetTime", stateReader.RuntimeGetTime)
	stateReader.Register("Neo.Runtime.Serialize", stateReader.RuntimeSerialize)
	stateReader.Register("Neo.Runtime.Deserialize", stateReader.RuntimeDerialize)

	stateReader.Register("Neo.Blockchain.GetHeight", stateReader.BlockChainGetHeight)
	stateReader.Register("Neo.Blockchain.GetHeader", stateReader.BlockChainGetHeader)
	stateReader.Register("Neo.Blockchain.GetBlock", stateReader.BlockChainGetBlock)
	stateReader.Register("Neo.Blockchain.GetTransaction", stateReader.BlockChainGetTransaction)
	stateReader.Register("Neo.Blockchain.GetTransactionHeight", stateReader.BlockchainGetTransactionHeight)
	stateReader.Register("Neo.Blockchain.GetAccount", stateReader.BlockChainGetAccount)
	stateReader.Register("Neo.Blockchain.GetValidators", stateReader.BlockChainGetValidators)
	stateReader.Register("Neo.Blockchain.GetAsset", stateReader.BlockChainGetAsset)

	stateReader.Register("Neo.Header.GetIndex", stateReader.HeaderGetHeight);
	stateReader.Register("Neo.Header.GetHash", stateReader.HeaderGetHash)
	stateReader.Register("Neo.Header.GetVersion", stateReader.HeaderGetVersion)
	stateReader.Register("Neo.Header.GetPrevHash", stateReader.HeaderGetPrevHash)
	stateReader.Register("Neo.Header.GetMerkleRoot", stateReader.HeaderGetMerkleRoot)
	stateReader.Register("Neo.Header.GetTimestamp", stateReader.HeaderGetTimestamp)
	stateReader.Register("Neo.Header.GetConsensusData", stateReader.HeaderGetConsensusData)
	stateReader.Register("Neo.Header.GetNextConsensus", stateReader.HeaderGetNextConsensus)

	stateReader.Register("Neo.Block.GetTransactionCount", stateReader.BlockgetTransactionCount)
	stateReader.Register("Neo.Block.GetTransactions", stateReader.BlockGetTransactions)
	stateReader.Register("Neo.Block.GetTransaction", stateReader.BlockGetTransaction)

	stateReader.Register("Neo.Transaction.GetHash", stateReader.TransactionGetHash)
	stateReader.Register("Neo.Transaction.GetType", stateReader.TransactionGetType)
	stateReader.Register("Neo.Transaction.GetAttributes", stateReader.TransactionGetAttributes)
	stateReader.Register("Neo.Transaction.GetInputs", stateReader.TransactionGetInputs)
	stateReader.Register("Neo.Transaction.GetOutputs", stateReader.TransactionGetOutputs)
	stateReader.Register("Neo.Transaction.GetReferences", stateReader.TransactionGetReferences)
	stateReader.Register("Neo.Transaction.GetUnspentCoins", stateReader.TransactionGetUnspentCoins)
	stateReader.Register("Neo.InvocationTransaction.GetScript", stateReader.InvocationTransactionGetScript)

	stateReader.Register("Neo.Attribute.GetUsage", stateReader.AttributeGetUsage)
	stateReader.Register("Neo.Attribute.GetData", stateReader.AttributeGetData)

	stateReader.Register("Neo.Input.GetHash", stateReader.InputGetHash)
	stateReader.Register("Neo.Input.GetIndex", stateReader.InputGetIndex)

	stateReader.Register("Neo.Output.GetAssetId", stateReader.OutputGetAssetId)
	stateReader.Register("Neo.Output.GetValue", stateReader.OutputGetValue)
	stateReader.Register("Neo.Output.GetScriptHash", stateReader.OutputGetCodeHash)

	stateReader.Register("Neo.Account.GetScriptHash", stateReader.AccountGetCodeHash)
	stateReader.Register("Neo.Account.GetBalance", stateReader.AccountGetBalance)
	stateReader.Register("Neo.Account.GetVotes", stateReader.AccountGetVotes)

	stateReader.Register("Neo.Asset.GetAssetId", stateReader.AssetGetAssetId)
	stateReader.Register("Neo.Asset.GetAssetType", stateReader.AssetGetAssetType)
	stateReader.Register("Neo.Asset.GetAmount", stateReader.AssetGetAmount)
	stateReader.Register("Neo.Asset.GetAvailable", stateReader.AssetGetAvailable)
	stateReader.Register("Neo.Asset.GetPrecision", stateReader.AssetGetPrecision)
	stateReader.Register("Neo.Asset.GetOwner", stateReader.AssetGetOwner)
	stateReader.Register("Neo.Asset.GetAdmin", stateReader.AssetGetAdmin)
	stateReader.Register("Neo.Asset.GetIssuer", stateReader.AssetGetIssuer)

	stateReader.Register("Neo.Contract.GetScript", stateReader.ContractGetCode)
	stateReader.Register("Neo.Contract.IsPayable", stateReader.ContractIsPayable)

	stateReader.Register("Neo.Storage.GetContext", stateReader.StorageGetContext)
	stateReader.Register("Neo.Storage.GetReadOnlyContext", stateReader.StorageGetReadOnlyContext)
	stateReader.Register("Neo.StorageContext.AsReadOnly", stateReader.StorageContextAsReadOnly);

	stateReader.Register("Neo.Iterator.Key", stateReader.IteratorKey);
	stateReader.Register("Neo.Iterator.Next", stateReader.EnumeratorNext);
	stateReader.Register("Neo.Iterator.Value", stateReader.EnumeratorValue);
	stateReader.Register("Neo.Iterator.Keys", stateReader.IteratorKeys);
	stateReader.Register("Neo.Iterator.Values", stateReader.IteratorValues);


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
	if s.NotifyEvent != nil && s.NotifyEvent.HashSubscriber(events.EventRunTimeNotify) {
		s.NotifyEvent.Notify(events.EventRunTimeNotify, item)
	}
	return true
}

func (s *StateReader) RuntimeLog(e *vm.ExecutionEngine) bool {
	data := vm.PopStackItem(e)
	if s.LogEvent != nil && s.LogEvent.HashSubscriber(events.EventRunTimeLog) {
		s.LogEvent.Notify(events.EventRunTimeLog, data)
	}
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

func (s *StateReader) RuntimeSerialize(e *vm.ExecutionEngine) bool {
	buf := new(bytes.Buffer)
	item := vm.PopStackItem(e)
	s.SerializeStackItem(item, buf)
	vm.PushData(e, buf.Bytes())
	return true;
}

func (s *StateReader) RuntimeDerialize (e *vm.ExecutionEngine) bool {
	data := vm.PopStackItem(e).GetByteArray()
	reader := bytes.NewReader(data)
	item, err := s.DerializeStackItem(reader)
	if err != nil {
		return false
	}
	vm.PushData(e, item)
	return true;
}

func (s *StateReader) DerializeStackItem(r io.Reader) (types.StackItem, error) {
	var itemType = make([]byte, 1)
	_, err := r.Read(itemType)
	if err != nil {
		return nil, err
	}

	switch (types.StackItemType(itemType[0])) {
	case types.TYPE_ByteArray:
		bytes, err := common.ReadVarBytes(r)
		if err != nil {
			return nil, err
		}
		return types.NewByteArray(bytes), nil
	case types.TYPE_Boolean:
		data, err := common.ReadUint8(r)
		if err != nil {
			return nil, err
		}
		bytes := true
		if data == 0 {
			bytes = false
		}
		if err != nil {
			return nil, err
		}
		return types.NewBoolean(bytes), nil
	case types.TYPE_Integer:
		data, err := common.ReadUint64(r)
		if err != nil {
			return nil, err
		}
		return types.NewInteger(big.NewInt(int64(data))), nil
	case types.TYPE_Array, types.TYPE_Struct:
		len, err := common.ReadVarUint(r, 0x00)
		if err != nil {
			return nil, err
		}
		var items = make([]types.StackItem, len)
		for i := 0; i < int(len); i++ {
			items[i], err = s.DerializeStackItem(r)
			if err != nil {
				return nil, err
			}
		}
		return types.NewArray(items), nil
	case types.TYPE_Map:
		dictionary := types.NewDictionary()
		len, err := common.ReadVarUint(r, 0x00)
		if err != nil {
			return nil, err
		}
		for i := 0; i < int(len); i++ {
			key, err := s.DerializeStackItem(r)
			if err != nil {
				return nil, err
			}
			value, err := s.DerializeStackItem(r)
			if err != nil {
				return nil, err
			}
			dictionary.PutStackItem(key, value)
		}
		return dictionary, nil
	}
	return nil, errors.New("error type")
}

func (s *StateReader) SerializeStackItem(item types.StackItem, w io.Writer) {
	switch item.(type) {
	case *types.Boolean:
		w.Write([]byte{byte(types.TYPE_Boolean)})
		w.Write(item.GetByteArray())
	case *types.Integer:
		w.Write([]byte{byte(types.TYPE_Integer)})
		common.WriteUint64(w, item.GetBigInteger().Uint64())
	case *types.ByteArray:
		w.Write([]byte{byte(types.TYPE_ByteArray)})
		common.WriteVarBytes(w, item.GetByteArray())
	case *types.GeneralInterface:
		w.Write([]byte{byte(types.TYPE_InteropInterface)})
		w.Write(item.GetByteArray());
	case *types.Array:
		w.Write([]byte{byte(types.TYPE_Array)})
		items := item.GetArray();
		common.WriteVarUint(w, (uint64(len(items))))
		for i := 0; i < len(items); i++ {
			s.SerializeStackItem(items[i], w)
		}
	case *types.Dictionary:
		dict := item.(*types.Dictionary)
		w.Write([]byte{byte(types.TYPE_Map)})
		dictMap := dict.GetMap();
		common.WriteVarUint(w, (uint64(len(dictMap))))
		for key := range dictMap {
			s.SerializeStackItem(key, w)
			s.SerializeStackItem(dictMap[key], w)
		}
	}
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

func (s *StateReader) BlockchainGetTransactionHeight(e *vm.ExecutionEngine) bool {
	d := vm.PopByteArray(e)
	hash, err := common.Uint256FromBytes(d)
	if err != nil {
		return false
	}
	_, height, err := DefaultLedger.Store.GetTransaction(*hash)
	if err != nil {
		return false
	}
	vm.PushData(e, height)
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

func (s *StateReader) BlockChainGetValidators(e *vm.ExecutionEngine) bool {
	//note ela chain is not have Validators data. because consensus is pow
	pkList := make([]types.StackItem, 0)
	vm.PushData(e, pkList)
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

func (s *StateReader) HeaderGetHeight(e *vm.ExecutionEngine) bool {
	d := vm.PopInteropInterface(e)
	if d == nil {
		return false
	}

	var height uint32
	switch d.(type) {
	case *core.Header:
		height = d.(*core.Header).Height
	case *core.Block:
		height = d.(*core.Block).Header.Height
	default:
		return false
	}

	vm.PushData(e, height)
	return true
}

func (s *StateReader) HeaderGetHash(e *vm.ExecutionEngine) bool {
	d := vm.PopInteropInterface(e)
	if d == nil {
		return false
	}
	var hash common.Uint256
	switch d.(type) {
	case *core.Header:
		hash = d.(*core.Header).Hash()
	case *core.Block:
		hash = d.(*core.Block).Header.Hash()
	default:
		return false
	}
	vm.PushData(e, hash.Bytes())
	return true
}

func (s *StateReader) HeaderGetVersion(e *vm.ExecutionEngine) bool {
	d := vm.PopInteropInterface(e)
	if d == nil {
		return false
	}
	var version uint32
	switch d.(type) {
	case *core.Header:
		version = d.(*core.Header).Version
	case *core.Block:
		version = d.(*core.Block).Header.Version
	default:
		return false
	}

	vm.PushData(e, version)
	return true
}

func (s *StateReader) HeaderGetPrevHash(e *vm.ExecutionEngine) bool {
	d := vm.PopInteropInterface(e)
	if d == nil {
		return false
	}

	var preHash common.Uint256
	switch d.(type) {
	case *core.Header:
		preHash = d.(*core.Header).Previous
	case *core.Block:
		preHash = d.(*core.Block).Header.Previous
	default:
		return false
	}

	vm.PushData(e, preHash.Bytes())
	return true
}

func (s *StateReader) HeaderGetMerkleRoot(e *vm.ExecutionEngine) bool {
	d := vm.PopInteropInterface(e)
	if d == nil {
		return false
	}

	var root common.Uint256
	switch d.(type) {
	case *core.Header:
		root = d.(*core.Header).MerkleRoot
	case *core.Block:
		root = d.(*core.Block).Header.MerkleRoot
	default:
		return false
	}

	vm.PushData(e, root.Bytes())
	return true
}

func (s *StateReader) HeaderGetTimestamp(e *vm.ExecutionEngine) bool {
	d := vm.PopInteropInterface(e)
	if d == nil {
		return false
	}

	var timeStamp uint32
	switch d.(type) {
	case *core.Header:
		timeStamp = d.(*core.Header).Timestamp
	case *core.Block:
		timeStamp = d.(*core.Block).Header.Timestamp
	default:
		return false
	}

	vm.PushData(e, timeStamp)
	return true
}

func (s *StateReader) HeaderGetConsensusData(e *vm.ExecutionEngine) bool {
	d := vm.PopInteropInterface(e)
	if d == nil {
		return false
	}

	var consensusData uint32
	switch d.(type) {
	case *core.Header:
		consensusData = d.(*core.Header).Nonce
	case *core.Block:
		consensusData = d.(*core.Block).Header.Nonce
	default:
		return false
	}

	vm.PushData(e, consensusData)
	return true
}

func (s *StateReader) HeaderGetNextConsensus(e *vm.ExecutionEngine) bool {
	d := vm.PopInteropInterface(e)
	if d == nil {
		return false
	}
	//note ela chain is not have NextConsensus data. because consensus is pow
	vm.PushData(e, 0)
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
	d := vm.PopInteropInterface(e)
	if d == nil {
		return false
	}
	index := vm.PopInt(e)
	if index < 0 {
		return false
	}
	transactions := d.(*core.Block).Transactions
	if index >= len(transactions) {
		return false
	}
	vm.PushData(e, transactions[index])
	return true
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

func (s *StateReader) TransactionGetUnspentCoins(e *vm.ExecutionEngine) bool {
	d := vm.PopInteropInterface(e)
	if d == nil {
		return false
	}
	tx := d.(*core.Transaction)
	unspentCoins, err := DefaultLedger.Store.GetUnspents(tx.Hash())
	if err != nil {
		return false
	}

	list := make([]types.StackItem, 0)
	for _, v := range unspentCoins {
		list = append(list, types.NewGeneralInterface(v))
	}
	vm.PushData(e, list)
	return true;
}

func (s *StateReader) InvocationTransactionGetScript(e *vm.ExecutionEngine) bool {
	d := vm.PopInteropInterface(e)
	if d == nil {
		return false
	}
	txtype := d.(*core.Transaction).TxType
	if txtype != core.Invoke {
		return false
	}
	payload := d.(*core.Transaction).Payload
	script := payload.(*core.PayloadInvoke).Code
	vm.PushData(e, script)
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

func (s *StateReader) AccountGetVotes(e *vm.ExecutionEngine) bool {
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
	//note ela chain is not have votes data. because consensus is pow
	pkList := make([]types.StackItem, 0)
	vm.PushData(e, pkList)
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
	if assetState == nil {
		return false
	}
	vm.PushData(e, assetState.Code.Code)
	return true
}

func (s *StateReader) ContractIsPayable(e *vm.ExecutionEngine) bool {
	d := vm.PopInteropInterface(e)
	if d == nil {
		return false
	}
	assetState := d.(*states.ContractState)
	if assetState == nil {
		return false
	}
	vm.PushData(e, true)
	return true
}

func (s *StateReader) StorageGetContext(e *vm.ExecutionEngine) bool {
	codeHash, err := common.Uint168FromBytes(e.Hash168(e.ExecutingScript()))
	if err != nil {
		return false
	}
	vm.PushData(e, NewStorageContext(codeHash))
	return true
}

func (s *StateReader) StorageGetReadOnlyContext(e *vm.ExecutionEngine) bool {

	codeHash, err := common.Uint168FromBytes(e.Hash168(e.ExecutingScript()))
	if err != nil {
		return false
	}
	context := NewStorageContext(codeHash)
	context.IsReadOnly = true
	vm.PushData(e, context)
	return true
}

func (s *StateReader) StorageContextAsReadOnly(e *vm.ExecutionEngine) bool {
	opInterface := vm.PopInteropInterface(e)
	if opInterface == nil {
		return false;
	}
	context := opInterface.(*StorageContext)
	if !context.IsReadOnly {
		newContext := NewStorageContext(context.codeHash)
		newContext.IsReadOnly = true
		vm.PushData(e, newContext)
	}
	return true;
}

func (s *StateReader) IteratorKey(e *vm.ExecutionEngine) bool {
	opInterface := vm.PopInteropInterface(e)
	if opInterface == nil {
		return false;
	}
	iter := opInterface.(store.IIterator)
	vm.PushData(e, iter.Key())
	return true
}

func (s *StateReader) EnumeratorNext(e *vm.ExecutionEngine) bool {
	opInterface := vm.PopInteropInterface(e)
	if opInterface == nil {
		return false;
	}
	iter := opInterface.(store.IIterator)
	vm.PushData(e, iter.Next())
	return true
}

func (s *StateReader) EnumeratorValue(e *vm.ExecutionEngine) bool {
	opInterface := vm.PopInteropInterface(e)
	if opInterface == nil {
		return false;
	}
	iter := opInterface.(store.IIterator)
	vm.PushData(e, iter.Value())
	return true
}

func (s *StateReader) IteratorKeys(e *vm.ExecutionEngine) bool {

	opInterface := vm.PopInteropInterface(e)
	if opInterface == nil {
		return false;
	}
	iter := opInterface.(store.IIterator)
	iterKeys := enumerators.NewIteratorKeys(iter)
	vm.PushData(e, iterKeys)
	return true
}

func (s *StateReader) IteratorValues(e *vm.ExecutionEngine) bool {

	opInterface := vm.PopInteropInterface(e)
	if opInterface == nil {
		return false;
	}
	iter := opInterface.(store.IIterator)
	iterValues := enumerators.NewIteratorValues(iter)
	vm.PushData(e, iterValues)
	return true
}
