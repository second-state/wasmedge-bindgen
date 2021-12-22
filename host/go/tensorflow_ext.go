// +build tensorflow

package host

import (
	"github.com/second-state/WasmEdge-go/wasmedge"
)
func (h *Host) InstallTensorflowExt() {
	/// Register WasmEdge-tensorflow
	tfImports := wasmedge.NewTensorflowImportObject()
	tfliteImports := wasmedge.NewTensorflowLiteImportObject()
	h.vm.executor.RegisterImport(h.vm.store, tfImports)
	h.vm.executor.RegisterImport(h.vm.store, tfliteImports)

	h.vm.extImports = append(h.vm.extImports, tfImports)
	h.vm.extImports = append(h.vm.extImports, tfliteImports)
}