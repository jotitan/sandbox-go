package main
import (
    "net/http"
    "fmt"
    "strconv"
    "resize"
    "runtime"
)

/* Launch a server to treat resize image */


func root(response http.ResponseWriter, request *http.Request){
    fmt.Println("root")
}

func resizeReq(response http.ResponseWriter, request *http.Request){
    from := request.FormValue("from")
    to := request.FormValue("to")
    height,width := int64(0),int64(0)
    if h := request.FormValue("height") ; h!="" {
        if h2,err := strconv.ParseInt(h,10,0) ; err == nil {
            height = h2
        }
    }
    if w := request.FormValue("width") ; w!="" {
        if w2,err := strconv.ParseInt(w,10,0) ; err == nil {
            width = w2
        }
    }
    resize.ResizeMany(from,to,0,0)
    fmt.Println("Resize",from,to,width,height)
}


func createServer(){
    mux := http.NewServeMux()
    mux.HandleFunc("/resize",resizeReq)
    mux.HandleFunc("/",root)

    fmt.Println("Runner ok on :9010")
    http.ListenAndServe(":9010",mux)
    fmt.Println("Runner ko")
}

func main(){
    runtime.GOMAXPROCS(runtime.NumCPU())
    createServer()
}
