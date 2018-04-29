package servers

import (
	"bytes"
	"errors"
	"fmt"
	"time"

	chain "github.com/elastos/Elastos.ELA.SideChain/blockchain"
	"github.com/elastos/Elastos.ELA.SideChain/config"
	. "github.com/elastos/Elastos.ELA.SideChain/core"
	. "github.com/elastos/Elastos.ELA.SideChain/errors"
	"github.com/elastos/Elastos.ELA.SideChain/log"
	. "github.com/elastos/Elastos.ELA.SideChain/protocol"

	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	. "github.com/elastos/Elastos.ELA.Utility/common"
	ela "github.com/elastos/Elastos.ELA/core"
)

const (
	AUXBLOCK_GENERATED_INTERVAL_SECONDS = 60
)

var NodeForServers Noder
var Pow *chain.PowService
var PreChainHeight uint64
var PreTime int64
var PreTransactionCount int

func GetTransactionInfo(header *Header, tx *ela.Transaction) *TransactionInfo {
	inputs := make([]InputInfo, len(tx.Inputs))
	for i, v := range tx.Inputs {
		inputs[i].TxID = BytesToHexString(v.Previous.TxID.Bytes())
		inputs[i].VOut = v.Previous.Index
		inputs[i].Sequence = v.Sequence
	}

	outputs := make([]OutputInfo, len(tx.Outputs))
	for i, v := range tx.Outputs {
		outputs[i].Value = v.Value.String()
		outputs[i].Index = uint32(i)
		address, _ := v.ProgramHash.ToAddress()
		outputs[i].Address = address
		outputs[i].AssetID = BytesToHexString(v.AssetID.Bytes())
		outputs[i].OutputLock = v.OutputLock
	}

	attributes := make([]AttributeInfo, len(tx.Attributes))
	for i, v := range tx.Attributes {
		attributes[i].Usage = v.Usage
		attributes[i].Data = BytesToHexString(v.Data)
	}

	programs := make([]ProgramInfo, len(tx.Programs))
	for i, v := range tx.Programs {
		programs[i].Code = BytesToHexString(v.Code)
		programs[i].Parameter = BytesToHexString(v.Parameter)
	}

	var txHash = tx.Hash()
	var txHashStr = txHash.String()
	var size = uint32(tx.GetSize())
	var blockHash Uint256
	var confirmations uint32
	var time uint32
	var blockTime uint32
	if header != nil {
		confirmations = chain.DefaultLedger.Blockchain.GetBestHeight() - header.Height + 1
		time = header.Timestamp
		blockTime = header.Timestamp
	}

	return &TransactionInfo{
		TxId:           txHashStr,
		Hash:           txHashStr,
		Size:           size,
		VSize:          size,
		Version:        header.Version,
		LockTime:       tx.LockTime,
		Inputs:         inputs,
		Outputs:        outputs,
		BlockHash:      blockHash.String(),
		Confirmations:  confirmations,
		Time:           time,
		BlockTime:      blockTime,
		TxType:         tx.TxType,
		PayloadVersion: tx.PayloadVersion,
		Payload:        getPayloadInfo(tx.Payload),
		Attributes:     attributes,
		Programs:       programs,
	}
}

func PayloadInfoToTransPayload(p PayloadInfo) (ela.Payload, error) {

	switch object := p.(type) {
	case RegisterAssetInfo:
		obj := new(ela.PayloadRegisterAsset)
		obj.Asset = object.Asset
		amount, err := StringToFixed64(object.Amount)
		if err != nil {
			return nil, err
		}
		obj.Amount = *amount
		bytes, err := HexStringToBytes(object.Controller)
		if err != nil {
			return nil, err
		}
		controller, err := Uint168FromBytes(bytes)
		obj.Controller = *controller
		return obj, nil
	case TransferCrossChainAssetInfo:
		obj := new(ela.PayloadTransferCrossChainAsset)
		obj.AddressesMap = object.AddressesMap
		return obj, nil
	}

	return nil, errors.New("Invalid payload type.")
}

