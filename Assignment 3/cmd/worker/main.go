package main

import (
	"flag"
	"log"

	distpow "example.org/cpsc416/a2"
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

	log.Println(config)
}
