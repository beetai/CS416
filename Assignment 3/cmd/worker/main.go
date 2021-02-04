package main

import (
	distpow "example.org/cpsc416/a2"
	"flag"
	"log"
	"net"
	"net/http"
	"net/rpc"
)

type API int

//type Args struct {
//	A, B int
//}
//
//type Quotient struct {
//	Quo, Rem int
//}
//
//type Arith int
//
//func (t *Arith) Multiply(args *Args, reply *int) error {
//	*reply = args.A * args.B
//	return nil
//}
//
//func (t *Arith) Divide(args *Args, quo *Quotient) error {
//	if args.B == 0 {
//		return errors.New("divide by zero")
//	}
//	quo.Quo = args.A / args.B
//	quo.Rem = args.A % args.B
//	return nil
//}

//type mineJob struct {
//	tracer  *tracing.Tracer
//	thrId   int
//	nonce   []uint8
//	cmpStr  string
//	thrBits uint
//	answer  chan []uint8
//	cancel  chan bool
//}
//
//func (mj *mineJob) work(args *Args, quo *Quotient) error {
//	defer group.Done()
//
//	threadByte := uint8(threadId)
//
//	tracer.RecordAction(WorkerStart{threadByte})
//
//	guessPrefix := []uint8{0}
//
//	for {
//		select {
//		case <-cancel:
//			tracer.RecordAction(WorkerCancelled{threadByte})
//			answer <- []uint8{0}
//			return
//		default:
//			start := threadId << (8 - threadBits)
//			finish := (threadId + 1) << (8 - threadBits)
//			for i := start; i < finish; i++ {
//				guess := []uint8{uint8(i)}
//				guess = append(guess, guessPrefix...)
//				appendedGuess := append(nonce, guess...)
//				checksum := md5.Sum(appendedGuess)
//				if checkTrailingZeros(checksum, cmpStr) {
//					tracer.RecordAction(WorkerSuccess{threadByte, guess})
//					answer <- guess
//					<-cancel
//					return
//				}
//			}
//			increment(&guessPrefix)
//		}
//	}
//}

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

	//var arith = new(distpow.Arith)
	//var worker = new(distpow.Worker)
	err = rpc.Register(worker)
	if err != nil {
		log.Fatal("error registering API: ", err)
	}

	rpc.HandleHTTP()

	//listener, err := net.Listen("tcp", config.CoordAddr)
	listener, err := net.Listen("tcp", ":33042")
	if err != nil {
		log.Fatal("Listener error: ", err)
	}

	//log.Printf("serving rpc on port %s", config.CoordAddr)
	log.Printf("serving rpc on port %s", ":33042")

	err = http.Serve(listener, nil)
	if err != nil {
		log.Fatal("error serving: ", err)
	}

	log.Println(config)
}
