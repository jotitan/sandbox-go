package main
import (
    "net/http"
	"runtime"
	"arguments"
	"logger"
	"encoding/json"
	"net"
	"strings"
	"time"
	"io"
	"longhorn"
	"strconv"
	"fmt"
)

/* Launch a server to treat resize image */

var gameManager longhorn.GameManager

type SSEWriter struct{
	w io.Writer
	f http.Flusher
}

func (sse SSEWriter)Write(message string){
	sse.w.Write([]byte("data: " + message + "\n\n"))
	sse.f.Flush()
}

// getTasksAsSSE return in a server sent event. All new tasks are notified
func getTasksAsSSE(response http.ResponseWriter,request *http.Request){
	logger.GetLogger().Info("Get request tasks SSE")
	sse := SSEWriter{response,response.(http.Flusher)}

	infos := "DATA"
	dataInfos,_ := json.Marshal(infos)
	// Send all init tasks, then send only updated tasks

	createSSEHeader(response)

	sse.Write(string(dataInfos))
	/*for data := range stream {
		sse.Write(string(data))
	} */
}

// Use to find node with very short timeout
func status(response http.ResponseWriter, request *http.Request){
	response.Write([]byte("Up"))
}

func createSSEHeader(response http.ResponseWriter){
	response.Header().Set("Content-Type","text/event-stream")
	response.Header().Set("Cache-Control","no-cache")
	response.Header().Set("Connection","keep-alive")
	response.Header().Set("Access-Control-Allow-Origin","*")
}

// Return stats by server side event
func statsAsSSE(response http.ResponseWriter, request *http.Request){
	createSSEHeader(response)
	sendStats(response)
}

func sendStats(r http.ResponseWriter){
	defer func(){
		if err := recover() ; err != nil {}
	}()
	stop := false
	go func(){
		<-r.(http.CloseNotifier).CloseNotify()
		stop=true
	}()

	for {
		stats := "DATA"
		data, _ := json.Marshal(stats)
		r.Write([]byte("data: " + string(data) + "\n\n"))
		if stop == true{
			break
		}
		r.(http.Flusher).Flush()
		time.Sleep(1 * time.Second)
	}
}

func dialog(response http.ResponseWriter, request *http.Request){

}


func root(response http.ResponseWriter, request *http.Request){
	url := request.RequestURI
	http.SetCookie(response,&http.Cookie{Name:"jsessionid",Value:"VALUE TEST"})
	fmt.Println(request.Cookies())
	http.ServeFile(response,request,"resources/" + url[1:])
}

// Join a game
func join(response http.ResponseWriter, r *http.Request){
    var g *longhorn.Game
	if idGame := r.FormValue("idGame") ; idGame != "" {
		if id,err := strconv.ParseInt(idGame,10,32);err == nil {
			g, _ = gameManager.GetGame(int(id))
		}
	}
	if g == nil {
		g = gameManager.CreateGame()
	}
	http.SetCookie(response,&http.Cookie{Name:"gameid",Value:fmt.Sprintf("%d",g.Board.GetId())})
	m := longhorn.NewServerMessage(g.Board)
	str,_ := json.Marshal(m)
	response.Header().Set("Content-type","application/json")
	response.Write(str)
}

func findExposedURL()string{
	adr,_ := net.InterfaceAddrs()
	for _,a := range adr {
		if a.String() != "0.0.0.0" && !strings.Contains(a.String(),"127.0.0.1"){
			if idx := strings.Index(a.String(),"/"); idx != -1 {
				return a.String()[:idx]
			}
			return a.String()
		}
	}
	return "localhost"
}

func createServer(port string){
	if port == ""{
		logger.GetLogger().Fatal("Impossible to run node, port is not defined")
	}
	localIP := findExposedURL()
	gameManager = longhorn.NewGameManager()
    mux := createRoutes()
	logger.GetLogger().Info("Runner ok on :",localIP,port)
    http.ListenAndServe(":" + port,mux)

	logger.GetLogger().Error("Runner ko")
}

func createRoutes()*http.ServeMux{
	mux := http.NewServeMux()
	mux.HandleFunc("/join",join)
	mux.HandleFunc("/dialog",dialog)
	mux.HandleFunc("/",root)
	return mux
}

func main(){
	runtime.GOMAXPROCS(runtime.NumCPU())
	args := arguments.ParseArgs()
	port := args["port"]

	createServer(port)
}
