use wasmedge_bindgen_macro::*;

#[wasmedge_bindgen]
fn say_hi(name: String) -> String {
    return name;
}
