package host

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"

	"github.com/second-state/WasmEdge-go/wasmedge"
)

// VM is a wrapper of objects of WasmEdge API
type VM struct {
	executor      *wasmedge.Executor
	store         *wasmedge.Store
	wasiImports   *wasmedge.ImportObject
	funcImports   *wasmedge.ImportObject
	tfImports     *wasmedge.ImportObject
	tfliteImports *wasmedge.ImportObject
}

type Host struct {
	vm         *VM
	wasmFile   string
	enableTF   bool

	resultChan chan []byte
	errChan    chan error
}

func NewHost(wasmFile string) *Host {
	return &Host {
		wasmFile: wasmFile,
	}
}

func NewHostWithTF(wasmFile string) *Host {
	return &Host {
		wasmFile: wasmFile,
		enableTF: true,
	}
}

func (h *Host) Init() error {
	loader := wasmedge.NewLoader()
	defer loader.Release()

	ast, err := loader.LoadFile(h.wasmFile)
	if err != nil {
		return err
	}
	defer ast.Release()

	val := wasmedge.NewValidator()
	defer val.Release()
	if err = val.Validate(ast); err != nil {
		return err
	}

	h.resultChan = make(chan []byte, 1)
	h.errChan = make(chan error, 1)

	store := wasmedge.NewStore()
	executor := wasmedge.NewExecutor()

	wasiImports := wasmedge.NewWasiImportObject(nil, nil, nil)
	executor.RegisterImport(store, wasiImports)

	var tfImports *wasmedge.ImportObject
	var tfliteImports *wasmedge.ImportObject

	if h.enableTF {
		/// Register WasmEdge-tensorflow
		tfImports = wasmedge.NewTensorflowImportObject()
		tfliteImports = wasmedge.NewTensorflowLiteImportObject()
		executor.RegisterImport(store, tfImports)
		executor.RegisterImport(store, tfliteImports)
	}

	funcImports := wasmedge.NewImportObject("env")

	resultFn := h.newHostFns(h.return_result, "return_result")
	funcImports.AddFunction("return_result", resultFn)
	errorFn := h.newHostFns(h.return_error, "return_error")
	funcImports.AddFunction("return_error", errorFn)

	executor.RegisterImport(store, funcImports)

	executor.Instantiate(store, ast)

	h.vm = &VM {
		executor:      executor,
		store:         store,
		wasiImports:   wasiImports,
		funcImports:   funcImports,
		tfImports:     tfImports,
		tfliteImports: tfliteImports,
	}

	return nil
}

func (h *Host) Run(inputs... interface{}) error {
	inputsCount := len(inputs)
	
	// allocate new frame for passing pointers
	allocateResult, err := h.vm.executor.Invoke(h.vm.store, "allocate", int32(inputsCount * 4 * 2))
	if err != nil {
		return err
	}
	pointerOfPointers := allocateResult[0].(int32)
	defer h.vm.executor.Invoke(h.vm.store, "deallocate", pointerOfPointers, int32(inputsCount * 4 * 2))
	
	memory := h.vm.store.FindMemory("memory")
	if memory == nil {
		return errors.New("Memory not found")
	}

	for idx, inp := range inputs {
		var pointer, lengthOfInput int32
		var err error
		switch input := inp.(type) {
		case []byte:
			pointer, lengthOfInput, err = h.settleByteSlice(memory, input)
		case []int8:
			pointer, lengthOfInput, err = h.settleI8Slice(memory, input)
		case []uint16:
			pointer, lengthOfInput, err = h.settleU16Slice(memory, input)
		case []int16:
			pointer, lengthOfInput, err = h.settleI16Slice(memory, input)
		case []uint32:
			pointer, lengthOfInput, err = h.settleU32Slice(memory, input)
		case []int32:
			pointer, lengthOfInput, err = h.settleI32Slice(memory, input)
		case []uint64:
			pointer, lengthOfInput, err = h.settleU64Slice(memory, input)
		case []int64:
			pointer, lengthOfInput, err = h.settleI64Slice(memory, input)
		case bool:
			pointer, lengthOfInput, err = h.settleBool(memory, input)
		case int8:
			pointer, lengthOfInput, err = h.settleI8(memory, input)
		case uint8:
			pointer, lengthOfInput, err = h.settleU8(memory, input)
		case int16:
			pointer, lengthOfInput, err = h.settleI16(memory, input)
		case uint16:
			pointer, lengthOfInput, err = h.settleU16(memory, input)
		case int32:
			pointer, lengthOfInput, err = h.settleI32(memory, input)
		case uint32:
			pointer, lengthOfInput, err = h.settleU32(memory, input)
		case int64:
			pointer, lengthOfInput, err = h.settleI64(memory, input)
		case uint64:
			pointer, lengthOfInput, err = h.settleU64(memory, input)
		case float32:
			pointer, lengthOfInput, err = h.settleF32(memory, input)
		case float64:
			pointer, lengthOfInput, err = h.settleF64(memory, input)
		case string:
			pointer, lengthOfInput, err = h.settleString(memory, input)
		default:
			return errors.New(fmt.Sprintf("Unsupported arg type %T", input))
		}
		if err != nil {
			return err
		}
		h.putPointerOfPointer(pointerOfPointers, memory, idx, pointer, lengthOfInput)
		defer h.vm.executor.Invoke(h.vm.store, "deallocate", pointer, lengthOfInput)
	}
	
	_, err = h.vm.executor.Invoke(h.vm.store, "run_e", pointerOfPointers, int32(inputsCount))

	return err
}

