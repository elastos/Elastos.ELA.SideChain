package core

import (
	"testing"
	"bytes"

	"github.com/elastos/Elastos.ELA.SideChain/contract"

	"github.com/stretchr/testify/assert"
	"github.com/elastos/Elastos.ELA.Utility/common"
)

func TestPayloadDeploy_Serialize(t *testing.T) {
	payload := PayloadDeploy{}
	code := contract.FunctionCode{}
	code.Code = []byte{1, 2, 3, 4, 3, 1, 2, 3, 4, 3, 1, 2, 3, 4, 3, 1, 2, 3, 4, 3, 3}
	code.ParameterTypes = []contract.ContractParameterType{contract.Signature}
	code.ReturnType = contract.Boolean
	code.CodeHash()
	payload.Code = &code
	payload.Params = []byte{1,3,4}
	payload.Name = "testName"
	payload.CodeVersion = "1.0"
	payload.Author = "author"
	payload.Email = "2@3.com"
	payload.Description = "description"
	payload.ProgramHash = common.Uint168{1,2,3}

	buf := new(bytes.Buffer)
	err := payload.Serialize(buf, 1)

	assert.NoError(t, err)

	r := bytes.NewReader(buf.Bytes())
	payload2 := PayloadDeploy{}
	err = payload2.Deserialize(r, 1)

	assert.NoError(t, err)
	assert.True(t, len(payload2.Code.Code) == len(payload.Code.Code))
	for i := 0; i < len(payload2.Code.Code); i++ {
		assert.True(t, payload2.Code.Code[i] == payload.Code.Code[i])
	}

	assert.True(t, len(payload2.Params) == len(payload.Params))
	for i := 0; i < len(payload.Params); i++ {
		assert.True(t, payload2.Params[i] == payload.Params[i])
	}

	assert.True(t, payload2.Name == payload.Name)
	assert.True(t, payload2.CodeVersion == payload.CodeVersion)
	assert.True(t, payload2.Author == payload.Author)
	assert.True(t, payload2.Email == payload.Email)
	assert.True(t, payload2.Description == payload.Description)
	assert.True(t, payload2.ProgramHash == payload.ProgramHash)
}
