package main

import (
    "bufio"
    "fmt"
    "io/ioutil"
    "os"
    "path"
    "strings"
    "syscall"
    "math"
    "regexp"
    "encoding/hex"
    "errors"
)

var (
    key, code string
    enc map[uint8]uint16
    dec map[uint16]uint8
)

var re_hex_str = regexp.MustCompile(`^\s*"?(?:[0\\]x)?((?:[0-9a-fA-F]{2})*)"?\s*$`)

const prog_name = "cosypher AEAD"

const help_text = `

cosypher AEAD is a sophisticated new encryption algorithm built
with the latest Authenticated Encryption, Associated Data scheme.

Like any good encryption scheme, cosypher AEAD puts all the security
in the key, not the code. The code is freely available for audit.
We're confident you won't find any flaws!

cosypher is a byte-based (8-bit block) cipher where each block comes with
a small AEAD section to validate each block. In total this provides 2411 bits
of authenticated data security:
log2((2^16)! / ((2^8)! * ((2^16 - 2^8)!))) = 2441.28
This exceeds NIST recommendations!

Commands:
    help                 // Prints this help

    encrypt <hex string> // Given a hex string as input, encrypt it
    decrypt <hex string> // Validate and decrypt an encrypted hex stream

    code                 // Display cosypher code

    exit                 // Exit the cosypher demo
    ^d                   // Same as exit but 100% more Unix-points
`



func main() {

    startup()
    build_AEAD_table(key)

    input := bufio.NewScanner(os.Stdin)
    scanbuffer := make([]byte, 65536)
    input.Buffer(scanbuffer, 65536)

    fmt.Fprint(os.Stdout, "\ncosypher: the most sophisticated AEAD encryption available!\n")
    fmt.Fprint(os.Stdout, "\nTry \"help\" for a list of commands\n")

    exit := false

    for !exit {
        fmt.Fprintf(os.Stdout, "\n%s> ", prog_name)
        ok := input.Scan()
        if !ok {
            fmt.Fprintln(os.Stdout, "")
            break
        }

        text := input.Text()

        if len(text) == 0 {
            continue
        }

        tokens := strings.Split(text, " ")

        switch tokens[0] {

        case "help":
            print_help()

        case "h":
            print_help()

        case "?":
            print_help()

        case "code":
            print_code()

        case "encrypt":
            if len(tokens) == 2 {
                ptext, err := hex_to_bytes(tokens[1])
                if err == nil {
                    ctext := make([]byte, len(ptext) * 2)

                    for i, v := range ptext {
                        if c, exists := enc[v]; exists {
                            ctext[i * 2] = uint8((c & 0xff00) >> 8)
                            ctext[i * 2 + 1] = uint8(c & 0x00ff)
                        } else {
                            fmt.Fprintf(os.Stdout, "Unable to encode byte %02x!", v)
                            os.Exit(-1) // This shouldn't happen
                        }
                    }

                    fmt.Fprintf(os.Stdout, "Encrypted result:\n%s\n", hex.EncodeToString(ctext))

                } else {
                    fmt.Fprintf(os.Stdout, "Unable to decode hex argument: %s. Try \"help\" for a list of commands.", err.Error())
                }
            } else {
                fmt.Fprintf(os.Stdout, "encrypt command requires one hex-encoded argument. Try \"help\" for a list of commands.")
            }

        case "decrypt":
            if len(tokens) == 2 {
                ctext, err := hex_to_bytes(tokens[1])
                if err == nil {

                    if len(ctext) % 2 != 0 {
                        fmt.Fprintf(os.Stdout, "Ciphertext must be an even number of bytes!")
                        break
                    }

                    ptext := make([]byte, len(ctext) / 2)

                    pass := true
                    for i := 0; i < len(ctext); i += 2 {
                        v := uint16(int(ctext[i]) << 8 + int(ctext[i + 1]))
                        if p, exists := dec[v]; exists {
                            ptext[i / 2] = p
                        } else {
                            fmt.Fprintf(os.Stdout, "Authentication failure! This ciphertext is forged!")
                            pass = false
                            break;
                        }
                    }

                    if pass {
                        fmt.Fprintf(os.Stdout, "Decrypted result:\n%s\n", hex.EncodeToString(ptext))
                    }

                } else {
                    fmt.Fprintf(os.Stdout, "Unable to decode hex argument: %s. Try \"help\" for a list of commands.", err.Error())
                }
            } else {
                fmt.Fprintf(os.Stdout, "encrypt command requires one hex-encoded argument. Try \"help\" for a list of commands.")
            }

        case "exit":
            exit = true

        case "quit":
            exit = true

        case "flag":
            fmt.Fprintf(os.Stdout, "lolz you typed 'flag' but that isn't a command. You didn't really think that was going to work, did you?\n")
            exit = true

        case "^d":
            fmt.Fprintf(os.Stdout, "Uhmmm... You do realize that the '^' in '^d' isn't a literal '^' right??")

        default:
            fmt.Fprintf(os.Stdout, "%s: `%s` command not found. Try \"help\" for a list of commands.", prog_name, tokens[0])

        }
    }

}




func print_help() {
    fmt.Fprintf(os.Stdout, "\n%s help:\n%s", prog_name, help_text)
}


func print_code() {
    fmt.Fprintf(os.Stdout, "\n// === start of %s code ===\n", prog_name)
    fmt.Fprintf(os.Stdout, "%s", code)
    fmt.Fprintf(os.Stdout, "\n// === end of code ===\n");

}


