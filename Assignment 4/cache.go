package distpow

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/DistributedClocks/tracing"
)

type CacheAdd struct {
	Nonce            []uint8
	NumTrailingZeros uint
	Secret           []uint8
}

type CacheRemove struct {
	Nonce            []uint8
	NumTrailingZeros uint
	Secret           []uint8
}

type CacheHit struct {
	Nonce            []uint8
	NumTrailingZeros uint
	Secret           []uint8
}

type CacheMiss struct {
	Nonce            []uint8
	NumTrailingZeros uint
}

type CacheValue struct {
	secret        []uint8
	trailingZeros uint
}

type Cache struct {
	cacheMap map[string]CacheValue
}

//key := strings.Trim(strings.Join(strings.Fields(fmt.Sprint(args.Nonce)), ""), "[]")
func NewCache() Cache {
	return Cache{
		cacheMap: make(map[string]CacheValue),
	}
}

func (c *Cache) Exists(trace *tracing.Trace, nonce []uint8, trailingZeros uint) bool {
	key := generateCacheKey(nonce)
	val, ok := c.cacheMap[key]
	if !ok || trailingZeros > val.trailingZeros {
		trace.RecordAction(CacheMiss{
			Nonce:            nonce,
			NumTrailingZeros: trailingZeros,
		})
		return false
	}
	trace.RecordAction(CacheHit{
		Nonce:            nonce,
		NumTrailingZeros: trailingZeros,
	})
	return true
}

func (c *Cache) Store(trace *tracing.Trace, nonce []uint8, trailingZeros uint, secret []uint8) {
	key := generateCacheKey(nonce)
	val, ok := c.cacheMap[key]
	if !ok {
		//trace.RecordAction(CacheMiss{
		//	Nonce:            nonce,
		//	NumTrailingZeros: trailingZeros,
		//})
		trace.RecordAction(CacheAdd{
			Nonce:            nonce,
			NumTrailingZeros: trailingZeros,
			Secret:           secret,
		})
		c.cacheMap[key] = CacheValue{
			secret:        secret,
			trailingZeros: trailingZeros,
		}
		return
	}

	if bytes.Compare(secret, val.secret) > 0 {
		trace.RecordAction(CacheRemove{
			Nonce:            nonce,
			NumTrailingZeros: val.trailingZeros,
			Secret:           val.secret,
		})
		trace.RecordAction(CacheAdd{
			Nonce:            nonce,
			NumTrailingZeros: trailingZeros,
			Secret:           secret,
		})
		c.cacheMap[key] = CacheValue{
			secret:        secret,
			trailingZeros: trailingZeros,
		}
	}
}

func (c *Cache) Load(nonce []uint8) []uint8 {
	key := generateCacheKey(nonce)
	if val, ok := c.cacheMap[key]; ok {
		return val.secret
	} else {
		return nil
	}
}

func generateCacheKey(nonce []uint8) string {
	return fmt.Sprintf("%s", hex.EncodeToString(nonce))
}
