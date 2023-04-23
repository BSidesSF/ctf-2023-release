package main

import (
	"fmt"
	"os"
	"io/ioutil"
	"math/big"
)


func main() {

	fbuf, err := ioutil.ReadFile(os.Args[1])

	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to read flag: %s\n", err.Error())
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "Got balanced ternary: %s\n", string(fbuf))

	place := big.NewInt(1)
	fileint := big.NewInt(0)
	for i := int(len(fbuf)) - 1; i >= 0; i-- {
		var d int64

		switch fbuf[i] {
		case 0x30:
			d = 0
		case 0x31:
			d = 1
		case 0x32:
			d = -1
		default:
			continue
		}

		//fmt.Fprintf(os.Stderr, "Processing digit %d\n", d)

		fileint = fileint.Add(fileint, new(big.Int).Mul(place, big.NewInt(d)))

		place = place.Mul(place, big.NewInt(3))

		//fmt.Fprintf(os.Stderr, "On place %s\n", place.Text(3))
	}

	fmt.Fprintf(os.Stderr, "Computed ternary: %s\n", fileint.Text(3))

	fmt.Fprint(os.Stderr, "Writing file\n")
	err = ioutil.WriteFile("out.bin", fileint.Bytes(), 0644)

}
