package server

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
)

var counter = 0
var firstInstanceHost string
var secondInstanceHost string

func Proxy(proxyPort int, onePort int, twoPort int) {

	firstInstanceHost = "http://localhost:" + strconv.Itoa(onePort)
	secondInstanceHost = "http://localhost:" + strconv.Itoa(twoPort)

	log.Println("Run proxy - proxyPort: ", proxyPort)

	http.HandleFunc("/", handle)
	log.Fatalln(http.ListenAndServe(":"+strconv.Itoa(proxyPort), nil))
}

// handle - sends request to the instances
func handle(w http.ResponseWriter, r *http.Request) {
	host := chooseInstanceHost()

	log.Println("use host: ", host)

	target, err := url.Parse(host)

	if err != nil {
		log.Fatalln(err)
	}

	proxy := httputil.NewSingleHostReverseProxy(target)

	proxy.ServeHTTP(w, r)
}

func chooseInstanceHost() (instance string) {
	if counter == 0 {
		counter++
		return firstInstanceHost
	}

	counter--
	return secondInstanceHost
}
