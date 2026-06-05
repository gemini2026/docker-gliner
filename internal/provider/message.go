// Package provider implements the Docker Compose provider protocol: parsing the
// up/down invocation Compose makes and emitting the line-delimited JSON messages
// it reads back from stdout. See docs/PROTOCOL.md.
package provider

import (
	"encoding/json"
	"fmt"
	"io"
)

// Message is one line-delimited JSON object written to stdout for Compose.
type Message struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

// Emitter writes protocol messages to a stream (Compose reads stdout).
type Emitter struct {
	w   io.Writer
	enc *json.Encoder
}

// NewEmitter returns an Emitter that writes newline-delimited JSON to w.
func NewEmitter(w io.Writer) *Emitter {
	return &Emitter{w: w, enc: json.NewEncoder(w)}
}

func (e *Emitter) emit(typ, msg string) {
	// json.Encoder.Encode appends a newline, giving us line-delimited output.
	_ = e.enc.Encode(Message{Type: typ, Message: msg})
}

// Info reports progress shown in the Compose UI.
func (e *Emitter) Info(msg string) { e.emit("info", msg) }

// Infof is Info with printf formatting.
func (e *Emitter) Infof(format string, a ...any) { e.Info(fmt.Sprintf(format, a...)) }

// Error reports a failure reason rendered as the service's failure message.
func (e *Emitter) Error(msg string) { e.emit("error", msg) }

// Errorf is Error with printf formatting.
func (e *Emitter) Errorf(format string, a ...any) { e.Error(fmt.Sprintf(format, a...)) }

// Debug is only shown when Compose runs with --verbose.
func (e *Emitter) Debug(msg string) { e.emit("debug", msg) }

// Setenv injects NAME=value into dependent services. Compose prefixes NAME with
// the provider service name, uppercased (e.g. service "ner" + "ENDPOINT" ->
// NER_ENDPOINT). See docs/PROTOCOL.md §3.
func (e *Emitter) Setenv(name, value string) {
	e.emit("setenv", fmt.Sprintf("%s=%s", name, value))
}
