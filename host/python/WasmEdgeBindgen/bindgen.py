import WasmEdge


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
            memory = self.vm.GetStoreContext().GetMemory("memory")
            size = size.Value
            res, address = memory.GetData(size * 4, ptr.Value)
            if __debug__:
                assert res

            res, ptr_size = memory.GetData(size * 4, ptr.Value + 8)
            if __debug__:
                assert res

            res, data = memory.GetData(
                int.from_bytes(bytes(ptr_size), byteorder="little"),
                int.from_bytes(bytes(address), "little"),
            )
            assert res
            self.result = res
            self.output = data
            return res, []

        def _return_error(ptr, size):
            if __debug__:
                print("Error Func")

            memory = self.vm.GetStoreContext().GetMemory("memory")
            _, data = memory.GetData(size.Value, ptr.Value)

            rets = [WasmEdge.Value(i, WasmEdge.Type.I32) for i in bytes(data)]
            self.result = _
            self.output = rets
            return _, []

        self.resultFn = WasmEdge.Function(
            result_function_type, _return_result, 0
        )
        self.errorFn = WasmEdge.Function(
            result_function_type, _return_error, 0
        )
        self.funcImports.AddFunction(self.resultFn, "return_result")
        self.funcImports.AddFunction(self.errorFn, "return_error")
        self.vm.RegisterModuleFromImport(self.funcImports)
        self.vm.Instantiate()

        self.pop = []
        self.length_pop = []

    def execute(self, function_name, args):
        args = bytes(args, "UTF-8")
        res, pointer_of_pointers = self.vm.Execute(
            "allocate", tuple([WasmEdge.Value(8, WasmEdge.Type.I32)]), 1
        )
        assert res

        self.pop.extend(pointer_of_pointers)
        self.length_pop.append(8)

        memory = self.vm.GetStoreContext().GetMemory("memory")
        res, ptr = self.vm.Execute(
            "allocate",
            tuple(
                [
                    WasmEdge.Value(4 * len(args), WasmEdge.Type.I32),
                ]
            ),
            1,
        )
        assert res

        self.pop.extend(ptr)
        self.length_pop.append(4 * len(args))

        assert memory.SetData(tuple(args), ptr[0].Value)

        assert memory.SetData(
            tuple(ptr[0].Value.to_bytes(4, "little")),
            pointer_of_pointers[0].Value,
        )

        assert memory.SetData(
            tuple(len(args).to_bytes(4, "little")),
            pointer_of_pointers[0].Value + 4,
        )

        _, d = memory.GetData(len(args), ptr[0].Value)
        assert _
        assert bytes(d) == args
        _, ptr_of_ptr_data = memory.GetData(8, pointer_of_pointers[0].Value)
        assert _
        assert (
            int.from_bytes(bytes(ptr_of_ptr_data[:4]), "little")
            == ptr[0].Value
        )
        assert int.from_bytes(bytes(ptr_of_ptr_data[4:]), "little") == len(
            args
        )

        res, _ = self.vm.Execute(
            function_name,
            tuple(
                [pointer_of_pointers[0], WasmEdge.Value(1, WasmEdge.Type.I32)]
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
