from unittest import result
import WasmEdge


class Bindgen:
    def __init__(self, VM):
        self.vm = VM
        self.result = None
        self.output = None
        funcImports = WasmEdge.ImportObject("wasmedge-bindgen")

        result_function_type = WasmEdge.FunctionType(
            [WasmEdge.Type.I32, WasmEdge.Type.I32], [])

        def _return_result(a, b):
            memory = self.vm.GetStoreContext().GetMemory()
            res, data = memory.GetData(len(tuple(a)), 4*3*b.Value)
            return res, data

        resultFn = WasmEdge.Function(result_function_type, _return_result, 0)
        errorFn = WasmEdge.Function(result_function_type, _return_result, 0)
        funcImports.AddFunction(resultFn, "return_result")
        funcImports.AddFunction(errorFn, "return_error")
        self.vm.RegisterModuleFromImport(funcImports)
        self.vm.Instantiate()

    def execute(self, function_name, args):
        res, pointerOfpointer = self.vm.Execute(
            "allocate", (WasmEdge.Value(len(args), WasmEdge.Type.I32),), 1)
        assert res

        memory = self.vm.GetStoreContext().GetMemory("memory")

        memory.SetData(tuple(bytes(args, 'UTF-8')), pointerOfpointer[0].Value)

        res, data = self.vm.Execute(function_name, tuple((WasmEdge.Value(
            i, WasmEdge.Type.I32) for i in bytes(args, 'UTF-8'))), len(args)+32)
        assert res
        return data
