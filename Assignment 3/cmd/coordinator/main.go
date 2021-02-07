package main

import (
	distpow "example.org/cpsc416/a2"
	"log"
	"net"
	"net/http"
	"net/rpc"
)

func main() {
	var config distpow.CoordinatorConfig
	err := distpow.ReadJSONConfig("config/coordinator_config.json", &config)
	if err != nil {
		log.Fatal(err)
	}

	var coordinator = new(distpow.Coordinator)
	if err := coordinator.Initialize(config); err != nil {
		log.Fatal(err)
	}

	err = rpc.Register(coordinator)
	if err != nil {
		log.Fatal("error registering API: ", err)
	}

	rpc.HandleHTTP()

	clientListener, err := net.Listen("tcp", config.ClientAPIListenAddr)
	if err != nil {
		log.Fatal("Listener error: ", err)
	}
	workerListener, err := net.Listen("tcp", config.WorkerAPIListenAddr)
	if err != nil {
		log.Fatal("Listener error: ", err)
	}

	log.Printf("serving rpc on port %s", config.ClientAPIListenAddr)
	log.Printf("serving rpc on port %s", config.WorkerAPIListenAddr)

	go http.Serve(clientListener, nil)
	err = http.Serve(workerListener, nil)
	if err != nil {
		log.Fatal("error serving: ", err)
	}

	log.Println(config)
}