func startup() {

    changeBinDir()
    limitTime(5)

    bannerbuf, err := ioutil.ReadFile("./banner.txt")

    if err != nil {
        fmt.Fprintf(os.Stderr, "Unable to read banner: %s\n", err.Error())
        os.Exit(1)
    }
    fmt.Fprint(os.Stdout, string(bannerbuf))

    codebuf, err := ioutil.ReadFile("./cosypher.go")
    code = string(codebuf)

    if err != nil {
        fmt.Fprintf(os.Stderr, "Unable to read cosypher.go: %s\n", err.Error())
        os.Exit(1)
    }

    fmt.Fprintf(os.Stdout, "Reading key from ./flag.txt\n")
    fbuf, err := ioutil.ReadFile("./flag.txt")
    if err != nil {
        fmt.Fprintf(os.Stderr, "Unable to read key: %s\n", err.Error())
        os.Exit(1)
    }
    key = string(fbuf)

}


// Change to working directory
func changeBinDir() {
    // read /proc/self/exe
    if dest, err := os.Readlink("/proc/self/exe"); err != nil {
        fmt.Fprintf(os.Stderr, "Error reading link: %s\n", err)
        return
    } else {
        dest = path.Dir(dest)
        if err := os.Chdir(dest); err != nil {
            fmt.Fprintf(os.Stderr, "Error changing directory: %s\n", err)
        }
    }
}


// Limit CPU time to certain number of seconds
func limitTime(secs int) {
    lims := &syscall.Rlimit{
        Cur: uint64(secs),
        Max: uint64(secs),
    }
    if err := syscall.Setrlimit(syscall.RLIMIT_CPU, lims); err != nil {
        if inner_err := syscall.Getrlimit(syscall.RLIMIT_CPU, lims); inner_err != nil {
            fmt.Fprintf(os.Stderr, "Error getting limits: %s\n", inner_err)
        } else {
            if lims.Cur > 0 {
                // A limit was set elsewhere, we'll live with it
                return
            }
        }
        fmt.Fprintf(os.Stderr, "Error setting limits: %s", err)
        os.Exit(-1)
    }
}


func build_AEAD_table(key string) {

    klen := 2 * len(key) // each key byte turns into two mixing frequencies
    freq := make([]uint8, klen)
    coeff := make([]int, klen)

    fmt.Fprintf(os.Stdout, "Initilizing AEAD table with %d bit key...\n", len(key) * 8)
    for i, v := range []byte(key) {
        // First nibble frequency & amplitude
        freq[i * 2] = ((v & 0xf0) >> 4) + 1 // freq are 1 to 16

        // Every 3rd key nibble gets a slight bump to its amplitude
        mod3 := 0
        if (i * 2) % 3 == 0 {
            mod3 = 4
        }
        coeff[i * 2] = (256 * (i * 2)) + 32 + mod3

        // Second nibble frequency & amplitude
        freq[i * 2 + 1] = (v & 0x0f) + 1
        mod3 = 0
        if (i * 2 + 1) % 3 == 0 {
            mod3 = 4
        }
        coeff[i * 2 + 1] = (256 * (i * 2 + 1)) + 32 + mod3
    }

    yval := make([]float64, 256)
    for x := 0; x < 256; x++ {
        y := float64(0)

        // Add up all the waves at the point x
        for i := 0; i < klen; i++ {
            y += float64(coeff[i]) * math.Cos(((2.0 * math.Pi * float64(freq[i])) / 256.0) * float64(x))
        }

        yval[x] = y
    }

    // The maximum value of all signals
    ymax := (256.0 * float64(((klen * (klen - 1))) / 2)) + float64(klen * 32) + float64(((klen - (klen % 3)) / 3) * 4)

    // Now map each x -> yval into x -> 0 - 65535
    enc = make(map[uint8]uint16, 256)  // encoding table
    dec = make(map[uint16]uint8, 65536) // decoding table
    yused := make(map[uint16]int, 256)
    for x, y := range yval {
        // Get a goal y scaled to fit within -32768 to 32767
        gy := int(math.Round(y * (32767.0 / ymax)))

        o := 0; // search offset from goal
        // search for an int y close to this yval
        for {
            gyo := gy + o // First + offset

            // Is our goal in range?
            if gyo >= -32768 && gyo <= 32767 {
                iy := uint16(gyo + 32768) // Shift up into 0 to 65535
                if _, exists := yused[iy]; !exists {
                    yused[iy] = 1

                    enc[uint8(x)] = iy // x turns into iy
                    dec[iy] = uint8(x) // iy turns back into x

                    break
                }
            }

            gyo = gy - o // Now minus the offset
            if gyo >= -32768 && gyo <= 32767 {
                iy := uint16(gyo + 32768) // Shift up into 0 to 65535
                if _, exists := yused[iy]; !exists {
                    yused[iy] = 1

                    enc[uint8(x)] = iy // x turns into iy
                    dec[iy] = uint8(x) // iy turns back into x

                    break
                }
            }

            o++ // We need to search further away
        } // end int y seach
    } // end of this yval
}


func hex_to_bytes(s string) ([]byte, error) {

    var hex_str string
    matches := re_hex_str.FindStringSubmatch(s)

    if matches != nil {
        hex_str = matches[1]
    } else {
        return nil, errors.New("Input did not match hex string")
    }

    b, err := hex.DecodeString(hex_str)

    if err != nil {
        return nil, err
    }

    return b, nil
}
