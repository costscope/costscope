package azure

import (
	"context"
	"io"

	"github.com/costscope/costscope/internal/core/focus/types"
	"github.com/costscope/costscope/internal/core/monitoring/telemetry"
)

// ProcessJSON streams Azure JSON (array or NDJSON via JSONStream) and applies the provided
// mapping and post-processing delegates, preserving legacy behavior and metrics.
// It returns (recordCount, processedRecords, errorRecords, error).
func ProcessJSON(
	ctx context.Context,
	reader JSONStream, // provider-scoped JSON stream abstraction
	config *types.ConversionConfig,
	dw types.DataWriter,
	pathLabel string,
	mapObjectToFocus func(obj map[string]any) types.FocusRecord,
	buildRecMap func(obj map[string]any) map[string]string,
	postProcess func(rec map[string]string, fr *types.FocusRecord, config *types.ConversionConfig),
	enforceJSONParity func(fr *types.FocusRecord, obj map[string]any),
	onDecodeError func(err error),
) (int64, int64, int64, error) {
	var recordCount, processedRecords, errorRecords int64

	for {
		obj, err := reader.NextObject(ctx)
		if err != nil {
			if err == io.EOF {
				break
			}
			// Count JSON decode error and continue (legacy behavior)
			if onDecodeError != nil {
				onDecodeError(err)
			}
			errorRecords++
			continue
		}
		recordCount++

		fr := mapObjectToFocus(obj)
		if enforceJSONParity != nil {
			enforceJSONParity(&fr, obj)
		}
		if buildRecMap != nil || postProcess != nil {
			recMap := buildRecMap(obj)
			if postProcess != nil {
				postProcess(recMap, &fr, config)
			}
		}

		// Classifier decision metric parity
		telemetry.ClassifierDecisions.WithLabelValues("azure", pathLabel, fr.ChargeCategory).Inc()

		if werr := dw.Write(ctx, []types.FocusRecord{fr}); werr != nil {
			return 0, 0, 0, werr
		}
		processedRecords++
	}

	return recordCount, processedRecords, errorRecords, nil
}
