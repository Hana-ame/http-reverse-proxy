package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/cookiejar"

	"os"

	_ "github.com/joho/godotenv/autoload"

	myclient "github.com/Hana-ame/http-reverse-proxy/my_client"
	"github.com/ssgelm/cookiejarparser"
)

var Client *http.Client

func httpHandler(trueHost string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		newUrl := r.URL
		newUrl.Host = trueHost
		newUrl.Scheme = "https"

		// make request
		req, err := http.NewRequest(http.MethodGet, newUrl.String(), r.Body)
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
		resp, err := client().Do(req)
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
				w.Header().Add(k, vv)
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

// cookie
var defaultJar http.CookieJar

func initCoolieJar() http.CookieJar {
	jar, err := cookiejarparser.LoadCookieJarFile("cookies.txt")
	if err != nil {
		log.Fatal(err)
	}
	// defaultJar = jar
	return jar
}

// clients
var defaultPool *myclient.ClientPool = myclient.NewClientPool()

func initIPv6Clients() *myclient.ClientPool {
	pool := myclient.NewClientPool()
	for _, ipv6addr := range openfile() {
		ip := net.ParseIP(ipv6addr)
		client := myclient.NewClient(ip, defaultJar.(*cookiejar.Jar))
		pool.Add(client)
	}
	// defaultPool = pool
	return pool
}

func client() *http.Client {
	return defaultPool.Get()
}

func main() {
	var laddr = flag.String("l", os.Getenv("HTTP_PROXY_LOCAL"), "listen address")
	var saddr = flag.String("s", os.Getenv("HTTP_PROXY_HOST"), "server address")
	// var ipv4 = flag.Bool("4", true, "use ipv4")
	var ipv6 = flag.Bool("6", false, "use ipv6")
	flag.Parse()

	defaultJar = initCoolieJar()
	if *ipv6 {
		defaultPool = initIPv6Clients()
	}

	handler := http.NewServeMux()
	handler.HandleFunc("/", httpHandler(*saddr))
	server := &http.Server{Addr: *laddr, Handler: handler}

	fmt.Printf("listen on %v, server is %v \n", *laddr, *saddr)

	err := server.ListenAndServe()
	log.Println(err)
}
