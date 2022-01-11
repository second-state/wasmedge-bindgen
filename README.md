## About

Let WebAssembly's exported function support more data types for its parameters and return values.

## Example

Export Rust things to host program.

```rust
use wasmedge_bindgen::*;
use wasmedge_bindgen_macro::*;

// Export a `say` function from Rust, that returns a hello message.
#[wasmedge_bindgen]
pub fn say(s: String) -> String {
	let r = String::from("hello ");
	return r + s.as_str();
}
```

Use exported Rust things from [WasmEdge-go](https://github.com/second-state/WasmEdge-go)!

```go
import (
	"github.com/second-state/WasmEdge-go/wasmedge"
	bindgen "github.com/second-state/wasmedge-bindgen/host/go"
)

func main() {
	.
	.
	.

	// Instantiate the bindgen and vm
	bg := bindgen.Instantiate(vm)

		/// say: string -> string
	res, _ := bg.Execute("say", "wasmedge-bindgen")
	fmt.Println("Run bindgen -- say:", res[0].(string))
}
```

## Guide

[**Find all the data types you can use!**](bindgen/rust/macro)

You can find how to call the exported function from go host [here](host/go).

