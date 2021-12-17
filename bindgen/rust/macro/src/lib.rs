extern crate proc_macro;
use proc_macro::TokenStream;

#[proc_macro]
pub fn build_run(_item: TokenStream) -> TokenStream {
    r#"
    extern "C" {
        fn return_result(result_pointer: *const u8, result_size: i32);
        // fn return_error(code: i32, result_pointer: *const u8, result_size: i32);
    }

    #[no_mangle]
    pub unsafe extern "C" fn run_e(pointer: *mut u8, size: i32) {
        // rebuild the memory into something usable
        let in_bytes = Vec::from_raw_parts(pointer, size as usize, size as usize);

        match run(in_bytes) {
            Ok(data) => {
                return_result(data.as_ptr(), data.len() as i32);
            }
            Err(_e) => {
                // return_error(code, message.as_ptr(), message.len() as i32);
            }
        }
    }"#.parse().unwrap()
}