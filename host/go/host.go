package host

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"

	"github.com/second-state/WasmEdge-go/wasmedge"
)

type Bindgen struct {
	vm         *wasmedge.VM

	resultChan chan []byte
	errChan    chan error

	funcImports *wasmedge.ImportObject
}

func NewBindgen(vm *wasmedge.VM) *Bindgen {
	b := &Bindgen {
		vm:          vm,
		resultChan:  make(chan []byte, 1),
		errChan:     make(chan error, 1),
	}

	b.init()

	return b
}

func (b *Bindgen) init() {
	funcImports := wasmedge.NewImportObject("env")

	resultFn := b.newHostFns(b.return_result, "return_result")
	funcImports.AddFunction("return_result", resultFn)
	errorFn := b.newHostFns(b.return_error, "return_error")
	funcImports.AddFunction("return_error", errorFn)

	b.vm.RegisterImport(funcImports)

	b.funcImports = funcImports
}

func (b *Bindgen) Execute(funcName string, inputs... interface{}) ([]byte, error) {
	inputsCount := len(inputs)
	
	// allocate new frame for passing pointers
	allocateResult, err := b.vm.Execute("allocate", int32(inputsCount * 4 * 2))
	if err != nil {
		return nil, err
	}
	pointerOfPointers := allocateResult[0].(int32)
	defer b.vm.Execute("deallocate", pointerOfPointers, int32(inputsCount * 4 * 2))
	
	memory := b.vm.GetStore().FindMemory("memory")
	if memory == nil {
		return nil, errors.New("Memory not found")
	}

	for idx, inp := range inputs {
		var pointer, lengthOfInput, byteLengthOfInput int32
		var err error
		switch input := inp.(type) {
		case []byte:
			pointer, lengthOfInput, err = b.settleByteSlice(memory, input)
			byteLengthOfInput = lengthOfInput
		case []int8:
			pointer, lengthOfInput, err = b.settleI8Slice(memory, input)
			byteLengthOfInput = lengthOfInput
		case []uint16:
			pointer, lengthOfInput, err = b.settleU16Slice(memory, input)
			byteLengthOfInput = lengthOfInput * 2
		case []int16:
			pointer, lengthOfInput, err = b.settleI16Slice(memory, input)
			byteLengthOfInput = lengthOfInput * 2
		case []uint32:
			pointer, lengthOfInput, err = b.settleU32Slice(memory, input)
			byteLengthOfInput = lengthOfInput * 4
		case []int32:
			pointer, lengthOfInput, err = b.settleI32Slice(memory, input)
			byteLengthOfInput = lengthOfInput * 4
		case []uint64:
			pointer, lengthOfInput, err = b.settleU64Slice(memory, input)
			byteLengthOfInput = lengthOfInput * 8
		case []int64:
			pointer, lengthOfInput, err = b.settleI64Slice(memory, input)
			byteLengthOfInput = lengthOfInput * 8
		case bool:
			pointer, lengthOfInput, err = b.settleBool(memory, input)
			byteLengthOfInput = lengthOfInput
		case int8:
			pointer, lengthOfInput, err = b.settleI8(memory, input)
			byteLengthOfInput = lengthOfInput
		case uint8:
			pointer, lengthOfInput, err = b.settleU8(memory, input)
			byteLengthOfInput = lengthOfInput
		case int16:
			pointer, lengthOfInput, err = b.settleI16(memory, input)
			byteLengthOfInput = lengthOfInput * 2
		case uint16:
			pointer, lengthOfInput, err = b.settleU16(memory, input)
			byteLengthOfInput = lengthOfInput * 2
		case int32:
			pointer, lengthOfInput, err = b.settleI32(memory, input)
			byteLengthOfInput = lengthOfInput * 4
		case uint32:
			pointer, lengthOfInput, err = b.settleU32(memory, input)
			byteLengthOfInput = lengthOfInput * 4
		case int64:
			pointer, lengthOfInput, err = b.settleI64(memory, input)
			byteLengthOfInput = lengthOfInput * 8
		case uint64:
			pointer, lengthOfInput, err = b.settleU64(memory, input)
			byteLengthOfInput = lengthOfInput * 8
		case float32:
			pointer, lengthOfInput, err = b.settleF32(memory, input)
			byteLengthOfInput = lengthOfInput * 4
		case float64:
			pointer, lengthOfInput, err = b.settleF64(memory, input)
			byteLengthOfInput = lengthOfInput * 8
		case string:
			pointer, lengthOfInput, err = b.settleString(memory, input)
			byteLengthOfInput = lengthOfInput
		default:
			return nil, errors.New(fmt.Sprintf("Unsupported arg type %T", input))
		}
		if err != nil {
			return nil, err
		}
		b.putPointerOfPointer(pointerOfPointers, memory, idx, pointer, lengthOfInput)
		defer b.vm.Execute("deallocate", pointer, byteLengthOfInput)
	}
	
	if _, err = b.vm.Execute(funcName, pointerOfPointers, int32(inputsCount)); err != nil {
		return nil, err
	}

	return b.executionResult()
}

