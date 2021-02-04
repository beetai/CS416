package main

import (
	distpow "example.org/cpsc416/a2"
	"log"
	"net"
	"net/http"
	"net/rpc"
)

//type Args struct {
//	A, B int
//}
//
//type Quotient struct {
//	Quo, Rem int
//}

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

	//coordinator, err := rpc.DialHTTP("tcp", "localhost"+config.WorkerAPIListenAddr)
	//if err != nil {
	//	log.Fatal("Connection error: ", err)
	//}

	//args := distpow.WorkerMine{[]uint8{1, 2, 3, 4}, 7, 0}

	//var multReply int
	//var divReply distpow.Quotient
	//var workerReply []uint8
	//coordinator.Call("Arith.Multiply", args, &multReply)
	//coordinator.Call("Arith.Divide", args, &divReply)
	//log.Println("start")
	//coordinator.Call("Worker.Mine", args, &workerReply)

	//log.Println(multReply)
	//log.Println(divReply)
	//log.Println("finish: ", workerReply)

	//var worker = new(distpow.Worker)
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
