package main

import (
	"io"
	"os"

	"github.com/elastos/Elastos.ELA.SideChain/blockchain"
	"github.com/elastos/Elastos.ELA.SideChain/config"
	"github.com/elastos/Elastos.ELA.SideChain/mempool"
	"github.com/elastos/Elastos.ELA.SideChain/netsync"
	"github.com/elastos/Elastos.ELA.SideChain/peer"
	"github.com/elastos/Elastos.ELA.SideChain/pow"
	"github.com/elastos/Elastos.ELA.SideChain/server"
	"github.com/elastos/Elastos.ELA.SideChain/service"

	"github.com/elastos/Elastos.ELA.Utility/elalog"
	"github.com/elastos/Elastos.ELA.Utility/http/jsonrpc"
	"github.com/elastos/Elastos.ELA.Utility/http/restful"
	"github.com/elastos/Elastos.ELA.Utility/p2p/addrmgr"
	"github.com/elastos/Elastos.ELA.Utility/p2p/connmgr"
)

const (
	logsPath = "./logs/"

	defaultMaxPerLogFileSize int64 = elalog.MBSize * 20
	defaultMaxLogsFolderSize int64 = elalog.GBSize * 2
)

// configFileWriter returns the configured parameters for log file writer.
func configFileWriter(params *config.Configuration) (string, int64, int64) {
	maxPerLogFileSize := defaultMaxPerLogFileSize
	maxLogsFolderSize := defaultMaxLogsFolderSize
	if params.MaxPerLogSize > 0 {
		maxPerLogFileSize = int64(params.MaxPerLogSize) * elalog.MBSize
	}
	if params.MaxLogsSize > 0 {
		maxLogsFolderSize = int64(params.MaxLogsSize) * elalog.MBSize
	}
	return logsPath, maxPerLogFileSize, maxLogsFolderSize
}

// log is a logger that is initialized with no output filters.  This
// means the package will not perform any logging by default until the caller
// requests it.
var (
	fileWriter = elalog.NewFileWriter(configFileWriter(config.Parameters.Configuration))
	logWriter  = io.MultiWriter(os.Stdout, fileWriter)
	backend    = elalog.NewBackend(logWriter, elalog.Llongfile)
	level      = elalog.Level(config.Parameters.PrintLevel)

	admrlog = backend.Logger("ADMR", level)
	cmgrlog = backend.Logger("CMGR", level)
	bcdblog = backend.Logger("BCDB", level)
	txmplog = backend.Logger("TXMP", level)
	synclog = backend.Logger("SYNC", level)
	peerlog = backend.Logger("PEER", level)
	minrlog = backend.Logger("MINR", level)
	spvslog = backend.Logger("SPVS", level)
	srvrlog = backend.Logger("SRVR", level)
	httplog = backend.Logger("HTTP", level)
	rpcslog = backend.Logger("RPCS", level)
	restlog = backend.Logger("REST", level)
	eladlog = backend.Logger("ELAD", level)
)

// The default amount of logging is none.
func init() {
	addrmgr.UseLogger(admrlog)
	connmgr.UseLogger(cmgrlog)
	blockchain.UseLogger(bcdblog)
	mempool.UseLogger(txmplog)
	netsync.UseLogger(synclog)
	peer.UseLogger(peerlog)
	server.UseLogger(srvrlog)
	pow.UseLogger(minrlog)
	service.UseLogger(httplog)
	jsonrpc.UseLogger(rpcslog)
	restful.UseLogger(restlog)
}