func (b *Bindgen) Release() {
	b.funcImports.Release()
}

func (b *Bindgen) executionResult() ([]byte, error) {
	select {
	case res := <-b.resultChan:
		return res, nil
	case err := <-b.errChan:
		return nil, err
	}

	return nil, nil
}

func (b *Bindgen) return_result(pointer int32, size int32) {
	memory := b.vm.GetStore().FindMemory("memory")
	if memory == nil {
		return
	}

	data, err := memory.GetData(uint(pointer), uint(size))
	if err != nil {
		b.errChan <- err
		return
	}

	result := make([]byte, size)

	copy(result, data)

	if result != nil {
		b.resultChan <- result
	}
}

func (b *Bindgen) return_error(pointer int32, size int32) {
	memory := b.vm.GetStore().FindMemory("memory")
	if memory == nil {
		return
	}

	data, err := memory.GetData(uint(pointer), uint(size))
	if err != nil {
		b.errChan <- err
		return
	}

	result := make([]byte, size)

	copy(result, data)

	if result != nil {
		b.errChan <- errors.New(string(result))
	}
}

func (b *Bindgen) newHostFns(fn func(int32, int32) , importName string) *wasmedge.Function {
	wasmHostFn := func(data interface{}, mem *wasmedge.Memory, params []interface{}) ([]interface{}, wasmedge.Result) {
		fn(params[0].(int32), params[1].(int32))
		return nil, wasmedge.Result_Success
	}

	argCount := 2

	argsType := make([]wasmedge.ValType, argCount)
	for i := 0; i < argCount; i++ {
		argsType[i] = wasmedge.ValType_I32
	}

	retType := []wasmedge.ValType{}
	funcType := wasmedge.NewFunctionType(argsType, retType)

	return wasmedge.NewFunction(funcType, wasmHostFn, nil, 0)
}

func (b *Bindgen) putPointerOfPointer(pointerOfPointers int32, memory *wasmedge.Memory, inputIdx int, pointer int32, lengthOfInput int32) {
	// set data for pointer of pointer
	pointerBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(pointerBytes, uint32(pointer))
	memory.SetData(pointerBytes, uint(pointerOfPointers) + uint(inputIdx * 4 * 2), uint(4))
	lengthBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(lengthBytes, uint32(lengthOfInput))
	memory.SetData(lengthBytes, uint(pointerOfPointers) + uint(inputIdx * 4 * 2 + 4), uint(4))
}

func (b *Bindgen) settleByteSlice(memory *wasmedge.Memory, input []byte) (int32, int32, error) {
	lengthOfInput := int32(len(input))
	allocateResult, err := b.vm.Execute("allocate", lengthOfInput)
	if err != nil {
		return 0, 0, err
	}
	pointer := allocateResult[0].(int32)
	
	memory.SetData(input, uint(pointer), uint(lengthOfInput))

	return pointer, lengthOfInput, nil
}

func (b *Bindgen) settleI8Slice(memory *wasmedge.Memory, input []int8) (int32, int32, error) {
	lengthOfInput := int32(len(input))
	allocateResult, err := b.vm.Execute("allocate", lengthOfInput)
	if err != nil {
		return 0, 0, err
	}
	pointer := allocateResult[0].(int32)

	data := make([]byte, lengthOfInput)
	for i := 0; i < int(lengthOfInput); i++ {
		data[i] = byte(input[i])
	}
	
	memory.SetData(data, uint(pointer), uint(lengthOfInput))

	return pointer, lengthOfInput, nil
}

func (b *Bindgen) settleU16Slice(memory *wasmedge.Memory, input []uint16) (int32, int32, error) {
	lengthOfInput := int32(len(input))
	allocateResult, err := b.vm.Execute("allocate", lengthOfInput * 2)
	if err != nil {
		return 0, 0, err
	}
	pointer := allocateResult[0].(int32)

	data := make([]byte, lengthOfInput * 2)
	for i := 0; i < int(lengthOfInput); i++ {
		binary.LittleEndian.PutUint16(data[i*2:(i+1)*2], input[i])
	}
	
	memory.SetData(data, uint(pointer), uint(lengthOfInput * 2))

	return pointer, lengthOfInput, nil
}

