## About

This is a tiny library lets you call the WebAssembly exported functions that are tweaked by wasmedge-bindgen.

## Guide

Originally, you can call the wasm function from [WasmEdge-go](https://github.com/second-state/WasmEdge-go) like this:

```go
import (
	"github.com/second-state/WasmEdge-go/wasmedge"
)

func main() {
    vm := wasmedge.NewVM()
    vm.LoadWasmFile("say.wasm")
	vm.Validate()
	vm.Instantiate()
    vm.Execute("say")
}
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
```go
import (
	"github.com/second-state/WasmEdge-go/wasmedge"
	bindgen "github.com/second-state/wasmedge-bindgen/host/go"
)

func main() {
    vm := wasmedge.NewVM()
    vm.LoadWasmFile("say.wasm")
	vm.Validate()

	// Instantiate the bindgen and vm
	bg := bindgen.Instantiate(vm)

    /// say: string -> string
	result, err := bg.Execute("say", "wasmedge-bindgen")
    if err == nil {
        fmt.Println("Run bindgen -- say:", result[0].(string))
    }
}
```

## Explain

The result is an interface array ([]interface{}).
Normally, the length of the array is 1.
When the return value of wasm(rust) is (Tuple,) or Result<(Tuple,), String>,
the array length will be the same with the members' count of tuple.

If the function returns Err(String), you will get it by err.