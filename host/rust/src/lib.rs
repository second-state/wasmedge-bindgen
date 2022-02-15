use std::any::Any;
use std::sync::mpsc::{channel, Sender, Receiver};

use num_derive::FromPrimitive;    
use num_traits::FromPrimitive;
use wasmedge_sys::*;

pub enum Param {
	I8(i8),
	U8(u8),
	I16(i16),
	U16(u16),
	I32(i32),
	U32(u32),
	I64(i64),
	U64(u64),
	F32(f32),
	F64(f64),
	Bool(bool),
	VecI8(Vec<i8>),
	VecU8(Vec<u8>),
	VecI16(Vec<i16>),
	VecU16(Vec<u16>),
	VecI32(Vec<i32>),
	VecU32(Vec<u32>),
	VecI64(Vec<i64>),
	VecU64(Vec<u64>),
	String(String),
}

impl Param {
	fn settle(&self, vm: &mut Vm, mem: &mut Memory) -> Result<(i32, i32, i32), WasmEdgeError> {
		match self {
			Param::I8(v) => {
				let length = 1;
				let pointer = allocate(vm, length)?;
				mem.set_data(vec![*v as u8], pointer as u32)?;
				Ok((pointer, length, length))
			}
			Param::U8(v) => {
				let length = 1;
				let pointer = allocate(vm, length)?;
				mem.set_data(vec![*v], pointer as u32)?;
				Ok((pointer, length, length))
			}
			Param::I16(v) => {
				let length = 1;
				let pointer = allocate(vm, length * 2)?;
				let bytes = v.to_le_bytes();
				mem.set_data(bytes, pointer as u32)?;
				Ok((pointer, length, length * 2))
			}
			Param::U16(v) => {
				let length = 2;
				let pointer = allocate(vm, length * 2)?;
				let bytes = v.to_le_bytes();
				mem.set_data(bytes, pointer as u32)?;
				Ok((pointer, length, length * 2))
			}
			Param::I32(v) => {
				let length = 1;
				let pointer = allocate(vm, length * 4)?;
				let bytes = v.to_le_bytes();
				mem.set_data(bytes, pointer as u32)?;
				Ok((pointer, length, length * 4))
			}
			Param::U32(v) => {
				let length = 1;
				let pointer = allocate(vm, length * 4)?;
				let bytes = v.to_le_bytes();
				mem.set_data(bytes, pointer as u32)?;
				Ok((pointer, length, length * 4))
			}
			Param::I64(v) => {
				let length = 1;
				let pointer = allocate(vm, length * 8)?;
				let bytes = v.to_le_bytes();
				mem.set_data(bytes, pointer as u32)?;
				Ok((pointer, length, length * 8))
			}
			Param::U64(v) => {
				let length = 1;
				let pointer = allocate(vm, length * 8)?;
				let bytes = v.to_le_bytes();
				mem.set_data(bytes, pointer as u32)?;
				Ok((pointer, length, length * 8))
			}
			Param::F32(v) => {
				let length = 1;
				let pointer = allocate(vm, length * 4)?;
				let bytes = v.to_le_bytes();
				mem.set_data(bytes, pointer as u32)?;
				Ok((pointer, length, length * 4))
			}
			Param::F64(v) => {
				let length = 1;
				let pointer = allocate(vm, length * 8)?;
				let bytes = v.to_le_bytes();
				mem.set_data(bytes, pointer as u32)?;
				Ok((pointer, length, length * 8))
			}
			Param::Bool(v) => {
				let length = 1;
				let pointer = allocate(vm, length)?;
				let byte: u8 = match v {
					true => 1,
					false => 0
				};
				mem.set_data(vec![byte], pointer as u32)?;
				Ok((pointer, length, length))
			}
			Param::VecI8(v) => {
				let length = v.len() as i32;
				let pointer = allocate(vm, length)?;
				let mut bytes = vec![0; length as usize];
				for (pos, iv) in v.iter().enumerate() {
					bytes[pos] = *iv as u8;
				}
				mem.set_data(bytes, pointer as u32)?;
				Ok((pointer, length, length))
			}
			Param::VecU8(v) => {
				let length = v.len() as i32;
				let pointer = allocate(vm, length)?;
				let mut bytes = vec![0; length as usize];
				for (pos, iv) in v.iter().enumerate() {
					bytes[pos] = *iv;
				}
				mem.set_data(bytes, pointer as u32)?;
				Ok((pointer, length, length))
			}
			Param::VecI16(v) => {
				let length = v.len() as i32;
				let pointer = allocate(vm, length * 2)?;
				let mut bytes = vec![0; length as usize * 2];
				for (pos, iv) in v.iter().enumerate() {
					let b = iv.to_le_bytes();
					for i in 0..2 {
						bytes[pos * 2 + i] = b[i];
					}
				}
				mem.set_data(bytes, pointer as u32)?;
				Ok((pointer, length, length * 2))
			}
			Param::VecU16(v) => {
				let length = v.len() as i32;
				let pointer = allocate(vm, length * 2)?;
				let mut bytes = vec![0; length as usize * 2];
				for (pos, iv) in v.iter().enumerate() {
					let b = iv.to_le_bytes();
					for i in 0..2 {
						bytes[pos * 2 + i] = b[i];
					}
				}
				mem.set_data(bytes, pointer as u32)?;
				Ok((pointer, length, length * 2))
			}
			Param::VecI32(v) => {
				let length = v.len() as i32;
				let pointer = allocate(vm, length * 4)?;
				let mut bytes = vec![0; length as usize * 4];
				for (pos, iv) in v.iter().enumerate() {
					let b = iv.to_le_bytes();
					for i in 0..4 {
						bytes[pos * 4 + i] = b[i];
					}
				}
				mem.set_data(bytes, pointer as u32)?;
				Ok((pointer, length, length * 4))
			}
			Param::VecU32(v) => {
				let length = v.len() as i32;
				let pointer = allocate(vm, length * 4)?;
				let mut bytes = vec![0; length as usize * 4];
				for (pos, iv) in v.iter().enumerate() {
					let b = iv.to_le_bytes();
					for i in 0..4 {
						bytes[pos * 4 + i] = b[i];
					}
				}
				mem.set_data(bytes, pointer as u32)?;
				Ok((pointer, length, length * 4))
			}
			Param::VecI64(v) => {
				let length = v.len() as i32;
				let pointer = allocate(vm, length * 8)?;
				let mut bytes = vec![0; length as usize * 8];
				for (pos, iv) in v.iter().enumerate() {
					let b = iv.to_le_bytes();
					for i in 0..8 {
						bytes[pos * 8 + i] = b[i];
					}
				}
				mem.set_data(bytes, pointer as u32)?;
				Ok((pointer, length, length * 8))
			}
			Param::VecU64(v) => {
				let length = v.len() as i32;
				let pointer = allocate(vm, length * 8)?;
				let mut bytes = vec![0; length as usize * 8];
				for (pos, iv) in v.iter().enumerate() {
					let b = iv.to_le_bytes();
					for i in 0..8 {
						bytes[pos * 8 + i] = b[i];
					}
				}
				mem.set_data(bytes, pointer as u32)?;
				Ok((pointer, length, length * 8))
			}
			Param::String(v) => {
				let bytes = v.as_bytes().to_vec();
				let length = bytes.len() as i32;
				let pointer = allocate(vm, length)?;
				mem.set_data(bytes, pointer as u32)?;
				Ok((pointer, length, length))
			}
		}
	}
}

