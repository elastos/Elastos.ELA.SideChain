package spv

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"os"

	"github.com/elastos/Elastos.ELA.SideChain/config"
	"github.com/elastos/Elastos.ELA.SideChain/log"

	spv "github.com/elastos/Elastos.ELA.SPV/interface"
	spvlog "github.com/elastos/Elastos.ELA.SPV/log"
)

var SpvService spv.SPVService

func SpvInit(listener spv.TransactionListener) error {
	var err error
	spvlog.Init(config.Parameters.SpvPrintLevel, 20, 1024)

	var id = make([]byte, 8)
	var clientId uint64
	rand.Read(id)
	binary.Read(bytes.NewReader(id), binary.LittleEndian, &clientId)

	SpvService, err = spv.NewSPVService(config.Parameters.SpvMagic, config.Parameters.MainChainFoundationAddress, clientId,
		config.Parameters.SpvSeedList, config.Parameters.SpvMinOutbound, config.Parameters.SpvMaxConnections)
	if err != nil {
		return err
	}

	err = SpvService.RegisterTransactionListener(listener)
	if err != nil {
		return err
	}
	log.Info("[SpvInit] add listerner to addr:", listener.Address())

	go func() {
		if err := SpvService.Start(); err != nil {
			log.Info("Spv service start failed ：", err)
		}
		log.Info("Spv service stoped")
		os.Exit(-1)
	}()
	return nil
}
