package servers

import (
	. "Elastos.ELA.SideChain/common"
	"Elastos.ELA.SideChain/common/log"
	"Elastos.ELA.SideChain/consensus/pow"
	"Elastos.ELA.SideChain/core/asset"
	. "Elastos.ELA.SideChain/core/transaction"
	tx "Elastos.ELA.SideChain/core/transaction"
	"Elastos.ELA.SideChain/core/transaction/payload"
	. "Elastos.ELA.SideChain/errors"
	. "Elastos.ELA.SideChain/net/protocol"
	"io"
)

const TlsPort = 443

var NodeForServers Noder
var Pow *pow.PowService

type TxAttributeInfo struct {
	Usage TransactionAttributeUsage
	Data  string
}

type UTXOTxInputInfo struct {
	ReferTxID          string
	ReferTxOutputIndex uint16
	Sequence           uint32
	Address            string
	Value              string
}

type BalanceTxInputInfo struct {
	AssetID     string
	Value       Fixed64
	ProgramHash string
}

type TxoutputInfo struct {
	AssetID    string
	Value      string
	Address    string
	OutputLock uint32
}

type TxoutputMap struct {
	Key   Uint256
	Txout []TxoutputInfo
}

type AmountMap struct {
	Key   Uint256
	Value Fixed64
}

type ProgramInfo struct {
	Code      string
	Parameter string
}

type Transactions struct {
	TxType         TransactionType
	PayloadVersion byte
	Payload        PayloadInfo
	Attributes     []TxAttributeInfo
	UTXOInputs     []UTXOTxInputInfo
	BalanceInputs  []BalanceTxInputInfo
	Outputs        []TxoutputInfo
	LockTime       uint32
	Programs       []ProgramInfo

	AssetOutputs      []TxoutputMap
	AssetInputAmount  []AmountMap
	AssetOutputAmount []AmountMap
	Timestamp         uint32 `json:",omitempty"`
	Confirmations     uint32 `json:",omitempty"`
	TxSize            uint32 `json:",omitempty"`
	Hash              string
}

type AuxInfo struct {
	Version    int32
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

type PayloadInfo interface {
	Data(version byte) string
	Serialize(w io.Writer, version byte) error
	Deserialize(r io.Reader, version byte) error
}

type CoinbaseInfo struct {
	CoinbaseData string
}

type RegisterAssetInfo struct {
	Asset      *asset.Asset
	Amount     string
	Controller string
}

type TransferAssetInfo struct {
}

type IssueTokenInfo struct {
}

func TransPayloadToHex(p Payload) PayloadInfo {
	switch object := p.(type) {
	case *payload.CoinBase:
		obj := new(CoinbaseInfo)
		obj.CoinbaseData = string(object.CoinbaseData)
		return obj
	case *payload.RegisterAsset:
		obj := new(RegisterAssetInfo)
		obj.Asset = object.Asset
		obj.Amount = object.Amount.String()
		obj.Controller = BytesToHexString(object.Controller.ToArrayReverse())
		return obj
	case *payload.TransferAsset:
	case *payload.Record:
	case *payload.DeployCode:
	}
	return nil
}

func VerifyAndSendTx(txn *tx.Transaction) ErrCode {
	// if transaction is verified unsucessfully then will not put it into transaction pool
	if errCode := NodeForServers.AppendToTxnPool(txn); errCode != Success {
		log.Warn("Can NOT add the transaction to TxnPool")
		log.Info("[httpjsonrpc] VerifyTransaction failed when AppendToTxnPool.")
		return errCode
	}
	if err := NodeForServers.Xmit(txn); err != nil {
		log.Error("Xmit Tx Error:Xmit transaction failed.", err)
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