func GetTransaction(txInfo *TransactionInfo) (*ela.Transaction, error) {
	txPaload, err := PayloadInfoToTransPayload(txInfo.Payload)
	if err != nil {
		return nil, err
	}

	var txAttribute []*ela.Attribute
	for _, att := range txInfo.Attributes {
		attData, err := HexStringToBytes(att.Data)
		if err != nil {
			return nil, err
		}
		txAttr := &ela.Attribute{
			Usage: att.Usage,
			Data:  attData,
			Size:  0,
		}
		txAttribute = append(txAttribute, txAttr)
	}

	var txUTXOTxInput []*ela.Input
	for _, input := range txInfo.Inputs {
		txID, err := HexStringToBytes(input.TxID)
		if err != nil {
			return nil, err
		}
		referID, err := Uint256FromBytes(txID)
		if err != nil {
			return nil, err
		}
		utxoInput := &ela.Input{
			Previous: ela.OutPoint{
				TxID:  *referID,
				Index: input.VOut,
			},
			Sequence: input.Sequence,
		}
		txUTXOTxInput = append(txUTXOTxInput, utxoInput)
	}

	var txOutputs []*ela.Output
	for _, output := range txInfo.Outputs {
		assetIdBytes, err := HexStringToBytes(output.AssetID)
		if err != nil {
			return nil, err
		}
		assetId, err := Uint256FromBytes(assetIdBytes)
		if err != nil {
			return nil, err
		}
		value, err := StringToFixed64(output.Value)
		if err != nil {
			return nil, err
		}
		programHash, err := Uint168FromAddress(output.Address)
		if err != nil {
			return nil, err
		}
		output := &ela.Output{
			AssetID:     *assetId,
			Value:       *value,
			OutputLock:  output.OutputLock,
			ProgramHash: *programHash,
		}
		txOutputs = append(txOutputs, output)
	}

	var txPrograms []*ela.Program
	for _, pgrm := range txInfo.Programs {
		code, err := HexStringToBytes(pgrm.Code)
		if err != nil {
			return nil, err
		}
		parameter, err := HexStringToBytes(pgrm.Parameter)
		if err != nil {
			return nil, err
		}
		txProgram := &ela.Program{
			Code:      code,
			Parameter: parameter,
		}
		txPrograms = append(txPrograms, txProgram)
	}

	txTransaction := &ela.Transaction{
		TxType:         txInfo.TxType,
		PayloadVersion: txInfo.PayloadVersion,
		Payload:        txPaload,
		Attributes:     txAttribute,
		Inputs:         txUTXOTxInput,
		Outputs:        txOutputs,
		Programs:       txPrograms,
	}
	return txTransaction, nil
}

// Input JSON string examples for getblock method as following:
func GetRawTransaction(param Params) map[string]interface{} {
	str, ok := param.String("hash")
	if !ok {
		return ResponsePack(InvalidParams, "")
	}

	hex, err := HexStringToBytes(str)
	if err != nil {
		return ResponsePack(InvalidParams, "")
	}
	var hash Uint256
	err = hash.Deserialize(bytes.NewReader(hex))
	if err != nil {
		return ResponsePack(InvalidTransaction, "")
	}
	tx, height, err := chain.DefaultLedger.Store.GetTransaction(hash)
	if err != nil {
		return ResponsePack(UnknownTransaction, "")
	}
	bHash, err := chain.DefaultLedger.Store.GetBlockHash(height)
	if err != nil {
		return ResponsePack(UnknownTransaction, "")
	}
	header, err := chain.DefaultLedger.Store.GetHeader(bHash)
	if err != nil {
		return ResponsePack(UnknownTransaction, "")
	}

	decoded, ok := param.Bool("decoded")
	if decoded {
		return ResponsePack(Success, GetTransactionInfo(header, tx))
	} else {
		buf := new(bytes.Buffer)
		tx.Serialize(buf)
		return ResponsePack(Success, BytesToHexString(buf.Bytes()))
	}
}

func GetNeighbors(param Params) map[string]interface{} {
	addr, _ := NodeForServers.GetNeighborAddrs()
	return ResponsePack(Success, addr)
}

func GetNodeState(param Params) map[string]interface{} {
	n := NodeInfo{
		State:    uint(NodeForServers.State()),
		Time:     NodeForServers.GetTime(),
		Port:     NodeForServers.Port(),
		ID:       NodeForServers.ID(),
		Version:  NodeForServers.Version(),
		Services: NodeForServers.Services(),
		Relay:    NodeForServers.IsRelay(),
		Height:   NodeForServers.Height(),
		TxnCnt:   NodeForServers.GetTxnCnt(),
		RxTxnCnt: NodeForServers.GetRxTxnCnt(),
	}
	return ResponsePack(Success, n)
}