fn allocate(vm: &mut Vm, size: i32) -> Result<i32, WasmEdgeError> {
	match vm.run_function("allocate", vec![Value::from_i32(size)]) {
		Ok(rv) => {
			Ok(rv[0].to_i32())
		}
		Err(e) => {
			Err(e)
		}
	}
}

#[derive(FromPrimitive)]
enum RetTypes {
	U8 = 1,
	I8 = 2,
	U16 = 3,
	I16 = 4,
	U32 = 5,
	I32 = 6,
	U64 = 7,
	I64 = 8,
	F32 = 9,
	F64 = 10,
	Bool = 11,
	Char = 12,
	U8Array = 21,
	I8Array = 22,
	U16Array = 23,
	I16Array = 24,
	U32Array = 25,
	I32Array = 26,
	U64Array = 27,
	I64Array = 28,
	String = 31,
}

#[derive(Debug)]
pub struct Bindgen<'a> {
	vm: &'a mut Vm,
	tx: Sender<Result<Vec<Box<dyn Any>>, String>>,
	rx: Receiver<Result<Vec<Box<dyn Any>>, String>>,
}

impl<'a> Bindgen<'a> {
	pub fn new(vm: &'a mut Vm) -> Self {
		let (tx, rx): (Sender<Result<Vec<Box<dyn Any>>, String>>, Receiver<Result<Vec<Box<dyn Any>>, String>>) = channel();

		let mut b = Bindgen {
			vm: vm,
			tx: tx,
			rx: rx,
		};

		b.init();

		let _ = b.vm.instantiate();

		b
	}

