package core

import (
	"bytes"
	"github.com/elastos/Elastos.ELA.Utility/common"
	"testing"
)

func TestPayloadRegisterIdentification_Deserialize(t *testing.T) {
	rawData, _ := common.HexStringToBytes("22696a387266623641345269376335435245316e4456645643554d7555786b6b3263360002176b79632f706572736f6e2f6964656e746974794361726401bd117820c4cf30b0ad9ce68fe92b0117ca41ac2b6a49235fabd793fc3a9413c0ad227369676e6174757265223a2233303435303232303439396135646533663834653765393139633236623661383534336664323431323936333463363565653464333866653265333338366563386135646165353730323231303062373637396465386431383161343534653264656638663535646534323365396531356265626364653563353865383731643230616130643931313632666636222c226e6f74617279223a22434f4f495822106b79632f706572736f6e2f70686f6e6501cf985a76d1eef80aa7aa4ce144edad2c042c217ecabb0f86cd91bd4dae3b7215af227369676e6174757265223a22333034363032323130306538383830343033383864306635363931383365656138653666363038653030663338613065643338653164326632323761383633386566336566643866366330323231303063613163326563323061396139333363303732663665346536623936363061356161316263616365386166333834616261613330393431303130643461396366222c226e6f74617279223a22434f4f495822")

	r := bytes.NewReader(rawData)
	payload := PayloadRegisterIdentification{}
	payload.Deserialize(r, RegisterIdentificationVersion)

	if payload.ID != "ij8rfb6A4Ri7c5CRE1nDVdVCUMuUxkk2c6" {
		t.Error("ID deserialize error!")
	}

	if len(payload.Contents) != 2 {
		t.Error("ID contents deserialize error!")
	}

	if payload.Contents[0].Path != "kyc/person/identityCard" ||
		payload.Contents[1].Path != "kyc/person/phone" {
		t.Error("ID content path deserialize error!")
	}

	if len(payload.Contents[0].Values) != 1 || len(payload.Contents[1].Values) != 1 {
		t.Error("ID content value deserialize error!")
	}
}