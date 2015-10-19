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
	"music"
	"strconv"
	"sort"
)

/* Launch a server to treat resize image */

type SSEWriter struct{
	w io.Writer
	f http.Flusher
}

type MusicServer struct{
	folder string
	dico music.MusicDictionnary
}

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

type sortByArtist []map[string]string

func (a sortByArtist)Len() int{return len(a)}
func (a sortByArtist)Less(i, j int) bool{return a[i]["name"] < a[j]["name"]}
func (a sortByArtist)Swap(i, j int) {a[i],a[j] = a[j],a[i]}

func (ms MusicServer)getAllArtists(response http.ResponseWriter){
	logger.GetLogger().Info("Get all artists")
	// Response with nampe and url
	artists := music.LoadArtistIndex(ms.folder).FindAll()
	artistsData := make([]map[string]string,0,len(artists))
	for artist,id := range artists{
		artistsData = append(artistsData,map[string]string{"name":artist,"url":fmt.Sprintf("id=%d",id)})
	}
	sort.Sort(sortByArtist(artistsData))
	bdata,_ := json.Marshal(artistsData)
	response.Write(bdata)
}

func (ms MusicServer)getMusics(response http.ResponseWriter,musicsIds []int){
	musics := make([]map[string]string,len(musicsIds))
	for i,musicId := range musicsIds{
		m := ms.dico.GetMusicFromId(musicId)
		delete(m,"path")	// Cause no need to return
		musics[i] = map[string]string{"name":m["title"],"info":m["length"],"id":fmt.Sprintf("%d",musicId)}
	}
	data,_:= json.Marshal(musics)
	response.Write(data)
}

func (ms MusicServer)listByArtist(response http.ResponseWriter, request *http.Request){
	if id := request.FormValue("id") ; id == "" {
		ms.getAllArtists(response)
	}else{
		logger.GetLogger().Info("Load music of artist",id)
		artistId,_ := strconv.ParseInt(id,10,32)
		musicsIds := music.LoadArtistMusicIndex(ms.folder).MusicsByArtist[int(artistId)]
		ms.getMusics(response,musicsIds)
	}
}

func (ms MusicServer)listByAlbum(response http.ResponseWriter, request *http.Request){
	 switch{
		 // return albums of artist
	  case request.FormValue("id") != "" :
	  logger.GetLogger().Info("Get all albums")
	  idArtist,_:= strconv.ParseInt(request.FormValue("id"),10,32)
		albums := music.NewAlbumByArtist().GetAlbums(ms.folder,int(idArtist))
	  	albumsData := make([]map[string]string,0,len(albums))
		for _,album := range albums{
			albumsData = append(albumsData,map[string]string{"name":album.Name,"url":fmt.Sprintf("idAlbum=%d",album.Id)})
		}
		  sort.Sort(sortByArtist(albumsData))
		  bdata,_ := json.Marshal(albumsData)
		  response.Write(bdata)
	  case request.FormValue("idAlbum") != "" :
	  	idAlbum,_ := strconv.ParseInt(request.FormValue("idAlbum"),10,32)
	  	musics := music.MusicByAlbum{}.GetMusics(ms.folder,int(idAlbum))
	  	ms.getMusics(response,musics)

	  default : ms.getAllArtists(response)

	}

}

// Return info about music
func (ms MusicServer)musicInfo(response http.ResponseWriter, request *http.Request){
	id,_ := strconv.ParseInt(request.FormValue("id"),10,32)
	logger.GetLogger().Info("Load music info with id",id)
	music := ms.dico.GetMusicFromId(int(id))
	delete(music,"path")
	music["src"] = fmt.Sprintf("music?id=%d",id)
	bdata,_ := json.Marshal(music)
	response.Write(bdata)
}

func (ms MusicServer)browse(response http.ResponseWriter, request *http.Request){
	folder := request.FormValue("folder")
	ms.dico.Browse(folder)
}

// Return music content
// TODO MOCK
func (ms MusicServer)music(response http.ResponseWriter, request *http.Request){
	id,_ := strconv.ParseInt(request.FormValue("id"),10,32)
	logger.GetLogger().Info("Get music id",id)
	music := ms.dico.GetMusicFromId(int(id))

	m,_ := os.Open(music["path"])
	info,_ := m.Stat()
	response.Header().Set("Content-type","audio/mpeg")
	response.Header().Set("Content-Length",fmt.Sprintf("%d",info.Size()))
	io.Copy(response,m)
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
	ms.folder = folder
	ms.dico = music.LoadDictionnary(ms.folder)
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
	mux.HandleFunc("/browse",ms.browse)
	mux.HandleFunc("/",ms.root)
	return mux
}

func main(){
	runtime.GOMAXPROCS(runtime.NumCPU())
	args := arguments.ParseArgs()
	port := args["port"]

	if logFolder, ok := args["log"] ; ok {
		logger.InitLogger(filepath.Join(logFolder, "music_"+port+".log"), true)
	}

	ms := MusicServer{}
	ms.create(port,args["folder"])
}