func SetLogLevel(param Params) map[string]interface{} {
	level, ok := param["level"].(float64)
	if !ok || level < 0 {
		return ResponsePack(InvalidParams, "level must be an integer in 0-6")
	}

	if err := log.Log.SetDebugLevel(int(level)); err != nil {
		return ResponsePack(InvalidParams, err.Error())
	}
	return ResponsePack(Success, fmt.Sprint("log level has been set to ", level))
}

func SubmitAuxBlock(param Params) map[string]interface{} {
	blockHash, ok := param.String("blockhash")
	if !ok {
		return ResponsePack(InvalidParams, "")
	}
	if _, ok := Pow.MsgBlock.BlockData[blockHash]; !ok {
		log.Trace("[json-rpc:SubmitAuxBlock] receive invalid block hash value:", blockHash)
		return ResponsePack(InvalidParams, "")
	}

	sideAuxPow, ok := param.String("sideauxpow")
	if !ok {
		return ResponsePack(InvalidParams, "")
	}

	buf, _ := HexStringToBytes(sideAuxPow)
	err := Pow.MsgBlock.BlockData[blockHash].Header.SideAuxPow.Deserialize(bytes.NewReader(buf))
	if err != nil {
		log.Trace(err)
		return ResponsePack(InternalError, "[json-rpc:SubmitAuxBlock] deserialize side aux pow failed")
	}

	_, _, err = chain.DefaultLedger.Blockchain.AddBlock(Pow.MsgBlock.BlockData[blockHash])
	if err != nil {
		log.Trace(err)
		return ResponsePack(InternalError, "")
	}

	Pow.MsgBlock.Mutex.Lock()
	for key := range Pow.MsgBlock.BlockData {
		delete(Pow.MsgBlock.BlockData, key)
	}
	Pow.MsgBlock.Mutex.Unlock()
	log.Trace("AddBlock called finished and Pow.MsgBlock.BlockData has been deleted completely")

	log.Info(sideAuxPow, blockHash)
	return ResponsePack(Success, "")
}

func GenerateAuxBlock(addr string) (*Block, string, bool) {
	msgBlock := &Block{}
	if NodeForServers.Height() == 0 || PreChainHeight != NodeForServers.Height() ||
		time.Now().Unix()-PreTime > AUXBLOCK_GENERATED_INTERVAL_SECONDS {
		if PreChainHeight != NodeForServers.Height() {
			PreChainHeight = NodeForServers.Height()
			PreTime = time.Now().Unix()
			PreTransactionCount = Pow.GetTransactionCount()
		}

		currentTxsCount := Pow.CollectTransactions(msgBlock)
		if 0 == currentTxsCount {
			// return nil, "currentTxs is nil", false
		}

		msgBlock, err := Pow.GenerateBlock(addr)
		if nil != err {
			return nil, "msgBlock generate err", false
		}

		curHash := msgBlock.Hash()
		curHashStr := BytesToHexString(curHash.Bytes())

		Pow.MsgBlock.Mutex.Lock()
		Pow.MsgBlock.BlockData[curHashStr] = msgBlock
		Pow.MsgBlock.Mutex.Unlock()

		PreChainHeight = NodeForServers.Height()
		PreTime = time.Now().Unix()
		PreTransactionCount = currentTxsCount // Don't Call GetTransactionCount()

		return msgBlock, curHashStr, true
	}
	return nil, "", false
}

