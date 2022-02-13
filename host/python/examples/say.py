from WasmEdgeBindgen import bindgen
import WasmEdge

WasmEdge.Logging.debug()
cfx = WasmEdge.Configure()
cfx.add(WasmEdge.Host.Wasi)
vm = WasmEdge.VM(cfx)
vm.LoadWasmFromFile(
    "examples/rust/target/wasm32-wasi/debug/rust_bindgen_funcs_lib.wasm"
)
vm.Validate()
b = bindgen.Bindgen(vm)
data = b.execute(function_name="say", args="hello")
print(data)
b.deallocator()
