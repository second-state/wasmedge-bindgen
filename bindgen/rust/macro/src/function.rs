use crate::signature::{parse_params, parse_returns};
use proc_macro::TokenStream;
use quote::{quote, ToTokens};

static mut FUNC_NUMBER: i32 = 0;

pub fn codegen_function_definition(mut ast: syn::ItemFn) -> TokenStream {
    let func_ident = ast.sig.ident;

    let ori_run: String;
    unsafe {
        ori_run = format!("run{}", FUNC_NUMBER);
        FUNC_NUMBER += 1;
    }
    let ori_run_ident = proc_macro2::Ident::new(ori_run.as_str(), proc_macro2::Span::call_site());
    ast.sig.ident = ori_run_ident.clone();

    let (arg_names, arg_values) = parse_params(&ast.sig);
    let (ret_names, ret_pointers, ret_types, ret_sizes, is_rust_result) = parse_returns(&ast.sig);
    let ret_len = ret_names.len();
    let ret_i = (0..ret_len).map(syn::Index::from);

    let params_len = arg_names.len();
    let i = (0..params_len).map(syn::Index::from);

    let ret_result = match is_rust_result {
        true => quote! {
            match #ori_run_ident(#(#arg_names),*) {
                Ok((#(#ret_names),*)) => {
                    let mut result_vec = vec![0; #ret_len * 3];
                    #(
                        result_vec[#ret_i * 3 + 2] = #ret_sizes;
                        result_vec[#ret_i * 3] = #ret_pointers;
                        result_vec[#ret_i * 3 + 1] = #ret_types;
                    )*
                    let result_vec = std::mem::ManuallyDrop::new(result_vec);
                    // return_result
                    let mut rvec = vec![0 as u8; 9];
                    rvec.splice(1..5, (result_vec.as_ptr() as i32).to_le_bytes());
                    rvec.splice(5..9, (#ret_len as i32).to_le_bytes());
                    let rvec = std::mem::ManuallyDrop::new(rvec);
                    return rvec.as_ptr() as i32;
                }
                Err(message) => {
                    let message = std::mem::ManuallyDrop::new(message);
                    // return_error
                    let mut rvec = vec![1 as u8; 9];
                    rvec.splice(1..5, (message.as_ptr() as i32).to_le_bytes());
                    rvec.splice(5..9, (message.len() as i32).to_le_bytes());
                    let rvec = std::mem::ManuallyDrop::new(rvec);
                    return rvec.as_ptr() as i32;
                }
            }
        },
        false => quote! {
            let (#(#ret_names),*) = #ori_run_ident(#(#arg_names),*);
            let mut result_vec = vec![0; #ret_len * 3];
            #(
                result_vec[#ret_i * 3 + 2] = #ret_sizes;
                result_vec[#ret_i * 3] = #ret_pointers;
                result_vec[#ret_i * 3 + 1] = #ret_types;
            )*
            let result_vec = std::mem::ManuallyDrop::new(result_vec);
            // return_result
            let mut rvec = vec![0 as u8; 9];
            rvec.splice(1..5, (result_vec.as_ptr() as i32).to_le_bytes());
            rvec.splice(5..9, (#ret_len as i32).to_le_bytes());
            let rvec = std::mem::ManuallyDrop::new(rvec);
            return rvec.as_ptr() as i32;
        },
    };

    let gen = quote! {

        #[no_mangle]
        pub unsafe extern "C" fn #func_ident(params_pointer: *mut u32, params_count: i32) -> i32 {
            if #params_len != params_count as usize {
                let err_msg = format!("Invalid params count, expect {}, got {}", #params_len, params_count);
                let err_msg = std::mem::ManuallyDrop::new(err_msg);
                // return_error
                let mut rvec = vec![1 as u8; 9];
                rvec.splice(1..5, (err_msg.as_ptr() as i32).to_le_bytes());
                rvec.splice(5..9, (err_msg.len() as i32).to_le_bytes());
                let rvec = std::mem::ManuallyDrop::new(rvec);
                return rvec.as_ptr() as i32;
            }

            #(
            let pointer = *params_pointer.offset(#i * 2) as *mut u8;
            let size= *params_pointer.offset(#i * 2 + 1);
            let #arg_names = #arg_values;
            )*

            #ret_result;
        }
    };

    let ori_run_str = ast.to_token_stream().to_string();
    let x = gen.to_string() + &ori_run_str;
    x.parse().unwrap()
}
