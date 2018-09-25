package blockchain

import (
	"errors"
	"math"
	"math/big"
	"time"

	"github.com/elastos/Elastos.ELA.SideChain/config"
	. "github.com/elastos/Elastos.ELA.SideChain/core"
	. "github.com/elastos/Elastos.ELA.SideChain/errors"

	. "github.com/elastos/Elastos.ELA.Utility/common"
	"github.com/elastos/Elastos.ELA.Utility/crypto"
)

const (
	PowCheckBlockSanity        = "powcheckblocksanity"
	PowCheckHeader             = "powcheckheader"
	PowCheckTransactionsCount  = "powchecktransactionscount"
	PowCheckBlockSize          = "powcheckblocksize"
	PowCheckTransactionsFee    = "powchecktransactionsfee"
	PowCheckTransactionsMerkle = "powchecktransactionsmerkle"
	PowCheckBlockContext       = "powcheckblockcontext"
	CheckProofOfWork           = "checkproofofwork"
	IsFinalizedTransaction     = "isfinalizedtransaction"

	MaxTimeOffsetSeconds = 2 * 60 * 60
)

var BlockValidator *BlockValidate

type BlockValidateType string

type BlockValidateParameter struct {
	Block       *Block
	PowLimit    *big.Int
	TimeSource  MedianTimeSource
	PrevNode    *BlockNode
	Header      *Header
	MsgTx       *Transaction
	BlockHeight uint32
}

type BlockValidate struct {
	CheckFunctions       map[BlockValidateType]func(param *BlockValidateParameter) error
	CheckSanityFunctions map[BlockValidateType]func(param *BlockValidateParameter) error
}

func InitBlockValidator() {
	BlockValidator = &BlockValidate{}
	BlockValidator.Init()
}

func (v *BlockValidate) Init() {
	v.CheckFunctions[PowCheckBlockContext] = v.PowCheckBlockContext
	v.CheckFunctions[CheckProofOfWork] = v.CheckProofOfWork
	v.CheckFunctions[IsFinalizedTransaction] = v.CheckFinalizedTransaction

	v.CheckSanityFunctions[PowCheckBlockSanity] = v.PowCheckBlockSanity
	v.CheckSanityFunctions[PowCheckHeader] = v.PowCheckHeader
	v.CheckSanityFunctions[PowCheckTransactionsCount] = v.PowCheckTransactionsCount
	v.CheckSanityFunctions[PowCheckBlockSize] = v.PowCheckBlockSize
	v.CheckSanityFunctions[PowCheckTransactionsFee] = v.PowCheckTransactionsFee
	v.CheckSanityFunctions[PowCheckTransactionsMerkle] = v.PowCheckTransactionsMerkle
}

//block *Block, powLimit *big.Int, timeSource MedianTimeSource
func (v *BlockValidate) PowCheckBlockSanity(param *BlockValidateParameter) error {
	for _, checkFunc := range v.CheckSanityFunctions {
		if err := checkFunc(param); err != nil {
			return errors.New("[PowCheckBlockSanity] error:" + err.Error())
		}
	}
	return nil
}

//block *Block, powLimit *big.Int, timeSource MedianTimeSource
func (v *BlockValidate) PowCheckHeader(param *BlockValidateParameter) error {
	header := param.Block.Header

	// A block's main chain block header must contain in spv module
	//mainChainBlockHash := header.SideAuxPow.MainBlockHeader.Hash()
	//if err := spv.VerifyElaHeader(&mainChainBlockHash); err != nil {
	//	return err
	//}

	if !header.SideAuxPow.SideAuxPowCheck(header.Hash()) {
		return errors.New("[PowCheckHeader] block check proof is failed")
	}
	if v.CheckFunctions[CheckProofOfWork](param) != nil {
		return errors.New("[PowCheckHeader] block check proof is failed.")
	}

	// A block timestamp must not have a greater precision than one second.
	tempTime := time.Unix(int64(header.Timestamp), 0)
	if !tempTime.Equal(time.Unix(tempTime.Unix(), 0)) {
		return errors.New("[PowCheckHeader] block timestamp of has a higher precision than one second")
	}

	// Ensure the block time is not too far in the future.
	maxTimestamp := param.TimeSource.AdjustedTime().Add(time.Second * MaxTimeOffsetSeconds)
	if tempTime.After(maxTimestamp) {
		return errors.New("[PowCheckHeader] block timestamp of is too far in the future")
	}

	return nil
}