func CreateAuxBlock(param Params) map[string]interface{} {
	msgBlock, curHashStr, _ := GenerateAuxBlock(config.Parameters.PowConfiguration.PayToAddr)
	if nil == msgBlock {
		return ResponsePack(UnknownBlock, "")
	}

	addr, ok := param.String("paytoaddress")
	if !ok {
		return ResponsePack(InvalidParams, "")
	}

	type SideAuxBlock struct {
		GenesisHash       string `json:"genesishash"`
		Height            uint64 `json:"height"`
		Bits              string `json:"bits"`
		Hash              string `json:"hash"`
		PreviousBlockHash string `json:"previousblockhash"`
	}

	Pow.PayToAddr = addr

	genesisHash, err := chain.DefaultLedger.Store.GetBlockHash(uint32(0))
	if err != nil {
		return ResponsePack(Error, "Get genesis hash failed")
	}
	genesisHashStr := BytesToHexString(genesisHash.Bytes())

	preHash := chain.DefaultLedger.Blockchain.CurrentBlockHash()
	preHashStr := BytesToHexString(preHash.Bytes())

	SendToAux := SideAuxBlock{
		GenesisHash:       genesisHashStr,
		Height:            NodeForServers.Height(),
		Bits:              fmt.Sprintf("%x", msgBlock.Bits), //difficulty
		Hash:              curHashStr,
		PreviousBlockHash: preHashStr,
	}
	return ResponsePack(Success, &SendToAux)
}

func GetInfo(param Params) map[string]interface{} {
	RetVal := struct {
		Version        int    `json:"version"`
		Balance        int    `json:"balance"`
		Blocks         uint64 `json:"blocks"`
		Timeoffset     int    `json:"timeoffset"`
		Connections    uint   `json:"connections"`
		Testnet        bool   `json:"testnet"`
		Keypoololdest  int    `json:"keypoololdest"`
		Keypoolsize    int    `json:"keypoolsize"`
		Unlocked_until int    `json:"unlocked_until"`
		Paytxfee       int    `json:"paytxfee"`
		Relayfee       int    `json:"relayfee"`
		Errors         string `json:"errors"`
	}{
		Version:        config.Parameters.Version,
		Balance:        0,
		Blocks:         NodeForServers.Height(),
		Timeoffset:     0,
		Connections:    NodeForServers.GetConnectionCnt(),
		Testnet:        config.Parameters.PowConfiguration.TestNet,
		Keypoololdest:  0,
		Keypoolsize:    0,
		Unlocked_until: 0,
		Paytxfee:       0,
		Relayfee:       0,
		Errors:         "Tobe written"}
	return ResponsePack(Success, &RetVal)
}

func AuxHelp(param Params) map[string]interface{} {

	//TODO  and description for this rpc-interface
	return ResponsePack(Success, "createauxblock==submitauxblock")
}

func ToggleMining(param Params) map[string]interface{} {
	mining, ok := param.Bool("mining")
	if !ok {
		return ResponsePack(InvalidParams, "")
	}

	var message string
	if mining {
		go Pow.Start()
		message = "mining started"
	} else {
		go Pow.Halt()
		message = "mining stopped"
	}

	return ResponsePack(Success, message)
}

func ManualMining(param Params) map[string]interface{} {
	count, ok := param.Uint("count")
	if !ok {
		return ResponsePack(InvalidParams, "")
	}

	ret := make([]string, count)

	blockHashes, err := Pow.ManualMining(uint32(count))
	if err != nil {
		return ResponsePack(Error, err)
	}

	for i, hash := range blockHashes {
		//ret[i] = hash.ToString()
		w := new(bytes.Buffer)
		hash.Serialize(w)
		ret[i] = BytesToHexString(w.Bytes())
	}

	return ResponsePack(Success, ret)
}

// A JSON example for submitblock method as following:
//   {"jsonrpc": "2.0", "method": "submitblock", "params": ["raw block in hex"], "id": 0}
func SubmitBlock(param Params) map[string]interface{} {
	str, ok := param.String("block")
	if !ok {
		return ResponsePack(InvalidParams, "")
	}

	hex, _ := HexStringToBytes(str)
	var block Block
	if err := block.Deserialize(bytes.NewReader(hex)); err != nil {
		return ResponsePack(UnknownBlock, "")
	}
	if _, _, err := chain.DefaultLedger.Blockchain.AddBlock(&block); err != nil {
		return ResponsePack(UnknownBlock, "")
	}
	if err := NodeForServers.Relay(nil, &block); err != nil {
		return ResponsePack(InternalError, "")
	}

	return ResponsePack(Success, "")
}

func GetConnectionCount(param Params) map[string]interface{} {
	return ResponsePack(Success, NodeForServers.GetConnectionCnt())
}

//Block
func GetCurrentHeight(param Params) map[string]interface{} {
	return ResponsePack(Success, chain.DefaultLedger.Blockchain.BlockHeight)
}

