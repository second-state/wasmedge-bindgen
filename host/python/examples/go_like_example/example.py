import math
import os
import sys
from hashlib import sha3_256

import WasmEdge
from Crypto.Hash import keccak

from WasmEdgeBindgen import bindgen
from WasmEdgeBindgen.utils import uint_from_bytes

WasmEdge.Logging().error()

# Create configure
conf = WasmEdge.Configure()
conf.AddHostRegistration(WasmEdge.Host.Wasi)

# Create VM with configure
vm = WasmEdge.VM(conf)

# Init WASI
wasi = vm.GetImportModuleContext(WasmEdge.Host.Wasi)
wasi.InitWASI(
    tuple(
        sys.argv,
    ),  # The args
    tuple([f"{k}={v}" for k, v in os.environ.items()]),  # The envs
    tuple(
        [
            ".:.",
        ]
    ),  # The mapping preopens
)

# Load and validate the wasm
vm.LoadWasmFromFile(sys.argv[1])
vm.Validate()

# Instantiate the bindgen and vm
bg = bindgen.Bindgen(vm)

# create_line: string, string, string -> string (inputs are JSON stringified)
err, res = bg.execute(
    "create_line",
    '{"x":2.5,"y":7.8}',
    '{"x":2.5,"y":5.8}',
    "A thin red line",
)
assert err
print("Run bindgen -- create_line:" + bytes(res).decode("utf-8"))
bg.deallocator()

# say: string -> string
err, res = bg.execute("say", "bindgen funcs test")
assert err
print("Run bindgen -- say:" + bytes(res).decode("utf-8"))
bg.deallocator()

# obfusticate: string -> string
err, res = bg.execute(
    "obfusticate", "A quick brown fox jumps over the lazy dog"
)
assert err
print("Run bindgen -- obfusticate:" + bytes(res).decode("utf-8"))
bg.deallocator()

# lowest_common_multiple: i32, i32 -> i32
err, res = bg.execute("lowest_common_multiple", 123, 2)
assert err
lcm = uint_from_bytes(res)
assert lcm == abs(123 * 2) // math.gcd(123, 2)
print("Run bindgen -- lowest_common_multiple:" + str(lcm))
bg.deallocator()

# sha3_digest: array -> array
err, res = bg.execute("sha3_digest", "This is an important message")
assert err
assert bytes(res) == sha3_256(b"This is an important message").digest()
print("Run bindgen -- sha3_digest:" + bytes(res).hex())
bg.deallocator()


# keccak_digest: array -> array
err, res = bg.execute("keccak_digest", "This is an important message")
assert err
assert (
    bytes(res)
    == keccak.new(digest_bits=256)
    .update(b"This is an important message")
    .digest()
)
print("Run bindgen -- keccak_digest:" + bytes(res).hex())
bg.deallocator()