	fn init(&mut self) {
		let mut imp_obj = ImportObject::create("wasmedge-bindgen").unwrap();

		// create a FuncType
		let func_ty = FuncType::create(vec![ValType::I32; 2], vec![]).expect("fail to create a FuncType");
		// create a Function instance
		let func = Function::create(func_ty, Box::new(return_result(self.tx.clone(), self.vm.store_mut().unwrap())), 0).expect("fail to create a Function instance");
		imp_obj.add_func("return_result", func);

		// create a FuncType
		let func_ty = FuncType::create(vec![ValType::I32; 2], vec![]).expect("fail to create a FuncType");
		// create a Function instance
		let func = Function::create(func_ty, Box::new(return_error(self.tx.clone(), self.vm.store_mut().unwrap())), 0).expect("fail to create a Function instance");
		imp_obj.add_func("return_error", func);

		let _ = self.vm.register_wasm_from_import(imp_obj);
	}

	pub fn run_wasm(&mut self, func_name: impl AsRef<str>, inputs: Vec<Param>) -> Result<Result<Vec<Box<dyn Any>>, String>, WasmEdgeError> {
		let inputs_count = inputs.len() as i32;

		// allocate new frame for passing pointers
		let pointer_of_pointers = match self.vm.run_function("allocate", vec![Value::from_i32(inputs_count * 4 * 2)]) {
			Ok(rv) => {
				rv[0].to_i32()
			}
			Err(e) => {
				return Err(e);
			}
		};

		let mut memory = self.vm.store_mut().unwrap().find_memory("memory").unwrap();

		let mut pv: Vec<(i32, i32)> = vec![];

		for (pos, inp) in inputs.iter().enumerate() {
			let sr = inp.settle(self.vm, &mut memory);

			let (pointer, length_of_input, byte_length_of_input) = match sr {
				Ok(r) => {
					(r.0, r.1, r.2)
				}
				Err(e) => {
					return Err(e);
				}
			};

			memory.set_data(pointer.to_le_bytes(), pointer_of_pointers as u32 + pos as u32 * 4 * 2)?;
			memory.set_data(length_of_input.to_le_bytes(), pointer_of_pointers as u32 + pos as u32 * 4 * 2 + 4)?;
			pv.push((pointer, byte_length_of_input));
		}

		self.vm.run_function(func_name, vec![Value::from_i32(pointer_of_pointers), Value::from_i32(inputs_count)])?;

		for p in pv {
			self.vm.run_function("deallocate", vec![Value::from_i32(p.0), Value::from_i32(p.1)])?;
		}
		self.vm.run_function("deallocate", vec![Value::from_i32(pointer_of_pointers), Value::from_i32(inputs_count * 4 * 2)])?;

		let result = self.rx.recv().unwrap();

		return Ok(result);
	}
}

