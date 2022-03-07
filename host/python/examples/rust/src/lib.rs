use wasmedge_bindgen::*;
use wasmedge_bindgen_macro::*;

#[wasmedge_bindgen]
pub fn say_veci32(s: Vec<i32>) -> String {
    let stuff_str: String = format!("{:?}", s);
    // let k: String = map().to_string();
    let r = String::from("Rust:Hello--> ");
    return r + &stuff_str;
}

#[wasmedge_bindgen]
pub fn say_string(s: String) -> String {
    let r = String::from("Rust:Hello--> ");
    return r + &s;
}

#[wasmedge_bindgen]
pub fn say_number(s: i32) -> i32 {
    return s;
}
