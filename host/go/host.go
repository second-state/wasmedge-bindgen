package host

import (
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

	funcImports := h.addHostFns()
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

func (h *Host) Run(input []byte) error {
	lengthOfInput := len(input)
	allocateResult, err := h.vm.executor.Invoke(h.vm.store, "allocate", int32(lengthOfInput))
	if err != nil {
		return err
	}
	pointer := allocateResult[0].(int32)
	defer h.vm.executor.Invoke(h.vm.store, "deallocate", pointer, int32(lengthOfInput))
	
	memory := h.vm.store.FindMemory("memory")
	if memory == nil {
		return errors.New("Memory not found")
	}
	
	memory.SetData(input, uint(pointer), uint(lengthOfInput))
	
	_, err = h.vm.executor.Invoke(h.vm.store, "run_e", pointer, int32(lengthOfInput))

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
	default:
		// do nothing and fall through
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

func (h *Host) addHostFns() *wasmedge.ImportObject {
	wasmHostFn := func(data interface{}, mem *wasmedge.Memory, params []interface{}) ([]interface{}, wasmedge.Result) {
		h.return_result(params[0].(int32), params[1].(int32))
		return nil, wasmedge.Result_Success
	}

	argCount := 2

	argsType := make([]wasmedge.ValType, argCount)
	for i := 0; i < argCount; i++ {
		argsType[i] = wasmedge.ValType_I32
	}

	retType := []wasmedge.ValType{}
	funcType := wasmedge.NewFunctionType(argsType, retType)

	wasmEdgeHostFn := wasmedge.NewFunction(funcType, wasmHostFn, nil, 0)

	imports := wasmedge.NewImportObject("env")
	imports.AddFunction("return_result", wasmEdgeHostFn)
	return imports
}