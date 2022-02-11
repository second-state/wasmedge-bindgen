## About

This is a Python library lets you call the WebAssembly exported functions that are tweaked by wasmedge-bindgen.

## Guide

Originally, you can call the wasm function from [WasmEdge-Python](https://github.com/SAtacker/WasmEdge) like this:

Install WasmEdge python module:

```
pip install -i https://test.pypi.org/simple/ WasmEdge>=0.2.2
```


```python
import WasmEdge

vm = WasmEdge.VM()
vm.LoadWasmFile("say.wasm")
vm.Validate()
vm.Instantiate()
vm.Execute("say")

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
import WasmEdge
import WasmEdgeBindgen

vm = WasmEdge.VM()
vm.LoadWasmFile("say.wasm")
vm.Validate()

// Instantiate the bindgen and vm
bg = bindgen.Instantiate(vm)

// say: string -> string
no_error, result = bg.Execute("say", "wasmedge-bindgen")

if no_error:
    print("Run bindgen -- say:", result[0])
else:
    print(no_error.message())
```

## Explain

The result is a WasmEdge.Value array.
When the return value of wasm(rust) is (Tuple,) or Result<(Tuple,), String>,
the array length will be the same with the members' count of tuple.

If the function returns Err(String), you will get it by err.