package distpow

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"github.com/DistributedClocks/tracing"
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

// https://www.reddit.com/r/golang/comments/a2b4iu/how_to_storeretrieve_channels_in_sync_map/
//type Map struct {
//	syncMap sync.Map
//}
//
//func (m *Map) Load(key int) chan bool {
//	val, ok := m.syncMap.Load(key)
//	if ok {
//		return val.(chan bool)
//	} else {
//		return nil
//	}
//}
//
//func (m *Map) Exists(key int) bool {
//	_, ok := m.syncMap.Load(key)
//	return ok
//}
//
//func (m *Map) Store(key int, value chan bool) {
//	m.syncMap.Store(key, value)
//}
//
//func (m *Map) Delete(key int) {
//	m.syncMap.Delete(key)
//}

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

type Worker struct {
	config  WorkerConfig
	doneMap map[string]chan bool
	tracer  *tracing.Tracer
}

func (w *Worker) Initialize(config WorkerConfig) error {
	tracerConfig := tracing.TracerConfig{
		ServerAddress:  config.TracerServerAddr,
		TracerIdentity: config.WorkerID,
		Secret:         config.TracerSecret,
	}
	w.config = config
	w.tracer = tracing.NewTracer(tracerConfig)
	w.doneMap = make(map[string]chan bool)
	return nil
}

func (w *Worker) Mine(args *WorkerMineArgs, unused *uint) error {
	w.tracer.RecordAction(WorkerMine{
		args.Nonce,
		args.NumTrailingZeros,
		args.WorkerByte,
	})

	// Create job hash
	jobHash := md5.Sum(append(args.Nonce, uint8(args.NumTrailingZeros)))
	jobHashStr := hex.EncodeToString(jobHash[:])
	w.doneMap[jobHashStr] = make(chan bool, 10)

	var buffer bytes.Buffer
	for i := 0; i < int(args.NumTrailingZeros); i++ {
		buffer.WriteString("0")
	}
	cmpStr := buffer.String()

	guessPrefix := []uint8{0}

	for {
		select {
		case <-w.doneMap[jobHashStr]:
			//w.tracer.RecordAction(WorkerCancel{
			//	args.Nonce,
			//	args.NumTrailingZeros,
			//	args.WorkerByte,
			//})
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
					coordArgs := WorkerResult{
						args.Nonce,
						args.NumTrailingZeros,
						args.WorkerByte,
						guess,
					}
					w.tracer.RecordAction(coordArgs)
					workerToCoord, err := rpc.DialHTTP("tcp", w.config.CoordAddr)
					if err != nil {
						log.Println("Connection error: ", err)
						return err
					}
					workerToCoord.Go("Coordinator.Result", coordArgs, nil, nil)
					return nil
				}
			}
			increment(&guessPrefix)
		}
	}
}

func (w *Worker) Cancel(args *WorkerCancel, unused *uint) error {
	w.tracer.RecordAction(*args)
	jobHash := md5.Sum(append(args.Nonce, uint8(args.NumTrailingZeros)))
	jobHashStr := hex.EncodeToString(jobHash[:])
	w.doneMap[jobHashStr] <- true
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
