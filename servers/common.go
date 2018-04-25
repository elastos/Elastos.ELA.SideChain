package servers

import (
	"github.com/elastos/Elastos.ELA.SideChain/blockchain"
	"github.com/elastos/Elastos.ELA.SideChain/log"
	. "github.com/elastos/Elastos.ELA.SideChain/errors"
	. "github.com/elastos/Elastos.ELA.SideChain/protocol"

	ela "github.com/elastos/Elastos.ELA/core"
	. "github.com/elastos/Elastos.ELA.Utility/common"
)

const TlsPort = 443

var NodeForServers Noder
var Pow *blockchain.PowService

type TxAttributeInfo struct {
	Usage ela.AttributeUsage
	Data  string
}

type UTXOTxInputInfo struct {
	ReferTxID          string
	ReferTxOutputIndex uint16
	Sequence           uint32
	Address            string
	Value              string
}

type TxoutputInfo struct {
	AssetID    string
	Value      string
	Address    string
	OutputLock uint32
}

type ProgramInfo struct {
	Code      string
	Parameter string
}

type Transactions struct {
	TxType         ela.TransactionType
	PayloadVersion byte
	Payload        PayloadInfo
	Attributes     []TxAttributeInfo
	UTXOInputs     []UTXOTxInputInfo
	Outputs        []TxoutputInfo
	LockTime       uint32
	Programs       []ProgramInfo

	Timestamp     uint32 `json:",omitempty"`
	Confirmations uint32 `json:",omitempty"`
	TxSize        uint32 `json:",omitempty"`
	Hash          string
}

type AuxInfo struct {
	Version    uint32
	PrevBlock  string
	MerkleRoot string
	Timestamp  uint32
	Bits       uint32
	Nonce      uint32
}

type BlockHead struct {
	Version          uint32
	PrevBlockHash    string
	TransactionsRoot string
	Timestamp        uint32
	Bits             uint32
	Height           uint32
	Nonce            uint32
	AuxPow           *AuxInfo
	Difficulty       string
	BlockSize        int
	Hash             string
}

type BlockInfo struct {
	Hash          string
	BlockData     *BlockHead
	Transactions  []*Transactions
	Confirmations uint32
	MinerInfo     string
}

type NodeInfo struct {
	State    uint   // NodeForServers status
	Port     uint16 // The nodes's port
	ID       uint64 // The nodes's id
	Time     int64
	Version  uint32 // The network protocol the NodeForServers used
	Services uint64 // The services the NodeForServers supplied
	Relay    bool   // The relay capability of the NodeForServers (merge into capbility flag)
	Height   uint64 // The NodeForServers latest block height
	TxnCnt   uint64 // The transactions be transmit by this NodeForServers
	RxTxnCnt uint64 // The transaction received by this NodeForServers
}

type ArbitratorGroupInfo struct {
	OnDutyArbitratorIndex int
	Arbitrators           []string
}

type PayloadInfo interface{}

type CoinbaseInfo struct {
	CoinbaseData string
}

type RegisterAssetInfo struct {
	Asset      ela.Asset
	Amount     string
	Controller string
}

type SideMiningInfo struct {
	SideBlockHash string
}

type TransferCrossChainAssetInfo struct {
	AddressesMap map[string]uint64
}

type WithdrawAssetInfo struct {
	BlockHeight uint32
}

func TransPayloadToHex(p ela.Payload) PayloadInfo {
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
