use wasmedge_sys::*;
use std::env;
use std::path::Path;
use wasmedge_bindgen_host::*;

fn main() {
	let mut config = Config::create().unwrap();
	config.wasi(true);

	let mut vm = Vm::create(Some(config), None).unwrap();

	let args: Vec<String> = env::args().collect();
	let wasm_path = Path::new(&args[1]);
	let _ = vm.load_wasm_from_file(wasm_path);
	let _ = vm.validate();

	let mut bg = Bindgen::new(vm);

	// create_line: string, string, string -> string (inputs are JSON stringified)	
	let params = vec![Param::String("{\"x\":2.5,\"y\":7.8}".to_string()), Param::String("{\"x\":2.5,\"y\":5.8}".to_string()), Param::String("A thin red line".to_string())];
	match bg.run_wasm("create_line", params) {
		Ok(rv) => {
			println!("Run bindgen -- create_line: {:?}", rv.unwrap().pop().unwrap().downcast::<String>().unwrap());
		}
		Err(e) => {
			println!("Run bindgen -- create_line FAILED {:?}", e);
		}
	}

	let params = vec![Param::String("bindgen funcs test".to_string())];
	match bg.run_wasm("say", params) {
		Ok(rv) => {
			println!("Run bindgen -- say: {:?}", rv.unwrap().pop().unwrap().downcast::<String>().unwrap());
		}
		Err(e) => {
			println!("Run bindgen -- say FAILED {:?}", e);
		}
	}

	let params = vec![Param::String("A quick brown fox jumps over the lazy dog".to_string())];
	match bg.run_wasm("obfusticate", params) {
		Ok(rv) => {
			println!("Run bindgen -- obfusticate: {:?}", rv.unwrap().pop().unwrap().downcast::<String>().unwrap());
		}
		Err(e) => {
			println!("Run bindgen -- obfusticate FAILED {:?}", e);
		}
	}

	let params = vec![Param::I32(123), Param::I32(2)];
	match bg.run_wasm("lowest_common_multiple", params) {
		Ok(rv) => {
			println!("Run bindgen -- lowest_common_multiple: {:?}", rv.unwrap().pop().unwrap().downcast::<i32>().unwrap());
		}
		Err(e) => {
			println!("Run bindgen -- lowest_common_multiple FAILED {:?}", e);
		}
	}

	let params = vec![Param::VecU8("This is an important message".as_bytes().to_vec())];
	match bg.run_wasm("sha3_digest", params) {
		Ok(rv) => {
			println!("Run bindgen -- sha3_digest: {:?}", rv.unwrap().pop().unwrap().downcast::<Vec<u8>>().unwrap());
		}
		Err(e) => {
			println!("Run bindgen -- sha3_digest FAILED {:?}", e);
		}
	}

	let params = vec![Param::VecU8("This is an important message".as_bytes().to_vec())];
	match bg.run_wasm("keccak_digest", params) {
		Ok(rv) => {
			println!("Run bindgen -- keccak_digest: {:?}", rv.unwrap().pop().unwrap().downcast::<Vec<u8>>().unwrap());
		}
		Err(e) => {
			println!("Run bindgen -- keccak_digest FAILED {:?}", e);
		}
	}
}