func GetTransactionPool(param Params) map[string]interface{} {
	txs := make([]*TransactionInfo, 0)
	for _, t := range NodeForServers.GetTxnPool(false) {
		txs = append(txs, GetTransactionInfo(nil, t))
	}
	return ResponsePack(Success, txs)
}

func GetBlockInfo(block *Block, verbose bool) BlockInfo {
	hash := block.Hash()

	var txs []interface{}
	if verbose {
		for _, tx := range block.Transactions {
			txs = append(txs, GetTransactionInfo(&block.Header, tx))
		}
	} else {
		for _, tx := range block.Transactions {
			hash := tx.Hash()
			txs = append(txs, BytesToHexString(hash.Bytes()))
		}
	}
	var versionBytes [4]byte
	binary.BigEndian.PutUint32(versionBytes[:], block.Header.Version)

	var chainWork [4]byte
	binary.BigEndian.PutUint32(chainWork[:], chain.DefaultLedger.Blockchain.GetBestHeight()-block.Header.Height)

	nextBlockHash, _ := chain.DefaultLedger.Store.GetBlockHash(block.Header.Height + 1)

	auxPow := new(bytes.Buffer)
	block.Header.SideAuxPow.Serialize(auxPow)

	return BlockInfo{
		Hash:              hash.String(),
		Confirmations:     chain.DefaultLedger.Blockchain.GetBestHeight() - block.Header.Height + 1,
		StrippedSize:      uint32(block.GetSize()),
		Size:              uint32(block.GetSize()),
		Weight:            uint32(block.GetSize() * 4),
		Height:            block.Header.Height,
		Version:           block.Header.Version,
		VersionHex:        hex.EncodeToString(versionBytes[:]),
		MerkleRoot:        block.Header.MerkleRoot.String(),
		Tx:                txs,
		Time:              block.Header.Timestamp,
		MedianTime:        block.Header.Timestamp,
		Nonce:             block.Header.Nonce,
		Bits:              block.Header.Bits,
		Difficulty:        chain.CalcCurrentDifficulty(block.Header.Bits),
		ChainWork:         hex.EncodeToString(chainWork[:]),
		PreviousBlockHash: block.Header.Previous.String(),
		NextBlockHash:     nextBlockHash.String(),
		AuxPow:            BytesToHexString(auxPow.Bytes()),
	}
}

func getBlock(hash Uint256, format uint32) (interface{}, ErrCode) {
	block, err := chain.DefaultLedger.Store.GetBlock(hash)
	if err != nil {
		return "", UnknownBlock
	}
	switch format {
	case 0:
		w := new(bytes.Buffer)
		block.Serialize(w)
		return BytesToHexString(w.Bytes()), Success
	case 2:
		return GetBlockInfo(block, true), Success
	}
	return GetBlockInfo(block, false), Success
}

func GetBlockByHash(param Params) map[string]interface{} {
	str, ok := param.String("hash")
	if !ok {
		return ResponsePack(InvalidParams, "block hash not found")
	}

	var hash Uint256
	hex, err := HexStringToBytes(str)
	if err != nil {
		return ResponsePack(InvalidParams, "invalid block hash")
	}
	if err := hash.Deserialize(bytes.NewReader(hex)); err != nil {
		ResponsePack(InvalidParams, "invalid block hash")
	}

	format, ok := param.Uint("format")
	if !ok {
		format = 1
	}

	result, error := getBlock(hash, format)

	return ResponsePack(error, result)
}

func SendTransactionInfo(param Params) map[string]interface{} {

	infoStr, ok := param.String("Info")
	if !ok {
		return ResponsePack(InvalidParams, "Info not found")
	}

	inforBytes, err := HexStringToBytes(infoStr)
	if err != nil {
		return ResponsePack(InvalidParams, "")
	}

	var txInfo TransactionInfo
	err = json.Unmarshal(inforBytes, &txInfo)
	if err != nil {
		return ResponsePack(InvalidParams, "")
	}

	txn, err := GetTransaction(&txInfo)
	if err != nil {
		return ResponsePack(InvalidParams, "")
	}
	var hash Uint256
	hash = txn.Hash()
	if errCode := VerifyAndSendTx(txn); errCode != Success {
		return ResponsePack(errCode, "")
	}
	return ResponsePack(Success, hash.String())
}

