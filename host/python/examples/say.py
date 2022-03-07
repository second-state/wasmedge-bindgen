from random import randint

import WasmEdge

from WasmEdgeBindgen import bindgen

WasmEdge.Logging.error()
cfx = WasmEdge.Configure()
cfx.AddHostRegistration(WasmEdge.Host.Wasi)
vm = WasmEdge.VM(cfx)
vm.LoadWasmFromFile(
    "examples/rust/target/wasm32-wasi/debug/rust_bindgen_funcs_lib.wasm"
)
vm.Validate()
b = bindgen.Bindgen(vm)

res, data = b.execute(
    function_name="say_veci32", args=tuple([i for i in range(0, 10)])
)
print(bytes(data))
b.deallocator()

res, data = b.execute(function_name="say_string", args="hello from python")
print(bytes(data))
b.deallocator()

num = randint(0, 100)
res, data = b.execute(function_name="say_number", args=num)
assert int.from_bytes(data, "little") == num
print(int.from_bytes(data, "little"))
b.deallocator()
