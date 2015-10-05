package main
import (
    "net/http"
	"runtime"
	"arguments"
	"logger"
	"fmt"
	"net"
	"strings"
	"time"
	"io"
	"path/filepath"
	"os"
	"encoding/json"
)

/* Launch a server to treat resize image */

type SSEWriter struct{
	w io.Writer
	f http.Flusher
}

type MusicServer struct{}

func (sse SSEWriter)Write(message string){
	sse.w.Write([]byte("data: " + message + "\n\n"))
	sse.f.Flush()
}


func (ms MusicServer)root(response http.ResponseWriter, request *http.Request){
	url := request.RequestURI
	http.ServeFile(response,request,"resources/" + url[1:])
}

// Use to find node with very short timeout
func (ms MusicServer)status(response http.ResponseWriter, request *http.Request){
	response.Write([]byte("Up"))
}

func (ms MusicServer)createSSEHeader(response http.ResponseWriter){
	response.Header().Set("Content-Type","text/event-stream")
	response.Header().Set("Cache-Control","no-cache")
	response.Header().Set("Connection","keep-alive")
	response.Header().Set("Access-Control-Allow-Origin","*")
}

// Return stats by server side event
func (ms MusicServer)statsAsSSE(response http.ResponseWriter, request *http.Request){
	ms.createSSEHeader(response)
	ms.sendStats(response)
}

func (ms MusicServer)listByArtist(response http.ResponseWriter, request *http.Request){
	// MOCK
	data := make(map[string]string)
	for i := 0 ; i < 20 ; i++{
		data[fmt.Sprintf("Artist %d",i)] = fmt.Sprintf("%d",i)
	}
	bdata,_ := json.Marshal(data)
	response.Write(bdata)
}

func (ms MusicServer)listByAlbum(response http.ResponseWriter, request *http.Request){
	//MOCK

}

// Return info about music
func (ms MusicServer)musicInfo(response http.ResponseWriter, request *http.Request){
	id := request.FormValue("id")
	logger.GetLogger().Info("=>",id,"BLA")
	data := map[string]string{"id":id,"src":"music?id=" + id,"title":"Title " + id,"time":"0"}
	bdata,_ := json.Marshal(data)
	response.Write(bdata)
}

// Return music content
func (ms MusicServer)music(response http.ResponseWriter, request *http.Request){
	id := request.FormValue("id")
	logger.GetLogger().Info("Get music id",id)
	// MOCK
	mockPath := ""
	switch id {
		case "3" : mockPath = "D:\\TORRENT\\Lenny Kravitz\\01.Are you gonna go my way.mp3"
		default : mockPath = "D:\\TORRENT\\Lenny Kravitz\\12.I belong to you.mp3"
	}
	music,_ := os.Open(mockPath)
	info,_ := music.Stat()
	response.Header().Set("Content-type","audio/mpeg")
	response.Header().Set("Content-Length",fmt.Sprintf("%d",info.Size()))
	io.Copy(response,music)
}


func (ms MusicServer)sendStats(r http.ResponseWriter){
	defer func(){
		if err := recover() ; err != nil {}
	}()
	stop := false
	go func(){
		<-r.(http.CloseNotifier).CloseNotify()
		stop=true
	}()

	for {
		r.Write([]byte("data: " + "hello" + "\n\n"))
		if stop == true{
			break
		}
		r.(http.Flusher).Flush()
		time.Sleep(1 * time.Second)
	}
}

func (ms MusicServer)findExposedURL()string{
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

func (ms MusicServer)create(port string,folder string){
	if port == ""{
		logger.GetLogger().Fatal("Impossible to run node, port is not defined")
	}
	localIP := ms.findExposedURL()

    mux := ms.createRoutes()
	logger.GetLogger().Info("Runner ok on :",localIP,port)
    http.ListenAndServe(":" + port,mux)

	logger.GetLogger().Error("Runner ko")
}

func (ms MusicServer)createRoutes()*http.ServeMux{
	mux := http.NewServeMux()

	mux.HandleFunc("/status",ms.status)
	mux.HandleFunc("/statsAsSSE",ms.statsAsSSE)
	mux.HandleFunc("/music",ms.music)
	mux.HandleFunc("/musicInfo",ms.musicInfo)
	mux.HandleFunc("/listByArtist",ms.listByArtist)
	mux.HandleFunc("/listByAlbum",ms.listByAlbum)
	mux.HandleFunc("/",ms.root)
	return mux
}

func main(){
	runtime.GOMAXPROCS(runtime.NumCPU())
	args := arguments.ParseArgs()
	port := args["port"]

	if logFolder, ok := args["log"] ; ok {
		logger.InitLogger(filepath.Join(logFolder, "tasker_"+port+".log"), true)
	}

	ms := MusicServer{}
	ms.create(port,args["folder"])
}
