package main
import (
    "fmt"
    "arguments"
    "resize"
    "runtime"
)


type Request struct {
    from string
    to string
}

// Launch resize image
func main(){
    args := arguments.NewArguments()
    if !args.Check([]string{"from","to"}){
        fmt.Println("Need more arguments")
        return
    }

    runtime.GOMAXPROCS(runtime.NumCPU())

    resize.ResizeMany(args.GetString("from"),args.GetString("to"),args.GetUInt("width"),args.GetUInt("height"))
}
