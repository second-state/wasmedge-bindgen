use wasmedge_bindgen_macro::*;

#[wasmedge_bindgen("book")]
extern "C" {
    fn say_hi(name: String) -> String;
}

fn main() {
    unsafe {
        book_say_hi("A book".into());
    }
}
