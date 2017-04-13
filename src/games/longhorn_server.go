package main
import (
    "net/http"
	"runtime"
	"arguments"
	"logger"
	"net"
	"strings"
	"time"
	"io"
	"longhorn"
	"strconv"
	"fmt"
	"crypto/md5"
	"math/rand"
	"encoding/hex"
)

/* Launch a server to treat resize image */

var gameManager longhorn.GameManager
var webFolder = "resources/"

type SSEWriter struct{
	w io.Writer
	f http.Flusher
}

func (sse SSEWriter)Write(message string){
	sse.w.Write([]byte("data: " + message + "\n\n"))
	sse.f.Flush()
}

// Use to find node with very short timeout
func status(response http.ResponseWriter, request *http.Request){
	response.Write([]byte("Up"))
}

func event(response http.ResponseWriter, request *http.Request){
	sid := getSessionID(*request)
	eventMessage := request.FormValue("event")
	if sm,err := longhorn.EventFromJSON([]byte(eventMessage)) ; err == nil {
		// Get game and check with sessionId
		if g,e := gameManager.GetGame(sm.GameId,sid,"") ; e == nil {
			g.Workflow(sm)
		}
	}
}

func getSessionID(request http.Request)string {
	for _, c := range request.Cookies() {
		if c.Name == "jsessionid" {
			return c.Value
		}
	}
	return ""
}

func setSessionID(response http.ResponseWriter,request http.Request)string{
	if id := getSessionID(request) ; id != ""{
		return id
	}
	h := md5.New()
	h.Write([]byte(fmt.Sprintf("%d-%d",time.Now().Nanosecond(),rand.Int())))
	hash := h.Sum(nil)
	hexValue := hex.EncodeToString(hash)
	logger.GetLogger().Info("Set cookie session",hexValue)
	http.SetCookie(response,&http.Cookie{Name:"jsessionid",Value:hexValue})
	return hexValue
}

func root(response http.ResponseWriter, request *http.Request){
	setSessionID(response,*request)
	url := request.RequestURI
	http.ServeFile(response,request,webFolder + url[1:])
}

// connect to a party to listen server message
func connect(response http.ResponseWriter, r *http.Request) {
	sessionId := setSessionID(response,*r)
	if idGame,err := strconv.ParseInt(r.FormValue("idGame"),10,32) ; err == nil {
		if g,e := gameManager.GetGame(int(idGame),sessionId,"") ; e == nil {
			if p, e := g.ConnectPlayer(response, sessionId); e == nil {
				for {
					if !p.IsConnected() {
						break
					}
					time.Sleep(5*time.Second)
				}
			}
		}
	}
	// Check if player can play on this game
}

// Join or create a game
func join(response http.ResponseWriter, r *http.Request){
    var g *longhorn.Game
	sessionId := setSessionID(response,*r)
	name := r.FormValue("name")
	if idGame := r.FormValue("idGame") ; idGame != "" {
		if id,err := strconv.ParseInt(idGame,10,32);err == nil {
			g, _ = gameManager.GetGame(int(id),sessionId,name)
		}
	}
	if g == nil {
		g = gameManager.CreateGame(sessionId,name)
	}
	http.SetCookie(response,&http.Cookie{Name:"gameid",Value:fmt.Sprintf("%d",g.Board.GetId())})
	m := longhorn.NewServerMessage(g.Board)
	response.Header().Set("Content-type","application/json")
	response.Write(m.ToJSON())
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
		logger.GetLogger().Fatal("Port is not defined")
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
	mux.HandleFunc("/connect",connect)
	mux.HandleFunc("/event",event)
	mux.HandleFunc("/",root)
	return mux
}

func main(){
	runtime.GOMAXPROCS(runtime.NumCPU())
	args := arguments.ParseArgs()
	port := args["port"]
	if resourcesFolder,exist := args["webFolder"] ; exist {
		webFolder = resourcesFolder
	}

	createServer(port)
}
