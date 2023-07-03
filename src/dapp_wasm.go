package main

import (
  "fmt"

  "dapp/processor"
  "syscall/js"
)

func EmptyCellValue(this js.Value, args []js.Value) interface{} {
  if len(args) == 0 {
    return nil
  }
  value, err := processor.CsvBlankCellPermillionage(args[0].String(),"na")
  if err != nil {
    fmt.Println("Error:",err)
    return nil
  }
  return value
}

func GetDataCid(this js.Value, args []js.Value) interface{} {
  if len(args) == 0 {
    return nil
  }
  value, err := processor.GetDataCid(args[0].String())
  if err != nil {
    fmt.Println("Error:",err)
    return nil
  }
  return value.String()
}

func PrepareData(this js.Value, args []js.Value) interface{} {
  if len(args) == 0 {
    return nil
  }
  value, err := processor.PrepareDataToSend([]byte(args[0].String()),uint64(args[1].Int()))
  if err != nil {
    fmt.Println("Error:",err)
    return nil
  }
  valueInterface := make([]interface{}, len(value))
  for i, v := range value {
    valueInterface[i] = v
  }
  return valueInterface
}

func main() {
  wait := make(chan struct{},0)
  fmt.Println("DAPP WASM initialized")
  js.Global().Set("emptyCellValue", js.FuncOf(EmptyCellValue))
  js.Global().Set("getDataCid", js.FuncOf(GetDataCid))
  js.Global().Set("prepareData", js.FuncOf(PrepareData))
  <- wait
}