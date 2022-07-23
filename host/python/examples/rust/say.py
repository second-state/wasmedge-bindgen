import WasmEdge

from WasmEdgeBindgen import bindgen
from WasmEdgeBindgen.utils import int_from_bytes, uint_from_bytes

WasmEdge.Logging.error()
cfx = WasmEdge.Configure()
cfx.AddHostRegistration(WasmEdge.Host.Wasi)
vm = WasmEdge.VM(cfx)
vm.LoadWasmFromFile(
    "examples/rust/target/wasm32-wasi/debug/rust_bindgen_funcs_lib.wasm"
)
vm.Validate()
b = bindgen.Bindgen(vm)

res, data = b.execute("say_veci32", [i for i in range(0, 10)])
print(bytes(data))
b.deallocator()

res, data = b.execute("say_string", "hello from python")
print(bytes(data))
b.deallocator()

res, data = b.execute("say_three_strings", "hello", "from", "python")
print(bytes(data))
b.deallocator()

num = 2**32 - 1
res, data = b.execute("say_number", num)
assert uint_from_bytes(data) == num
b.deallocator()

num = -(2**16) + 1
res, data = b.execute("say_number", num)
assert int_from_bytes(data) == num
b.deallocator()
