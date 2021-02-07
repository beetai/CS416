package distpow

import (
	"crypto/md5"
	"encoding/hex"
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

// Coordinator custom structs
type Coordinator struct {
	config     CoordinatorConfig
	threadBits uint
	answerMap  map[string]chan CoordinatorWorkerResult
	tracer     *tracing.Tracer
}

func (c *Coordinator) Initialize(config CoordinatorConfig) error {
	tracerConfig := tracing.TracerConfig{
		ServerAddress:  config.TracerServerAddr,
		TracerIdentity: "coordinator",
		Secret:         config.TracerSecret,
	}
	c.config = config
	c.tracer = tracing.NewTracer(tracerConfig)
	c.answerMap = make(map[string]chan CoordinatorWorkerResult)
	c.threadBits = uint(math.Log2(float64(len(c.config.Workers))))
	return nil
}

func (c *Coordinator) Mine(args *CoordinatorMine, secret *[]uint8) error {
	c.tracer.RecordAction(*args)

	jobHash := md5.Sum(append(args.Nonce, uint8(args.NumTrailingZeros)))
	jobHashStr := hex.EncodeToString(jobHash[:])
	c.answerMap[jobHashStr] = make(chan CoordinatorWorkerResult)

	var coordinatorToWorkers []*rpc.Client
	for _, port := range c.config.Workers {
		coordinatorToWorker, err := rpc.DialHTTP("tcp", string(port))
		for err != nil {
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

	// CoordinatorWorkerResult
	cwr := <-c.answerMap[jobHashStr]

	c.tracer.RecordAction(cwr)

	stopped := make(chan *rpc.Call, len(c.config.Workers)-1)

	for i, port := range c.config.Workers {
		coordinatorToWorker, err := rpc.DialHTTP("tcp", string(port))
		if err != nil {
			log.Println("Connection error: ", err)
			return err
		}

		if uint8(i) != cwr.WorkerByte {
			workerArgs := CoordinatorWorkerCancel{
				args.Nonce,
				args.NumTrailingZeros,
				uint8(i),
			}
			c.tracer.RecordAction(workerArgs)
			coordinatorToWorker.Go("Worker.Cancel", workerArgs, nil, stopped)
		}
	}

	for i := 0; i < len(c.config.Workers)-1; i++ {
		<-stopped
	}

	*secret = cwr.Secret

	c.tracer.RecordAction(CoordinatorSuccess{
		Nonce:            args.Nonce,
		NumTrailingZeros: args.NumTrailingZeros,
		Secret:           *secret,
	})

	return nil
}

func (c *Coordinator) Result(args *CoordinatorWorkerResult, unused *uint) error {
	//c.tracer.RecordAction(*args)
	jobHash := md5.Sum(append(args.Nonce, uint8(args.NumTrailingZeros)))
	jobHashStr := hex.EncodeToString(jobHash[:])
	c.answerMap[jobHashStr] <- *args

	return nil
}
