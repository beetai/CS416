package main

import (
	distpow "example.org/cpsc416/a2"
	"flag"
	"log"
	"net"
	"net/http"
	"net/rpc"
)

func main() {
	var config distpow.WorkerConfig
	err := distpow.ReadJSONConfig("config/worker_config.json", &config)
	if err != nil {
		log.Fatal(err)
	}
	flag.StringVar(&config.WorkerID, "id", config.WorkerID, "Worker ID, e.g. worker1")
	flag.StringVar(&config.ListenAddr, "listen", config.ListenAddr, "Listen address, e.g. 127.0.0.1:5000")
	flag.Parse()

	var worker = new(distpow.Worker)

	if err := worker.Initialize(config); err != nil {
		log.Fatal(err)
	}

	err = rpc.Register(worker)
	if err != nil {
		log.Fatal("error registering API: ", err)
	}

	rpc.HandleHTTP()

	listener, err := net.Listen("tcp", config.ListenAddr)
	if err != nil {
		log.Fatal("Listener error: ", err)
	}

	log.Printf("serving rpc on port %s", config.ListenAddr)

	err = http.Serve(listener, nil)
	if err != nil {
		log.Fatal("error serving: ", err)
	}

	log.Println(config)
}
