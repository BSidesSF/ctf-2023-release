package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/ecdsa"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	_ "embed"

	"github.com/google/uuid"

	"github.com/BSidesSF/ctf-2023/sefi/sample"
	"github.com/BSidesSF/ctf-2023/sefi/types"
	apitypes "github.com/BSidesSF/ctf-2023/sefi/types"
)

const (
	ENDPOINT        = "0.0.0.0:4000"
	UIDIR           = "/app/out/ui"
	DEFAULT_FLAG    = "CTF{line_encodings_enable_clock_recover}"
	SAMPLES_PER_BIT = 32
	SPECTRUM_WIDTH  = 2048
	FLAG_FREQ       = 1337
	SIG_HEADER      = "X-Request-Signature"
	NUM_PIECES      = 5
)

var (
	sampleCache      []*sample.Sample
	sampleCacheOnce  sync.Once
	flagVal          = []byte(DEFAULT_FLAG)
	requestKey       = "13375dea789b1337"
	allowMissingSig  = false
	configSigningKey *ecdsa.PrivateKey
	indexHtml        []byte
)

//go:embed privkey.pem
var privkey_pem_bytes []byte

func init() {
	http.HandleFunc("/download/sefi.tar.gz", SEFITarHandler)
	http.HandleFunc("/api/workUnit", WorkUnitHandler)
	http.HandleFunc("/", IndexHandler)

	if v := os.Getenv("FLAG"); v != "" {
		flagVal = []byte(v)
	}
	switch strings.ToLower(os.Getenv("ALLOW_MISSING_SIG")) {
	case "true":
		allowMissingSig = true
	case "false":
		allowMissingSig = false
	default:
	}
	// load the privkey
	block, _ := pem.Decode(privkey_pem_bytes)
	if block == nil {
		panic("Unable to locate block in privkey_pem_bytes!")
	}
	if block.Type != "EC PRIVATE KEY" {
		panic("Expected EC PRIVATE KEY!")
	}
	if parsedPrivKey, err := x509.ParseECPrivateKey(block.Bytes); err != nil {
		panic(err)
	} else {
		configSigningKey = parsedPrivKey
	}

	if fp, err := os.Open("index.html"); err != nil {
		panic(err)
	} else {
		defer fp.Close()
		buf, err := io.ReadAll(fp)
		if err != nil {
			panic(err)
		}
		indexHtml = buf[:]
	}
}

func main() {
	log.Printf("starting listening on %v", ENDPOINT)
	if err := http.ListenAndServe(ENDPOINT, nil); err != nil {
		log.Fatalf("error listening: %v", err)
	}
}

type SigError struct {
	estr string
	hstr string
	code int
}

func (err *SigError) Error() string {
	return err.estr
}

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	w.Write(indexHtml)
}

func WorkUnitHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	var workUnit types.WorkUnitRequest
	if err := getSignedBody(r, &workUnit); err != nil {
		log.Printf("Error in getSignedBody: %v", err)
		if serr, ok := err.(*SigError); ok {
			http.Error(w, serr.hstr, serr.code)
		} else {
			http.Error(w, "", http.StatusInternalServerError)
		}
		return
	}
	if workUnit.ClientID == "" {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}
	sample := getSampleForClient(workUnit.ClientID, workUnit.UnitsFinished)
	resp := types.WorkUnitResponse{
		ClientID: workUnit.ClientID,
		WorkUnit: sample.Prepare(),
	}
	w.Header().Add("Content-type", "application/json")
	enc := json.NewEncoder(w)
	if err := enc.Encode(&resp); err != nil {
		log.Printf("Error encoding: %v", err)
	}
}

func getSignedBody(r *http.Request, data interface{}) error {
	var buf bytes.Buffer
	defer r.Body.Close()
	if _, err := io.Copy(&buf, r.Body); err != nil {
		return err
	}
	h := hmac.New(sha256.New, []byte(requestKey))
	h.Write(buf.Bytes())
	expected := h.Sum(nil)
	got, err := hex.DecodeString(r.Header.Get(SIG_HEADER))
	if err != nil {
		return err
	}
	if !hmac.Equal(got, expected) {
		if !allowMissingSig || len(got) != 0 {
			return &SigError{
				code: http.StatusForbidden,
				hstr: "Not authorized",
				estr: "Invalid signature header",
			}
		}
	}
	dec := json.NewDecoder(&buf)
	if err := dec.Decode(data); err != nil {
		return err
	}
	return nil
}

