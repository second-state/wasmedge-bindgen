extern crate proc_macro;
use proc_macro::TokenStream;
use quote::quote;
use syn;

#[proc_macro_attribute]
pub fn build_run(_: TokenStream, item: TokenStream) -> TokenStream {
    let input_str = item.to_string();
    let ast: syn::ItemFn = syn::parse(item).unwrap();

    let run_name = ast.sig.ident;

    let mut parsed_params = Vec::<&str>::new();

    let params_iter = ast.sig.inputs.iter();
    for param in params_iter {
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
                                                                parsed_params.push("vec<u8>");
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

    let params_len = parsed_params.len();
    let i = (0..params_len).rev().map(syn::Index::from);
    let i2 = (0..params_len).rev().map(syn::Index::from);

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
            let mut bytes: Vec<Vec<u8>> = vec!();
            #(
            let pointer = *params_pointer.offset(#i * 2) as *mut u8;
            let size= *params_pointer.offset(#i * 2 + 1);
            bytes.push(Vec::from_raw_parts(pointer, size as usize, size as usize));
            )*

            match #run_name(#(bytes.remove(#i2)),*) {
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