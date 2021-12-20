extern crate proc_macro;
use proc_macro::TokenStream;
use quote::quote;
use syn;

#[proc_macro_attribute]
pub fn build_run(_: TokenStream, item: TokenStream) -> TokenStream {
    let input_str = item.to_string();
    let ast: syn::ItemFn = syn::parse(item).unwrap();

    let run_name = ast.sig.ident;

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
        extern "C" {
            fn return_result(result_pointer: *const u8, result_size: i32);
            fn return_error(result_pointer: *const u8, result_size: i32);
        }

        #[no_mangle]
        pub unsafe extern "C" fn run_e(params_pointer: *mut u32, params_count: i32) {
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


            match #run_name(#(#arg_names),*) {
                Ok(data) => {
                    return_result(data.as_ptr(), data.len() as i32);
                }
                Err(message) => {
                    return_error(message.as_ptr(), message.len() as i32);
                }
            }
        }
    };

    let x = gen.to_string() + &input_str;
    x.parse().unwrap()
}