use crate::signature::{parse_params, parse_returns};
use proc_macro::TokenStream;
use quote::{quote, ToTokens};

pub fn codegen_foreign_module(
    _import_module_name: String,
    ast: syn::ItemForeignMod,
) -> TokenStream {
    for item in &ast.items {
        match item {
            syn::ForeignItem::Fn(f) => {
                let (arg_names, arg_values) = parse_params(&f.sig);
                let (ret_names, ret_pointers, ret_types, ret_sizes, is_rust_result) =
                    parse_returns(&f.sig);
                let ret_len = ret_names.len();
                let ret_i = (0..ret_len).map(syn::Index::from);

                let params_len = arg_names.len();
                let i = (0..params_len).map(syn::Index::from);
            }
            _ => unreachable!(),
        }
        // "String" => {
        // 	ret_pointers.push(quote! {
        // 		std::mem::ManuallyDrop::new(#ret_name).as_ptr() as i32
        // 	});
        // 	ret_types.push(RetTypes::String as i32);
        // 	ret_sizes.push(quote! {
        // 		#ret_name.len() as i32
        // 	});
        // 	ret_names.push(ret_name);
        // }
    }

    let gen = quote! { fn useless() {}};
    let x = gen.to_string() + &ast.to_token_stream().to_string();
    x.parse().unwrap()
}
