package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

var charset = []byte("abcdefghijklmnopqrstuvwxyz0123456789")

const targetHash = "b3b366a6a156ecd6bc2e7f36aee253796dc04969a23152c97be368862c1c85b3"
const known = "trans"

func getByteValues(targetLen int) <-chan []byte {
	ch := make(chan []byte, 1)
	go func() {
		defer close(ch)
		var gen func([]byte)
		gen = func(part []byte) {
			for _, c := range charset {
				p := append(part, c)
				if len(p) == targetLen {
					ch <- p
				} else {
					gen(p)
				}
			}
		}
		gen([]byte{})
	}()
	return ch
}

func main() {
	targetBytes, err := hex.DecodeString(targetHash)
	if err != nil {
		panic(err)
	}
	if !testVal([]byte(known), targetBytes) {
		panic("failed test!!!")
	}
	start := time.Now()
	byteChan := getByteValues(5)
	for v := range byteChan {
		if testVal(v, targetBytes) {
			fmt.Printf("FOUND: %s\n", string(v))
			break
		}
	}
	end := time.Now()
	delta := end.Sub(start)
	fmt.Printf("Time: %0.2fs\n", delta.Seconds())
}

func testVal(val, targetBytes []byte) bool {
	seed := []byte("admin")
	for i, c := range val {
		seed[i] ^= c
	}
	h := sha256.New()
	h.Write(seed)
	keyVal := h.Sum(nil)
	h.Reset()
	h.Write(keyVal)
	keyHash := h.Sum(nil)
	return bytes.Equal(targetBytes, keyHash)
}