fn return_result(tx: Sender<Result<Vec<Box<dyn Any>>, String>>, store: Store) -> impl Fn(Vec<Value>) -> Result<Vec<Value>, u8> {
	let f = move |inputs: Vec<Value>| -> Result<Vec<Value>, u8> {
		let memory = store.find_memory("memory").unwrap();
		let size = inputs[1].to_i32() as usize;
		let p_data = memory.get_data(inputs[0].to_i32() as u32, size as u32 * 3 * 4).unwrap();

		let mut p_values = vec![0; size * 3];

		for i in 0..size * 3 {
			p_values[i] = i32::from_le_bytes(p_data[i*4..(i+1)*4].try_into().unwrap());
		}

		let mut results: Vec<Box<dyn Any>> = Vec::with_capacity(size);

		for i in 0..size {
			let bytes = memory.get_data(p_values[i*3] as u32, p_values[i*3+2] as u32).unwrap();
			match FromPrimitive::from_i32(p_values[i*3+1]) {
				Some(RetTypes::U8) => {
					results.push(Box::new(bytes[0]));
				}
				Some(RetTypes::I8) => {
					results.push(Box::new(bytes[0] as i8));
				}
				Some(RetTypes::U16) => {
					let v = u16::from_le_bytes(bytes.try_into().unwrap());
					results.push(Box::new(v));
				}
				Some(RetTypes::I16) => {
					let v = i16::from_le_bytes(bytes.try_into().unwrap());
					results.push(Box::new(v));
				}
				Some(RetTypes::U32) => {
					let v = u32::from_le_bytes(bytes.try_into().unwrap());
					results.push(Box::new(v));
				}
				Some(RetTypes::I32) => {
					let v = i32::from_le_bytes(bytes.try_into().unwrap());
					results.push(Box::new(v));
				}
				Some(RetTypes::U64) => {
					let v = u64::from_le_bytes(bytes.try_into().unwrap());
					results.push(Box::new(v));
				}
				Some(RetTypes::I64) => {
					let v = i64::from_le_bytes(bytes.try_into().unwrap());
					results.push(Box::new(v));
				}
				Some(RetTypes::F32) => {
					let v = f32::from_le_bytes(bytes.try_into().unwrap());
					results.push(Box::new(v));
				}
				Some(RetTypes::F64) => {
					let v = f64::from_le_bytes(bytes.try_into().unwrap());
					results.push(Box::new(v));
				}
				Some(RetTypes::Bool) => {
					results.push(Box::new(bytes[0] == 1 as u8));
				}
				Some(RetTypes::Char) => {
					let v = u32::from_le_bytes(bytes.try_into().unwrap());
					results.push(Box::new(char::from_u32(v)));
				}
				Some(RetTypes::U8Array) => {
					results.push(Box::new(bytes));
				}
				Some(RetTypes::I8Array) => {
					let len = bytes.len();
					let mut v = vec![0; len];
					for i in 0..len {
						v[i] = bytes[i] as i8;
					}
					results.push(Box::new(v));
				}
				Some(RetTypes::U16Array) => {
					let len = bytes.len() / 2;
					let mut v = vec![0; len];
					for i in 0..len {
						v[i] = u16::from_le_bytes(bytes[i*2..(i+1)*2].try_into().unwrap());
					}
					results.push(Box::new(v));
				}
				Some(RetTypes::I16Array) => {
					let len = bytes.len() / 2;
					let mut v = vec![0; len];
					for i in 0..len {
						v[i] = i16::from_le_bytes(bytes[i*2..(i+1)*2].try_into().unwrap());
					}
					results.push(Box::new(v));
				}
				Some(RetTypes::U32Array) => {
					let len = bytes.len() / 4;
					let mut v = vec![0; len];
					for i in 0..len {
						v[i] = u32::from_le_bytes(bytes[i*4..(i+1)*4].try_into().unwrap());
					}
					results.push(Box::new(v));
				}
				Some(RetTypes::I32Array) => {
					let len = bytes.len() / 4;
					let mut v = vec![0; len];
					for i in 0..len {
						v[i] = i32::from_le_bytes(bytes[i*4..(i+1)*4].try_into().unwrap());
					}
					results.push(Box::new(v));
				}
				Some(RetTypes::U64Array) => {
					let len = bytes.len() / 8;
					let mut v = vec![0; len];
					for i in 0..len {
						v[i] = u64::from_le_bytes(bytes[i*8..(i+1)*8].try_into().unwrap());
					}
					results.push(Box::new(v));
				}
				Some(RetTypes::I64Array) => {
					let len = bytes.len() / 8;
					let mut v = vec![0; len];
					for i in 0..len {
						v[i] = i64::from_le_bytes(bytes[i*8..(i+1)*8].try_into().unwrap());
					}
					results.push(Box::new(v));
				}
				Some(RetTypes::String) => {
					results.push(Box::new(String::from_utf8(bytes).unwrap()));
				}
				None => {}
			}
		}

		tx.send(Ok(results)).unwrap();

		Ok(vec![])
	};
	return f;
}

fn return_error(tx: Sender<Result<Vec<Box<dyn Any>>, String>>, store: Store) -> impl Fn(Vec<Value>) -> Result<Vec<Value>, u8> {
	let f = move |inputs: Vec<Value>| -> Result<Vec<Value>, u8> {
		let memory = store.find_memory("memory").unwrap();
		let err_bytes = memory.get_data(inputs[0].to_i32() as u32, inputs[1].to_i32() as u32).unwrap();
		tx.send(Err(String::from_utf8_lossy(&err_bytes).into_owned())).unwrap();

		Ok(vec![])
	};
	return f;
}