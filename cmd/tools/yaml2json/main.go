package main

// yaml2json: convert a YAML document to pretty-printed JSON.
// Usage:
//   yaml2json < input.yaml > output.json
//   yaml2json path/to/input.yaml > output.json

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"

	yaml "gopkg.in/yaml.v3"
)

func run() error {
	flag.Parse()
	var r io.Reader
	if flag.NArg() > 0 {
		f, err := os.Open(flag.Arg(0))
		if err != nil {
			return err
		}
		defer func() {
			if cerr := f.Close(); cerr != nil {
				// Log close error to stderr; primary operation already completed.
				fmt.Fprintf(os.Stderr, "yaml2json: close error: %v\n", cerr)
			}
		}()
		r = f
	} else {
		r = os.Stdin
	}

	dec := yaml.NewDecoder(r)
	var v any
	if err := dec.Decode(&v); err != nil {
		return fmt.Errorf("decode yaml: %w", err)
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		return fmt.Errorf("encode json: %w", err)
	}
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "yaml2json error: %v\n", err)
		os.Exit(1)
	}
}
