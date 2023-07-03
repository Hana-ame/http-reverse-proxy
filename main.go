package main

import (
	"flag"
	"io"
	"log"
	"net"
	"net/http"
	"time"
)

var Client *http.Client
var channel = make(chan struct{}, 5)

func httpHandler(trueHost, cookie string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		channel <- struct{}{}
		defer func() { <-channel }()

		newUrl := r.URL
		newUrl.Host = trueHost
		newUrl.Scheme = "https"

		req, err := http.NewRequest("GET", newUrl.String(), r.Body)
		if err != nil {
			log.Println(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		for k, v := range r.Header {
			req.Header.Set(k, v[0])
		}
		if cookie != "" {
			req.Header.Set("Cookie", cookie)
		}
		req.Header.Set("Referer", trueHost)

		resp, err := Client.Do(req)
		if err != nil {
			log.Println(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		statusCode := resp.StatusCode
		for k, v := range resp.Header {
			w.Header().Set(k, v[0])
		}

		w.WriteHeader(statusCode)

		_, err = io.Copy(w, resp.Body)
		if err != nil {
			log.Println(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

	}
}

func InitClient(dualStack bool) {
	log.Printf("[log]init client with dualStack:%v\n", dualStack)
	tr := &http.Transport{
		Dial: (&net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: dualStack, // Forces IPv4 only when set to `false`
		}).Dial,
	}

	client := &http.Client{
		Transport: tr,
	}

	Client = client
}

func main() {
	var laddr = flag.String("l", "127.0.0.1:8080", "listen address")
	var saddr = flag.String("s", HOST, "server address")
	var cookie = flag.String("cookie", COOKIE, "cookie")
	var dualStack = flag.Bool("-dual-stack", false, "Forces IPv4 only when not set this flag")
	// var reExp = flag.String("r", ".*", "regex to match")
	flag.Parse()

	InitClient(*dualStack)

	handler := http.NewServeMux()
	handler.HandleFunc("/", httpHandler(*saddr, *cookie))
	server := &http.Server{Addr: *laddr, Handler: handler}

	err := server.ListenAndServe()
	log.Println(err)
}
