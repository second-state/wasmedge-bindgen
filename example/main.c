#include "wasmedge/wasmedge.h"
#include <stdio.h>

int main() {
  WasmEdge_ConfigureContext *ConfCxt = WasmEdge_ConfigureCreate();
  WasmEdge_VMContext *VMCxt = WasmEdge_VMCreate(ConfCxt, NULL);
  WasmEdge_String ModNameBook = WasmEdge_StringCreateByCString("book");
  WasmEdge_String ModNameApp = WasmEdge_StringCreateByCString("app");
  WasmEdge_String FuncNameStart = WasmEdge_StringCreateByCString("start");
  WasmEdge_VMRegisterModuleFromFile(
      VMCxt, ModNameBook, "./book/target/wasm32-wasi/debug/book.wasm");
  WasmEdge_VMRegisterModuleFromFile(VMCxt, ModNameApp,
                                    "./app/target/wasm32-wasi/debug/app.wasm");

  WasmEdge_Value Params[0] = {};
  WasmEdge_Value Returns[1];
  WasmEdge_VMExecuteRegistered(VMCxt, ModNameApp, FuncNameStart, Params, 0,
                               Returns, 1);
  printf("Get the result: %d\n", WasmEdge_ValueGetI32(Returns[0]));

  WasmEdge_StringDelete(ModNameBook);
  WasmEdge_StringDelete(ModNameApp);
  WasmEdge_StringDelete(FuncNameStart);
  WasmEdge_VMDelete(VMCxt);
  WasmEdge_ConfigureDelete(ConfCxt);
  return 0;
}
