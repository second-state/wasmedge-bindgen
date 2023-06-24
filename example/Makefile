all:
	@cargo build --target wasm32-wasi --manifest-path ./app/Cargo.toml
	@cargo build --target wasm32-wasi --manifest-path ./book/Cargo.toml
	@clang -I/usr/local/include -L/usr/local/lib -lwasmedge main.c
	@./a.out
