package server

import (
	"fmt"
	"io"
	"log"
	"net/http"
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
	log.Fatalln(http.ListenAndServe("localhost:"+strconv.Itoa(proxyPort), nil))
}

// handle - sends request to the instances
func handle(w http.ResponseWriter, r *http.Request) {
	host := fmt.Sprintf("%s%v", chooseInstanceHost(), r.URL)
	log.Println(host)

	request, err := http.NewRequest(r.Method, host, r.Body)
	if err != nil {
		log.Fatalln(err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(r.Body)

	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		log.Fatalln(err)
	}

	out, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)

	if _, err := w.Write(out); err != nil {
		log.Fatalln(err)
	}
}

func chooseInstanceHost() (instance string) {
	if counter == 0 {
		instance = firstInstanceHost
		counter++
		return
	}

	instance = secondInstanceHost
	counter--
	return
}
