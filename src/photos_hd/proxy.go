package main

import "net/http"
import "net/http/httputil"
import "net/url"
import "strings"
import "os"
import "io/ioutil"
import "encoding/json"
import "log"

var rep *httputil.ReverseProxy

var proxyRoutes map[string]*httputil.ReverseProxy

// Structure of conf file : Json like {[{route:,host:},{route:,host:}]}

func main(){
	if len(os.Args) != 3 {
		log.Println("Need parameters <port> <conf>")
		os.Exit(1)
	}
	routes := extractRoutes(os.Args[2])
	proxyRoutes = make(map[string]*httputil.ReverseProxy,len(routes))
	for route,host := range routes {
		u,_ := url.Parse(host)
		proxyRoutes[route] = httputil.NewSingleHostReverseProxy(u)
	}

	server := http.NewServeMux()
	server.HandleFunc("/",routing)
	port := os.Args[1]
	log.Println("Start proxy on port",port,"with",len(routes),"routes")
	http.ListenAndServe(":" + port,server)

}

func extractRoutes(path string)map[string]string{
	routes := make(map[string]string,0)
	if data,err := ioutil.ReadFile(path) ; err == nil {
		rawRoutes := make(map[string]interface{},0)
		json.Unmarshal(data,&rawRoutes)
		for _,route := range rawRoutes["routes"].([]interface{}) {
			routeDetail := route.(map[string]interface{})
			routes[routeDetail["route"].(string)] = routeDetail["host"].(string)
		}
	}
	return routes
}

func routing(w http.ResponseWriter, r * http.Request) {
	if pos := strings.Index(r.URL.Path[1:], "/"); pos != -1 {
		subPath := r.URL.Path[1 : pos+1]
		if route, exist := proxyRoutes[subPath]; exist {
			// Redirect
			//fmt.Println(r.URL.Path[1+pos:])
			serve(w,r,subPath,r.URL.Path[1+pos:],route)
		} else {
			log.Println("Unknown route", subPath, "=>", r.URL.Path)
			w.Write([]byte("Unknown route"))
		}
	} else {
		if route, exist := proxyRoutes[r.URL.Path[1:]]; exist {
			serve(w,r,r.URL.Path[1:],"/",route)
		} else{
			log.Println("No route", r.URL.Path)
			w.Write([]byte("No route"))
		}
	}
}

func serve(w http.ResponseWriter, r * http.Request, routeName, path string,rp * httputil.ReverseProxy){
	r.URL.Path = path
	r.Header.Set("proxy-redirect",routeName+ "/")
	rp.ServeHTTP(w, r)
}

