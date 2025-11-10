package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/smtp"
	"net/url"
	"os"
	"strings"
	"time"
)

/* Micro service which detect when ip is changing and send email with the new ip */

var lastIpCheck time.Time
var previousIP string

var conf config

type config struct {
	passwordLB   string
	usernameSMTP string
	passwordSMTP string
	hostSMTP     string
	fromEmail    string
	toEmail      string
	port         string
}

func main() {
	if err := checkStartup(); err != nil {
		panic(err)
	}
	go func() {
		for {
			checkAndMail()
			time.Sleep(15 * time.Minute)
		}
	}()
	runServerHealth()
}

func checkStartup() error {
	if len(os.Args) != 8 {
		return errors.New("need to specify configuration <password livebox> <SMTP user> <SMTP password> < SMTP host> <from email> <to email> <port> ")
	}
	conf = config{
		passwordLB:   os.Args[1],
		usernameSMTP: os.Args[2],
		passwordSMTP: os.Args[3],
		hostSMTP:     os.Args[4],
		fromEmail:    os.Args[5],
		toEmail:      os.Args[6],
		port:         os.Args[7],
	}
	return nil
}

func runServerHealth() {
	s := http.ServeMux{}
	s.HandleFunc("/health", health)
	log.Println("Server started on port " + conf.port)
	log.Fatal(http.ListenAndServe(":"+conf.port, &s))
}

func health(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(fmt.Sprintf("Up, last check : %s", lastIpCheck.Format("02-01-2006 15:04:05"))))
}

// Check if IP change and send IP if so
func checkAndMail() {
	if newIP := getCompleteIP(); previousIP != newIP {
		previousIP = newIP
		sendEmail(newIP)
	}
}

func getCompleteIP() string {
	lastIpCheck = time.Now()
	key, sessionName, sessionId, err := login(conf.passwordLB)
	if err != nil {
		return "not-found"
	}
	ip, err := getIP(key, sessionName, sessionId)
	if err != nil {
		return "not-found"
	}
	return ip
}

// Login into livebox and get auth token
func login(password string) (string, string, string, error) {
	buf := bytes.NewBufferString(fmt.Sprintf(`{"service":"sah.Device.Information","method":"createContext","parameters":{"applicationName":"webui","username":"admin","password":"%s"}}`, password))
	request, _ := http.NewRequest(http.MethodPost, "http://192.168.1.1/ws", buf)

	request.Header.Add("Content-Type", "application/json")
	request.Header.Add("Authorization", "X-Sah-Login")

	r, err := http.DefaultClient.Do(request)
	if err == nil {
		data, _ := io.ReadAll(r.Body)
		var m map[string]interface{}
		json.Unmarshal(data, &m)
		key := m["data"].(map[string]interface{})["contextID"].(string)
		r.Cookies()
		cookie := r.Header.Get("Set-Cookie")
		idx := strings.Index(cookie, "; path")
		cookie = cookie[:idx]
		subs := strings.Split(cookie, "=")

		return key, subs[0], subs[1], nil
	} else {
		return "", "", "", err
	}
}

// Call Livebox API to retrieve public IP
func getIP(key, sessionName, sessionId string) (string, error) {
	buf := bytes.NewBufferString(`{"service": "NMC", "method": "getWANStatus", "parameters": {}}`)
	request, _ := http.NewRequest(http.MethodPost, "http://192.168.1.1/ws", buf)

	request.Header.Set("Content-Type", "application/x-sah-ws-4-call+json")
	request.Header.Set("Authorization", fmt.Sprintf("X-Sah %s", key))
	request.Header.Set("X-Context", key)
	request.AddCookie(&http.Cookie{
		Name: sessionName, Value: sessionId,
	})
	request.AddCookie(&http.Cookie{
		Name: "sah/contextId", Value: url.QueryEscape(key),
	})

	r, err := http.DefaultClient.Do(request)
	if err == nil {
		data, _ := io.ReadAll(r.Body)
		var m map[string]interface{}
		json.Unmarshal(data, &m)
		return m["data"].(map[string]interface{})["IPAddress"].(string), nil
	} else {
		return "", err
	}
}

func sendEmail(ip string) {
	auth := smtp.PlainAuth("", conf.usernameSMTP, conf.passwordSMTP, conf.hostSMTP)
	message := fmt.Sprintf("Voici l'adresse IP de la maison : %s", ip)
	smtp.SendMail(conf.hostSMTP+":587", auth, conf.fromEmail, []string{conf.toEmail}, []byte("Subject:Message serveur maison\n\n"+message))
}
