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

func extractInt(r *http.Request,name string)int{
	if intValue := r.FormValue(name) ; intValue!="" {
		if value,err := strconv.ParseInt(intValue,10,0) ; err == nil {
			return value
		}
	}
	return 0
}

func resizeReq(_ http.ResponseWriter, r *http.Request){
    from := r.FormValue("from")
    to := r.FormValue("to")
	height,width := extractInt(r,"height"),extractInt(r,"width")

    resize.ResizeMany(from,to,uint(width),uint(height))
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
