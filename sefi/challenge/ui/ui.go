package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	_ "embed"

	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"

	_ "github.com/BSidesSF/ctf-2023/sefi/encoders"
	"github.com/BSidesSF/ctf-2023/sefi/types"
)

const (
	frameInterval = 200 * time.Millisecond
)

var (
	configPublicKey *ecdsa.PublicKey
)

//go:embed public.pem
var public_key_bytes []byte

func main() {
	// unpack the key
	block, _ := pem.Decode(public_key_bytes)
	if block == nil {
		panic("Unable to locate block in public_key_bytes!")
	}
	if block.Type != "PUBLIC KEY" {
		panic("Wrong key type in public_key_bytes!")
	}
	if parsedPubKey, err := x509.ParsePKIXPublicKey(block.Bytes); err != nil {
		panic(err)
	} else {
		if parsedEcKey, ok := parsedPubKey.(*ecdsa.PublicKey); ok {
			configPublicKey = parsedEcKey
		} else {
			panic("Expected ecdsa.PublicKey")
		}
	}

	// start the ui
	ui := NewUI()
	ui.Run()
}

type SEFIUI struct {
	mainWin        *gtk.ApplicationWindow
	app            *gtk.Application
	workUnit       *types.Sample
	config         *types.ClientConfig
	exeDir         string
	frameId        int
	processed      bool
	lock           sync.Mutex
	counts         []int
	averagesum     []int
	averages       []int
	rawDrawingArea *gtk.DrawingArea
	avgDrawingArea *gtk.DrawingArea
	cntDrawingArea *gtk.DrawingArea
	rawGraph       GraphRenderer
	avgGraph       GraphRenderer
	cntGraph       GraphRenderer
	frameDenom     int
	statusBar      *gtk.Statusbar
	statusBarctx   uint
	unitsDone      int
}

func NewUI() *SEFIUI {
	exeDir, err := getExeDir()
	if err != nil {
		panic(err)
	}
	app, err := gtk.ApplicationNew("net.bsidessf.ctf.sefi", glib.APPLICATION_FLAGS_NONE)
	if err != nil {
		panic(err)
	}
	rv := &SEFIUI{
		app:    app,
		exeDir: exeDir,
		counts: make([]int, 256),
	}
	if err := rv.LoadConfig(); err != nil {
		panic(err)
	}
	return rv
}

func (ui *SEFIUI) LoadConfig() error {
	ui.lock.Lock()
	defer ui.lock.Unlock()
	sigPath := filepath.Join(ui.exeDir, "config.json.sig")
	var sigData []byte
	if fp, err := os.Open(sigPath); err != nil {
		return fmt.Errorf("error opening config.json.sig: %w", err)
	} else {
		defer fp.Close()
		readData, err := io.ReadAll(fp)
		if err != nil {
			return fmt.Errorf("error reading config.json.sig: %w", err)
		}
		sigData = readData
	}
	configPath := filepath.Join(ui.exeDir, "config.json")
	if fp, err := os.Open(configPath); err != nil {
		return fmt.Errorf("error opening config.json: %w", err)
	} else {
		defer fp.Close()
		data, err := io.ReadAll(fp)
		if err != nil {
			return fmt.Errorf("error reading config.json: %w", err)
		}
		if err := verifySignature(data, sigData); err != nil {
			return fmt.Errorf("error verifying config signature: %w", err)
		}
		buf := bytes.NewReader(data)
		dec := json.NewDecoder(buf)
		var config types.ClientConfig
		if err := dec.Decode(&config); err != nil {
			return err
		}
		ui.config = &config
	}
	return nil
}

func (ui *SEFIUI) Run() error {
	ui.app.Connect("activate", ui.activate)
	if code := ui.app.Run(os.Args); code > 0 {
		return fmt.Errorf("gtk app exited with code %d", code)
	}
	return nil
}

