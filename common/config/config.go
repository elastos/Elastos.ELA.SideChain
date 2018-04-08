package config

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"

	. "github.com/elastos/Elastos.ELA.Core/common/config"
)

var SideParameters sideConfigParams

type SideConfiguration struct {
	Configuration
	SpvSeedList         []string         `json:"SpvSeedList"`
	DestroyAddr         string           `json:"DestroyAddr"`
}

type SideConfigFile struct {
	ConfigFile SideConfiguration `json:"Configuration"`
}

type sideConfigParams struct {
	*SideConfiguration
	ChainParam *ChainParams
}

func init() {
	file, e := ioutil.ReadFile(DefaultConfigFilename)
	if e != nil {
		log.Fatalf("File error: %v\n", e)
		os.Exit(1)
	}
	// Remove the UTF-8 Byte Order Mark
	file = bytes.TrimPrefix(file, []byte("\xef\xbb\xbf"))

	config := SideConfigFile{}
	e = json.Unmarshal(file, &config)
	if e != nil {
		log.Fatalf("Unmarshal json file erro %v", e)
		os.Exit(1)
	}
	//	Parameters = &(config.ConfigFile)
	Parameters.Configuration = &(config.ConfigFile.Configuration)
	SideParameters.SideConfiguration = &(config.ConfigFile)
}
