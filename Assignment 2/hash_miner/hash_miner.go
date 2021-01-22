package hash_miner

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"github.com/DistributedClocks/tracing"
	"math"
	"sync"
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

func worker(tracer *tracing.Tracer, threadId int, nonce []uint8, cmpStr string, threadBits uint, answer chan []uint8, done *bool, group *sync.WaitGroup) {
	defer group.Done()

	threadByte := uint8(threadId)

	tracer.RecordAction(WorkerStart{threadByte})

	guessPrefix := []uint8{0}

	for !*done {
		start := threadId << (8 - threadBits)
		finish := (threadId + 1) << (8 - threadBits)
		for i := start; i < finish; i++ {
			guess := []uint8{uint8(i)}
			guess = append(guess, guessPrefix...)
			appendedGuess := append(nonce, guess...)
			checksum := md5.Sum(appendedGuess)
			if checkTrailingZeros(checksum, cmpStr) {
				tracer.RecordAction(WorkerSuccess{threadByte, guess})
				*done = true
				answer <- guess
				return
			}
		}
		increment(&guessPrefix)
	}

	tracer.RecordAction(WorkerCancelled{threadByte})
}

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

	var buffer bytes.Buffer
	for i := 0; i < int(numTrailingZeroes); i++ {
		buffer.WriteString("0")
	}
	cmpStr := buffer.String()

	// concurrency implementation
	answer := make(chan []uint8)
	done := false

	var group sync.WaitGroup
	for i := 0; i < int(math.Pow(2, float64(threadBits))); i++ {
		group.Add(1)
		go worker(tracer, i, nonce, cmpStr, threadBits, answer, &done, &group)
	}

	result := <-answer
	group.Wait()

	// synchronous implementation
	//guess := []uint8{0}
	//
	//for {
	//	appendedGuess := append(nonce, guess...)
	//	checksum := md5.Sum(appendedGuess)
	//	if checkTrailingZeros(checksum, cmpStr) {
	//		break
	//	}
	//	increment(&guess)
	//}
	//
	//result := guess

	tracer.RecordAction(MiningComplete{result})

	return result, nil
}