func SendRawTransaction(param Params) map[string]interface{} {
	str, ok := param.String("Data")
	if !ok {
		return ResponsePack(InvalidParams, "")
	}

	bys, err := HexStringToBytes(str)
	if err != nil {
		return ResponsePack(InvalidParams, "")
	}
	var txn ela.Transaction
	if err := txn.Deserialize(bytes.NewReader(bys)); err != nil {
		return ResponsePack(InvalidTransaction, "")
	}
	var hash Uint256
	hash = txn.Hash()
	if errCode := VerifyAndSendTx(&txn); errCode != Success {
		return ResponsePack(errCode, "")
	}

	return ResponsePack(Success, BytesToHexString(hash.Bytes()))
}

func GetBlockHeight(param Params) map[string]interface{} {
	return ResponsePack(Success, chain.DefaultLedger.Blockchain.BlockHeight)
}

func GetBestBlockHash(param Params) map[string]interface{} {
	bestHeight := chain.DefaultLedger.Blockchain.BlockHeight
	return GetBlockHash(map[string]interface{}{"index": float64(bestHeight)})
}

func GetBlockCount(param Params) map[string]interface{} {
	return ResponsePack(Success, chain.DefaultLedger.Blockchain.BlockHeight+1)
}

func GetBlockHash(param Params) map[string]interface{} {
	height, ok := param.Uint("index")
	if !ok {
		return ResponsePack(InvalidParams, "index parameter should be a positive integer")
	}

	hash, err := chain.DefaultLedger.Store.GetBlockHash(height)
	if err != nil {
		return ResponsePack(InvalidParams, "")
	}
	return ResponsePack(Success, BytesToHexString(hash.Bytes()))
}

func GetBlockTransactions(block *Block) interface{} {
	trans := make([]string, len(block.Transactions))
	for i := 0; i < len(block.Transactions); i++ {
		h := block.Transactions[i].Hash()
		trans[i] = BytesToHexString(h.Bytes())
	}
	hash := block.Hash()
	type BlockTransactions struct {
		Hash         string
		Height       uint32
		Transactions []string
	}
	b := BlockTransactions{
		Hash:         BytesToHexString(hash.Bytes()),
		Height:       block.Header.Height,
		Transactions: trans,
	}
	return b
}

func GetTransactionsByHeight(param Params) map[string]interface{} {
	height, ok := param.Uint("index")
	if !ok {
		return ResponsePack(InvalidParams, "index parameter should be a positive integer")
	}

	hash, err := chain.DefaultLedger.Store.GetBlockHash(uint32(height))
	if err != nil {
		return ResponsePack(UnknownBlock, "")

	}
	block, err := chain.DefaultLedger.Store.GetBlock(hash)
	if err != nil {
		return ResponsePack(UnknownBlock, "")
	}
	return ResponsePack(Success, GetBlockTransactions(block))
}

func GetBlockByHeight(param Params) map[string]interface{} {
	height, ok := param.Uint("height")
	if !ok {
		return ResponsePack(InvalidParams, "height parameter should be a positive integer")
	}

	hash, err := chain.DefaultLedger.Store.GetBlockHash(uint32(height))
	if err != nil {
		return ResponsePack(UnknownBlock, "")
	}

	result, errCode := getBlock(hash, 2)

	return ResponsePack(errCode, result)
}

//Asset
func GetAssetByHash(param Params) map[string]interface{} {
	str, ok := param.String("hash")
	if !ok {
		return ResponsePack(InvalidParams, "")
	}
	hex, err := HexStringToBytes(str)
	if err != nil {
		return ResponsePack(InvalidParams, "")
	}
	var hash Uint256
	err = hash.Deserialize(bytes.NewReader(hex))
	if err != nil {
		return ResponsePack(InvalidAsset, "")
	}
	asset, err := chain.DefaultLedger.Store.GetAsset(hash)
	if err != nil {
		return ResponsePack(UnknownAsset, "")
	}
	if false {
		w := new(bytes.Buffer)
		asset.Serialize(w)
		return ResponsePack(Success, BytesToHexString(w.Bytes()))
	}
	return ResponsePack(Success, asset)
}

