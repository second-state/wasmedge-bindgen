use std::mem;

/// We hand over the the pointer to the allocated memory.
/// Caller has to ensure that the memory gets freed again.
#[no_mangle]
pub unsafe extern fn allocate(size: i32) -> *const u8 {
	// TODO
	// Allocate with capacity of new should be equivalent but not
	// let buffer = Vec::with_capacity(size as usize);
	let buffer = vec![0; size as usize];
	assert_eq!(size, buffer.capacity() as i32);

	let mut buffer = mem::ManuallyDrop::new(buffer);
	let pointer = buffer.as_mut_ptr();

	pointer as *const u8
}

#[no_mangle]
pub unsafe extern fn deallocate(pointer: *mut u8, size: i32) {
	drop(Vec::from_raw_parts(pointer, size as usize, size as usize));
}

