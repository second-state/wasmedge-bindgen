extern crate proc_macro;
mod extern_mod;
mod function;
mod signature;

use extern_mod::codegen_foreign_module;
use function::codegen_function_definition;
use proc_macro::TokenStream;
use syn;

#[proc_macro_attribute]
pub fn wasmedge_bindgen(attr: TokenStream, item: TokenStream) -> TokenStream {
    let ast: syn::Item = syn::parse(item).unwrap();
    match ast {
        syn::Item::Fn(ast) => codegen_function_definition(ast),
        syn::Item::ForeignMod(fmod) => codegen_foreign_module(format!("{}", attr), fmod),
        _ => {
            unreachable!()
        }
    }
}
