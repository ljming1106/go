//获取当前程序的PID
package main

import(
  "fmt"
  "os"
)

func main(){
  fmt.Print("%d",os.Getpid())
}