func (h *Host) Release() {
	if h.vm.tfImports != nil {
		h.vm.tfImports.Release()
	}
	if h.vm.tfliteImports != nil {
		h.vm.tfliteImports.Release()
	}
	h.vm.funcImports.Release()
	h.vm.wasiImports.Release()
	h.vm.store.Release()
	h.vm.executor.Release()
}

func (h *Host) ExecutionResult() ([]byte, error) {
	select {
	case res := <-h.resultChan:
		return res, nil
	case err := <-h.errChan:
		return nil, err
	}

	return nil, nil
}

func (h *Host) return_result(pointer int32, size int32) {
	memory := h.vm.store.FindMemory("memory")
	if memory == nil {
		return
	}

	data, err := memory.GetData(uint(pointer), uint(size))
	if err != nil {
		h.errChan <- err
		return
	}

	result := make([]byte, size)

	copy(result, data)

	if result != nil {
		h.resultChan <- result
	}
}

func (h *Host) return_error(pointer int32, size int32) {
	memory := h.vm.store.FindMemory("memory")
	if memory == nil {
		return
	}

	data, err := memory.GetData(uint(pointer), uint(size))
	if err != nil {
		h.errChan <- err
		return
	}

	result := make([]byte, size)

	copy(result, data)

	if result != nil {
		h.errChan <- errors.New(string(result))
	}
}

