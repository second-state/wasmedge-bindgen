import WasmEdge


class Bindgen:
    def __init__(self, VM):
        self.vm = VM
        self.result = None
        self.output = None
        funcImports = WasmEdge.ImportObject("wasmedge-bindgen")

        result_function_type = WasmEdge.FunctionType(
            [WasmEdge.Type.I32, WasmEdge.Type.I32], []
        )

        def _return_result(a, b):
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
        args_w = args[:]
        args = [WasmEdge.Value(i, WasmEdge.Type.I32) for i in args]
        res, pointer_of_pointers = self.vm.Execute(
            "allocate", tuple([WasmEdge.Value(8, WasmEdge.Type.I32)]), 1
        )

        self.pop.extend(pointer_of_pointers)
        self.length_pop.append(8)

        memory = self.vm.GetStoreContext().GetMemory("memory")
        res, ptr = self.vm.Execute(
            "allocate",
            tuple(
                [
                    WasmEdge.Value(len(args_w) * 8, WasmEdge.Type.I32),
                ]
            ),
            1,
        )
        assert res

        self.pop.extend(ptr)
        self.length_pop.append(len(args_w) * 8)

        assert memory.SetData(tuple(args_w), ptr[0].Value)
        # assert memory.SetData(
        #     tuple(ptr[0].Value.to_bytes(4, 'little')), pointer_of_pointers[0].Value+4)
        assert memory.SetData(
            tuple(ptr[0].Value.to_bytes(4, "little")), pointer_of_pointers[0].Value + 8
        )

        assert memory.SetData(
            tuple(len(args_w).to_bytes(4, "little")), pointer_of_pointers[0].Value + 12
        )

        res, data = self.vm.Execute(
            function_name,
            (pointer_of_pointers[0], ptr[0]),
            len(args_w),
        )
        assert res
        return data

    def deallocator(self):
        for ptr, len in zip(self.pop, self.length_pop):
            res, _ = self.vm.Execute("deallocate", (ptr,), len)
            assert res
