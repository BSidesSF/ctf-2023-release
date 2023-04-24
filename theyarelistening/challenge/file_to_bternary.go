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

	fmt.Fprint(os.Stderr, "Converting file to big.Int\n")
	fileint := new(big.Int)
	fileint.SetBytes(fbuf)

	fmt.Fprint(os.Stderr, "Converting big.Int into trinary\n")
	trinary := fileint.Text(3)

	fmt.Fprintf(os.Stderr, "File in trinary: %s\n", trinary)

	btern := make([]byte, len(trinary) + 1)

	btern[0] = 0;
	for i := 0; i < len(trinary); i++ {
		btern[i + 1] = trinary[i] - 0x30
	}

	for i := len(btern) - 1; i >= 0; i-- {
		if btern[i] >= 2 {
			switch btern[i] {
			case 3:
				btern[i] = 0
				btern[i - 1] += 1
			case 2:
				btern[i - 1] += 1
			default:
				fmt.Fprintf(os.Stderr, "Impossible digit! Exiting\n")
				os.Exit(-1)
			}
		}
	}


	for i, v := range btern {
		if i == 0 && v == 0 {
			continue
		}

		fmt.Fprintf(os.Stdout, "%d", v)
	}

	fmt.Fprintf(os.Stdout, "\n")
}
