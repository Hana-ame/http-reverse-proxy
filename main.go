package main

import (
	"flag"
	"io"
	"log"
	"net/http"
)

var Client *http.Client = &http.Client{}

func httpHandler(trueHost string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		newUrl := r.URL
		newUrl.Host = trueHost
		newUrl.Scheme = "https"

		req, err := http.NewRequest("GET", newUrl.String(), nil)
		if err != nil {
			log.Println(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if COOKIE != "" {
			req.Header.Set("Cookie", COOKIE)
		}

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

func main() {
	var laddr = flag.String("l", "127.0.0.1:8080", "listen address")
	var saddr = flag.String("s", "s.exhentai.org", "server address")
	// var reExp = flag.String("r", ".*", "regex to match")
	flag.Parse()

	handler := http.NewServeMux()
	handler.HandleFunc("/", httpHandler(*saddr))
	server := &http.Server{Addr: *laddr, Handler: handler}

	err := server.ListenAndServe()
	log.Println(err)
}
