package main

import (
	"encoding/hex"
	"testing"
)

func BenchmarkTestVal(b *testing.B) {
	targetBytes, _ := hex.DecodeString(targetHash)
	//src := getByteValues(6)
	for n := 0; n < b.N; n++ {
		//testVal(<-src, targetBytes)
		testVal([]byte("aaaaa"), targetBytes)
	}
}
