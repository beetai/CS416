package hash_miner

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"github.com/DistributedClocks/tracing"
)

type WorkerStart struct {
	ThreadByte uint8
}

type WorkerSuccess struct {
	ThreadByte uint8
	Secret     []uint8
}

type WorkerCancelled struct {
	ThreadByte uint8
}

type MiningBegin struct{}

type MiningComplete struct {
	Secret []uint8
}

//func worker(tracer *tracing.Tracer, threadId uint8, nonce []uint8, numTrailingZeroes, threadBits uint, group *sync.WaitGroup) {
//	//defer group.Done()
//
//	tracer.RecordAction(WorkerStart{threadId})
//
//	for {
//		start :=
//		for i := start; i < finish; i++ {
//
//		}
//	}
//
//	tracer.RecordAction(WorkerStart{})
//}

func checkTrailingZeros(checksum [16]byte, cmpStr string) bool {
	str := hex.EncodeToString(checksum[:])
	str = str[len(str)-len(cmpStr):]

	return str == cmpStr
}

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

func Mine(tracer *tracing.Tracer, nonce []uint8, numTrailingZeroes, threadBits uint) (secret []uint8, err error) {
	tracer.RecordAction(MiningBegin{})

	// TODO
	//var group sync.WaitGroup
	//for i := 0; i < int(math.Pow(2, float64(threadBits))); i++ {
	//	group.Add(1)
	//	go worker(tracer, uint8(i), nonce, numTrailingZeroes, threadBits, &group)
	//}

	//guess := []uint8{194, 170, 210, 13}
	guess := []uint8{0}
	//guess := []uint8{169, 113, 171, 12}

	var buffer bytes.Buffer
	for i := 0; i < int(numTrailingZeroes); i++ {
		buffer.WriteString("0")
	}
	cmpStr := buffer.String()

	for {
		appendedGuess := append(nonce, guess...)
		checksum := md5.Sum(appendedGuess)
		if checkTrailingZeros(checksum, cmpStr) {
			break
		}
		increment(&guess)
	}

	result := guess

	tracer.RecordAction(MiningComplete{result})

	return result, nil
}
