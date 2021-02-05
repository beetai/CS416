package distpow

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"github.com/DistributedClocks/tracing"
	"log"
	"net/rpc"
	"sync"
)

type WorkerConfig struct {
	WorkerID         string
	ListenAddr       string
	CoordAddr        string
	TracerServerAddr string
	TracerSecret     []byte
}

type WorkerMine struct {
	Nonce            []uint8
	NumTrailingZeros uint
	WorkerByte       uint8
}

type WorkerResult struct {
	Nonce            []uint8
	NumTrailingZeros uint
	WorkerByte       uint8
	Secret           []uint8
}

type WorkerCancel struct {
	Nonce            []uint8
	NumTrailingZeros uint
	WorkerByte       uint8
}

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

// Worker code
//type WorkerMineArgs struct {
//	//tracer *tracing.Tracer
//	Nonce                         []uint8
//	numTrailingZeroes, workerByte uint8
//	//CmpStr                        string
//}

// https://www.reddit.com/r/golang/comments/a2b4iu/how_to_storeretrieve_channels_in_sync_map/
type Map struct {
	syncMap sync.Map
}

func (m *Map) Load(key int) chan bool {
	val, ok := m.syncMap.Load(key)
	if ok {
		return val.(chan bool)
	} else {
		return nil
	}
}

func (m *Map) Exists(key int) bool {
	_, ok := m.syncMap.Load(key)
	return ok
}

func (m *Map) Store(key int, value chan bool) {
	m.syncMap.Store(key, value)
}

func (m *Map) Delete(key int) {
	m.syncMap.Delete(key)
}

//var KillMap Map
//
//type KillChannel chan bool

// Worker custom structs
type WorkerMineArgs struct {
	Nonce            []uint8
	NumTrailingZeros uint
	WorkerByte       uint8
	ThreadBits       uint
}

type WorkerCancelArgs struct {
	Nonce            []uint8
	NumTrailingZeros uint
	WorkerByte       uint8
	JobId            int
}

type Worker struct {
	config        WorkerConfig
	doneMap       map[int]chan bool
	nextJobId     int
	workerToCoord *rpc.Client
	tracer        *tracing.Tracer
}

func (w *Worker) Initialize(config WorkerConfig) error {
	tracerConfig := tracing.TracerConfig{
		ServerAddress:  config.TracerServerAddr,
		TracerIdentity: config.WorkerID,
		Secret:         config.TracerSecret,
	}
	w.config = config
	workerToCoord, err := rpc.DialHTTP("tcp", w.config.CoordAddr)
	if err != nil {
		log.Fatal("Connection error: ", err)
		return err
	}
	w.tracer = tracing.NewTracer(tracerConfig)
	w.workerToCoord = workerToCoord
	w.doneMap = make(map[int]chan bool)
	w.nextJobId = 0
	return nil
	//return errors.New("not implemented")
}

func (w *Worker) Mine(args *WorkerMineArgs, unused *uint) error {
	//log.Println("Worker.Mine start")
	w.tracer.RecordAction(WorkerMine{
		args.Nonce,
		args.NumTrailingZeros,
		args.WorkerByte,
	})
	//guess := []uint8{0}
	jobId := w.nextJobId
	w.doneMap[jobId] = make(chan bool)
	w.nextJobId++

	var buffer bytes.Buffer
	for i := 0; i < int(args.NumTrailingZeros); i++ {
		buffer.WriteString("0")
	}
	cmpStr := buffer.String()

	//for {
	//	appendedGuess := append(args.Nonce, guess...)
	//	checksum := md5.Sum(appendedGuess)
	//	hashStrBuf.Reset()
	//	fmt.Fprintf(hashStrBuf, "%x", checksum)
	//	if hasNumZeroesSuffix(hashStrBuf.Bytes(), args.NumTrailingZeros) {
	//		break
	//	}
	//	increment(&guess)
	//}

	guessPrefix := []uint8{0}

	for {
		select {
		//case <-w.doneMap.Load(jobId):
		case <-w.doneMap[jobId]:
			//tracer.RecordAction(WorkerCancelled{threadByte})
			//answer <- []uint8{0}
			//log.Println("Worker.Cancelled")
			return nil
		default:
			start := int(args.WorkerByte) << (8 - args.ThreadBits)
			finish := int(args.WorkerByte+1) << (8 - args.ThreadBits)
			for i := start; i < finish; i++ {
				guess := []uint8{uint8(i)}
				guess = append(guess, guessPrefix...)
				appendedGuess := append(args.Nonce, guess...)
				checksum := md5.Sum(appendedGuess)
				if checkTrailingZeros(checksum, cmpStr) {
					//tracer.RecordAction(WorkerSuccess{threadByte, guess})
					//answer <- guess
					//<- cancel
					//log.Printf("Worker.Mine answer found: jobId is %d\n", jobId)
					coordArgs := CoordinatorResultArgs{args.Nonce, args.NumTrailingZeros, args.WorkerByte, guess, jobId}
					w.tracer.RecordAction(WorkerResult{
						args.Nonce,
						args.NumTrailingZeros,
						args.WorkerByte,
						guess,
					})
					w.workerToCoord.Go("Coordinator.Result", coordArgs, nil, nil)
					return nil
				}
			}
			increment(&guessPrefix)
		}
	}

	//*secret = guess
	//log.Println("Worker.Mine answer found")

	//workerToCoord, err := rpc.DialHTTP("tcp", w.config.CoordAddr)
	//if err != nil {
	//	log.Fatal("Connection error: ", err)
	//	return err
	//}
	//coordArgs := CoordinatorWorkerResult{args.Nonce, args.NumTrailingZeros, args.WorkerByte, guess}
	//w.workerToCoord.Call("Coordinator.Result", coordArgs, nil)

	//return nil
}

func (w *Worker) Cancel(args *WorkerCancelArgs, unused *uint) error {
	//log.Printf("Worker.Cancel called: jobId is %d\n", args.JobId)
	w.tracer.RecordAction(WorkerCancel{
		args.Nonce,
		args.NumTrailingZeros,
		args.WorkerByte,
	})
	w.doneMap[args.JobId] <- true
	//charleneIsTheSmartest := make(chan bool)
	//w.doneMap.Store(args.jobId, charleneIsTheSmartest)
	//charleneIsTheSmartest <- true
	return nil
}

// helpers
func increment(guess *[]uint8) {
	var acc uint8
	one := []uint8{1}
	for i := 0; i < len(*guess) || acc != 0; i++ {
		var tmp uint16 = uint16(acc)
		if i < len(*guess) {
			tmp += uint16((*guess)[i])
		}
		if i < len(one) {
			tmp += uint16(one[i])
		}
		acc = uint8(tmp >> 8)
		if i >= len(*guess) {
			*guess = append(*guess, uint8(tmp&0xFF))
		} else {
			(*guess)[i] = uint8(tmp & 0xFF)
		}
	}
}

func checkTrailingZeros(checksum [16]byte, cmpStr string) bool {
	str := hex.EncodeToString(checksum[:])
	str = str[len(str)-len(cmpStr):]

	return str == cmpStr
}
