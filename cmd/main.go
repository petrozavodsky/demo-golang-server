package cmd

import (
	"flag"
	"user_service/pkg/server"
)

func Execute() {

	appType := "app"
	var port int
	var onePort int
	var twoPort int

	flag.IntVar(&port, "port", 8080, "Port")
	flag.StringVar(&appType, "type", "app", "Service type")
	flag.IntVar(&onePort, "port-one", 8080, "Application port")
	flag.IntVar(&twoPort, "port-two", 8081, "Application port")
	flag.Parse()

	if "proxy" == appType {
		server.Proxy(port, onePort, twoPort)
		return
	}

	server.WebService(port)
	return

}
