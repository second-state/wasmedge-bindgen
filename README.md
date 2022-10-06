## About

Let WebAssembly's exported function support more data types for its parameters and return values.

## Example


## Export Rust things to host program

See demo program in [rust_bindgen_func](/test/BindgenFuncs/wasm/rust_bindgen_funcs/).

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

Run with command will compile the above code to a WASM bytes file.

```bash
cargo build --target wasm32-wasi --release
```

### Rust SDK 
Use exported functions in [Rust SDK](https://github.com/WasmEdge/WasmEdge/tree/master/bindings/rust/wasmedge-sdk).

Refer to the completed exmaple in [rust-sdk-test](/test/BindgenFuncs/host/rust-sdk-test/src/main.rs)

```rust

fn main() {
	....

    let vm = Vm::new(Some(config)).unwrap();
    let args: Vec<String> = env::args().collect();
    let wasm_path = Path::new(&args[1]);
    let module = Module::from_file(None, wasm_path).unwrap();
    let vm = vm.register_module(None, module).unwrap();
    let mut bg = Bindgen::new(vm);

    // create_line: string, string, string -> string (inputs are JSON stringified)
    let params = vec![
        Param::String("{\"x\":2.5,\"y\":7.8}"),
        Param::String("{\"x\":2.5,\"y\":5.8}"),
        Param::String("A thin red line"),
    ];
    match bg.run_wasm("create_line", params) {
        Ok(rv) => {
            println!(
                "Run bindgen -- create_line: {:?}",
                rv.unwrap().pop().unwrap().downcast::<String>().unwrap()
            );
        }
        Err(e) => {
            println!("Run bindgen -- create_line FAILED {:?}", e);
        }
    }

	...
}

```

### Go SDK 
Use exported Rust things from [WasmEdge-go](https://github.com/second-state/WasmEdge-go)!

Refer to the completed example in [bindgen_funcs.go](./test/BindgenFuncs/host/go/bindgen_funcs.go)

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

