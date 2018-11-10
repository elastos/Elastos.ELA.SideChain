package smartcontract

import (
	"math/big"
	"bytes"
	"strconv"

	"github.com/elastos/Elastos.ELA.SideChain/contract"
	"github.com/elastos/Elastos.ELA.SideChain/vm"
	"github.com/elastos/Elastos.ELA.SideChain/vm/interfaces"
	"github.com/elastos/Elastos.ELA.SideChain/vm/types"
	"github.com/elastos/Elastos.ELA.SideChain/smartcontract/storage"
	"github.com/elastos/Elastos.ELA.SideChain/core"
	"github.com/elastos/Elastos.ELA.SideChain/servers"
	. "github.com/elastos/Elastos.ELA.SideChain/common"

	"github.com/elastos/Elastos.ELA.Utility/common"
	"github.com/elastos/Elastos.ELA.SideChain/blockchain"
)

type SmartContract struct {
	Engine         Engine
	Code           []byte
	Input          []byte
	ParameterTypes []contract.ContractParameterType
	Caller         common.Uint168
	CodeHash       common.Uint168
	ReturnType     contract.ContractParameterType
}

type Context struct {
	Caller         common.Uint168
	Code           []byte
	Input          []byte
	CodeHash       common.Uint168
	CacheCodeTable interfaces.IScriptTable
	Time           *big.Int
	BlockNumber    *big.Int
	SignableData   SignableData
	StateMachine   *blockchain.StateMachine
	DBCache        storage.DBCache
	Gas            common.Fixed64
	ReturnType     contract.ContractParameterType
	ParameterTypes []contract.ContractParameterType
	Trigger        vm.TriggerType
}

type Engine interface {
	Create(caller common.Uint168, code []byte) ([]byte, error)
	Call(caller common.Uint168, codeHash common.Uint168, input []byte) ([]byte, error)
}

func NewSmartContract(context *Context) (*SmartContract, error) {
	e := vm.NewExecutionEngine(context.SignableData,
		new(vm.CryptoECDsa),
		vm.MAXSTEPS,
		context.CacheCodeTable,
		context.StateMachine,
		context.Gas,
		context.Trigger,
		false,
	)

	return &SmartContract{
		Engine:         e,
		Code:           context.Code,
		CodeHash:       context.CodeHash,
		Input:          context.Input,
		Caller:         context.Caller,
		ReturnType:     context.ReturnType,
		ParameterTypes: context.ParameterTypes,
	}, nil
}

func (sc *SmartContract) DeployContract(payload *core.PayloadDeploy) ([]byte, error) {
	buffer := new(bytes.Buffer)
	paramBuilder := vm.NewParamsBuider(buffer)
	var parameterTypes []byte
	parameterTypes = contract.ContractParameterTypeToByte(payload.Code.ParameterTypes)
	returnType := byte(payload.Code.ReturnType)
	paramBuilder.EmitSysCall("Neo.Contract.Create", payload.Code.Code, parameterTypes, returnType, payload.Name,
		payload.CodeVersion, payload.Author, payload.Email, payload.Description)
	_ , err := sc.Engine.Call(sc.Caller, sc.CodeHash, paramBuilder.Bytes())
	if err != nil {
		return nil, err
	}
	return sc.Code, nil
}

func (sc *SmartContract) InvokeContract() (interface{}, error) {
	_, err := sc.Engine.Call(sc.Caller, sc.CodeHash, sc.Input)
	if err != nil {
		return nil, err
	}
	return sc.InvokeResult()
}

func (sc *SmartContract) InvokeResult() (interface{}, error) {
	engine := sc.Engine.(*vm.ExecutionEngine)
	if engine.GetEvaluationStack().Count() > 0 && vm.Peek(engine) != nil {
		switch sc.ReturnType {
		case contract.Boolean:
			return vm.PopBoolean(engine), nil
		case contract.Integer:
			return vm.PopBigInt(engine), nil
		case contract.ByteArray:
			bs := vm.PopByteArray(engine)
			return BytesToInt(bs), nil
		case contract.String:
			return string(vm.PopByteArray(engine)), nil
		case contract.Hash160, contract.Hash256:
			data := vm.PopByteArray(engine)
			return common.BytesToHexString(common.BytesReverse(data)), nil
		case contract.PublicKey:
			return common.BytesToHexString(vm.PopByteArray(engine)), nil
		case contract.Object:
			data := vm.PeekStackItem(engine)
			switch data.(type) {
			case *types.Boolean:
				return data.GetBoolean(), nil
			case *types.Integer:
				return data.GetBigInteger(), nil
			case *types.ByteArray:
				return common.BytesToHexString(data.GetByteArray()), nil
			case *types.GeneralInterface:
				interop := data.GetInterface()
				switch interop.(type) {
				case *core.Header:
					return servers.GetHeaderInfo(interop.(*core.Header)), nil
				case *core.Block:
					return servers.GetBlockInfo(interop.(*core.Block), true), nil
				case *core.Transaction:
					return servers.GetTransactionInfo(nil, interop.(*core.Transaction)), nil
				case *core.Asset:
					return servers.GetAssetInfo(interop.(*core.Asset)), nil
				}
			}

		}
	}
	return nil, nil
}

func (sc *SmartContract) InvokeParamsTransform() ([]byte, error) {
	builder := vm.NewParamsBuider(new(bytes.Buffer))
	b := bytes.NewBuffer(sc.Input)
	for _, k := range sc.ParameterTypes {
		switch k {
		case contract.Boolean:
			p, err := common.ReadUint8(b)
			if err != nil {
				return nil, err
			}
			if p >= 1 {
				builder.EmitPushBool(true)
			} else {
				builder.EmitPushBool(false)
			}
		case contract.Integer:
			p, err := common.ReadVarBytes(b)
			if err != nil {
				return nil, err
			}
			i, err := strconv.ParseInt(string(p), 10, 64)
			if err != nil {
				return nil, err
			}
			builder.EmitPushInteger(int64(i))
		case contract.Hash160, contract.Hash256:
			p, err := common.ReadVarBytes(b)
			if err != nil {
				return nil, err
			}
			builder.EmitPushByteArray(common.BytesReverse(p))
		case contract.ByteArray, contract.String:
			p, err := common.ReadVarBytes(b)
			if err != nil {
				return nil, err
			}
			builder.EmitPushByteArray(p)

		}
		builder.EmitPushCall(sc.CodeHash.Bytes())
		return builder.Bytes(), nil
	}
	return nil, nil
}