func (b *Bindgen) settleI16Slice(memory *wasmedge.Memory, input []int16) (int32, int32, error) {
	lengthOfInput := int32(len(input))
	allocateResult, err := b.vm.Execute("allocate", lengthOfInput * 2)
	if err != nil {
		return 0, 0, err
	}
	pointer := allocateResult[0].(int32)

	data := make([]byte, lengthOfInput * 2)
	for i := 0; i < int(lengthOfInput); i++ {
		binary.LittleEndian.PutUint16(data[i*2:(i+1)*2], uint16(input[i]))
	}
	
	memory.SetData(data, uint(pointer), uint(lengthOfInput * 2))

	return pointer, lengthOfInput, nil
}

func (b *Bindgen) settleU32Slice(memory *wasmedge.Memory, input []uint32) (int32, int32, error) {
	lengthOfInput := int32(len(input))
	allocateResult, err := b.vm.Execute("allocate", lengthOfInput * 4)
	if err != nil {
		return 0, 0, err
	}
	pointer := allocateResult[0].(int32)

	data := make([]byte, lengthOfInput * 4)
	for i := 0; i < int(lengthOfInput); i++ {
		binary.LittleEndian.PutUint32(data[i*4:(i+1)*4], input[i])
	}
	
	memory.SetData(data, uint(pointer), uint(lengthOfInput * 4))

	return pointer, lengthOfInput, nil
}

func (b *Bindgen) settleI32Slice(memory *wasmedge.Memory, input []int32) (int32, int32, error) {
	lengthOfInput := int32(len(input))
	allocateResult, err := b.vm.Execute("allocate", lengthOfInput * 4)
	if err != nil {
		return 0, 0, err
	}
	pointer := allocateResult[0].(int32)

	data := make([]byte, lengthOfInput * 4)
	for i := 0; i < int(lengthOfInput); i++ {
		binary.LittleEndian.PutUint32(data[i*4:(i+1)*4], uint32(input[i]))
	}
	
	memory.SetData(data, uint(pointer), uint(lengthOfInput * 4))

	return pointer, lengthOfInput, nil
}

func (b *Bindgen) settleU64Slice(memory *wasmedge.Memory, input []uint64) (int32, int32, error) {
	lengthOfInput := int32(len(input))
	allocateResult, err := b.vm.Execute("allocate", lengthOfInput * 8)
	if err != nil {
		return 0, 0, err
	}
	pointer := allocateResult[0].(int32)

	data := make([]byte, lengthOfInput * 8)
	for i := 0; i < int(lengthOfInput); i++ {
		binary.LittleEndian.PutUint64(data[i*8:(i+1)*8], input[i])
	}
	
	memory.SetData(data, uint(pointer), uint(lengthOfInput * 8))

	return pointer, lengthOfInput, nil
}

func (b *Bindgen) settleI64Slice(memory *wasmedge.Memory, input []int64) (int32, int32, error) {
	lengthOfInput := int32(len(input))
	allocateResult, err := b.vm.Execute("allocate", lengthOfInput * 8)
	if err != nil {
		return 0, 0, err
	}
	pointer := allocateResult[0].(int32)

	data := make([]byte, lengthOfInput * 8)
	for i := 0; i < int(lengthOfInput); i++ {
		binary.LittleEndian.PutUint64(data[i*8:(i+1)*8], uint64(input[i]))
	}
	
	memory.SetData(data, uint(pointer), uint(lengthOfInput * 8))

	return pointer, lengthOfInput, nil
}

func (b *Bindgen) settleBool(memory *wasmedge.Memory, input bool) (int32, int32, error) {
	lengthOfInput := int32(1)
	allocateResult, err := b.vm.Execute("allocate", lengthOfInput)
	if err != nil {
		return 0, 0, err
	}
	pointer := allocateResult[0].(int32)
	
	var bt byte = 0
	if input {
		bt = 1
	}
	memory.SetData([]byte{bt}, uint(pointer), uint(lengthOfInput))

	return pointer, lengthOfInput, nil
}

func (b *Bindgen) settleRune(memory *wasmedge.Memory, input rune) (int32, int32, error) {
	lengthOfInput := int32(4)
	allocateResult, err := b.vm.Execute("allocate", lengthOfInput)
	if err != nil {
		return 0, 0, err
	}
	pointer := allocateResult[0].(int32)

	bytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(bytes, uint32(input))
	memory.SetData(bytes, uint(pointer), uint(lengthOfInput))

	return pointer, lengthOfInput, nil
}

func (b *Bindgen) settleI8(memory *wasmedge.Memory, input int8) (int32, int32, error) {
	lengthOfInput := int32(1)
	allocateResult, err := b.vm.Execute("allocate", lengthOfInput)
	if err != nil {
		return 0, 0, err
	}
	pointer := allocateResult[0].(int32)
	
	memory.SetData([]byte{byte(input)}, uint(pointer), uint(lengthOfInput))

	return pointer, lengthOfInput, nil
}

