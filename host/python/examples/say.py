from WasmEdgeBindgen import bindgen
import WasmEdge

WasmEdge.Logging.error()
cfx = WasmEdge.Configure()
cfx.add(WasmEdge.Host.Wasi)
vm = WasmEdge.VM(cfx)
vm.LoadWasmFromFile(
    "examples/rust/target/wasm32-wasi/debug/rust_bindgen_funcs_lib.wasm"
)
vm.Validate()
b = bindgen.Bindgen(vm)
res, data = b.execute(function_name="say", args="hello")
print(bytes(data))
b.deallocator()
