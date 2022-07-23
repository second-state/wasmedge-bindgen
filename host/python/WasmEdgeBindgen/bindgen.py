import WasmEdge

from .consts import I32, ByteArray, I32Array, String
from .utils import int_to_bytes, uint_from_bytes, uint_to_bytes


class Bindgen:
    def __init__(
        self,
        VM,
    ):
        self.vm = VM
        self.result = None
        self.output = None
        self.funcImports = WasmEdge.ImportObject("wasmedge-bindgen")

        result_function_type = WasmEdge.FunctionType(
            [WasmEdge.Type.I32, WasmEdge.Type.I32], []
        )

        def _return_result(ptr, size):
            memory = self.vm.GetStoreContext().FindMemory("memory")
            size = size.Value
            res, address = memory.GetData(size * 4, ptr.Value)
            if __debug__:
                assert res

            res, type = memory.GetData(size * 4, ptr.Value + 4)
            if __debug__:
                assert res

            if uint_from_bytes(type) in [String, ByteArray]:
                res, ptr_size = memory.GetData(size * 4, ptr.Value + 8)
                if __debug__:
                    assert res

                res, data = memory.GetData(
                    uint_from_bytes(ptr_size),
                    uint_from_bytes(address),
                )
                assert res
                self.result = res
                self.output = data
            elif uint_from_bytes(type) == I32:
                res, ptr_size = memory.GetData(size * 4, ptr.Value + 8)
                if __debug__:
                    assert res

                res, data = memory.GetData(
                    uint_from_bytes(ptr_size),
                    uint_from_bytes(address),
                )
                assert res
                self.result = res
                self.output = data
            else:
                print("Unknown type: " + str(uint_from_bytes(type)))

            return res, []

        def _return_error(ptr, size):
            if __debug__:
                print("Error Func")

            memory = self.vm.GetStoreContext().FindMemory("memory")
            _, data = memory.GetData(size.Value, ptr.Value)

            self.result = _
            self.output = data
            return _, []

        self.resultFn = WasmEdge.Function(
            result_function_type, _return_result, 0
        )
        self.errorFn = WasmEdge.Function(
            result_function_type, _return_error, 0
        )
        self.funcImports.AddFunction("return_result", self.resultFn)
        self.funcImports.AddFunction("return_error", self.errorFn)
        self.vm.RegisterModuleFromImport(self.funcImports)
        self.vm.Instantiate()

        self.pop = []
        self.length_pop = []

    def execute(self, function_name, *args):
        execution_type = 1
        function_arg_len = len(args)
        memory = self.vm.GetStoreContext().FindMemory("memory")

        res, pointer_of_pointers = self.vm.Execute(
            "allocate",
            tuple([WasmEdge.Value(8 * function_arg_len, WasmEdge.Type.I32)]),
            1,
        )
        assert res

        self.pop.extend(pointer_of_pointers)
        self.length_pop.append(8 * function_arg_len)

        for i, arg in enumerate(args):
            offset = 4 * 2 * i
            len_offset = offset + 4
            if isinstance(arg, str):
                arg = bytes(arg, "UTF-8")
                execution_type = String
            elif isinstance(arg, int):
                arg = int_to_bytes(arg)
                execution_type = I32
            elif isinstance(arg, (tuple, list)) and all(
                x == int for x in list(map(type, arg))
            ):
                arg = [int_to_bytes(i) for i in arg]
                execution_type = I32Array
            else:
                raise TypeError(f"Unsupported type:{type(arg)}")

            if execution_type == I32Array:
                res, ptr = self.vm.Execute(
                    "allocate",
                    tuple(
                        [
                            WasmEdge.Value(4 * len(arg), WasmEdge.Type.I32),
                        ]
                    ),
                    1,
                )
                assert res

                self.pop.extend(ptr)
                self.length_pop.append(4 * len(arg))

                prev_len = 0
                for i, data in enumerate(arg):
                    assert memory.SetData(tuple(data), ptr[0].Value + prev_len)
                    _, d = memory.GetData(len(data), ptr[0].Value + prev_len)
                    assert _
                    assert uint_from_bytes(d) == uint_from_bytes(data)
                    prev_len += len(data)

                assert memory.SetData(
                    tuple(uint_to_bytes(ptr[0].Value)),
                    pointer_of_pointers[0].Value + offset,
                )

                assert memory.SetData(
                    tuple(uint_to_bytes(len(arg))),
                    pointer_of_pointers[0].Value + len_offset,
                )

                _, ptr_of_ptr_data = memory.GetData(
                    8, pointer_of_pointers[0].Value + offset
                )
                assert _
                assert uint_from_bytes(ptr_of_ptr_data[:4]) == ptr[0].Value
                assert uint_from_bytes(ptr_of_ptr_data[4:]) == len(arg), (
                    uint_from_bytes(ptr_of_ptr_data[4:]),
                    len(arg),
                )

            elif execution_type in [String, I32]:
                res, ptr = self.vm.Execute(
                    "allocate",
                    tuple(
                        [
                            WasmEdge.Value(4 * len(arg), WasmEdge.Type.I32),
                        ]
                    ),
                    1,
                )
                assert res

                self.pop.extend(ptr)
                self.length_pop.append(4 * len(arg))

                assert memory.SetData(tuple(arg), ptr[0].Value)

                assert memory.SetData(
                    tuple(uint_to_bytes(ptr[0].Value)),
                    pointer_of_pointers[0].Value + offset,
                )

                assert memory.SetData(
                    tuple(uint_to_bytes(len(arg))),
                    pointer_of_pointers[0].Value + len_offset,
                )

                _, d = memory.GetData(len(arg), ptr[0].Value)
                assert _
                assert bytes(d) == arg

                _, ptr_of_ptr_data = memory.GetData(
                    4, pointer_of_pointers[0].Value + offset
                )
                assert _
                assert uint_from_bytes(ptr_of_ptr_data) == ptr[0].Value

                _, ptr_of_ptr_data_len = memory.GetData(
                    4, pointer_of_pointers[0].Value + len_offset
                )
                assert _
                assert uint_from_bytes(ptr_of_ptr_data_len) == len(arg)
            else:
                self.deallocator()
                raise TypeError("Unknown Type")

        res, _ = self.vm.Execute(
            function_name,
            tuple(
                [
                    pointer_of_pointers[0],
                    WasmEdge.Value(function_arg_len, WasmEdge.Type.I32),
                ]
            ),
            0,
        )
        assert res
        return self.result, self.output

    def deallocator(self):
        for ptr, len in zip(self.pop, self.length_pop):
            res, _ = self.vm.Execute(
                "deallocate", (ptr, WasmEdge.Value(len, WasmEdge.Type.I32)), 0
            )
            assert res
        self.pop.clear()
        self.length_pop.clear()
