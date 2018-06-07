package blockchain

import (
	"bytes"
	"errors"
	"sort"

	"github.com/elastos/Elastos.ELA.SideChain/mainchain"
	"github.com/elastos/Elastos.ELA.SideChain/spv"

	. "github.com/elastos/Elastos.ELA.Utility/common"
	"github.com/elastos/Elastos.ELA.Utility/crypto"
	ela "github.com/elastos/Elastos.ELA/core"
)

func VerifySignature(tx *ela.Transaction) (bool, error) {
	if tx.TxType == ela.RechargeToSideChain {
		if err := spv.VerifyTransaction(tx); err != nil {
			return false, errors.New("Issue token transaction validate failed.")
		}
		return true, nil
	}

	hashes, err := GetTxProgramHashes(tx)
	if err != nil {
		return false, err
	}

	programs := tx.Programs
	Length := len(hashes)
	if Length != len(programs) {
		return false, errors.New("The number of data hashes is different with number of programs.")
	}

	buf := new(bytes.Buffer)
	tx.SerializeUnsigned(buf)
	data := buf.Bytes()

	for i := 0; i < len(programs); i++ {

		code := programs[i].Code
		param := programs[i].Parameter

		programHash, err := crypto.ToProgramHash(code)
		if err != nil {
			return false, err
		}

		if !hashes[i].IsEqual(*programHash) {
			return false, errors.New("The data hashes is different with corresponding program code.")
		}
		// Get transaction type
		signType, err := crypto.GetScriptType(code)
		if err != nil {
			return false, err
		}
		if signType == STANDARD {
			// Remove length byte and sign type byte
			publicKeyBytes := code[1 : len(code)-1]

			return checkStandardSignature(publicKeyBytes, data, param)

		} else if signType == MULTISIG {
			publicKeys, err := crypto.ParseMultisigScript(code)
			if err != nil {
				return false, err
			}
			return checkMultiSignSignatures(code, param, data, publicKeys)

		} else {
			return false, errors.New("unknown signature type")
		}
	}

	return true, nil
}

func GetTxProgramHashes(tx *ela.Transaction) ([]Uint168, error) {
	if tx == nil {
		return nil, errors.New("[Transaction],GetProgramHashes transaction is nil.")
	}
	hashes := make([]Uint168, 0)
	uniqueHashes := make([]Uint168, 0)
	// add inputUTXO's transaction
	references, err := DefaultLedger.Store.GetTxReference(tx)
	if err != nil {
		return nil, errors.New("[Transaction], GetProgramHashes failed.")
	}
	for _, output := range references {
		programHash := output.ProgramHash
		hashes = append(hashes, programHash)
	}
	for _, attribute := range tx.Attributes {
		if attribute.Usage == ela.Script {
			dataHash, err := Uint168FromBytes(attribute.Data)
			if err != nil {
				return nil, errors.New("[Transaction], GetProgramHashes err.")
			}
			hashes = append(hashes, *dataHash)
		}
	}

	//remove dupilicated hashes
	uniq := make(map[Uint168]bool)
	for _, v := range hashes {
		uniq[v] = true
	}
	for k := range uniq {
		uniqueHashes = append(uniqueHashes, k)
	}
	sort.Sort(byProgramHashes(uniqueHashes))
	return uniqueHashes, nil
}

func checkStandardSignature(publicKeyBytes, content, signature []byte) (bool, error) {
	if len(signature) != crypto.SignatureScriptLength {
		return false, errors.New("Invalid signature length")
	}

	publicKey, err := crypto.DecodePoint(publicKeyBytes)
	if err != nil {
		return false, err
	}
	err = crypto.Verify(*publicKey, content, signature[1:])
	if err == nil {
		return false, err
	}
	return true, nil
}

func checkMultiSignSignatures(code, param, content []byte, publicKeys [][]byte) (bool, error) {
	// Get N parameter
	n := int(code[len(code)-2]) - crypto.PUSH1 + 1
	// Get M parameter
	m := int(code[0]) - crypto.PUSH1 + 1
	if m < 1 || m > n {
		return false, errors.New("invalid multi sign script code")
	}
	if len(publicKeys) != n {
		return false, errors.New("invalid multi sign public key script count")
	}

	signatureCount := 0
	for i := 0; i < len(param); i += crypto.SignatureScriptLength {
		// Remove length byte
		sign := param[i : i+crypto.SignatureScriptLength][1:]
		// Get signature index, if signature exists index will not be -1
		index := -1
		for i, publicKey := range publicKeys {
			pubKey, err := crypto.DecodePoint(publicKey[1:])
			if err != nil {
				return false, err
			}
			err = crypto.Verify(*pubKey, content, sign)
			if err == nil {
				index = i
			}
		}
		if index != -1 {
			signatureCount++
		}
	}
	// Check signature count
	if signatureCount != m {
		return false, errors.New("invalid signature count")
	}

	return true, nil
}

func checkCrossChainTransaction(txn *ela.Transaction) error {
	if !txn.IsRechargeToSideChainTx() {
		return nil
	}

	depositPayload, ok := txn.Payload.(*ela.PayloadRechargeToSideChain)
	if !ok {
		return errors.New("Invalid payload type.")
	}

	if mainchain.DbCache == nil {
		dbCache, err := mainchain.OpenDataStore()
		if err != nil {
			errors.New("Open data store failed")
		}
		mainchain.DbCache = dbCache
	}

	mainChainTransaction := new(ela.Transaction)
	reader := bytes.NewReader(depositPayload.MainChainTransaction)
	if err := mainChainTransaction.Deserialize(reader); err != nil {
		return errors.New("PayloadRechargeToSideChain mainChainTransaction deserialize failed")
	}

	ok, err := mainchain.DbCache.HasMainChainTx(mainChainTransaction.Hash().String())
	if err != nil {
		return err
	}
	if ok {
		return errors.New("Reduplicate withdraw transaction.")
	}
	err = mainchain.DbCache.AddMainChainTx(mainChainTransaction.Hash().String())
	if err != nil {
		return err
	}
	return nil
}

type byProgramHashes []Uint168

func (a byProgramHashes) Len() int      { return len(a) }
func (a byProgramHashes) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a byProgramHashes) Less(i, j int) bool {
	if a[i].Compare(a[j]) > 0 {
		return false
	} else {
		return true
	}
}