func SEFITarHandler(w http.ResponseWriter, r *http.Request) {
	ctype := "application/x-gzip"
	w.Header().Add("Content-Type", ctype)
	w.Header().Add("Content-Disposition",
		fmt.Sprintf("attachment; filename=\"%s\"", path.Base(r.URL.Path)))
	gzwr := gzip.NewWriter(w)
	defer gzwr.Close()
	tarw := tar.NewWriter(gzwr)
	defer tarw.Close()
	if err := sendTarDir(tarw, UIDIR); err != nil {
		log.Printf("Error in sendTarDir: %v", err)
		return
	}
	// Now add the config
	cfg := apitypes.ClientConfig{
		APIEndpoint: getAPIURL(r).String(),
		RequestKey:  requestKey,
		ClientID:    makeClientID(),
	}
	buf := bytes.Buffer{}
	jwriter := json.NewEncoder(&buf)
	jwriter.SetIndent("", "  ")
	if err := jwriter.Encode(&cfg); err != nil {
		log.Printf("Error converting to json: %v", err)
		return
	}
	// sign data
	sig, err := signData(buf.Bytes())
	if err != nil {
		log.Printf("Error signing JSON data: %w", err)
		return
	}
	header := tar.Header{
		Name:     "config.json",
		ModTime:  time.Now(),
		Mode:     int64(0644),
		Typeflag: tar.TypeReg,
		Size:     int64(buf.Len()),
	}
	tarw.WriteHeader(&header)
	if _, err := tarw.Write(buf.Bytes()); err != nil {
		log.Printf("Error writing json to tar: %v", err)
		return
	}
	header = tar.Header{
		Name:     "config.json.sig",
		ModTime:  time.Now(),
		Mode:     int64(0644),
		Typeflag: tar.TypeReg,
		Size:     int64(len(sig)),
	}
	tarw.WriteHeader(&header)
	if _, err := tarw.Write(sig); err != nil {
		log.Printf("Error writing sig to tar: %v", err)
		return
	}
}

// Note that this requires trusting x-forwarded-* headers!
func getAPIURL(r *http.Request) *url.URL {
	var apiURL url.URL
	apiURL.Scheme = r.URL.Scheme
	apiURL.Host = r.Host
	apiURL.Path = "/api/"
	forwardedProto := strings.ToLower(r.Header.Get("X-Forwarded-Proto"))
	switch forwardedProto {
	case "http", "https":
		apiURL.Scheme = forwardedProto
	default:
		if apiURL.Scheme == "" {
			apiURL.Scheme = "http"
		}
	}
	if apiURL.Host == "" {
		apiURL.Host = ENDPOINT
	}
	return &apiURL
}

func sendTarDir(w *tar.Writer, dirName string) error {
	if err := filepath.Walk(dirName, func(wpath string, info fs.FileInfo, err error) error {
		relpath, err := filepath.Rel(dirName, wpath)
		if err != nil {
			return err
		}
		if relpath == "." {
			return nil
		}
		log.Printf("tar file: %s\n", relpath)
		now := time.Now()
		header := tar.Header{
			Name:    relpath,
			Mode:    int64(info.Mode().Perm()),
			ModTime: now,
		}
		if info.Mode().IsDir() {
			header.Typeflag = tar.TypeDir
		} else {
			realfi, err := os.Stat(wpath)
			if err != nil {
				return fmt.Errorf("error stat in sendTarDir: %w", err)
			}
			header.Typeflag = tar.TypeReg
			header.Size = realfi.Size()
		}
		if fp, err := os.Open(wpath); err != nil {
			return err
		} else {
			if err := w.WriteHeader(&header); err != nil {
				return err
			}
			defer fp.Close()
			io.Copy(w, fp)
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func getSamples() []*sample.Sample {
	sampleCacheOnce.Do(func() {
		samplesPerChar := SAMPLES_PER_BIT * 10
		flagPadded := flagVal[:]
		if len(flagVal)%4 != 0 {
			log.Printf("flag length is not a multiple of 4, padding will not be even")
			padLen := 4 - len(flagVal)%4
			flagPadded := make([]byte, len(flagVal)+padLen)
			copy(flagPadded, flagVal)
			for i := padLen; i > 0; i-- {
				flagPadded[len(flagVal)+i-1] = ' '
			}
		}
		totalLen := roundUp(samplesPerChar*len(flagPadded), samplesPerChar*NUM_PIECES)
		pieceLen := totalLen / NUM_PIECES
		mainSample := sample.NewSample(SPECTRUM_WIDTH, totalLen)
		mainSample.FillWithNoise()
		mainSample.EncodeBytes(FLAG_FREQ, flagPadded, SAMPLES_PER_BIT)
		sampleCache = mainSample.Split(pieceLen)
		log.Printf("sample cache built, %d samples", len(sampleCache))
	})
	return sampleCache
}

func getSampleForClient(clientId string, unitsFinished int) *sample.Sample {
	samples := getSamples()
	h := sha256.New()
	h.Write([]byte(clientId))
	fmt.Fprintf(h, "::%d", unitsFinished)
	res := h.Sum(nil)
	v := int(res[0]) % len(samples)
	return samples[v]
}

// round val up to a multiple of int
func roundUp(val, mult int) int {
	r := val % mult
	if r == 0 {
		return val
	}
	return val + mult - r
}

// Sign config data
func signData(data []byte) ([]byte, error) {
	hfunc := sha256.New()
	if _, err := hfunc.Write(data); err != nil {
		return nil, fmt.Errorf("failed writing for hash: %w", err)
	}
	hval := hfunc.Sum(nil)
	sig, err := ecdsa.SignASN1(rand.Reader, configSigningKey, hval)
	if err != nil {
		return nil, fmt.Errorf("failed signing data: %w", err)
	}
	pemBlock := pem.Block{
		// Non-standard
		Type:  "ECDSA Signature",
		Bytes: sig,
	}
	return pem.EncodeToMemory(&pemBlock), nil
}

func makeClientID() string {
	if id, err := uuid.NewRandom(); err == nil {
		return id.String()
	}
	h := sha256.New()
	fmt.Fprintf(h, "%v", time.Now().String())
	dgst := h.Sum(nil)
	return hex.EncodeToString(dgst)
}
