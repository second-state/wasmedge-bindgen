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

                let (arg_names, arg_values) = parse_params(&f.sig);
                let (ret_names, ret_pointers, ret_types, ret_sizes, is_rust_result) =
                    parse_returns(&f.sig);
                let ret_len = ret_names.len();
                let ret_i = (0..ret_len).map(syn::Index::from);

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
                    fn #origin_ident(#(#arg_names : String),*) {
                        let mut params_pointer : u32 = 1;
                        let result = unsafe { #ori_run_ident(&mut params_pointer, #params_len as i32) };
                        println!("test result: {}", result);
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
