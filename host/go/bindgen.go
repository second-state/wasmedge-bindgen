package bindgen

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"math"

	"github.com/second-state/WasmEdge-go/wasmedge"
)

const (
	U8 int32 = 1
	I8 = 2
	U16 = 3
	I16 = 4
	U32 = 5
	I32 = 6
	U64 = 7
	I64 = 8
	F32 = 9
	F64 = 10
	Bool = 11
	Rune = 12
	ByteArray = 21
	I8Array = 22
	U16Array = 23
	I16Array = 24
	U32Array = 25
	I32Array = 26
	U64Array = 27
	I64Array = 28
	String = 31
)

type Bindgen struct {
	vm         *wasmedge.VM

	funcImports *wasmedge.Module
}

func New(vm *wasmedge.VM) *Bindgen {
	b := &Bindgen {
		vm:          vm,
	}

	return b
}

func (b *Bindgen) Instantiate() {
	b.vm.Instantiate()
}

func (b *Bindgen) GetVm() *wasmedge.VM {
	return b.vm
}

func (b *Bindgen) Execute(funcName string, inputs... interface{}) ([]interface{}, interface{}, error) {
	inputsCount := len(inputs)
	
	// allocate new frame for passing pointers
	allocateResult, err := b.vm.Execute("allocate", int32(inputsCount * 4 * 2))
	if err != nil {
		return nil, nil, err
	}
	pointerOfPointers := allocateResult[0].(int32)
	// Don't need to deallocate because the memory will be loaded and free in the wasm
	// defer b.vm.Execute("deallocate", pointerOfPointers, int32(inputsCount * 4 * 2))
	
	memory := b.vm.GetActiveModule().FindMemory("memory")
	if memory == nil {
		return nil, nil, errors.New("Memory not found")
	}

	for idx, inp := range inputs {
		var pointer, lengthOfInput int32
		var err error
		switch input := inp.(type) {
		case []byte:
			pointer, lengthOfInput, err = b.settleByteSlice(memory, input)
		case []int8:
			pointer, lengthOfInput, err = b.settleI8Slice(memory, input)
		case []uint16:
			pointer, lengthOfInput, err = b.settleU16Slice(memory, input)
		case []int16:
			pointer, lengthOfInput, err = b.settleI16Slice(memory, input)
		case []uint32:
			pointer, lengthOfInput, err = b.settleU32Slice(memory, input)
		case []int32:
			pointer, lengthOfInput, err = b.settleI32Slice(memory, input)
		case []uint64:
			pointer, lengthOfInput, err = b.settleU64Slice(memory, input)
		case []int64:
			pointer, lengthOfInput, err = b.settleI64Slice(memory, input)
		case bool:
			pointer, lengthOfInput, err = b.settleBool(memory, input)
		case int8:
			pointer, lengthOfInput, err = b.settleI8(memory, input)
		case uint8:
			pointer, lengthOfInput, err = b.settleU8(memory, input)
		case int16:
			pointer, lengthOfInput, err = b.settleI16(memory, input)
		case uint16:
			pointer, lengthOfInput, err = b.settleU16(memory, input)
		case int32:
			pointer, lengthOfInput, err = b.settleI32(memory, input)
		case uint32:
			pointer, lengthOfInput, err = b.settleU32(memory, input)
		case int64:
			pointer, lengthOfInput, err = b.settleI64(memory, input)
		case uint64:
			pointer, lengthOfInput, err = b.settleU64(memory, input)
		case float32:
			pointer, lengthOfInput, err = b.settleF32(memory, input)
		case float64:
			pointer, lengthOfInput, err = b.settleF64(memory, input)
		case string:
			pointer, lengthOfInput, err = b.settleString(memory, input)
		default:
			return nil, nil, errors.New(fmt.Sprintf("Unsupported arg type %T", input))
		}
		if err != nil {
			return nil, nil, err
		}
		b.putPointerOfPointer(pointerOfPointers, memory, idx, pointer, lengthOfInput)
	}
	
	var rets = make([]interface{}, 0);
	if rets, err = b.vm.Execute(funcName, pointerOfPointers, int32(inputsCount)); err != nil {
		return nil, nil, err
	}

	if len(rets) != 1 {
		return nil, nil, errors.New("Invalid return value")
	}

	revc, err := memory.GetData(uint(rets[0].(int32)), 9)
	if err != nil {
		return nil, nil, err
	}
	flag := revc[0]
	retPointer := b.getI32(revc[1:5])
	retLen := b.getI32(revc[5:9])

	if flag == 0 {
		return b.parse_result(retPointer, retLen);
	} else {
		return b.parse_error(retPointer, retLen);
	}
}

