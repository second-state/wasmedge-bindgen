// +build image

package host

import (
	"github.com/second-state/WasmEdge-go/wasmedge"
)
func (h *Host) InstallImageExt() {
	/// Register WasmEdge-image
	imgImports := wasmedge.NewImageImportObject()
	h.vm.executor.RegisterImport(h.vm.store, imgImports)

	h.vm.extImports = append(h.vm.extImports, imgImports)
}