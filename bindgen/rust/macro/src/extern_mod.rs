use crate::signature::{parse_params, parse_returns};
use proc_macro::TokenStream;
use quote::quote;
use syn;

pub fn codegen_foreign_module(import_module_name: String, ast: syn::ItemForeignMod) -> TokenStream {
    let import_module_name = import_module_name.trim_matches('"');

    let mut ffi_list: Vec<String> = vec!["extern \"C\" {".into()];
    let mut wrap_fn_list: Vec<String> = vec![];

    for item in &ast.items {
        match item {
            syn::ForeignItem::Fn(f) => {
                let ori_run: String;
                ori_run = format!("{}_{}", import_module_name, f.sig.ident.clone());
                let ori_run_ident =
                    proc_macro2::Ident::new(ori_run.as_str(), proc_macro2::Span::call_site());

                let (arg_names, _arg_values) = parse_params(&f.sig);
                let (ret_names, _ret_pointers, _ret_types, _ret_sizes, _is_rust_result) =
                    parse_returns(&f.sig);
                let ret_len = ret_names.len();
                let _ret_i = (0..ret_len).map(syn::Index::from);

                // foreign function
                let gen_ffi = quote! {
                    #[no_mangle]
                    fn #ori_run_ident(params_pointer: *mut u32, params_count: i32) -> i32;
                };

                let origin_ident = proc_macro2::Ident::new(
                    f.sig.ident.clone().to_string().as_str(),
                    proc_macro2::Span::call_site(),
                );
                let params_len = arg_names.len();
                // shim function that call foreign function and fill the gap
                let wrap_ffi = quote! {
                    #[no_mangle]
                    fn #origin_ident(#(#arg_names : String),*) -> String {
                        let mut params_pointer : u32 = 1;
                        let rvec_pointer: *const Vec<u8> = unsafe { #ori_run_ident(&mut params_pointer, #params_len as i32) } as *const Vec<u8>;
                        let rvec: Vec<u8> = (*rvec_pointer).clone();
                        let flag = rvec[0];
                        let ret_pointer = i32::from_le_bytes(rvec[1..5].try_into().unwrap());
                        let ret_len = i32::from_le_bytes(rvec[5..9].try_into().unwrap());
                        match flag {
                            0 => String::from_utf8(Vec::from_raw_parts(ret_pointer as *mut u8, ret_len as usize, ret_len as usize)).unwrap(),
                            _ => "".to_string(),
                        }
                    }
                };

                ffi_list.push(gen_ffi.to_string());
                wrap_fn_list.push(wrap_ffi.to_string());
            }
            _ => unreachable!(),
        }
    }

    ffi_list.push("}".into());
    ffi_list.extend(wrap_fn_list);

    println!("{}", ffi_list.join("\n"));

    ffi_list.join("\n").parse().unwrap()
}
