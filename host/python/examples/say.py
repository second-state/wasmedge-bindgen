from WasmEdgeBindgen import bindgen
import WasmEdge

cfx = WasmEdge.Configure()
cfx.add(WasmEdge.Host.Wasi)
vm = WasmEdge.VM(cfx)
vm.LoadWasmFromFile("rust/target/wasm32-wasi/release/rust_bindgen_funcs_lib.wasm")
vm.Validate()
b = bindgen.Bindgen(vm)
data  = b.execute(function_name="say",args="hello")
print(data[0].Value)