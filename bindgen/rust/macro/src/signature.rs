use quote::quote;
use syn;

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

pub fn parse_returns(
    sig: &syn::Signature,
) -> (
    Vec<syn::Ident>,
    Vec<proc_macro2::TokenStream>,
    Vec<i32>,
    Vec<proc_macro2::TokenStream>,
    bool,
) {
    let mut ret_names = Vec::<syn::Ident>::new();
    let mut ret_pointers = Vec::<proc_macro2::TokenStream>::new();
    let mut ret_types = Vec::<i32>::new();
    let mut ret_sizes = Vec::<proc_macro2::TokenStream>::new();
    let mut is_rust_result = false;

    let mut prep_types = |seg: &syn::PathSegment, pos: usize| {
        let ret_name = quote::format_ident!("ret{}", pos.to_string());
        match seg.ident.to_string().as_str() {
            "u8" => {
                ret_pointers.push(quote! {{
                    let x = #ret_name.to_le_bytes()[..].to_vec();
                    std::mem::ManuallyDrop::new(x).as_ptr() as *const u8 as i32
                }});
                ret_names.push(ret_name);
                ret_types.push(RetTypes::U8 as i32);
                ret_sizes.push(quote! {
                    1
                });
            }
            "i8" => {
                ret_pointers.push(quote! {{
                    let x = #ret_name.to_le_bytes()[..].to_vec();
                    std::mem::ManuallyDrop::new(x).as_ptr() as *const i8 as i32
                }});
                ret_names.push(ret_name);
                ret_types.push(RetTypes::I8 as i32);
                ret_sizes.push(quote! {
                    1
                });
            }
            "u16" => {
                ret_pointers.push(quote! {{
                    let x = #ret_name.to_le_bytes()[..].to_vec();
                    std::mem::ManuallyDrop::new(x).as_ptr() as *const u16 as i32
                }});
                ret_names.push(ret_name);
                ret_types.push(RetTypes::U16 as i32);
                ret_sizes.push(quote! {
                    2
                });
            }
            "i16" => {
                ret_pointers.push(quote! {{
                    let x = #ret_name.to_le_bytes()[..].to_vec();
                    std::mem::ManuallyDrop::new(x).as_ptr() as *const i16 as i32
                }});
                ret_names.push(ret_name);
                ret_types.push(RetTypes::I16 as i32);
                ret_sizes.push(quote! {
                    2
                });
            }
            "u32" => {
                ret_pointers.push(quote! {{
                    let x = #ret_name.to_le_bytes()[..].to_vec();
                    std::mem::ManuallyDrop::new(x).as_ptr() as *const u32 as i32
                }});
                ret_names.push(ret_name);
                ret_types.push(RetTypes::U32 as i32);
                ret_sizes.push(quote! {
                    4
                });
            }
            "i32" => {
                ret_pointers.push(quote! {{
                    let x = #ret_name.to_le_bytes()[..].to_vec();
                    std::mem::ManuallyDrop::new(x).as_ptr() as *const i32 as i32
                }});
                ret_names.push(ret_name);
                ret_types.push(RetTypes::I32 as i32);
                ret_sizes.push(quote! {
                    4
                });
            }
            "u64" => {
                ret_pointers.push(quote! {{
                    let x = #ret_name.to_le_bytes()[..].to_vec();
                    std::mem::ManuallyDrop::new(x).as_ptr() as *const u64 as i32
                }});
                ret_names.push(ret_name);
                ret_types.push(RetTypes::U64 as i32);
                ret_sizes.push(quote! {
                    8
                });
            }
            "i64" => {
                ret_pointers.push(quote! {{
                    let x = #ret_name.to_le_bytes()[..].to_vec();
                    std::mem::ManuallyDrop::new(x).as_ptr() as *const i64 as i32
                }});
                ret_names.push(ret_name);
                ret_types.push(RetTypes::I64 as i32);
                ret_sizes.push(quote! {
                    8
                });
            }
            "f32" => {
                ret_pointers.push(quote! {{
                    let x = #ret_name.to_le_bytes()[..].to_vec();
                    std::mem::ManuallyDrop::new(x).as_ptr() as *const f32 as i32
                }});
                ret_names.push(ret_name);
                ret_types.push(RetTypes::F32 as i32);
                ret_sizes.push(quote! {
                    4
                });
            }
            "f64" => {
                ret_pointers.push(quote! {{
                    let x = #ret_name.to_le_bytes()[..].to_vec();
                    std::mem::ManuallyDrop::new(x).as_ptr() as *const f64 as i32
                }});
                ret_names.push(ret_name);
                ret_types.push(RetTypes::F64 as i32);
                ret_sizes.push(quote! {
                    8
                });
            }
            "bool" => {
                ret_pointers.push(quote! {{
                    let x = #ret_name.to_le_bytes()[..].to_vec();
                    std::mem::ManuallyDrop::new(x).as_ptr() as *const bool as i32
                }});
                ret_names.push(ret_name);
                ret_types.push(RetTypes::Bool as i32);
                ret_sizes.push(quote! {
                    1
                });
            }
            "char" => {
                ret_pointers.push(quote! {{
                    let x = #ret_name.to_le_bytes()[..].to_vec();
                    std::mem::ManuallyDrop::new(x).as_ptr() as *const char as i32
                }});
                ret_names.push(ret_name);
                ret_types.push(RetTypes::Char as i32);
                ret_sizes.push(quote! {
                    4
                });
            }
            "String" => {
                ret_pointers.push(quote! {
                    std::mem::ManuallyDrop::new(#ret_name).as_ptr() as i32
                });
                ret_types.push(RetTypes::String as i32);
                ret_sizes.push(quote! {
                    #ret_name.len() as i32
                });
                ret_names.push(ret_name);
            }
            "Vec" => match &seg.arguments {
                syn::PathArguments::AngleBracketed(args) => match args.args.first().unwrap() {
                    syn::GenericArgument::Type(arg_type) => match arg_type {
                        syn::Type::Path(arg_type_path) => {
                            let arg_seg = arg_type_path.path.segments.first().unwrap();
                            match arg_seg.ident.to_string().as_str() {
                                "u8" => {
                                    ret_pointers.push(quote! {
                                        std::mem::ManuallyDrop::new(#ret_name).as_ptr() as i32
                                    });
                                    ret_sizes.push(quote! {
                                        #ret_name.len() as i32
                                    });
                                    ret_types.push(RetTypes::U8Array as i32);
                                    ret_names.push(ret_name);
                                }
                                "i8" => {
                                    ret_pointers.push(quote! {
                                        std::mem::ManuallyDrop::new(#ret_name).as_ptr() as i32
                                    });
                                    ret_sizes.push(quote! {
                                        #ret_name.len() as i32
                                    });
                                    ret_types.push(RetTypes::I8Array as i32);
                                    ret_names.push(ret_name);
                                }
                                "u16" => {
                                    ret_pointers.push(quote! {
                                        std::mem::ManuallyDrop::new(#ret_name).as_ptr() as i32
                                    });
                                    ret_sizes.push(quote! {
                                        #ret_name.len() as i32 * 2
                                    });
                                    ret_types.push(RetTypes::U16Array as i32);
                                    ret_names.push(ret_name);
                                }
                                "i16" => {
                                    ret_pointers.push(quote! {
                                        std::mem::ManuallyDrop::new(#ret_name).as_ptr() as i32
                                    });
                                    ret_sizes.push(quote! {
                                        #ret_name.len() as i32 * 2
                                    });
                                    ret_types.push(RetTypes::I16Array as i32);
                                    ret_names.push(ret_name);
                                }
                                "u32" => {
                                    ret_pointers.push(quote! {
                                        std::mem::ManuallyDrop::new(#ret_name).as_ptr() as i32
                                    });
                                    ret_sizes.push(quote! {
                                        #ret_name.len() as i32 * 4
                                    });
                                    ret_types.push(RetTypes::U32Array as i32);
                                    ret_names.push(ret_name);
                                }
                                "i32" => {
                                    ret_pointers.push(quote! {
                                        std::mem::ManuallyDrop::new(#ret_name).as_ptr() as i32
                                    });
                                    ret_sizes.push(quote! {
                                        #ret_name.len() as i32 * 4
                                    });
                                    ret_types.push(RetTypes::I32Array as i32);
                                    ret_names.push(ret_name);
                                }
                                "u64" => {
                                    ret_pointers.push(quote! {
                                        std::mem::ManuallyDrop::new(#ret_name).as_ptr() as i32
                                    });
                                    ret_sizes.push(quote! {
                                        #ret_name.len() as i32 * 8
                                    });
                                    ret_types.push(RetTypes::U64Array as i32);
                                    ret_names.push(ret_name);
                                }
                                "i64" => {
                                    ret_pointers.push(quote! {
                                        std::mem::ManuallyDrop::new(#ret_name).as_ptr() as i32
                                    });
                                    ret_sizes.push(quote! {
                                        #ret_name.len() as i32 * 8
                                    });
                                    ret_types.push(RetTypes::I64Array as i32);
                                    ret_names.push(ret_name);
                                }
                                _ => {}
                            }
                        }
                        _ => {}
                    },
                    _ => {}
                },
                _ => {}
            },
            _ => {}
        }
    };

    match sig.output {
        syn::ReturnType::Type(_, ref rt) => match &**rt {
            syn::Type::Path(type_path) => {
                let seg = &type_path.path.segments.first().unwrap();
                let seg_type = seg.ident.to_string();
                if seg_type == "Result" {
                    is_rust_result = true;
                    match &seg.arguments {
                        syn::PathArguments::AngleBracketed(args) => {
                            match args.args.first().unwrap() {
                                syn::GenericArgument::Type(arg_type) => match arg_type {
                                    syn::Type::Path(arg_type_path) => {
                                        let arg_seg = arg_type_path.path.segments.first().unwrap();
                                        prep_types(&arg_seg, 0)
                                    }
                                    syn::Type::Tuple(arg_type_tuple) => {
                                        for (pos, elem) in arg_type_tuple.elems.iter().enumerate() {
                                            match elem {
                                                syn::Type::Path(type_path) => {
                                                    let seg =
                                                        &type_path.path.segments.first().unwrap();
                                                    prep_types(&seg, pos);
                                                }
                                                _ => {}
                                            }
                                        }
                                    }
                                    _ => {}
                                },
                                _ => {}
                            }
                        }
                        _ => {}
                    }
                } else {
                    prep_types(&seg, 0);
                }
            }
            syn::Type::Tuple(type_tuple) => {
                for (pos, elem) in type_tuple.elems.iter().enumerate() {
                    match elem {
                        syn::Type::Path(type_path) => {
                            let seg = &type_path.path.segments.first().unwrap();
                            prep_types(&seg, pos);
                        }
                        _ => {}
                    }
                }
            }
            _ => {}
        },
        _ => {}
    }

    (
        ret_names,
        ret_pointers,
        ret_types,
        ret_sizes,
        is_rust_result,
    )
}

pub fn parse_params(sig: &syn::Signature) -> (Vec<syn::Ident>, Vec<proc_macro2::TokenStream>) {
    let mut arg_names = Vec::<syn::Ident>::new();
    let mut arg_values = Vec::<proc_macro2::TokenStream>::new();

    let params_iter = sig.inputs.iter();
    for (pos, param) in params_iter.enumerate() {
        match param {
            syn::FnArg::Typed(param_type) => match &*param_type.ty {
                syn::Type::Path(type_path) => {
                    let seg = &type_path.path.segments.first().unwrap();
                    match seg.ident.to_string().as_str() {
                        "Vec" => match &seg.arguments {
                            syn::PathArguments::AngleBracketed(args) => match args
                                .args
                                .first()
                                .unwrap()
                            {
                                syn::GenericArgument::Type(arg_type) => match arg_type {
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
                                },
                                _ => {}
                            },
                            _ => {}
                        },
                        "bool" => {
                            arg_names.push(quote::format_ident!("arg{}", pos));
                            arg_values.push(quote! {
								Vec::from_raw_parts(pointer as *mut bool, size as usize, size as usize)[0]
							})
                        }
                        "char" => {
                            arg_names.push(quote::format_ident!("arg{}", pos));
                            arg_values.push(quote! {
								Vec::from_raw_parts(pointer as *mut char, size as usize, size as usize)[0]
							})
                        }
                        "i8" => {
                            arg_names.push(quote::format_ident!("arg{}", pos));
                            arg_values.push(quote! {
								Vec::from_raw_parts(pointer as *mut i8, size as usize, size as usize)[0]
							})
                        }
                        "u8" => {
                            arg_names.push(quote::format_ident!("arg{}", pos));
                            arg_values.push(quote! {
								Vec::from_raw_parts(pointer as *mut u8, size as usize, size as usize)[0]
							})
                        }
                        "i16" => {
                            arg_names.push(quote::format_ident!("arg{}", pos));
                            arg_values.push(quote! {
								Vec::from_raw_parts(pointer as *mut i16, size as usize, size as usize)[0]
							})
                        }
                        "u16" => {
                            arg_names.push(quote::format_ident!("arg{}", pos));
                            arg_values.push(quote! {
								Vec::from_raw_parts(pointer as *mut u16, size as usize, size as usize)[0]
							})
                        }
                        "i32" => {
                            arg_names.push(quote::format_ident!("arg{}", pos));
                            arg_values.push(quote! {
								Vec::from_raw_parts(pointer as *mut i32, size as usize, size as usize)[0]
							})
                        }
                        "u32" => {
                            arg_names.push(quote::format_ident!("arg{}", pos));
                            arg_values.push(quote! {
								Vec::from_raw_parts(pointer as *mut u32, size as usize, size as usize)[0]
							})
                        }
                        "i64" => {
                            arg_names.push(quote::format_ident!("arg{}", pos));
                            arg_values.push(quote! {
								Vec::from_raw_parts(pointer as *mut i64, size as usize, size as usize)[0]
							})
                        }
                        "u64" => {
                            arg_names.push(quote::format_ident!("arg{}", pos));
                            arg_values.push(quote! {
								Vec::from_raw_parts(pointer as *mut u64, size as usize, size as usize)[0]
							})
                        }
                        "f32" => {
                            arg_names.push(quote::format_ident!("arg{}", pos));
                            arg_values.push(quote! {
								Vec::from_raw_parts(pointer as *mut f32, size as usize, size as usize)[0]
							})
                        }
                        "f64" => {
                            arg_names.push(quote::format_ident!("arg{}", pos));
                            arg_values.push(quote! {
								Vec::from_raw_parts(pointer as *mut f64, size as usize, size as usize)[0]
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
                syn::Type::Reference(_) => {}
                syn::Type::Slice(_) => {}
                _ => {}
            },
            _ => {}
        }
    }

    (arg_names, arg_values)
}
