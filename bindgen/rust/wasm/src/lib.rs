use std::mem;

/// We hand over the the pointer to the allocated memory.
/// Caller has to ensure that the memory gets freed again.
#[no_mangle]
pub unsafe extern fn allocate(size: i32) -> *const u8 {
	let buffer = Vec::with_capacity(size as usize);

	let buffer = mem::ManuallyDrop::new(buffer);
	buffer.as_ptr() as *const u8
}

#[no_mangle]
pub unsafe extern fn deallocate(pointer: *mut u8, size: i32) {
	drop(Vec::from_raw_parts(pointer, size as usize, size as usize));
}

