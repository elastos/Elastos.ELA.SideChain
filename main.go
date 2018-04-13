package main

import (
	"os"
	"runtime"
	"time"

	. "github.com/elastos/Elastos.ELA.SideChain/common/config"
	"github.com/elastos/Elastos.ELA.SideChain/consensus/pow"
	. "github.com/elastos/Elastos.ELA.SideChain/core/auxpow"
	. "github.com/elastos/Elastos.ELA.SideChain/core/ledger"
	. "github.com/elastos/Elastos.ELA.SideChain/core/transaction"
	"github.com/elastos/Elastos.ELA.SideChain/net/servers"
	"github.com/elastos/Elastos.ELA.SideChain/net/servers/httpjsonrpc"
	"github.com/elastos/Elastos.ELA.SideChain/net/servers/httpnodeinfo"
	"github.com/elastos/Elastos.ELA.SideChain/net/servers/httprestful"
	"github.com/elastos/Elastos.ELA.SideChain/net/servers/httpwebsocket"
	"github.com/elastos/Elastos.ELA.SideChain/spv"
	uti_tx "github.com/elastos/Elastos.ELA.Utility/core/transaction"
	"github.com/elastos/Elastos.ELA/common/config"
	"github.com/elastos/Elastos.ELA/common/log"
	core_auxpow "github.com/elastos/Elastos.ELA/core/auxpow"
	"github.com/elastos/Elastos.ELA/core/ledger"
	core_ledger "github.com/elastos/Elastos.ELA/core/ledger"
	"github.com/elastos/Elastos.ELA/core/store/ChainStore"
	core_tx "github.com/elastos/Elastos.ELA/core/transaction"
	"github.com/elastos/Elastos.ELA/net/node"
	"github.com/elastos/Elastos.ELA/net/protocol"
)

const (
	DefaultMultiCoreNum = 4
)

func init() {
	InitSideConfig()

	core_auxpow.AuxPowFactorySingleton = &SideAuxPowFactoryImpl{}
	core_ledger.Validator = &SideTransactionValidatorImpl{
		core_ledger.TransactionValidatorImpl{}}
	uti_tx.PayloadFactorySingleton = &PayloadFactorySideNodeImpl{&uti_tx.PayloadFactoryImpl{}}

	log.Init(log.Path, log.Stdout)
	var coreNum int
	if config.Parameters.MultiCoreNum > DefaultMultiCoreNum {
		coreNum = int(config.Parameters.MultiCoreNum)
	} else {
		coreNum = DefaultMultiCoreNum
	}
	log.Debug("The Core number is ", coreNum)
	runtime.GOMAXPROCS(coreNum)
}

func handleLogFile() {
	go func() {
		for {
			time.Sleep(6 * time.Second)
			log.Trace("BlockHeight = ", ledger.DefaultLedger.Blockchain.BlockHeight)
			ledger.DefaultLedger.Blockchain.DumpState()
			bc := ledger.DefaultLedger.Blockchain
			log.Info("[", len(bc.Index), len(bc.BlockCache), len(bc.Orphans), "]")
			//ledger.DefaultLedger.Blockchain.DumpState()
			isNeedNewFile := log.CheckIfNeedNewFile()
			if isNeedNewFile {
				log.ClosePrintLog()
				log.Init(log.Path, os.Stdout)
			}
		} //for end
	}()

}

func startConsensus(noder protocol.Noder) {
	servers.Pow = pow.NewPowService("logPow", noder)
	if config.Parameters.PowConfiguration.AutoMining {
		log.Info("Start POW Services")
		go servers.Pow.Start()
	}
}

func main() {
	//var blockChain *ledger.Blockchain
	var err error
	var noder protocol.Noder
	log.Info("1. BlockChain init")
	ledger.DefaultLedger = new(ledger.Ledger)
	ledger.DefaultLedger.Store, err = ChainStore.NewLedgerStore()
	if err != nil {
		log.Fatal("open LedgerStore err:", err)
		os.Exit(1)
	}
	defer ledger.DefaultLedger.Store.Close()

	ledger.DefaultLedger.Store.InitLedgerStore(ledger.DefaultLedger)
	core_tx.TxStore = ledger.DefaultLedger.Store
	_, err = ledger.NewBlockchainWithGenesisBlock()
	if err != nil {
		log.Fatal(err, "BlockChain generate failed")
		goto ERROR
	}

	log.Info("2. SPV module init")
	spv.SpvInit()

	log.Info("3. Start the P2P networks")
	noder = node.InitNode()
	noder.WaitForSyncFinish()

	servers.NodeForServers = noder
	startConsensus(noder)

	handleLogFile()

	log.Info("4. --Start the RPC service")
	go httpjsonrpc.StartRPCServer()
	go httprestful.StartServer()
	go httpwebsocket.StartServer()
	if config.Parameters.HttpInfoStart {
		go httpnodeinfo.StartServer()
	}
	select {}
ERROR:
	os.Exit(1)
}