func (b *Bindgen) settleU8(memory *wasmedge.Memory, input uint8) (int32, int32, error) {
	lengthOfInput := int32(1)
	allocateResult, err := b.vm.Execute("allocate", lengthOfInput)
	if err != nil {
		return 0, 0, err
	}
	pointer := allocateResult[0].(int32)
	
	memory.SetData([]byte{byte(input)}, uint(pointer), uint(lengthOfInput))

	return pointer, lengthOfInput, nil
}

func (b *Bindgen) settleI16(memory *wasmedge.Memory, input int16) (int32, int32, error) {
	lengthOfInput := int32(1)
	allocateResult, err := b.vm.Execute("allocate", lengthOfInput * 2)
	if err != nil {
		return 0, 0, err
	}
	pointer := allocateResult[0].(int32)

	bytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(bytes, uint16(input))
	memory.SetData(bytes, uint(pointer), uint(lengthOfInput * 2))

	return pointer, lengthOfInput, nil
}

func (b *Bindgen) settleU16(memory *wasmedge.Memory, input uint16) (int32, int32, error) {
	lengthOfInput := int32(1)
	allocateResult, err := b.vm.Execute("allocate", lengthOfInput * 2)
	if err != nil {
		return 0, 0, err
	}
	pointer := allocateResult[0].(int32)

	bytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(bytes, input)
	memory.SetData(bytes, uint(pointer), uint(lengthOfInput * 2))

	return pointer, lengthOfInput, nil
}

func (b *Bindgen) settleI32(memory *wasmedge.Memory, input int32) (int32, int32, error) {
	lengthOfInput := int32(1)
	allocateResult, err := b.vm.Execute("allocate", lengthOfInput * 4)
	if err != nil {
		return 0, 0, err
	}
	pointer := allocateResult[0].(int32)

	bytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(bytes, uint32(input))
	memory.SetData(bytes, uint(pointer), uint(lengthOfInput * 4))

	return pointer, lengthOfInput, nil
}

func (b *Bindgen) settleU32(memory *wasmedge.Memory, input uint32) (int32, int32, error) {
	lengthOfInput := int32(1)
	allocateResult, err := b.vm.Execute("allocate", lengthOfInput * 4)
	if err != nil {
		return 0, 0, err
	}
	pointer := allocateResult[0].(int32)

	bytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(bytes, input)
	memory.SetData(bytes, uint(pointer), uint(lengthOfInput * 4))

	return pointer, lengthOfInput, nil
}

func (b *Bindgen) settleI64(memory *wasmedge.Memory, input int64) (int32, int32, error) {
	lengthOfInput := int32(1)
	allocateResult, err := b.vm.Execute("allocate", lengthOfInput * 8)
	if err != nil {
		return 0, 0, err
	}
	pointer := allocateResult[0].(int32)

	bytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(bytes, uint64(input))
	memory.SetData(bytes, uint(pointer), uint(lengthOfInput * 8))

	return pointer, lengthOfInput, nil
}

func (b *Bindgen) settleU64(memory *wasmedge.Memory, input uint64) (int32, int32, error) {
	lengthOfInput := int32(1)
	allocateResult, err := b.vm.Execute("allocate", lengthOfInput * 8)
	if err != nil {
		return 0, 0, err
	}
	pointer := allocateResult[0].(int32)

	bytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(bytes, input)
	memory.SetData(bytes, uint(pointer), uint(lengthOfInput * 8))

	return pointer, lengthOfInput, nil
}

func (b *Bindgen) settleF32(memory *wasmedge.Memory, input float32) (int32, int32, error) {
	lengthOfInput := int32(1)
	allocateResult, err := b.vm.Execute("allocate", lengthOfInput * 4)
	if err != nil {
		return 0, 0, err
	}
	pointer := allocateResult[0].(int32)

	bytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(bytes, math.Float32bits(input))
	memory.SetData(bytes, uint(pointer), uint(lengthOfInput * 4))

	return pointer, lengthOfInput, nil
}

func (b *Bindgen) settleF64(memory *wasmedge.Memory, input float64) (int32, int32, error) {
	lengthOfInput := int32(1)
	allocateResult, err := b.vm.Execute("allocate", lengthOfInput * 8)
	if err != nil {
		return 0, 0, err
	}
	pointer := allocateResult[0].(int32)

	bytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(bytes, math.Float64bits(input))
	memory.SetData(bytes, uint(pointer), uint(lengthOfInput * 8))

	return pointer, lengthOfInput, nil
}

func (b *Bindgen) settleString(memory *wasmedge.Memory, input string) (int32, int32, error) {
	lengthOfInput := int32(len([]byte(input)))
	allocateResult, err := b.vm.Execute("allocate", lengthOfInput)
	if err != nil {
		return 0, 0, err
	}
	pointer := allocateResult[0].(int32)
	
	memory.SetData([]byte(input), uint(pointer), uint(lengthOfInput))

	return pointer, lengthOfInput, nil
}