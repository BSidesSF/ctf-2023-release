package main

import (
    "bufio"
    "fmt"
    "io/ioutil"
    "os"
    "path"
    "strings"
    "syscall"
    "math/big"
    cryptor "crypto/rand"
    "strconv"
)

type pubkey struct {
    fname, lname string
    messages []string
    n, e *big.Int
}


var (
    flag string
    keylist []*pubkey
)


const prog_name = "keyservice"

const help_text = `

Commands:
    help                 // Prints this help

    help types           // Display help about the key types

    listkeys             // Display all current public keys
    genkey               // Generate a new keypair

    listmsgs             // Get a list of unread messages
    sendmsg              // Send a user a message
    readmsg              // Read your message(s)

    exit                 // Exit the keyservice
`


const help_types_text = `
The keyservice has the ability to generate a "standard" key or a
signing-optimized key.  If you are unsure which to pick, go with a
standard key.

Standard keys are optimized for fast encryption and serve most users
well. Signing-optimized keys are much slower for encryption but enable
substantially faster signing.  Signing-optimized keys are still secure
by ensuring your private exponent has 128 bits of entropy. The only
drawback in that your resulting public exponent is much larger.
`



func main() {

    startup()

    input := bufio.NewScanner(os.Stdin)
    scanbuffer := make([]byte, 65536)
    input.Buffer(scanbuffer, 65536)


    // Make admin key
    akey, _ := gen_new_key(true)

    akey.fname = "Michael"
    akey.lname = "Wiener"
    akey.messages = make([]string, 0)

    akey.messages = append(akey.messages, fmt.Sprintf("Michael the flag you requested is %s", flag))

    keylist = append(keylist, akey)

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
            if len(tokens) > 1 {
                switch tokens[1] {
                    case "types":
                    fmt.Fprintf(os.Stdout, "%s", help_types_text)

                }
            } else {
                print_help()
            }

        case "h":
            print_help()

        case "?":
            print_help()

        case "listkeys":
            for i, k := range keylist {
                fmt.Fprintf(os.Stdout, "\nUser %2d. (%s, %s)\n", i, k.fname, k.lname)
                fmt.Fprintf(os.Stdout, "Public modulus: %s\n", (k.n).Text(10))
                fmt.Fprintf(os.Stdout, "Public exponent: %s\n", (k.e).Text(10))
            }

        case "listmsgs":
            for i, k := range keylist {
                fmt.Fprintf(os.Stdout, "User %2d. (%s, %s): %d unread messages\n", i, k.fname, k.lname, len(k.messages))
            }

        case "sendmsg":
            fmt.Fprint(os.Stdout, "\nEnter user (by number) to send a message? ")
            ok = input.Scan()
            if !ok {
                fmt.Fprintln(os.Stdout, "Error reading input!")
                break
            }
            unum := input.Text()

            intin, err := strconv.Atoi(unum)

            if err != nil {
                fmt.Fprintln(os.Stdout, "Error, could not interpret input as number!")
                break
            }

            if intin < 0 && intin >= len(keylist) {
                fmt.Fprintf(os.Stdout, "Error, user %d does not exist!", intin)
                break
            }

            fmt.Fprintf(os.Stdout, "\nMessage for user %d? ", intin)
            ok = input.Scan()
            if !ok {
                fmt.Fprintln(os.Stdout, "Error reading input!")
                break
            }
            umsg := input.Text()

            keylist[intin].messages = append(keylist[intin].messages, umsg)

            fmt.Fprintln(os.Stdout, "Message sent!")


        case "readmsg":
            fmt.Fprint(os.Stdout, "\nEnter user (by number) to read unread messages? ")
            ok = input.Scan()
            if !ok {
                fmt.Fprintln(os.Stdout, "Error reading input!")
                break
            }
            unum := input.Text()

            intin, err := strconv.Atoi(unum)

            if err != nil {
                fmt.Fprintln(os.Stdout, "Error, could not interpret input as number!")
                break
            }

            if intin < 0 || intin >= len(keylist) {
                fmt.Fprintf(os.Stdout, "Error, user %d does not exist!", intin)
                break
            }

            chal, err := cryptor.Int(cryptor.Reader, new(big.Int).Exp(big.NewInt(2), big.NewInt(128), nil))

            if err != nil {
                fmt.Fprintln(os.Stdout, "Unable to generate random message!")
                os.Exit(1);
            }

            fmt.Fprintf(os.Stdout, "\nWelcome %s(?), before we continue, you must prove it's you\n", keylist[intin].fname)
            fmt.Fprintf(os.Stdout, "\nPlease provide the signature for %s : ", chal.Text(10))
            ok = input.Scan()
            if !ok {
                fmt.Fprintln(os.Stdout, "Error reading input!")
                break
            }

            chal_sig, ok := new(big.Int).SetString(input.Text(), 10)
            if !ok || chal_sig == nil {
                fmt.Fprintln(os.Stdout, "Error parsing provided signature!")
                break
            }

            verify := new(big.Int).Exp(chal_sig, keylist[intin].e, keylist[intin].n)
            if chal.Cmp(verify) != 0 {
                fmt.Fprintln(os.Stdout, "Signature validation failed!")
                //fmt.Fprintln(os.Stdout, "debug: using n: %s\n", keylist[intin].n.Text(10))
                //fmt.Fprintln(os.Stdout, "debug: using e: %s\n", keylist[intin].e.Text(10))
                //fmt.Fprintln(os.Stdout, "debug: got verify of %s\n", verify.Text(10))
                break
            }

            if len(keylist[intin].messages) == 0 {
                fmt.Fprintf(os.Stdout, "%s doesn't have any unread messages\n", keylist[intin].fname)
                break
            }

            for i, m := range keylist[intin].messages {
                fmt.Fprintf(os.Stdout, "Message %d: %s\n", i + 1, m)
            }
            // clear messages
            keylist[intin].messages = make([]string, 0)


        case "genkey":
            fmt.Fprint(os.Stdout, "\nFirst name? ")
            ok = input.Scan()
            if !ok {
                fmt.Fprintln(os.Stdout, "Error reading input!")
                break
            }
            ufname := input.Text()

            fmt.Fprint(os.Stdout, "\nLast name? ")
            ok = input.Scan()
            if !ok {
                fmt.Fprintln(os.Stdout, "Error reading input!")
                break
            }
            ulname := input.Text()

            gotyn := false
            var gensmalld bool
            for !gotyn {
                fmt.Fprint(os.Stdout, "\nGenerate signing optimized key? (y/n) ")
                ok = input.Scan()
                if !ok {
                    fmt.Fprintln(os.Stdout, "Error reading input!")
                    break
                }
                yn := input.Text()

                switch yn {
                case "y":
                    gensmalld = true
                    gotyn = true
                case "n":
                    gensmalld = false
                    gotyn = true
                default:
                    fmt.Fprint(os.Stdout, "\nAnswer must be either y or n. See `help types` for an explanation.")
                }
            }

            fmt.Fprintf(os.Stdout, "\nGenerating key for %s...", ufname)
            ukey, ud := gen_new_key(gensmalld)

            ukey.fname = ufname
            ukey.lname = ulname
            ukey.messages = make([]string, 0)
            ukey.messages = append(ukey.messages, fmt.Sprintf("Welcome %s, Thank you for using the keyservice!", ufname))

            keylist = append(keylist, ukey)

            fmt.Fprintf(os.Stdout, "\n%s, your key is ready!\n", ufname)
            fmt.Printf("Public modulus: %s\n", (ukey.n).Text(10))
            fmt.Printf("Public exponent: %s\n", (ukey.e).Text(10))
            fmt.Printf("Private exponent: %s\n", (ud).Text(10))
            fmt.Fprint(os.Stdout, "\nRemember to save your private exponent! It is not stored by the keyservice and can never be recovered!")


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


func startup() {

    changeBinDir()
    limitTime(5)

    bannerbuf, err := ioutil.ReadFile("./banner.txt")

    if err != nil {
        fmt.Fprintf(os.Stderr, "Unable to read banner: %s\n", err.Error())
        os.Exit(1)
    }
    fmt.Fprint(os.Stdout, string(bannerbuf))

    fbuf, err := ioutil.ReadFile("./flag.txt")
    if err != nil {
        fmt.Fprintf(os.Stderr, "Unable to read flag: %s\n", err.Error())
        os.Exit(1)
    }
    flag = string(fbuf)

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


func gen_new_key(smalld bool) (*pubkey, *big.Int) {

    key := new(pubkey)

    fails := 0
retry_key:
    p, err := cryptor.Prime(cryptor.Reader, 1024)

    if err != nil {
        fmt.Fprintf(os.Stderr, "Error: unable to generate prime p\n")
        os.Exit(1)
    }

    q, err := cryptor.Prime(cryptor.Reader, 1024)

    if err != nil {
        fmt.Fprintf(os.Stderr, "Error: unable to generate prime q\n")
        os.Exit(1)
    }

    // Check p < q < 2p
    if p.Cmp(q) < 0 {
        if q.Cmp(new(big.Int).Mul(p, big.NewInt(2))) > 0 {
            goto retry_key
        }
    }

    // Now check q < p < 2q
    if q.Cmp(p) < 0 {
        if p.Cmp(new(big.Int).Mul(q, big.NewInt(2))) > 0 {
            goto retry_key
        }
    }

    n := new(big.Int).Mul(p, q)

    pm1 := new(big.Int).Add(p, big.NewInt(-1))
    qm1 := new(big.Int).Add(q, big.NewInt(-1))
    // Carmichael totient function
    carm := new(big.Int).Div(new(big.Int).Mul(pm1, qm1), new(big.Int).GCD(nil, nil, pm1, qm1))

    var e, d *big.Int

    // Make an intentionally weak small d key
    if smalld {
        d, err = cryptor.Prime(cryptor.Reader, 128)

        if err != nil {
            fmt.Fprintf(os.Stderr, "Error: unable to generate prime for d\n")
            os.Exit(1)
        }

        e = new(big.Int).ModInverse(d, carm)
    } else {
        // make a more traditional key
        e = big.NewInt(65537)

        d = new(big.Int).ModInverse(e, carm)
    }

    if d == nil || e == nil {
        if (fails > 5) {
            fmt.Fprintf(os.Stderr, "Error: unable to generate d! Probably (p - 1) or (q - 1) was a multiple of e or d\n")
            os.Exit(1)
        } else {
            fails++
            goto retry_key
        }
    }

    //fmt.Fprintf(os.Stdout, "debug: d: %s\n", d.Text(10))
    //fmt.Fprintf(os.Stdout, "debug: carm: %s\n", carm.Text(10))

    key.n = n
    key.e = e

    return key, d

}
