package MainChainStore

import (
	"database/sql"
	"os"
	"sync"

	"github.com/elastos/Elastos.ELA/common/log"
	_ "github.com/mattn/go-sqlite3"
)

const (
	DriverName      = "sqlite3"
	DBName          = "./mainChainCache.db"
	QueryHeightCode = 0
)

const (
	CreateMainChainTxsTable = `CREATE TABLE IF NOT EXISTS MainChainTxs (
				Id INTEGER NOT NULL PRIMARY KEY,
				TransactionHash VARCHAR
			);`
)

var (
	DbCache DataStore
)

type DataStore interface {
	AddSideChainTx(transactionHash string) error
	HashSideChainTx(transactionHash string) (bool, error)

	ResetDataStore() error
}

type DataStoreImpl struct {
	sideMux *sync.Mutex

	*sql.DB
}

func OpenDataStore() (DataStore, error) {
	db, err := initDB()
	if err != nil {
		return nil, err
	}
	dataStore := &DataStoreImpl{DB: db, sideMux: new(sync.Mutex)}

	// Handle system interrupt signals
	dataStore.catchSystemSignals()

	return dataStore, nil
}

func initDB() (*sql.DB, error) {
	db, err := sql.Open(DriverName, DBName)
	if err != nil {
		log.Error("Open data db error:", err)
		return nil, err
	}
	// Create MainChainTxs table
	_, err = db.Exec(CreateMainChainTxsTable)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func (store *DataStoreImpl) catchSystemSignals() {
	HandleSignal(func() {
		store.sideMux.Lock()
		store.Close()
		os.Exit(-1)
	})
}

func (store *DataStoreImpl) ResetDataStore() error {

	store.DB.Close()
	os.Remove(DBName)

	var err error
	store.DB, err = initDB()
	if err != nil {
		return err
	}

	return nil
}

func (store *DataStoreImpl) AddSideChainTx(transactionHash string) error {
	store.sideMux.Lock()
	defer store.sideMux.Unlock()

	// Prepare sql statement
	stmt, err := store.Prepare("INSERT INTO MainChainTxs(TransactionHash) values(?)")
	if err != nil {
		return err
	}
	// Do insert
	_, err = stmt.Exec(transactionHash)
	if err != nil {
		return err
	}
	return nil
}

func (store *DataStoreImpl) HashSideChainTx(transactionHash string) (bool, error) {
	store.sideMux.Lock()
	defer store.sideMux.Unlock()

	rows, err := store.Query(`SELECT * FROM MainChainTxs WHERE TransactionHash=?`, transactionHash)
	defer rows.Close()
	if err != nil {
		return false, err
	}

	return rows.Next(), nil
}
