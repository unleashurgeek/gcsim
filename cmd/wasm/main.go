package main

import (
	"encoding/json"
	"errors"
	"log"
	"strconv"
	"syscall/js"
	"time"

	"github.com/genshinsim/gcsim/pkg/agg"
	"github.com/genshinsim/gcsim/pkg/gcs/ast"
	"github.com/genshinsim/gcsim/pkg/simulation"
	"github.com/genshinsim/gcsim/pkg/simulator"
	"github.com/genshinsim/gcsim/pkg/stats"
)

const DefaultBufferLength = 1024 * 10

var (
	sha1ver   string // sha1 revision used to build the program
	buildTime string // when the executable was built
)

// shared variables
var cfg string
var simcfg *ast.ActionList
var buffer []byte

// Aggregator variables
var aggregators []agg.Aggregator

func main() {
	//GOOS=js GOARCH=wasm go build -o main.wasm
	ch := make(chan struct{}, 0)

	// Helper Functions (stateless, no init call needed)
	js.Global().Set("debug", js.FuncOf(debug))
	js.Global().Set("validateConfig", js.FuncOf(validateConfig))

	// Worker Functions
	js.Global().Set("initializeWorker", js.FuncOf(initializeWorker))
	js.Global().Set("simulate", js.FuncOf(simulate))

	// Aggregator Functions
	js.Global().Set("initializeAggregator", js.FuncOf(initializeAggregator))
	js.Global().Set("aggregate", js.FuncOf(aggregate))
	js.Global().Set("flush", js.FuncOf(flush))

	<-ch
}

// static helper functions (stateless)

// debug(cfg: string, seed: string) -> string
func debug(this js.Value, args []js.Value) interface{} {
	cfg := args[0].String()
	seed, _ := strconv.ParseUint(args[1].String(), 10, 64)

	parser := ast.New(cfg)
	simcfg, err := parser.Parse()
	if err != nil {
		return marshal(err)
	}

	data, err := simulator.GenerateDebugLogWithSeed(simcfg, int64(seed))
	if err != nil {
		return marshal(err)
	}

	out, err := json.Marshal(data)
	if err != nil {
		return marshal(err)
	}

	return string(out)
}

// validateConfig(cfg: string) -> string
func validateConfig(this js.Value, args []js.Value) interface{} {
	in := args[0].String()

	parser := ast.New(in)
	cfg, err := parser.Parse()
	if err != nil {
		return marshal(err)
	}

	for i, v := range cfg.Characters {
		log.Printf("%v: %v\n", i, v.Base.Key.String())
	}

	data, err := json.Marshal(cfg)
	if err != nil {
		return marshal(err)
	}
	return string(data)
}

// worker functions

// initializeWorker(cfg: string)
func initializeWorker(this js.Value, args []js.Value) interface{} {
	in := args[0].String()
	if err := initialize(in); err != nil {
		return marshal(err)
	}
	return nil
}

// simulate() -> js Uint8Array
func simulate(this js.Value, args []js.Value) interface{} {
	cpycfg := simcfg.Copy()
	core, err := simulation.NewCore(simulator.CryptoRandSeed(), false, cpycfg)
	if err != nil {
		return marshal(err)
	}

	sim, err := simulation.New(cpycfg, core)
	if err != nil {
		return marshal(err)
	}

	result, err := sim.Run()
	if err != nil {
		return marshal(err)
	}

	buffer, err = result.MarshalMsg(buffer[:0])
	if err != nil {
		return marshal(err)
	}

	dst := js.Global().Get("Uint8Array").New(len(buffer))
	copyLen := js.CopyBytesToJS(dst, buffer)
	if copyLen != len(buffer) {
		marshal(errors.New("BytesToJS: copied array was the incorrect size!"))
	}
	return dst
}

// aggregator functions

// initializeAggregator(cfg: string) -> string
func initializeAggregator(this js.Value, args []js.Value) interface{} {
	in := args[0].String()
	if err := initialize(in); err != nil {
		return marshal(err)
	}

	aggregators = aggregators[:0]
	for _, aggregator := range agg.Aggregators() {
		a, err := aggregator(simcfg)
		if err != nil {
			return marshal(err)
		}
		aggregators = append(aggregators, a)
	}

	opts := simulator.Options{
		Version:          sha1ver,
		BuildDate:        buildTime,
		Debug:            false,
		GZIPResult:       false,
		ResultSaveToPath: "",
		ConfigPath:       "",
	}
	result, err := simulator.GenerateResult(cfg, simcfg, opts)
	if err != nil {
		return marshal(err)
	}

	out, err := json.Marshal(result)
	if err != nil {
		return marshal(err)
	}
	return string(out)
}

// aggregate(src: Uint8Array)
func aggregate(this js.Value, args []js.Value) interface{} {
	src := args[0]
	var err error

	// golang wasm copy requires src and destination length to have enough capacity to copy
	// should be enforced by capacity and not length but whatev....
	// have to waste cycles here to make sure buffer has the right length
	length := src.Get("length").Int()
	if length > cap(buffer) {
		buffer = make([]byte, length)
	} else {
		buffer = buffer[:length]
	}

	copyLen := js.CopyBytesToGo(buffer, src)
	if copyLen != len(buffer) {
		marshal(errors.New("BytesToGo: copied array was the incorrect size!"))
	}

	result := stats.Result{}
	buffer, err = result.UnmarshalMsg(buffer)
	if err != nil {
		return marshal(err)
	}

	for _, a := range aggregators {
		a.Add(result)
	}
	return nil
}

// flush(startTime: int) -> string
func flush(this js.Value, args []js.Value) interface{} {
	startTime := args[0].Int()

	stats := agg.Result{}
	for _, a := range aggregators {
		a.Flush(&stats)
	}
	stats.Runtime = float64(time.Now().Nanosecond() - startTime)

	out, err := json.Marshal(stats)
	if err != nil {
		return marshal(err)
	}
	return string(out)
}

// internal helper functions

func initialize(raw string) error {
	parser := ast.New(raw)
	out, err := parser.Parse()
	if err != nil {
		return err
	}

	if cap(buffer) < DefaultBufferLength {
		buffer = make([]byte, 0, DefaultBufferLength)
	}

	cfg = raw
	simcfg = out
	return nil
}

func marshal(err error) string {
	d := struct {
		Err string `json:"error"`
	}{
		Err: err.Error(),
	}
	b, _ := json.Marshal(d)
	return string(b)
}
