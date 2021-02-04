package distpow

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"log"
	"net/rpc"
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

type Worker struct {
	config WorkerConfig
	done   chan bool
	//workerToCoord *rpc.Client
}

func (w *Worker) Initialize(config WorkerConfig) error {
	w.config = config
	//workerToCoord, err := rpc.DialHTTP("tcp", w.config.CoordAddr)
	//if err != nil {
	//	log.Fatal("Connection error: ", err)
	//	return err
	//}
	//w.workerToCoord = workerToCoord
	return nil
	//return errors.New("not implemented")
}

func (w *Worker) Mine(args *WorkerMine, unused *uint) error {
	log.Println("Worker.Mine start")
	guess := []uint8{0}

	hashStrBuf := new(bytes.Buffer)

	for {
		appendedGuess := append(args.Nonce, guess...)
		checksum := md5.Sum(appendedGuess)
		hashStrBuf.Reset()
		fmt.Fprintf(hashStrBuf, "%x", checksum)
		if hasNumZeroesSuffix(hashStrBuf.Bytes(), args.NumTrailingZeros) {
			break
		}
		increment(&guess)
	}

	//*secret = guess
	log.Println("Worker.Mine answer found")

	workerToCoord, err := rpc.DialHTTP("tcp", w.config.CoordAddr)
	if err != nil {
		log.Fatal("Connection error: ", err)
		return err
	}
	coordArgs := CoordinatorWorkerResult{args.Nonce, args.NumTrailingZeros, args.WorkerByte, guess}
	workerToCoord.Call("Coordinator.Result", coordArgs, nil)

	return nil
}

//func (w *Worker) Cancel(args *WorkerCancel, unused *uint) error {
//	if ()
//}

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

//func checkTrailingZeros(checksum [16]byte, cmpStr string) bool {
//	str := hex.EncodeToString(checksum[:])
//	str = str[len(str)-len(cmpStr):]
//
//	return str == cmpStr
//}

func hasNumZeroesSuffix(str []byte, numZeroes uint) bool {
	var trailingZeroesFound uint
	for i := len(str) - 1; i >= 0; i-- {
		if str[i] == '0' {
			trailingZeroesFound++
		} else {
			break
		}
	}
	return trailingZeroesFound >= numZeroes
}
