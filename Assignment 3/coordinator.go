package distpow

import (
	"log"
	"net/rpc"
)

type WorkerAddr string

type CoordinatorConfig struct {
	ClientAPIListenAddr string
	WorkerAPIListenAddr string
	Workers             []WorkerAddr
	TracerServerAddr    string
	TracerSecret        []byte
}

type CoordinatorMine struct {
	Nonce            []uint8
	NumTrailingZeros uint
}

type CoordinatorWorkerMine struct {
	Nonce            []uint8
	NumTrailingZeros uint
	WorkerByte       uint8
}

type CoordinatorWorkerResult struct {
	Nonce            []uint8
	NumTrailingZeros uint
	WorkerByte       uint8
	Secret           []uint8
}

type CoordinatorWorkerCancel struct {
	Nonce            []uint8
	NumTrailingZeros uint
	WorkerByte       uint8
}

type CoordinatorSuccess struct {
	Nonce            []uint8
	NumTrailingZeros uint
	Secret           []uint8
}

//type Args struct {
//	A, B int
//}
//
//type Quotient struct {
//	Quo, Rem int
//}

//type CoordinatorResultArgs struct {
//	Nonce                         []uint8
//	NumTrailingZeroes, workerByte uint8
//	secret                        []uint8
//}

type Coordinator struct {
	config CoordinatorConfig
	//done          chan bool
	coordToWorker []*rpc.Client
	answer        chan []uint8
}

func (c *Coordinator) Initialize(config CoordinatorConfig) error {
	c.config = config
	coordinatorToWorker, err := rpc.DialHTTP("tcp", string(c.config.Workers[0]))
	if err != nil {
		log.Fatal("Connection error: ", err)
		return err
	}
	c.coordToWorker = append(c.coordToWorker, coordinatorToWorker)
	c.answer = make(chan []uint8)
	return nil
	//return errors.New("not implemented")
}

func (c *Coordinator) Mine(args *CoordinatorMine, secret *[]uint8) error {
	//var config CoordinatorConfig
	//err := ReadJSONConfig("config/coordinator_config.json", &config)
	//if err != nil {
	//	log.Fatal(err)
	//}

	//coordinatorToWorker, err := rpc.DialHTTP("tcp", c.config.WorkerAPIListenAddr)
	//if err != nil {
	//	log.Fatal("Connection error: ", err)
	//}

	//var workerReply []uint8

	log.Println("Coordinator.Mine start")
	workerArgs := WorkerMine{args.Nonce, args.NumTrailingZeros, 0}
	c.coordToWorker[0].Go("Worker.Mine", workerArgs, nil, nil)
	//coordinatorToWorker.Go("Worker.Mine", workerArgs, nil, nil)

	//log.Println("finish: ", workerReply)

	//log.Println(c.config)

	*secret = <-c.answer

	log.Println(*secret)

	return nil
}

func (c *Coordinator) Result(args *CoordinatorWorkerResult, unused *uint) error {
	log.Println("Coordinator.Result called")
	workerArgs := WorkerMine{args.Nonce, args.NumTrailingZeros, args.WorkerByte}

	// for loop over all workers????
	//coordinator, err := rpc.DialHTTP("tcp", "localhost"+c.config.WorkerAPIListenAddr)
	//if err != nil {
	//	log.Fatal("Connection error: ", err)
	//}

	//coordinator.Go("Worker.Cancel", workerArgs, nil, nil)
	c.coordToWorker[0].Go("Worker.Cancel", workerArgs, nil, nil)

	c.answer <- args.Secret

	log.Println("Coordinator.Result end")

	return nil
}
