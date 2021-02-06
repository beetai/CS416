package distpow

import (
	"github.com/DistributedClocks/tracing"
	"log"
	"math"
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

// Coordinator custom structs
type Coordinator struct {
	//done          chan bool
	//coordToWorker []*rpc.Client
	config     CoordinatorConfig
	threadBits uint
	answer     chan []uint8
	tracer     *tracing.Tracer
}

//type CoordinatorResultArgs struct {
//	Nonce            []uint8
//	NumTrailingZeros uint
//	WorkerByte       uint8
//	Secret           []uint8
//	//JobId            int
//}

func (c *Coordinator) Initialize(config CoordinatorConfig) error {
	tracerConfig := tracing.TracerConfig{
		ServerAddress:  config.TracerServerAddr,
		TracerIdentity: "coordinator",
		Secret:         config.TracerSecret,
	}
	c.config = config
	//coordinatorToWorker, err := rpc.DialHTTP("tcp", string(c.config.Workers[0]))
	//if err != nil {
	//	log.Fatal("Connection error: ", err)
	//	return err
	//}
	//c.coordToWorker = append(c.coordToWorker, coordinatorToWorker)
	c.tracer = tracing.NewTracer(tracerConfig)
	c.answer = make(chan []uint8)
	c.threadBits = uint(math.Log2(float64(len(c.config.Workers))))
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

	//log.Println("Coordinator.Mine start")

	c.tracer.RecordAction(*args)

	//workerArgs := WorkerMine{args.Nonce, args.NumTrailingZeros, 0}
	//c.coordToWorker[0].Go("Worker.Mine", workerArgs, nil, nil)
	//coordinatorToWorker.Go("Worker.Mine", workerArgs, nil, nil)
	var coordinatorToWorkers []*rpc.Client
	for _, port := range c.config.Workers {
		coordinatorToWorker, err := rpc.DialHTTP("tcp", string(port))
		for err != nil {
			//log.Fatal("Connection error: ", err)
			coordinatorToWorker, err = rpc.DialHTTP("tcp", string(port))
		}
		coordinatorToWorkers = append(coordinatorToWorkers, coordinatorToWorker)
	}

	for i, coordinatorToWorker := range coordinatorToWorkers {
		workerArgs := WorkerMineArgs{args.Nonce, args.NumTrailingZeros, uint8(i), c.threadBits}
		c.tracer.RecordAction(CoordinatorWorkerMine{
			args.Nonce,
			args.NumTrailingZeros,
			uint8(i),
		})
		coordinatorToWorker.Go("Worker.Mine", workerArgs, nil, nil)
	}

	//log.Println("finish: ", workerReply)

	//log.Println(c.config)

	*secret = <-c.answer

	c.tracer.RecordAction(CoordinatorSuccess{
		Nonce:            args.Nonce,
		NumTrailingZeros: args.NumTrailingZeros,
		Secret:           *secret,
	})

	log.Println(*secret)

	return nil
}

func (c *Coordinator) Result(args *CoordinatorWorkerResult, unused *uint) error {
	//log.Println("Coordinator.Result called")
	//log.Printf("jobId: %d\n", args.JobId)
	c.tracer.RecordAction(*args)
	//workerArgs := WorkerCancelArgs{args.Nonce, args.NumTrailingZeros, args.WorkerByte, args.JobId}

	// for loop over all workers????
	//coordinator, err := rpc.DialHTTP("tcp", "localhost"+c.config.WorkerAPIListenAddr)
	//if err != nil {
	//	log.Fatal("Connection error: ", err)
	//}

	//coordinator.Go("Worker.Cancel", workerArgs, nil, nil)
	//c.coordToWorker[0].Go("Worker.Cancel", workerArgs, nil, nil)'

	stopped := make(chan *rpc.Call, len(c.config.Workers)-1)

	for i, port := range c.config.Workers {
		coordinatorToWorker, err := rpc.DialHTTP("tcp", string(port))
		if err != nil {
			log.Fatal("Connection error: ", err)
			return err
		}

		if uint8(i) != args.WorkerByte {
			workerArgs := CoordinatorWorkerCancel{
				args.Nonce,
				args.NumTrailingZeros,
				uint8(i),
			}
			c.tracer.RecordAction(workerArgs)
			//workerArgs := WorkerCancelArgs{args.Nonce, args.NumTrailingZeros, uint8(i), args.JobId}
			coordinatorToWorker.Go("Worker.Cancel", workerArgs, nil, stopped)
		}
		//coordinatorToWorker.Go("Worker.Cancel", workerArgs, nil, stopped)
	}

	for i := 0; i < len(c.config.Workers)-1; i++ {
		<-stopped
	}

	c.answer <- args.Secret

	//log.Println("Coordinator.Result end")

	return nil
}