//block *Block
func (v *BlockValidate) PowCheckTransactionsCount(param *BlockValidateParameter) error {
	// A block must have at least one transaction.
	numTx := len(param.Block.Transactions)
	if numTx == 0 {
		return errors.New("[PowCheckTransactionsCount]  block does not contain any transactions")
	}

	// A block must not have more transactions than the max block payload.
	if numTx > config.Parameters.MaxTxInBlock {
		return errors.New("[PowCheckTransactionsCount]  block contains too many transactions")
	}

	return nil
}

//block *Block
func (v *BlockValidate) PowCheckBlockSize(param *BlockValidateParameter) error {
	// A block must not exceed the maximum allowed block payload when serialized.
	blockSize := param.Block.GetSize()
	if blockSize > config.Parameters.MaxBlockSize {
		return errors.New("[PowCheckBlockSize] serialized block is too big")
	}

	return nil
}

//block *Block
func (v *BlockValidate) PowCheckTransactionsFee(param *BlockValidateParameter) error {
	transactions := param.Block.Transactions
	var rewardInCoinbase = Fixed64(0)
	var totalTxFee = Fixed64(0)
	for index, tx := range transactions {
		// The first transaction in a block must be a coinbase.
		if index == 0 {
			if !tx.IsCoinBaseTx() {
				return errors.New("[PowCheckTransactionsFee] first transaction in block is not a coinbase")
			}
			// Calculate reward in coinbase
			for _, output := range tx.Outputs {
				rewardInCoinbase += output.Value
			}
			continue
		}

		// A block must not have more than one coinbase.
		if tx.IsCoinBaseTx() {
			return errors.New("[PowCheckTransactionsFee] block contains second coinbase")
		}

		// Calculate transaction fee
		totalTxFee += TxFeeHelper.GetTxFee(tx, DefaultChain.AssetID)
	}

	// Reward in coinbase must match total transaction fee
	if rewardInCoinbase != totalTxFee {
		return errors.New("[PowCheckTransactionsFee] reward amount in coinbase not correct")
	}

	return nil
}

//block *Block
func (v *BlockValidate) PowCheckTransactionsMerkle(param *BlockValidateParameter) error {
	txIds := make([]Uint256, 0, len(param.Block.Transactions))
	existingTxIds := make(map[Uint256]struct{})
	existingTxInputs := make(map[string]struct{})
	existingMainTxs := make(map[Uint256]struct{})
	for _, txn := range param.Block.Transactions {
		txId := txn.Hash()
		// Check for duplicate transactions.
		if _, exists := existingTxIds[txId]; exists {
			return errors.New("[PowCheckTransactionsMerkle] block contains duplicate transaction")
		}
		existingTxIds[txId] = struct{}{}

		// Check for transaction sanity
		if errCode := TransactionValidator.CheckTransactionSanity(txn); errCode != Success {
			return errors.New("[PowCheckTransactionsMerkle] failed when verifiy block")
		}

		// Check for duplicate UTXO inputs in a block
		for _, input := range txn.Inputs {
			referKey := input.ReferKey()
			if _, exists := existingTxInputs[referKey]; exists {
				return errors.New("[PowCheckTransactionsMerkle] block contains duplicate UTXO")
			}
			existingTxInputs[referKey] = struct{}{}
		}

		if txn.IsRechargeToSideChainTx() {
			rechargePayload := txn.Payload.(*PayloadRechargeToSideChain)
			// Check for duplicate mainchain tx in a block
			hash, err := rechargePayload.GetMainchainTxHash()
			if err != nil {
				return err
			}
			if _, exists := existingMainTxs[*hash]; exists {
				return errors.New("[PowCheckTransactionsMerkle] block contains duplicate mainchain Tx")
			}
			existingMainTxs[*hash] = struct{}{}
		}

		// Append transaction to list
		txIds = append(txIds, txId)
	}
	calcTransactionsRoot, err := crypto.ComputeRoot(txIds)
	if err != nil {
		return errors.New("[PowCheckTransactionsMerkle] merkleTree compute failed")
	}
	if !param.Block.Header.MerkleRoot.IsEqual(calcTransactionsRoot) {
		return errors.New("[PowCheckTransactionsMerkle] block merkle root is invalid")
	}

	return nil
}

