## About

This is a Python library lets you call the WebAssembly exported functions that are tweaked by wasmedge-bindgen.

## Guide

Originally, you can call the wasm function from [WasmEdge-Python](https://github.com/SAtacker/WasmEdge) like this:

Install WasmEdge python module:

```sh
pip install -i https://test.pypi.org/simple/ WasmEdge>=0.2.2
```


```python
import WasmEdge

vm = WasmEdge.VM()
vm.LoadWasmFile("say.wasm")
vm.Validate()
vm.Instantiate()
res,data = vm.Execute("say","hello",1)
```

When your wasm function is retouched by #[wasmedge_bindgen]:
```rust
#[wasmedge_bindgen]
pub fn say(s: String) -> String {
  let r = String::from("hello ");
  return r + s.as_str();
}
```
Then you should call it using this library:
```python
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
```

## Explain

The result is a WasmEdge.Value array.
When the return value of wasm(rust) is (Tuple,) or Result<(Tuple,), String>,
the array length will be the same with the members' count of tuple.

If the function returns Err(String), you will get it by err.