func GetBalanceByAddr(param Params) map[string]interface{} {
	str, ok := param.String("addr")
	if !ok {
		return ResponsePack(InvalidParams, "")
	}

	programHash, err := Uint168FromAddress(str)
	if err != nil {
		return ResponsePack(InvalidParams, "")
	}
	unspends, err := chain.DefaultLedger.Store.GetUnspentsFromProgramHash(*programHash)
	var balance Fixed64 = 0
	for _, u := range unspends {
		for _, v := range u {
			balance = balance + v.Value
		}
	}
	return ResponsePack(Success, balance.String())
}

func GetBalanceByAsset(param Params) map[string]interface{} {
	addr, ok := param.String("addr")
	if !ok {
		return ResponsePack(InvalidParams, "")
	}

	programHash, err := Uint168FromAddress(addr)
	if err != nil {
		return ResponsePack(InvalidParams, "")
	}

	assetId, ok := param.String("assetid")
	if !ok {
		return ResponsePack(InvalidParams, "")
	}

	unspents, err := chain.DefaultLedger.Store.GetUnspentsFromProgramHash(*programHash)
	var balance Fixed64 = 0
	for k, u := range unspents {
		assid := BytesToHexString(k.Bytes())
		for _, v := range u {
			if assetId == assid {
				balance = balance + v.Value
			}
		}
	}
	return ResponsePack(Success, balance.String())
}

func GetUnspends(param Params) map[string]interface{} {
	addr, ok := param.String("addr")
	if !ok {
		return ResponsePack(InvalidParams, "")
	}

	programHash, err := Uint168FromAddress(addr)
	if err != nil {
		return ResponsePack(InvalidParams, "")
	}
	type UTXOUnspentInfo struct {
		Txid  string
		Index uint32
		Value string
	}
	type Result struct {
		AssetId   string
		AssetName string
		Utxo      []UTXOUnspentInfo
	}
	var results []Result
	unspends, err := chain.DefaultLedger.Store.GetUnspentsFromProgramHash(*programHash)

	for k, u := range unspends {
		assetid := BytesToHexString(k.Bytes())
		asset, err := chain.DefaultLedger.Store.GetAsset(k)
		if err != nil {
			return ResponsePack(InternalError, "")
		}
		var unspendsInfo []UTXOUnspentInfo
		for _, v := range u {
			unspendsInfo = append(unspendsInfo, UTXOUnspentInfo{BytesToHexString(v.TxId.Bytes()), v.Index, v.Value.String()})
		}
		results = append(results, Result{assetid, asset.Name, unspendsInfo})
	}
	return ResponsePack(Success, results)
}

func GetUnspendOutput(param Params) map[string]interface{} {
	addr, ok := param.String("addr")
	if !ok {
		return ResponsePack(InvalidParams, "")
	}
	programHash, err := Uint168FromAddress(addr)
	if err != nil {
		return ResponsePack(InvalidParams, "")
	}
	assetId, ok := param.String("assetid")
	if !ok {
		return ResponsePack(InvalidParams, "")
	}
	bys, err := HexStringToBytes(assetId)
	if err != nil {
		return ResponsePack(InvalidParams, "")
	}

	var assetHash Uint256
	if err := assetHash.Deserialize(bytes.NewReader(bys)); err != nil {
		return ResponsePack(InvalidParams, "")
	}
	type UTXOUnspentInfo struct {
		Txid  string
		Index uint32
		Value string
	}
	infos, err := chain.DefaultLedger.Store.GetUnspentFromProgramHash(*programHash, assetHash)
	if err != nil {
		return ResponsePack(InvalidParams, "")

	}
	var UTXOoutputs []UTXOUnspentInfo
	for _, v := range infos {
		UTXOoutputs = append(UTXOoutputs, UTXOUnspentInfo{Txid: BytesToHexString(v.TxId.Bytes()), Index: v.Index, Value: v.Value.String()})
	}
	return ResponsePack(Success, UTXOoutputs)
}