func (ui *SEFIUI) activate(app *gtk.Application) {
	log.Printf("starting ui...")
	mainWin, err := gtk.ApplicationWindowNew(app)
	if err != nil {
		log.Printf("error creating window: %v", err)
		return
	}
	ui.mainWin = mainWin
	ui.mainWin.SetTitle("SEFI - Search for Extra Flag Intelligence")
	ui.mainWin.SetDefaultSize(800, 600)
	ui.renderUI()
	ui.mainWin.Show()

	// ui should be valid below here
	ui.FetchWorkUnit()
	ui.animateOne()
	go ui.animate()
}

// should be called on own goroutine
func (ui *SEFIUI) animate() {
	for range time.Tick(frameInterval) {
		glib.IdleAdd(func() {
			ui.animateOne()
		})
	}
}

func (ui *SEFIUI) animateOne() {
	frameData := ui.nextFrame()
	ui.processFrame(frameData)
	frameInts := make([]int, len(frameData))
	for i, v := range frameData {
		frameInts[i] = int(v)
	}
	ui.rawGraph.SetData(frameInts)
	ui.rawGraph.SetYRange(0, 256)
	ui.cntGraph.SetData(logTransform(ui.counts))
	ui.avgGraph.SetData(ui.averages)
	ui.SetStatusMessage(fmt.Sprintf("Frame %d/%d", ui.frameId, ui.workUnit.TimeSlots))
	ui.Refresh()
}

func (ui *SEFIUI) FetchWorkUnit() error {
	if ui.config == nil {
		return fmt.Errorf("ui.config is nil!")
	}
	wureq := types.WorkUnitRequest{
		ClientID:      ui.getClientId(),
		UnitsFinished: ui.unitsDone,
	}
	ui.unitsDone += 1
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.Encode(&wureq)
	h := hmac.New(sha256.New, []byte(ui.config.RequestKey))
	h.Write(buf.Bytes())
	sig := hex.EncodeToString(h.Sum(nil))
	req, err := http.NewRequest(http.MethodPost, ui.getAPIURL("workUnit"), &buf)
	if err != nil {
		return err
	}
	req.Header.Set("X-Request-Signature", sig)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Search for Extra Flag Intelligence (SEFI)/1.0")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		log.Printf("Invalid http response code: %v %v", resp.StatusCode, resp.Status)
		return fmt.Errorf("bad http response code: %v %v", resp.StatusCode, resp.Status)
	}
	defer resp.Body.Close()
	dec := json.NewDecoder(resp.Body)
	var wuresp types.WorkUnitResponse
	if err := dec.Decode(&wuresp); err != nil {
		return err
	}
	ui.withLock(func() error {
		ui.workUnit = wuresp.WorkUnit
		return nil
	})
	log.Printf(
		"loaded workUnit, freqs: %d, timeSlots: %d, startTime: %d",
		wuresp.WorkUnit.Freqs, wuresp.WorkUnit.TimeSlots, wuresp.WorkUnit.StartTime)
	ui.Refresh()
	return nil
}

func (ui *SEFIUI) Refresh() {
	ui.avgDrawingArea.QueueDraw()
	ui.cntDrawingArea.QueueDraw()
	ui.rawDrawingArea.QueueDraw()
	// trigger actual redraw
	glib.IdleAdd(func() {})
}

func (ui *SEFIUI) getAPIURL(path string) string {
	apiPath := ui.config.APIEndpoint
	if !strings.HasSuffix(apiPath, "/") {
		apiPath += "/"
	}
	return apiPath + path
}

func (ui *SEFIUI) getClientId() string {
	return ui.config.ClientID
}

func (ui *SEFIUI) processFrame(frameData []byte) {
	ui.lock.Lock()
	defer ui.lock.Unlock()
	if ui.processed {
		return
	}
	if len(ui.averages) < len(frameData) {
		newAveragesum := make([]int, len(frameData))
		copy(newAveragesum, ui.averagesum)
		ui.averagesum = newAveragesum
		ui.averages = make([]int, len(frameData))
	}
	ui.frameDenom++
	for i, v := range frameData {
		ui.counts[int(v)]++
		ui.averagesum[i] += int(v)
		ui.averages[i] = ui.averagesum[i] / ui.frameDenom
	}
}