//block *Block, prevNode *BlockNode
func (v *BlockValidate) PowCheckBlockContext(param *BlockValidateParameter) error {
	// The genesis block is valid by definition.
	if param.PrevNode == nil {
		return nil
	}

	header := param.Block.Header
	expectedDifficulty, err := CalcNextRequiredDifficulty(param.PrevNode,
		time.Unix(int64(header.Timestamp), 0))
	if err != nil {
		return err
	}

	if header.Bits != expectedDifficulty {
		return errors.New("[PowCheckBlockContext] block difficulty is not the expected")
	}

	// Ensure the timestamp for the block header is after the
	// median time of the last several blocks (medianTimeBlocks).
	medianTime := CalcPastMedianTime(param.PrevNode)
	tempTime := time.Unix(int64(header.Timestamp), 0)

	if !tempTime.After(medianTime) {
		return errors.New("[PowCheckBlockContext] block timestamp is not after expected")
	}

	// The height of this block is one more than the referenced
	// previous block.
	blockHeight := param.PrevNode.Height + 1

	// Ensure all transactions in the block are finalized.
	for _, txn := range param.Block.Transactions[1:] {
		if err := v.CheckFunctions[IsFinalizedTransaction](&BlockValidateParameter{MsgTx: txn, BlockHeight: blockHeight}); err != nil {
			return errors.New("[PowCheckBlockContext] block contains unfinalized transaction")
		}
	}

	return nil
}

//header *Header, powLimit *big.Int
func (v *BlockValidate) CheckProofOfWork(param *BlockValidateParameter) error {
	// The target difficulty must be larger than zero.
	target := CompactToBig(param.Header.Bits)
	if target.Sign() <= 0 {
		return errors.New("[CheckProofOfWork], block target difficulty is too low.")
	}

	// The target difficulty must be less than the maximum allowed.
	if target.Cmp(param.PowLimit) > 0 {
		return errors.New("[CheckProofOfWork], block target difficulty is higher than max of limit.")
	}

	// The block hash must be less than the claimed target.
	hash := param.Header.SideAuxPow.MainBlockHeader.AuxPow.ParBlockHeader.Hash()

	hashNum := HashToBig(&hash)
	if hashNum.Cmp(target) > 0 {
		return errors.New("[CheckProofOfWork], block target difficulty is higher than expected difficulty.")
	}

	return nil
}

//msgTx *Transaction, blockHeight uint32
func (v *BlockValidate) CheckFinalizedTransaction(param *BlockValidateParameter) error {
	// Lock time of zero means the transaction is finalized.
	lockTime := param.MsgTx.LockTime
	if lockTime == 0 {
		return nil
	}

	//FIXME only height
	if lockTime < param.BlockHeight {
		return nil
	}

	// At this point, the transaction's lock time hasn't occurred yet, but
	// the transaction might still be finalized if the sequence number
	// for all transaction inputs is maxed out.
	for _, txIn := range param.MsgTx.Inputs {
		if txIn.Sequence != math.MaxUint16 {
			return errors.New("[CheckFinalizedTransaction] lock time check failed")
		}
	}
	return nil
}