//Transaction
func GetTransactionByHash(param Params) map[string]interface{} {
	str, ok := param.String("hash")
	if !ok {
		return ResponsePack(InvalidParams, "")
	}

	bys, err := HexStringToBytes(str)
	if err != nil {
		return ResponsePack(InvalidParams, "")
	}

	var hash Uint256
	err = hash.Deserialize(bytes.NewReader(bys))
	if err != nil {
		return ResponsePack(InvalidTransaction, "")
	}
	txn, height, err := chain.DefaultLedger.Store.GetTransaction(hash)
	if err != nil {
		return ResponsePack(UnknownTransaction, "")
	}
	if false {
		w := new(bytes.Buffer)
		txn.Serialize(w)
		return ResponsePack(Success, BytesToHexString(w.Bytes()))
	}
	bHash, err := chain.DefaultLedger.Store.GetBlockHash(height)
	if err != nil {
		return ResponsePack(UnknownBlock, "")
	}
	header, err := chain.DefaultLedger.Store.GetHeader(bHash)
	if err != nil {
		return ResponsePack(UnknownBlock, "")
	}

	return ResponsePack(Success, GetTransactionInfo(header, txn))
}

func getPayloadInfo(p ela.Payload) PayloadInfo {
	switch object := p.(type) {
	case *ela.PayloadCoinBase:
		obj := new(CoinbaseInfo)
		obj.CoinbaseData = string(object.CoinbaseData)
		return obj
	case *ela.PayloadRegisterAsset:
		obj := new(RegisterAssetInfo)
		obj.Asset = object.Asset
		obj.Amount = object.Amount.String()
		obj.Controller = BytesToHexString(BytesReverse(object.Controller.Bytes()))
		return obj
	case *ela.PayloadSideMining:
		obj := new(SideMiningInfo)
		obj.SideBlockHash = object.SideBlockHash.String()
		return obj
	case *ela.PayloadWithdrawAsset:
		obj := new(WithdrawAssetInfo)
		obj.BlockHeight = object.BlockHeight
		return obj
	case *ela.PayloadTransferCrossChainAsset:
		obj := new(TransferCrossChainAssetInfo)
		obj.AddressesMap = object.AddressesMap
		return obj
	case *ela.PayloadTransferAsset:
	case *ela.PayloadRecord:
	}
	return nil
}

func VerifyAndSendTx(txn *ela.Transaction) ErrCode {
	// if transaction is verified unsucessfully then will not put it into transaction pool
	if errCode := NodeForServers.AppendToTxnPool(txn); errCode != Success {
		log.Warn("Can NOT add the transaction to TxnPool")
		log.Info("[httpjsonrpc] VerifyTransaction failed when AppendToTxnPool.")
		return errCode
	}
	if err := NodeForServers.Relay(nil, txn); err != nil {
		log.Error("Xmit Tx Error:Relay transaction failed.", err)
		return ErrXmitFail
	}
	return Success
}

func ResponsePack(errCode ErrCode, result interface{}) map[string]interface{} {
	if errCode != 0 && (result == "" || result == nil) {
		result = ErrMap[errCode]
	}
	return map[string]interface{}{"Result": result, "Error": errCode}
}

func GetBlockTransactionsDetail(block *Block, filter func(*ela.Transaction) bool) interface{} {
	var trans []*TransactionInfo
	for i := 0; i < len(block.Transactions); i++ {
		if filter(block.Transactions[i]) {
			continue
		}

		trans = append(trans, GetTransactionInfo(&block.Header, block.Transactions[i]))
	}
	hash := block.Hash()
	type BlockTransactions struct {
		Hash         string
		Height       uint32
		Transactions []*TransactionInfo
	}
	b := BlockTransactions{
		Hash:         hash.String(),
		Height:       block.Height,
		Transactions: trans,
	}
	return b
}

func GetDestroyedTransactionsByHeight(param Params) map[string]interface{} {
	height, ok := param.Uint("height")
	if !ok {
		return ResponsePack(InvalidParams, "height parameter should be a positive integer")
	}

	hash, err := chain.DefaultLedger.Store.GetBlockHash(uint32(height))
	if err != nil {
		return ResponsePack(UnknownBlock, "")

	}
	block, err := chain.DefaultLedger.Store.GetBlock(hash)
	if err != nil {
		return ResponsePack(UnknownBlock, "")
	}

	destroyHash, err := Uint168FromAddress(config.Parameters.DestroyAddr)
	return ResponsePack(Success, GetBlockTransactionsDetail(block, func(tran *ela.Transaction) bool {
		for _, output := range tran.Outputs {
			if output.ProgramHash == *destroyHash {
				return false
			}
		}
		return true
	}))
}
