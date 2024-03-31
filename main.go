package main

import (
	"crypto/tls"
	"flag"
	"io"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/ssgelm/cookiejarparser"
)

var Client *http.Client

func httpHandler(trueHost string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		newUrl := r.URL
		newUrl.Host = trueHost
		newUrl.Scheme = "https"

		// make request
		req, err := http.NewRequest("GET", newUrl.String(), r.Body)
		if err != nil {
			log.Println(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		for k, v := range r.Header {
			req.Header.Set(k, v[0])
		}

		req.Header.Set("Referer", trueHost)

		// do request
		resp, err := Client.Do(req)
		if err != nil {
			log.Println(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		// output
		statusCode := resp.StatusCode
		for k, v := range resp.Header {
			for _, vv := range v {
				w.Header().Set(k, vv)
			}
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

func InitCoolieJar() http.CookieJar {
	cookies, err := cookiejarparser.LoadCookieJarFile("cookies.txt")
	if err != nil {
		log.Fatal(err)
	}
	return cookies
}

func InitClient(dualStack bool) {
	log.Printf("[log]init client with dualStack:%v\n", dualStack)

	var client *http.Client = func() *http.Client {
		tr := &http.Transport{
			Dial: (&net.Dialer{
				Timeout:   5 * time.Second,
				KeepAlive: 30 * time.Second,
				DualStack: dualStack, // Forces IPv4 only when set to `false`
			}).Dial,
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		return &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
			Transport: tr,
			Jar:       InitCoolieJar(),
		}
	}()

	Client = client
}

func main() {
	var laddr = flag.String("l", "127.0.0.1:8080", "listen address")
	var saddr = flag.String("s", HOST, "server address")
	var dualStack = flag.Bool("dual-stack", false, "Forces IPv4 only when not set this flag")
	flag.Parse()

	InitClient(*dualStack)

	handler := http.NewServeMux()
	handler.HandleFunc("/", httpHandler(*saddr))
	server := &http.Server{Addr: *laddr, Handler: handler}

	err := server.ListenAndServe()
	log.Println(err)
}
