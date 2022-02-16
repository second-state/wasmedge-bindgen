import WasmEdge


class Bindgen:
    def __init__(
        self,
        VM,
    ):
        self.vm = VM
        self.result = None
        self.output = None
        funcImports = WasmEdge.ImportObject("wasmedge-bindgen")

        result_function_type = WasmEdge.FunctionType(
            [WasmEdge.Type.I32, WasmEdge.Type.I32], []
        )

        def _return_result(a, b):
            if __debug__:
                print("hi")
            memory = self.vm.GetStoreContext().GetMemory()
            size = b.Value
            res, data = memory.GetData(len(bytes(a, "UTF-8")), 4 * 3 * size)
            rets = []
            for i in range(0, size * 3):
                rets.append(
                    int.from_bytes(bytes(data[i * 4 : (i + 1) * 4]), byteorder="little")
                )
            for i in range(0, size):
                _, data_ = memory.GetData(rets[i * 3], rets[i * 3 + 2])
                if __debug__:
                    print(data_.Value)
            # res, data = memory.GetData(len(bytes(a, "UTF-8")), 4 * 3 * b.Value)
            return res, data

        resultFn = WasmEdge.Function(result_function_type, _return_result, 0)
        errorFn = WasmEdge.Function(result_function_type, _return_result, 0)
        funcImports.AddFunction(resultFn, "return_result")
        funcImports.AddFunction(errorFn, "return_error")
        self.vm.RegisterModuleFromImport(funcImports)
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
            tuple(ptr[0].Value.to_bytes(4, "little")), pointer_of_pointers[0].Value
        )

        assert memory.SetData(
            tuple(len(args).to_bytes(4, "little")), pointer_of_pointers[0].Value + 4
        )

        _, d = memory.GetData(len(args), ptr[0].Value)
        assert _
        assert bytes(d) == args
        _, ptr_of_ptr_data = memory.GetData(8, pointer_of_pointers[0].Value)
        assert _
        assert int.from_bytes(bytes(ptr_of_ptr_data[:4]), "little") == ptr[0].Value
        assert int.from_bytes(bytes(ptr_of_ptr_data[4:]), "little") == len(args)

        res, data = self.vm.Execute(
            function_name,
            tuple([pointer_of_pointers[0], WasmEdge.Value(1, WasmEdge.Type.I32)]),
            1,
        )
        assert res
        return data

    def deallocator(self):
        for ptr, len in zip(self.pop, self.length_pop):
            res, _ = self.vm.Execute(
                "deallocate", (ptr, WasmEdge.Value(len, WasmEdge.Type.I32)), 1
            )
            assert res
