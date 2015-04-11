package main
import (
    "fmt"
    "arguments"
    "resize"
    "sync"
    "time"
    "os"
    "strings"
    "path/filepath"
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

    counter := sync.WaitGroup{}
    begin := time.Now()

    var requests []Request
    if file,err := os.Open(args.GetString("from")) ; err == nil {
        if info,_ := file.Stat() ; info.IsDir() {
            // List jpg file
            names,_ := file.Readdirnames(0)
            for _,name := range names {
                if strings.HasSuffix(name,".jpg") {
                    from := filepath.Join(args.GetString("from"),name)
                    to := fmt.Sprintf("%s_%s", args.GetString("to"), name)
                    requests = append(requests,Request{from,to})
                }
            }
        }else{
            requests = []Request{Request{args.GetString("from"),args.GetString("to")}}
        }
    }
    for _,r := range requests {
        counter.Add(1)
        go func(req Request) {
            if err := resize.Resize(req.from, req.to, args.GetUInt("width"), args.GetUInt("height")); err == nil {
                fmt.Println("Img resized", req.to)
            }else {
                fmt.Println("Impossible",err)
            }
            counter.Done()
        }(r)
    }
    counter.Wait()
    fmt.Println("Done",time.Now().Sub(begin))
}
