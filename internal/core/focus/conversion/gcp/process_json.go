package gcp

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"path/filepath"

	"github.com/costscope/costscope/internal/core/focus/types"
)

// ProcessJSON streams GCP JSON data. It supports both array-of-objects and NDJSON forms.
// Mapping of classifier metrics should be handled by the provided mapObjectToFocus function
// to preserve legacy behavior.
// Returns (recordCount, processedRecords, errorRecords, error).
func ProcessJSON(
	ctx context.Context,
	reader io.Reader,
	config *types.ConversionConfig,
	dw types.DataWriter,
	mapObjectToFocus func(obj map[string]interface{}) *types.FocusRecord,
) (int64, int64, int64, error) {
	br := bufio.NewReader(reader)
	// Detect top-level (array vs NDJSON); behavior mirrors detectGCPJSONTopLevel
	isArray, err := func() (bool, error) {
		for {
			b, e := br.Peek(1)
			if e != nil {
				if e == io.EOF {
					return false, nil
				}
				return false, e
			}
			if len(b) == 0 {
				return false, nil
			}
			if b[0] == ' ' || b[0] == '\n' || b[0] == '\r' || b[0] == '\t' {
				_, _ = br.ReadByte()
				continue
			}
			return b[0] == '[', nil
		}
	}()
	if err != nil {
		return 0, 0, 0, err
	}

	base := filepath.Base(config.InputPath)
	writeObj := func(obj map[string]interface{}) error {
		fr := mapObjectToFocus(obj)
		fr.SourceFileName = base
		return dw.Write(ctx, []types.FocusRecord{*fr})
	}

	if isArray {
		dec := json.NewDecoder(br)
		if t, _ := dec.Token(); t == nil { // must be '['
			return 0, 0, 0, io.ErrUnexpectedEOF
		}
		var rc, pc, ec int64
		for dec.More() {
			var obj map[string]interface{}
			if derr := dec.Decode(&obj); derr != nil {
				ec++
				continue
			}
			if err := writeObj(obj); err != nil {
				return 0, 0, 0, err
			}
			rc++
			pc++
		}
		_, _ = dec.Token() // consume ']'
		return rc, pc, ec, nil
	}

	// NDJSON path
	var rc, pc, ec int64
	scanner := bufio.NewScanner(br)
	buf := make([]byte, 0, 1024*1024)
	scanner.Buffer(buf, 10*1024*1024)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var obj map[string]interface{}
		if jerr := json.Unmarshal(line, &obj); jerr != nil {
			ec++
			continue
		}
		if err := writeObj(obj); err != nil {
			return 0, 0, 0, err
		}
		rc++
		pc++
	}
	if err := scanner.Err(); err != nil {
		return 0, 0, 0, err
	}
	return rc, pc, ec, nil
}