func (h *Host) newHostFns(fn func(int32, int32) , importName string) *wasmedge.Function {
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

func (h *Host) putPointerOfPointer(pointerOfPointers int32, memory *wasmedge.Memory, inputIdx int, pointer int32, lengthOfInput int32) {
	// set data for pointer of pointer
	pointerBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(pointerBytes, uint32(pointer))
	memory.SetData(pointerBytes, uint(pointerOfPointers) + uint(inputIdx * 4 * 2), uint(4))
	lengthBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(lengthBytes, uint32(lengthOfInput))
	memory.SetData(lengthBytes, uint(pointerOfPointers) + uint(inputIdx * 4 * 2 + 4), uint(4))
}

func (h *Host) settleByteSlice(memory *wasmedge.Memory, input []byte) (int32, int32, error) {
	lengthOfInput := int32(len(input))
	allocateResult, err := h.vm.executor.Invoke(h.vm.store, "allocate", lengthOfInput)
	if err != nil {
		return 0, 0, err
	}
	pointer := allocateResult[0].(int32)
	
	memory.SetData(input, uint(pointer), uint(lengthOfInput))

	return pointer, lengthOfInput, nil
}

func (h *Host) settleI8Slice(memory *wasmedge.Memory, input []int8) (int32, int32, error) {
	lengthOfInput := int32(len(input))
	allocateResult, err := h.vm.executor.Invoke(h.vm.store, "allocate", lengthOfInput)
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

func (h *Host) settleU16Slice(memory *wasmedge.Memory, input []uint16) (int32, int32, error) {
	lengthOfInput := int32(len(input))
	allocateResult, err := h.vm.executor.Invoke(h.vm.store, "allocate", lengthOfInput * 2)
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

func (h *Host) settleI16Slice(memory *wasmedge.Memory, input []int16) (int32, int32, error) {
	lengthOfInput := int32(len(input))
	allocateResult, err := h.vm.executor.Invoke(h.vm.store, "allocate", lengthOfInput * 2)
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

func (h *Host) settleU32Slice(memory *wasmedge.Memory, input []uint32) (int32, int32, error) {
	lengthOfInput := int32(len(input))
	allocateResult, err := h.vm.executor.Invoke(h.vm.store, "allocate", lengthOfInput * 4)
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

func (h *Host) settleI32Slice(memory *wasmedge.Memory, input []int32) (int32, int32, error) {
	lengthOfInput := int32(len(input))
	allocateResult, err := h.vm.executor.Invoke(h.vm.store, "allocate", lengthOfInput * 4)
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

func (h *Host) settleU64Slice(memory *wasmedge.Memory, input []uint64) (int32, int32, error) {
	lengthOfInput := int32(len(input))
	allocateResult, err := h.vm.executor.Invoke(h.vm.store, "allocate", lengthOfInput * 8)
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

func (h *Host) settleI64Slice(memory *wasmedge.Memory, input []int64) (int32, int32, error) {
	lengthOfInput := int32(len(input))
	allocateResult, err := h.vm.executor.Invoke(h.vm.store, "allocate", lengthOfInput * 8)
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

func (h *Host) settleBool(memory *wasmedge.Memory, input bool) (int32, int32, error) {
	lengthOfInput := int32(1)
	allocateResult, err := h.vm.executor.Invoke(h.vm.store, "allocate", lengthOfInput)
	if err != nil {
		return 0, 0, err
	}
	pointer := allocateResult[0].(int32)
	
	var b byte = 0
	if input {
		b = 1
	}
	memory.SetData([]byte{b}, uint(pointer), uint(lengthOfInput))

	return pointer, lengthOfInput, nil
}

func (h *Host) settleRune(memory *wasmedge.Memory, input rune) (int32, int32, error) {
	lengthOfInput := int32(4)
	allocateResult, err := h.vm.executor.Invoke(h.vm.store, "allocate", lengthOfInput)
	if err != nil {
		return 0, 0, err
	}
	pointer := allocateResult[0].(int32)

	bytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(bytes, uint32(input))
	memory.SetData(bytes, uint(pointer), uint(lengthOfInput))

	return pointer, lengthOfInput, nil
}

func (h *Host) settleI8(memory *wasmedge.Memory, input int8) (int32, int32, error) {
	lengthOfInput := int32(1)
	allocateResult, err := h.vm.executor.Invoke(h.vm.store, "allocate", lengthOfInput)
	if err != nil {
		return 0, 0, err
	}
	pointer := allocateResult[0].(int32)
	
	memory.SetData([]byte{byte(input)}, uint(pointer), uint(lengthOfInput))

	return pointer, lengthOfInput, nil
}

func (h *Host) settleU8(memory *wasmedge.Memory, input uint8) (int32, int32, error) {
	lengthOfInput := int32(1)
	allocateResult, err := h.vm.executor.Invoke(h.vm.store, "allocate", lengthOfInput)
	if err != nil {
		return 0, 0, err
	}
	pointer := allocateResult[0].(int32)
	
	memory.SetData([]byte{byte(input)}, uint(pointer), uint(lengthOfInput))

	return pointer, lengthOfInput, nil
}

func (h *Host) settleI16(memory *wasmedge.Memory, input int16) (int32, int32, error) {
	lengthOfInput := int32(2)
	allocateResult, err := h.vm.executor.Invoke(h.vm.store, "allocate", lengthOfInput)
	if err != nil {
		return 0, 0, err
	}
	pointer := allocateResult[0].(int32)

	bytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(bytes, uint16(input))
	memory.SetData(bytes, uint(pointer), uint(lengthOfInput))

	return pointer, lengthOfInput, nil
}

func (h *Host) settleU16(memory *wasmedge.Memory, input uint16) (int32, int32, error) {
	lengthOfInput := int32(2)
	allocateResult, err := h.vm.executor.Invoke(h.vm.store, "allocate", lengthOfInput)
	if err != nil {
		return 0, 0, err
	}
	pointer := allocateResult[0].(int32)

	bytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(bytes, input)
	memory.SetData(bytes, uint(pointer), uint(lengthOfInput))

	return pointer, lengthOfInput, nil
}

func (h *Host) settleI32(memory *wasmedge.Memory, input int32) (int32, int32, error) {
	lengthOfInput := int32(4)
	allocateResult, err := h.vm.executor.Invoke(h.vm.store, "allocate", lengthOfInput)
	if err != nil {
		return 0, 0, err
	}
	pointer := allocateResult[0].(int32)

	bytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(bytes, uint32(input))
	memory.SetData(bytes, uint(pointer), uint(lengthOfInput))

	return pointer, lengthOfInput, nil
}

func (h *Host) settleU32(memory *wasmedge.Memory, input uint32) (int32, int32, error) {
	lengthOfInput := int32(4)
	allocateResult, err := h.vm.executor.Invoke(h.vm.store, "allocate", lengthOfInput)
	if err != nil {
		return 0, 0, err
	}
	pointer := allocateResult[0].(int32)

	bytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(bytes, input)
	memory.SetData(bytes, uint(pointer), uint(lengthOfInput))

	return pointer, lengthOfInput, nil
}

func (h *Host) settleI64(memory *wasmedge.Memory, input int64) (int32, int32, error) {
	lengthOfInput := int32(8)
	allocateResult, err := h.vm.executor.Invoke(h.vm.store, "allocate", lengthOfInput)
	if err != nil {
		return 0, 0, err
	}
	pointer := allocateResult[0].(int32)

	bytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(bytes, uint64(input))
	memory.SetData(bytes, uint(pointer), uint(lengthOfInput))

	return pointer, lengthOfInput, nil
}

func (h *Host) settleU64(memory *wasmedge.Memory, input uint64) (int32, int32, error) {
	lengthOfInput := int32(8)
	allocateResult, err := h.vm.executor.Invoke(h.vm.store, "allocate", lengthOfInput)
	if err != nil {
		return 0, 0, err
	}
	pointer := allocateResult[0].(int32)

	bytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(bytes, input)
	memory.SetData(bytes, uint(pointer), uint(lengthOfInput))

	return pointer, lengthOfInput, nil
}

func (h *Host) settleF32(memory *wasmedge.Memory, input float32) (int32, int32, error) {
	lengthOfInput := int32(4)
	allocateResult, err := h.vm.executor.Invoke(h.vm.store, "allocate", lengthOfInput)
	if err != nil {
		return 0, 0, err
	}
	pointer := allocateResult[0].(int32)

	bytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(bytes, math.Float32bits(input))
	memory.SetData(bytes, uint(pointer), uint(lengthOfInput))

	return pointer, lengthOfInput, nil
}

func (h *Host) settleF64(memory *wasmedge.Memory, input float64) (int32, int32, error) {
	lengthOfInput := int32(8)
	allocateResult, err := h.vm.executor.Invoke(h.vm.store, "allocate", lengthOfInput)
	if err != nil {
		return 0, 0, err
	}
	pointer := allocateResult[0].(int32)

	bytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(bytes, math.Float64bits(input))
	memory.SetData(bytes, uint(pointer), uint(lengthOfInput))

	return pointer, lengthOfInput, nil
}

func (h *Host) settleString(memory *wasmedge.Memory, input string) (int32, int32, error) {
	lengthOfInput := int32(len([]byte(input)))
	allocateResult, err := h.vm.executor.Invoke(h.vm.store, "allocate", lengthOfInput)
	if err != nil {
		return 0, 0, err
	}
	pointer := allocateResult[0].(int32)
	
	memory.SetData([]byte(input), uint(pointer), uint(lengthOfInput))

	return pointer, lengthOfInput, nil
}