package host

import (
	"encoding/binary"
	"errors"

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
		input := inp.([]byte)
		lengthOfInput := int32(len(input))
		allocateResult, err = h.vm.executor.Invoke(h.vm.store, "allocate", lengthOfInput)
		if err != nil {
			return err
		}
		pointer := allocateResult[0].(int32)
		defer h.vm.executor.Invoke(h.vm.store, "deallocate", pointer, lengthOfInput)
		
		memory.SetData(input, uint(pointer), uint(lengthOfInput))

		// set data for pointer of pointer
		pointerBytes := make([]byte, 4)
		binary.LittleEndian.PutUint32(pointerBytes, uint32(pointer))
		memory.SetData(pointerBytes, uint(pointerOfPointers) + uint(idx * 4 * 2), uint(4))
		lengthBytes := make([]byte, 4)
		binary.LittleEndian.PutUint32(lengthBytes, uint32(lengthOfInput))
		memory.SetData(lengthBytes, uint(pointerOfPointers) + uint(idx * 4 * 2 + 4), uint(4))
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