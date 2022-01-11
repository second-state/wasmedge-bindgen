## About

This crate only export a macro named #[wasmedge_bindgen] that used for
retouching exporting functions to make it support more data types.

## Data Types

### Parameters

You can set the parameters to any one of the following types:

* Scalar Types: i8, u8, i16, u16, i32, u32, i64, u64, f32, f64, bool, char
* String
* Vec: Vec&lt;i8&gt;, Vec&lt;u8&gt;, Vec&lt;i16&gt;, Vec&lt;u16&gt;, Vec&lt;i32&gt;, Vec&lt;u32&gt;, Vec&lt;i64&gt;, Vec&lt;u64&gt;

### Return Values

You can set the return values to any one of the following types:

* Scalar Types: i8, u8, i16, u16, i32, u32, i64, u64, f32, f64, bool, char
* String
* Vec: Vec&lt;i8&gt;, Vec&lt;u8&gt;, Vec&lt;i16&gt;, Vec&lt;u16&gt;, Vec&lt;i32&gt;, Vec&lt;u32&gt;, Vec&lt;i64&gt;, Vec&lt;u64&gt;
* Tuple Type: compounded by any number of the above three types
* Result: Ok&lt;any one of the above four types&gt;, Err&lt;String&gt;

The only way to tell the host that the error has occurred is to return Err&lt;String&gt; of Result.

## Examples

```rust
#[wasmedge_bindgen]
pub fn create_line(p1: String, p2: String, desc: String) -> String

#[wasmedge_bindgen]
pub fn lowest_common_multiple(a: i32, b: i32) -> i32

#[wasmedge_bindgen]
pub fn sha3_digest(v: Vec<u8>) -> Vec<u8>

#[wasmedge_bindgen]
pub fn info(v: Vec<u8>) -> Result<(u8, String), String>
```