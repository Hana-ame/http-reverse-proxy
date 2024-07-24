package main

import (
	"context"
	"crypto/tls"
	"flag"
	"io"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
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

// generateRandomNumber generates a random integer between min and max (inclusive)
func generateRandomNumber(min, max int) int {
	return rand.Intn(max-min+1) + min
}

var list = []string{
	"2001:470:c:6c::2",
	"2001:470:c:6c::3",
	"2001:470:c:6c::4",
	"2001:470:c:6c::5",
}

func randomLocalAddr() *net.TCPAddr {
	randomNumber := generateRandomNumber(0, 3)
	ipString := list[randomNumber]
	addr, _ := net.ResolveTCPAddr("tcp6", ipString)
	return addr
}

func InitClient(ipv4 bool, ipv6 bool) {
	log.Printf("[log]init client with dualStack:%v\n", ipv4)

	var client *http.Client = func() *http.Client {
		tr := &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				// 只允许IPv6连接，但是没用
				if network == "tcp" {
					network = "tcp6"
				}
				return (&net.Dialer{
					Timeout:   5 * time.Second,
					KeepAlive: 30 * time.Second,
					LocalAddr: randomLocalAddr(),
					Resolver: &net.Resolver{
						PreferGo: true,
						Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
							d := net.Dialer{
								Timeout: 5 * time.Second,
							}
							// 使用Google的IPv6 DNS服务器
							return d.DialContext(ctx, "udp6", "1.1.1.1:53")
						},
					},
				}).DialContext(ctx, network, addr)
			},
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
	var laddr = flag.String("l", os.Getenv("HTTP_PROXY_LOCAL"), "listen address")
	var saddr = flag.String("s", os.Getenv("HTTP_PROXY_HOST"), "server address")
	// var ipv4 = flag.Bool("4", true, "use ipv4")
	var ipv6 = flag.Bool("6", false, "use ipv6")
	flag.Parse()

	InitClient(*ipv6, *ipv6)

	handler := http.NewServeMux()
	handler.HandleFunc("/", httpHandler(*saddr))
	server := &http.Server{Addr: *laddr, Handler: handler}

	err := server.ListenAndServe()
	log.Println(err)
}