func (b *Bindgen) Release() {
	b.vm.Release()
}


func (b *Bindgen) parse_result(pointer int32, size int32) ([]interface{}, interface{}, error) {
	memory := b.vm.GetActiveModule().FindMemory("memory")
	if memory == nil {
		return nil, nil, errors.New("Can't get memory object")
	}

	data, err := memory.GetData(uint(pointer), uint(size) * 3 * 4)
	if err != nil {
		return nil, nil, err
	}

	rets := make([]int32, size * 3)

	for i := 0; i < int(size * 3); i++ {
		buf := bytes.NewBuffer(data[i*4:(i+1)*4])
		var p int32
		binary.Read(buf, binary.LittleEndian, &p)
		rets[i] = p
	}

	result := make([]interface{}, size)
	for i := 0; i < int(size); i++ {
		bytes, err := memory.GetData(uint(rets[i * 3]), uint(rets[i * 3 + 2]))
		if err != nil {
			return nil, nil, err
		}
		switch rets[i * 3 + 1] {
		case U8:
			result[i] = interface{}(b.getU8(bytes))
		case I8:
			result[i] = interface{}(b.getI8(bytes))
		case U16:
			result[i] = interface{}(b.getU16(bytes))
		case I16:
			result[i] = interface{}(b.getI16(bytes))
		case U32:
			result[i] = interface{}(b.getU32(bytes))
		case I32:
			result[i] = interface{}(b.getI32(bytes))
		case U64:
			result[i] = interface{}(b.getU64(bytes))
		case I64:
			result[i] = interface{}(b.getI64(bytes))
		case F32:
			result[i] = interface{}(b.getF32(bytes))
		case F64:
			result[i] = interface{}(b.getF64(bytes))
		case Bool:
			result[i] = interface{}(b.getBool(bytes))
		case Rune:
			result[i] = interface{}(b.getRune(bytes))
		case String:
			result[i] = interface{}(b.getString(bytes))
		case ByteArray:
			result[i] = interface{}(b.getByteSlice(bytes))
		case I8Array:
			result[i] = interface{}(b.getI8Slice(bytes))
		case U16Array:
			result[i] = interface{}(b.getU16Slice(bytes))
		case I16Array:
			result[i] = interface{}(b.getI16Slice(bytes))
		case U32Array:
			result[i] = interface{}(b.getU32Slice(bytes))
		case I32Array:
			result[i] = interface{}(b.getI32Slice(bytes))
		case U64Array:
			result[i] = interface{}(b.getU64Slice(bytes))
		case I64Array:
			result[i] = interface{}(b.getI64Slice(bytes))
		}
	}

	return result, nil, nil
}