func (ui *SEFIUI) nextFrame() []byte {
	ui.lock.Lock()
	defer ui.lock.Unlock()
	frameLen := ui.workUnit.Freqs
	start := ui.frameId * frameLen
	if start >= len(ui.workUnit.Samples) {
		ui.frameId = 0
		start = 0
		ui.processed = true
	}
	end := start + frameLen
	if end >= len(ui.workUnit.Samples) {
		end = len(ui.workUnit.Samples)
	}
	ui.frameId++
	return ui.workUnit.Samples[start:end]
}

func (ui *SEFIUI) withLock(f func() error) error {
	ui.lock.Lock()
	defer ui.lock.Unlock()
	return f()
}

func (ui *SEFIUI) renderUI() {
	da, _ := gtk.DrawingAreaNew()
	ui.rawDrawingArea = da
	rawGraph := NewAveragingBarGraphRenderer(nil)
	rawGraph.xticksmajor = 128
	rawGraph.xticksminor = 32
	rawGraph.fgcolor = White
	rawGraph.bgcolor = Black
	rawGraph.barColorFunc = MakeGradientFunc(Blue, Red, 256)
	ui.rawGraph = rawGraph
	da.Connect("draw", rawGraph.DrawFunc)
	da.SetHExpand(true)
	da.SetVExpand(true)
	frame, _ := gtk.FrameNew("Raw")
	frame.Add(da)
	vbox, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 0)
	vbox.Add(frame)
	vbox.SetHExpand(true)
	vbox.SetVExpand(true)
	hbox, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 0)
	hbox.SetHomogeneous(true)
	cntGraph := NewAveragingBarGraphRenderer(nil)
	cntGraph.xticksmajor = 32
	cntGraph.xticksminor = 4
	cntGraph.fgcolor = White
	ui.cntGraph = cntGraph
	cntDA, _ := gtk.DrawingAreaNew()
	ui.cntDrawingArea = cntDA
	cntDA.Connect("draw", cntGraph.DrawFunc)
	cntDA.SetHExpand(true)
	cntDA.SetVExpand(true)
	frame, _ = gtk.FrameNew("Distribution")
	frame.Add(cntDA)
	hbox.Add(frame)
	avgGraph := NewAveragingBarGraphRenderer(nil)
	avgGraph.xticksmajor = 128
	avgGraph.xticksminor = 32
	avgGraph.fgcolor = White
	ui.avgGraph = avgGraph
	avgDA, _ := gtk.DrawingAreaNew()
	ui.avgDrawingArea = avgDA
	avgDA.Connect("draw", avgGraph.DrawFunc)
	avgDA.SetHExpand(true)
	avgDA.SetVExpand(true)
	frame, _ = gtk.FrameNew("Averages")
	frame.Add(avgDA)
	hbox.Add(frame)
	vbox.Add(hbox)
	sbar, _ := gtk.StatusbarNew()
	ui.statusBar = sbar
	ui.statusBarctx = sbar.GetContextId("ui")
	sbar.Push(ui.statusBarctx, "Loading")
	vbox.Add(sbar)
	ui.mainWin.Add(vbox)
	vbox.ShowAll()
}

func (ui *SEFIUI) SetStatusMessage(msg string) {
	ui.statusBar.Pop(ui.statusBarctx)
	ui.statusBar.Push(ui.statusBarctx, msg)
}

func logTransform(data []int) []int {
	rv := make([]int, len(data))
	for i, v := range data {
		if v != 0 {
			rv[i] = int(math.Log2(float64(v)) * 8.0)
		} else {
			rv[i] = 0
		}
	}
	return rv
}

func verifySignature(dataBytes, sigPEMBytes []byte) error {
	hfunc := sha256.New()
	if _, err := hfunc.Write(dataBytes); err != nil {
		return fmt.Errorf("failed writing for hash: %w", err)
	}
	hval := hfunc.Sum(nil)
	block, _ := pem.Decode(sigPEMBytes)
	if block == nil {
		return fmt.Errorf("invalid signature: no PEM")
	}
	if block.Type != "ECDSA Signature" {
		return fmt.Errorf("unexpected signature type")
	}
	if ecdsa.VerifyASN1(configPublicKey, hval, block.Bytes) {
		return nil
	}
	return fmt.Errorf("invalid signature!")
}
