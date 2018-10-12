package blockchain

import (
	"github.com/elastos/Elastos.ELA.SideChain/core"
	"github.com/elastos/Elastos.ELA.SideChain/smartcontract/states"
	. "github.com/elastos/Elastos.ELA.Utility/common"
	"github.com/elastos/Elastos.ELA.SideChain/store"
)

// IChainStore provides func with store package.
type IChainStore interface {
	store.IStore
	InitWithGenesisBlock(genesisblock *core.Block) (uint32, error)

	SaveBlock(b *core.Block) error
	GetBlock(hash Uint256) (*core.Block, error)
	GetBlockHash(height uint32) (Uint256, error)
	IsDoubleSpend(tx *core.Transaction) bool

	GetHeader(hash Uint256) (*core.Header, error)

	RollbackBlock(hash Uint256) error

	GetTransaction(txId Uint256) (*core.Transaction, uint32, error)
	GetTxReference(tx *core.Transaction) (map[*core.Input]*core.Output, error)

	PersistAsset(assetid Uint256, asset core.Asset) error
	GetAsset(hash Uint256) (*core.Asset, error)

	PersistMainchainTx(mainchainTxHash Uint256)
	GetMainchainTx(mainchainTxHash Uint256) (byte, error)

	GetCurrentBlockHash() Uint256
	GetHeight() uint32

	RemoveHeaderListElement(hash Uint256)

	GetUnspents(txid Uint256) ([]*core.Output, error)
	GetUnspent(txid Uint256, index uint16) (*core.Output, error)
	ContainsUnspent(txid Uint256, index uint16) (bool, error)
	GetUnspentFromProgramHash(programHash Uint168, assetid Uint256) ([]*UTXO, error)
	GetUnspentsFromProgramHash(programHash Uint168) (map[Uint256][]*UTXO, error)
	GetAssets() map[Uint256]*core.Asset
	GetAccount(programHash Uint168) (*states.AccountState, error)

	IsTxHashDuplicate(txhash Uint256) bool
	IsMainchainTxHashDuplicate(mainchainTxHash Uint256) bool
	IsBlockInStore(hash Uint256) bool
}