func (b *Bindgen) parse_error(pointer int32, size int32) ([]interface{}, interface{}, error) {
	memory := b.vm.GetActiveModule().FindMemory("memory")
	if memory == nil {
		return nil, nil, errors.New("Can't get memory object")
	}

	data, err := memory.GetData(uint(pointer), uint(size))
	if err != nil {
		return nil, nil, err
	}

	return nil, string(data), nil
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



func (b *Bindgen) getU8(d []byte) uint8 {
	return uint8(d[0])
}

func (b *Bindgen) getI8(d []byte) int8 {
	return int8(d[0])
}

func (b *Bindgen) getU16(d []byte) (r uint16) {
	buf := bytes.NewBuffer(d)
	binary.Read(buf, binary.LittleEndian, &r)
	return
}

func (b *Bindgen) getI16(d []byte) (r int16) {
	buf := bytes.NewBuffer(d)
	binary.Read(buf, binary.LittleEndian, &r)
	return
}

func (b *Bindgen) getU32(d []byte) (r uint32) {
	buf := bytes.NewBuffer(d)
	binary.Read(buf, binary.LittleEndian, &r)
	return
}

func (b *Bindgen) getI32(d []byte) (r int32) {
	buf := bytes.NewBuffer(d)
	binary.Read(buf, binary.LittleEndian, &r)
	return
}

func (b *Bindgen) getU64(d []byte) (r uint64) {
	buf := bytes.NewBuffer(d)
	binary.Read(buf, binary.LittleEndian, &r)
	return
}

func (b *Bindgen) getI64(d []byte) (r int64) {
	buf := bytes.NewBuffer(d)
	binary.Read(buf, binary.LittleEndian, &r)
	return
}

func (b *Bindgen) getF32(d []byte) float32 {
	buf := bytes.NewBuffer(d)
	var p uint32
	binary.Read(buf, binary.LittleEndian, &p)
	return math.Float32frombits(p)
}

func (b *Bindgen) getF64(d []byte) float64 {
	buf := bytes.NewBuffer(d)
	var p uint64
	binary.Read(buf, binary.LittleEndian, &p)
	return math.Float64frombits(p)
}

func (b *Bindgen) getBool(d []byte) bool {
	return d[0] == byte(1)
}

func (b *Bindgen) getRune(d []byte) rune {
	buf := bytes.NewBuffer(d)
	var p uint32
	binary.Read(buf, binary.LittleEndian, &p)
	return rune(p)
}

func (b *Bindgen) getString(d []byte) string {
	return string(d)
}

func (b *Bindgen) getByteSlice(d []byte) []byte {
	x := make([]byte, len(d))
	copy(x, d)
	return x
}

func (b *Bindgen) getI8Slice(d []byte) []int8 {
	r := make([]int8, len(d))
	for i, v := range d {
		r[i] = int8(v)
	}
	return r
}

func (b *Bindgen) getU16Slice(d []byte) []uint16 {
	r := make([]uint16, len(d) / 2)
	for i := 0; i < len(r); i++ {
		buf := bytes.NewBuffer(d[i*2 : (i+1)*2])
		binary.Read(buf, binary.LittleEndian, &r[i])
	}
	return r
}

func (b *Bindgen) getI16Slice(d []byte) []int16 {
	r := make([]int16, len(d) / 2)
	for i := 0; i < len(r); i++ {
		buf := bytes.NewBuffer(d[i*2 : (i+1)*2])
		binary.Read(buf, binary.LittleEndian, &r[i])
	}
	return r
	
}

func (b *Bindgen) getU32Slice(d []byte) []uint32 {
	r := make([]uint32, len(d) / 4)
	for i := 0; i < len(r); i++ {
		buf := bytes.NewBuffer(d[i*4 : (i+1)*4])
		binary.Read(buf, binary.LittleEndian, &r[i])
	}
	return r
	
}

func (b *Bindgen) getI32Slice(d []byte) []int32 {
	r := make([]int32, len(d) / 4)
	for i := 0; i < len(r); i++ {
		buf := bytes.NewBuffer(d[i*4 : (i+1)*4])
		binary.Read(buf, binary.LittleEndian, &r[i])
	}
	return r
	
}

func (b *Bindgen) getU64Slice(d []byte) []uint64 {
	r := make([]uint64, len(d) / 8)
	for i := 0; i < len(r); i++ {
		buf := bytes.NewBuffer(d[i*8 : (i+1)*8])
		binary.Read(buf, binary.LittleEndian, &r[i])
	}
	return r
	
}

func (b *Bindgen) getI64Slice(d []byte) []int64 {
	r := make([]int64, len(d) / 8)
	for i := 0; i < len(r); i++ {
		buf := bytes.NewBuffer(d[i*8 : (i+1)*8])
		binary.Read(buf, binary.LittleEndian, &r[i])
	}
	return r
	
}