extern crate proc_macro;

use proc_macro::TokenStream;
use quote::{quote, ToTokens};
use syn;

static mut FUNC_NUMBER: i32 = 0;

#[proc_macro_attribute]
pub fn wasmedge_bindgen(_: TokenStream, item: TokenStream) -> TokenStream {
    let mut ast: syn::ItemFn = syn::parse(item).unwrap();

    let func_ident = ast.sig.ident;

    let ori_run: String;
    unsafe {
        ori_run = format!("run{}", FUNC_NUMBER);
        FUNC_NUMBER += 1;
    }
    let ori_run_ident = proc_macro2::Ident::new(ori_run.as_str(), proc_macro2::Span::call_site());
    ast.sig.ident = ori_run_ident.clone();

    let mut arg_names = Vec::<syn::Ident>::new();
    let mut arg_values = Vec::<proc_macro2::TokenStream>::new();

    let params_iter = ast.sig.inputs.iter();
    for (pos, param) in params_iter.enumerate() {
        match param {
            syn::FnArg::Typed(param_type) => {
                match &*param_type.ty {
                    syn::Type::Path(type_path) => {
                        let seg = &type_path.path.segments.first().unwrap();
                        match seg.ident.to_string().as_str() {
                            "Vec" => {
                                match &seg.arguments {
                                    syn::PathArguments::AngleBracketed(args) => {
                                        match args.args.first().unwrap() {
                                            syn::GenericArgument::Type(arg_type) => {
                                                match arg_type {
                                                    syn::Type::Path(arg_type_path) => {
                                                        let arg_seg = arg_type_path.path.segments.first().unwrap();
                                                        match arg_seg.ident.to_string().as_str() {
                                                            "u8" => {
                                                                arg_names.push(quote::format_ident!("arg{}", pos));
                                                                arg_values.push(quote! {
                                                                    Vec::from_raw_parts(pointer, size as usize, size as usize)
                                                                })
                                                            }
                                                            "i8" => {
                                                                arg_names.push(quote::format_ident!("arg{}", pos));
                                                                arg_values.push(quote! {
                                                                    Vec::from_raw_parts(pointer as *mut i8, size as usize, size as usize)
                                                                })
                                                            }
                                                            "u16" => {
                                                                arg_names.push(quote::format_ident!("arg{}", pos));
                                                                arg_values.push(quote! {
                                                                    Vec::from_raw_parts(pointer as *mut u16, size as usize, size as usize)
                                                                })
                                                            }
                                                            "i16" => {
                                                                arg_names.push(quote::format_ident!("arg{}", pos));
                                                                arg_values.push(quote! {
                                                                    Vec::from_raw_parts(pointer as *mut i16, size as usize, size as usize)
                                                                })
                                                            }
                                                            "u32" => {
                                                                arg_names.push(quote::format_ident!("arg{}", pos));
                                                                arg_values.push(quote! {
                                                                    Vec::from_raw_parts(pointer as *mut u32, size as usize, size as usize)
                                                                })
                                                            }
                                                            "i32" => {
                                                                arg_names.push(quote::format_ident!("arg{}", pos));
                                                                arg_values.push(quote! {
                                                                    Vec::from_raw_parts(pointer as *mut i32, size as usize, size as usize)
                                                                })
                                                            }
                                                            "u64" => {
                                                                arg_names.push(quote::format_ident!("arg{}", pos));
                                                                arg_values.push(quote! {
                                                                    Vec::from_raw_parts(pointer as *mut u64, size as usize, size as usize)
                                                                })
                                                            }
                                                            "i64" => {
                                                                arg_names.push(quote::format_ident!("arg{}", pos));
                                                                arg_values.push(quote! {
                                                                    Vec::from_raw_parts(pointer as *mut i64, size as usize, size as usize)
                                                                })
                                                            }
                                                            _ => {}
                                                        }
                                                    }
                                                    _ => {}
                                                }
                                            }
                                            _ => {}
                                        }
                                    }
                                    _ => {}
                                }
                            }
                            "bool" => {
                                arg_names.push(quote::format_ident!("arg{}", pos));
                                arg_values.push(quote! {
                                    *(pointer as *const bool)
                                })
                            }
                            "char" => {
                                arg_names.push(quote::format_ident!("arg{}", pos));
                                arg_values.push(quote! {
                                    *(pointer as *const char)
                                })
                            }
                            "i8" => {
                                arg_names.push(quote::format_ident!("arg{}", pos));
                                arg_values.push(quote! {
                                    *(pointer as *const i8)
                                })
                            }
                            "u8" => {
                                arg_names.push(quote::format_ident!("arg{}", pos));
                                arg_values.push(quote! {
                                    *(pointer as *const u8)
                                })
                            }
                            "i16" => {
                                arg_names.push(quote::format_ident!("arg{}", pos));
                                arg_values.push(quote! {
                                    *(pointer as *const i16)
                                })
                            }
                            "u16" => {
                                arg_names.push(quote::format_ident!("arg{}", pos));
                                arg_values.push(quote! {
                                    *(pointer as *const u16)
                                })
                            }
                            "i32" => {
                                arg_names.push(quote::format_ident!("arg{}", pos));
                                arg_values.push(quote! {
                                    *(pointer as *const i32)
                                })
                            }
                            "u32" => {
                                arg_names.push(quote::format_ident!("arg{}", pos));
                                arg_values.push(quote! {
                                    *(pointer as *const u32)
                                })
                            }
                            "i64" => {
                                arg_names.push(quote::format_ident!("arg{}", pos));
                                arg_values.push(quote! {
                                    *(pointer as *const i64)
                                })
                            }
                            "u64" => {
                                arg_names.push(quote::format_ident!("arg{}", pos));
                                arg_values.push(quote! {
                                    *(pointer as *const u64)
                                })
                            }
                            "f32" => {
                                arg_names.push(quote::format_ident!("arg{}", pos));
                                arg_values.push(quote! {
                                    *(pointer as *const f32)
                                })
                            }
                            "f64" => {
                                arg_names.push(quote::format_ident!("arg{}", pos));
                                arg_values.push(quote! {
                                    *(pointer as *const f64)
                                })
                            }
                            "String" => {
                                arg_names.push(quote::format_ident!("arg{}", pos));
                                arg_values.push(quote! {
                                    std::str::from_utf8(&Vec::from_raw_parts(pointer, size as usize, size as usize)).unwrap().to_string()
                                })
                            }
                            _ => {}
                        }
                    }
                    syn::Type::Reference(_) => {

                    }
                    syn::Type::Slice(_) => {

                    }
                    _ => {}
                }
            }
            _ => {}
        }
    }

    let params_len = arg_names.len();
    let i = (0..params_len).map(syn::Index::from);

    let gen = quote! {

        #[no_mangle]
        pub unsafe extern "C" fn #func_ident(params_pointer: *mut u32, params_count: i32) {

            #[link(wasm_import_module = "wasmedge-bindgen")]
            extern "C" {
                fn return_result(result_pointer: *const u8, result_size: i32);
                fn return_error(result_pointer: *const u8, result_size: i32);
            }

            if #params_len != params_count as usize {
                let err_msg = format!("Invalid params count, expect {}, got {}", #params_len, params_count);
                return_error(err_msg.as_ptr(), err_msg.len() as i32);
                return;
            }

            #(
            let pointer = *params_pointer.offset(#i * 2) as *mut u8;
            let size= *params_pointer.offset(#i * 2 + 1);
            let #arg_names = #arg_values;
            )*


            match #ori_run_ident(#(#arg_names),*) {
                Ok(data) => {
                    return_result(data.as_ptr(), data.len() as i32);
                }
                Err(message) => {
                    return_error(message.as_ptr(), message.len() as i32);
                }
            }
        }
    };

    let ori_run_str = ast.to_token_stream().to_string();
    let x = gen.to_string() + &ori_run_str;
    x.parse().unwrap()
}