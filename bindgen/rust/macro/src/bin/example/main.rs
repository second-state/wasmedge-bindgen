use wasmedge_bindgen_macro::*;

#[wasmedge_bindgen("env")]
extern "C" {
    fn say_hi(name: String) -> String;
}

fn main() {}
