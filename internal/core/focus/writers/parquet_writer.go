package writers

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/xitongsys/parquet-go-source/writerfile"
	"github.com/xitongsys/parquet-go/parquet"
	pqwriter "github.com/xitongsys/parquet-go/writer"

	"local/costscope/internal/core/focus/types"
	"local/costscope/internal/core/monitoring/telemetry"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// syncFile is a small indirection over (*os.File).Sync used to enable test-time
// observation of fsync behavior (counts) without altering production code paths.
// In production, it points to f.Sync.
var syncFile = func(f *os.File) error { return f.Sync() }

// ParquetWriter implements types.DataWriter for writing FOCUS records to Parquet
type ParquetWriter struct {
	file        *os.File
	pw          *pqwriter.ParquetWriter
	metadata    *types.DataDestinationMetadata
	Options     *types.ParquetOptions
	basePath    string
	tmpPath     string
	seq         int
	rotateSize  int64
	rotateEvery time.Duration
	nextRotate  time.Time
	mu          sync.Mutex
}

// Open opens the parquet file for writing with the given schema
func (w *ParquetWriter) Open(_ context.Context, path string, _ *types.FocusSchema) error {
	w.basePath = path
	if err := w.configureRotation(); err != nil {
		return err
	}
	// Start first segment
	if err := w.startNewFile(); err != nil {
		return err
	}

	// Create parquet writer using the FocusRecord struct schema tags
	fw := writerfile.NewWriterFile(w.file)
	// Use an internal parquet schema struct compatible with parquet-go
	// Use single parallel writer to avoid race issues observed in tests with parquet-go
	pw, err := pqwriter.NewParquetWriter(fw, new(focusRecordParquet), 1)
	if err != nil {
		_ = w.file.Close()
		return fmt.Errorf("failed to create parquet writer: %w", err)
	}
	// Set defaults and allow overrides via Options
	pw.RowGroupSize = 128 * 1024 * 1024 // 128MB default
	pw.PageSize = 8 * 1024              // 8KB default
	if w.Options != nil {
		if w.Options.RowGroupSizeBytes > 0 {
			pw.RowGroupSize = w.Options.RowGroupSizeBytes
		}
		if w.Options.PageSizeBytes > 0 {
			pw.PageSize = w.Options.PageSizeBytes
		}
		switch strings.ToLower(w.Options.CompressionCodec) {
		case "uncompressed", "none":
			pw.CompressionType = parquet.CompressionCodec_UNCOMPRESSED
		case "gzip":
			pw.CompressionType = parquet.CompressionCodec_GZIP
		case "zstd":
			pw.CompressionType = parquet.CompressionCodec_ZSTD
		case "snappy", "":
			pw.CompressionType = parquet.CompressionCodec_SNAPPY
		default:
			pw.CompressionType = parquet.CompressionCodec_SNAPPY
		}
	}
	w.pw = pw

	w.metadata = &types.DataDestinationMetadata{
		FilePath: path,
		Created:  time.Now(),
		Format:   "parquet",
		Schema:   "FOCUS_1.2",
	}
	return nil
}

// configureRotation sets rotation thresholds from Options with sensible defaults
func (w *ParquetWriter) configureRotation() error {
	// Defaults
	const defaultRotateSize = int64(512 * 1024 * 1024) // 512MB
	w.rotateSize = 0
	w.rotateEvery = 0
	if w.Options != nil {
		// size
		if w.Options.RotateSizeBytes > 0 {
			w.rotateSize = w.Options.RotateSizeBytes
		} else if w.Options.RotateSizeBytes == 0 {
			// enable by default to 512MB when not explicitly disabled
			w.rotateSize = defaultRotateSize
		} else if w.Options.RotateSizeBytes < 0 {
			// negative value explicitly disables size-based rotation
			w.rotateSize = 0
		}
		// interval
		if w.Options.RotateInterval != "" {
			d, err := time.ParseDuration(w.Options.RotateInterval)
			if err != nil {
				return fmt.Errorf("invalid rotate interval: %w", err)
			}
			if d > 0 {
				w.rotateEvery = d
				w.nextRotate = time.Now().Add(d)
			}
		}
	} else {
		// If no options supplied, enable default size rotation
		w.rotateSize = defaultRotateSize
	}
	return nil
}

// startNewFile opens a new temporary file for the next segment
func (w *ParquetWriter) startNewFile() error {
	w.seq++
	dir := filepath.Dir(w.basePath)
	if err := os.MkdirAll(dir, 0750); err != nil { // #nosec G301 - controlled path
		return err
	}
	// If rotation is disabled (both size and time), write directly to base path
	if w.rotateSize == 0 && w.rotateEvery == 0 {
		// #nosec G304 - path validated by caller
		f, err := os.Create(w.basePath)
		if err != nil {
			return fmt.Errorf("failed to create parquet file: %w", err)
		}
		w.file = f
		w.tmpPath = ""
		return nil
	}
	// temp path ensures atomic finalization later
	tmpName := fmt.Sprintf(".%s.%d.tmp", filepath.Base(w.basePath), w.seq)
	tmpPath := filepath.Join(dir, tmpName)
	// #nosec G304 - path validated by caller
	f, err := os.Create(tmpPath)
	if err != nil {
		return fmt.Errorf("failed to create temp parquet file: %w", err)
	}
	w.file = f
	w.tmpPath = tmpPath
	return nil
}

// finalizeAndRotate closes current parquet writer and performs atomic rename
func (w *ParquetWriter) finalizeAndRotate() error {
	ctx := context.Background()
	_, span := otel.Tracer("costscope.converter").Start(ctx, "rotation", trace.WithSpanKind(trace.SpanKindInternal))
	defer span.End()
	// Stop writer if exists
	if w.pw != nil {
		if err := w.pw.WriteStop(); err != nil {
			span.RecordError(err)
			return fmt.Errorf("failed to finalize parquet file: %w", err)
		}
		w.pw = nil
	}
	// close file and atomically rename to final name
	if w.file != nil {
		// compute rotated final filename
		finalName := w.rotatedFileName()
		// Ensure file is synced before rename
		if err := syncFile(w.file); err != nil {
			_ = w.file.Close()
			return err
		}
		if err := w.file.Close(); err != nil {
			return err
		}
		// #nosec G305 - rename within controlled directory
		if err := os.Rename(w.tmpPath, finalName); err != nil {
			span.RecordError(err)
			return fmt.Errorf("atomic rename failed: %w", err)
		}
		// update metadata for last closed file
		w.metadata.FilePath = finalName
		if stat, err := os.Stat(finalName); err == nil {
			w.metadata.FileSize = stat.Size()
			// record rotation size metric
			telemetry.ParquetRotationSize.Observe(float64(stat.Size()))
		}
		w.file = nil
		w.tmpPath = ""
	}
	// open a new segment and new parquet writer bound to it
	if err := w.startNewFile(); err != nil {
		span.RecordError(err)
		return err
	}
	fw := writerfile.NewWriterFile(w.file)
	// Use single parallel writer to avoid race issues observed in tests with parquet-go
	pw, err := pqwriter.NewParquetWriter(fw, new(focusRecordParquet), 1)
	if err != nil {
		_ = w.file.Close()
		return fmt.Errorf("failed to create parquet writer: %w", err)
	}
	// carry forward options
	pw.RowGroupSize = 128 * 1024 * 1024
	pw.PageSize = 8 * 1024
	if w.Options != nil {
		if w.Options.RowGroupSizeBytes > 0 {
			pw.RowGroupSize = w.Options.RowGroupSizeBytes
		}
		if w.Options.PageSizeBytes > 0 {
			pw.PageSize = w.Options.PageSizeBytes
		}
		switch strings.ToLower(w.Options.CompressionCodec) {
		case "uncompressed", "none":
			pw.CompressionType = parquet.CompressionCodec_UNCOMPRESSED
		case "gzip":
			pw.CompressionType = parquet.CompressionCodec_GZIP
		case "zstd":
			pw.CompressionType = parquet.CompressionCodec_ZSTD
		case "snappy", "":
			pw.CompressionType = parquet.CompressionCodec_SNAPPY
		default:
			pw.CompressionType = parquet.CompressionCodec_SNAPPY
		}
	} else {
		pw.CompressionType = parquet.CompressionCodec_SNAPPY
	}
	w.pw = pw
	// reset time rotation
	if w.rotateEvery > 0 {
		w.nextRotate = time.Now().Add(w.rotateEvery)
	}
	span.SetAttributes(attribute.String("file.path", w.metadata.FilePath))
	return nil
}

// rotatedFileName builds final rotated filename with pattern <prefix>-YYYYMMDD-HHMM-<seq>.parquet
func (w *ParquetWriter) rotatedFileName() string {
	dir := filepath.Dir(w.basePath)
	base := filepath.Base(w.basePath)
	prefix := strings.TrimSuffix(base, filepath.Ext(base))
	if w.Options != nil && w.Options.FilePrefix != "" {
		prefix = w.Options.FilePrefix
	}
	ts := time.Now().Format("20060102-1504")
	name := fmt.Sprintf("%s-%s-%03d.parquet", prefix, ts, w.seq)
	return filepath.Join(dir, name)
}

// Write writes a batch of FOCUS records
func (w *ParquetWriter) Write(ctx context.Context, records []types.FocusRecord) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	for i := range records {
		pr := toParquetRecord(&records[i])
		if err := w.pw.Write(pr); err != nil {
			return fmt.Errorf("failed to write parquet row: %w", err)
		}
		w.metadata.RecordCount++
		// Rotation checks
		if w.shouldRotate() {
			if err := w.finalizeAndRotate(); err != nil {
				return err
			}
		}
	}
	if span := trace.SpanFromContext(ctx); span != nil {
		span.SetAttributes(attribute.Int("parquet.records_written", len(records)))
	}
	return nil
}

// WriteChunk is not supported for ParquetWriter; use Write instead
func (w *ParquetWriter) WriteChunk(_ context.Context, _ []byte) error {
	return fmt.Errorf("WriteChunk not supported for ParquetWriter; use Write(records)")
}

// Flush flushes any buffered data
func (w *ParquetWriter) Flush(_ context.Context) error {
	// parquet-go writes directly; nothing to flush other than ensuring file sync
	if w.file != nil {
		return syncFile(w.file)
	}
	return nil
}

// Close closes the writer and file
func (w *ParquetWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	// If rotation is enabled, finalize current temp file with atomic rename
	if w.rotateSize > 0 || w.rotateEvery > 0 {
		// gracefully stop writer, then rename
		if w.pw != nil {
			if err := w.pw.WriteStop(); err != nil {
				if w.file != nil {
					_ = w.file.Close()
				}
				return fmt.Errorf("failed to finalize parquet file: %w", err)
			}
			w.pw = nil
		}
		if w.file != nil {
			if err := syncFile(w.file); err != nil {
				_ = w.file.Close()
				return err
			}
			if err := w.file.Close(); err != nil {
				return err
			}
			finalName := w.rotatedFileName()
			if err := os.Rename(w.tmpPath, finalName); err != nil {
				return fmt.Errorf("atomic rename failed: %w", err)
			}
			if stat, err := os.Stat(finalName); err == nil {
				w.metadata.FileSize = stat.Size()
				telemetry.ParquetRotationSize.Observe(float64(stat.Size()))
			}
			w.metadata.FilePath = finalName
			w.file = nil
			w.tmpPath = ""
		}
		return nil
	}
	// No rotation: behave like previous implementation, the file is final already
	if w.pw != nil {
		if err := w.pw.WriteStop(); err != nil {
			if w.file != nil {
				_ = w.file.Close()
			}
			return fmt.Errorf("failed to finalize parquet file: %w", err)
		}
		w.pw = nil
	}
	if w.file != nil {
		if stat, err := w.file.Stat(); err == nil {
			w.metadata.FileSize = stat.Size()
		}
		if err := w.file.Close(); err != nil {
			return err
		}
		w.file = nil
	}
	return nil
}

// GetMetadata returns destination metadata
func (w *ParquetWriter) GetMetadata() *types.DataDestinationMetadata {
	return w.metadata
}

// shouldRotate checks file size and time thresholds
func (w *ParquetWriter) shouldRotate() bool {
	if w.file == nil {
		return false
	}
	// time-based rotation
	if w.rotateEvery > 0 && time.Now().After(w.nextRotate) {
		return true
	}
	// size-based rotation: check underlying file size
	if w.rotateSize > 0 {
		if sz, err := fileSize(w.file); err == nil {
			if sz >= w.rotateSize {
				return true
			}
		}
	}
	return false
}

func fileSize(f *os.File) (int64, error) {
	// Fast path: rely on the OS-reported file size. parquet-go writes directly
	// to the underlying file descriptor; forcing an fsync() on every row is
	// unnecessarily expensive and can degrade performance significantly when
	// rotation is enabled. Durability is ensured during rotation finalization
	// and on Close(), where we already Sync() before atomic rename.
	// Rationale: parquet-go writes to the underlying FD; syncing on rotation/close preserves durability with better performance.
	st, err := f.Stat()
	if err != nil {
		return 0, err
	}
	return st.Size(), nil
}

// focusRecordParquet mirrors types.FocusRecord but uses a Parquet-compatible representation for Tags.
type focusRecordParquet struct {
	BillingAccountId   string  `parquet:"name=billing_account_id, type=BYTE_ARRAY, convertedtype=UTF8"`
	BillingAccountName string  `parquet:"name=billing_account_name, type=BYTE_ARRAY, convertedtype=UTF8"`
	BillingCurrency    string  `parquet:"name=billing_currency, type=BYTE_ARRAY, convertedtype=UTF8"`
	BillingPeriodEnd   int64   `parquet:"name=billing_period_end, type=INT64, convertedtype=TIMESTAMP_MILLIS"`
	BillingPeriodStart int64   `parquet:"name=billing_period_start, type=INT64, convertedtype=TIMESTAMP_MILLIS"`
	ChargeCategory     string  `parquet:"name=charge_category, type=BYTE_ARRAY, convertedtype=UTF8"`
	ChargeClass        string  `parquet:"name=charge_class, type=BYTE_ARRAY, convertedtype=UTF8"`
	ChargeDescription  string  `parquet:"name=charge_description, type=BYTE_ARRAY, convertedtype=UTF8"`
	ChargeFrequency    string  `parquet:"name=charge_frequency, type=BYTE_ARRAY, convertedtype=UTF8"`
	ChargePeriodEnd    int64   `parquet:"name=charge_period_end, type=INT64, convertedtype=TIMESTAMP_MILLIS"`
	ChargePeriodStart  int64   `parquet:"name=charge_period_start, type=INT64, convertedtype=TIMESTAMP_MILLIS"`
	ChargeSubcategory  string  `parquet:"name=charge_subcategory, type=BYTE_ARRAY, convertedtype=UTF8"`
	EffectiveCost      float64 `parquet:"name=effective_cost, type=DOUBLE"`
	InvoiceIssuerName  string  `parquet:"name=invoice_issuer_name, type=BYTE_ARRAY, convertedtype=UTF8"`
	ListCost           float64 `parquet:"name=list_cost, type=DOUBLE"`
	ListUnitPrice      float64 `parquet:"name=list_unit_price, type=DOUBLE"`
	PricingCategory    string  `parquet:"name=pricing_category, type=BYTE_ARRAY, convertedtype=UTF8"`
	PricingQuantity    float64 `parquet:"name=pricing_quantity, type=DOUBLE"`
	PricingUnit        string  `parquet:"name=pricing_unit, type=BYTE_ARRAY, convertedtype=UTF8"`
	ProviderName       string  `parquet:"name=provider_name, type=BYTE_ARRAY, convertedtype=UTF8"`
	PublisherName      string  `parquet:"name=publisher_name, type=BYTE_ARRAY, convertedtype=UTF8"`
	ResourceId         string  `parquet:"name=resource_id, type=BYTE_ARRAY, convertedtype=UTF8"`
	ResourceName       string  `parquet:"name=resource_name, type=BYTE_ARRAY, convertedtype=UTF8"`
	ResourceType       string  `parquet:"name=resource_type, type=BYTE_ARRAY, convertedtype=UTF8"`
	ServiceCategory    string  `parquet:"name=service_category, type=BYTE_ARRAY, convertedtype=UTF8"`
	ServiceName        string  `parquet:"name=service_name, type=BYTE_ARRAY, convertedtype=UTF8"`
	SkuId              string  `parquet:"name=sku_id, type=BYTE_ARRAY, convertedtype=UTF8"`
	SkuPriceId         string  `parquet:"name=sku_price_id, type=BYTE_ARRAY, convertedtype=UTF8"`
	SubAccountId       string  `parquet:"name=sub_account_id, type=BYTE_ARRAY, convertedtype=UTF8"`
	SubAccountName     string  `parquet:"name=sub_account_name, type=BYTE_ARRAY, convertedtype=UTF8"`
	UsageQuantity      float64 `parquet:"name=usage_quantity, type=DOUBLE"`
	UsageUnit          string  `parquet:"name=usage_unit, type=BYTE_ARRAY, convertedtype=UTF8"`

	AvailabilityZone       *string  `parquet:"name=availability_zone, type=BYTE_ARRAY, convertedtype=UTF8, repetitiontype=OPTIONAL"`
	BilledCost             *float64 `parquet:"name=billed_cost, type=DOUBLE, repetitiontype=OPTIONAL"`
	CommitmentDiscountId   *string  `parquet:"name=commitment_discount_id, type=BYTE_ARRAY, convertedtype=UTF8, repetitiontype=OPTIONAL"`
	CommitmentDiscountName *string  `parquet:"name=commitment_discount_name, type=BYTE_ARRAY, convertedtype=UTF8, repetitiontype=OPTIONAL"`
	CommitmentDiscountType *string  `parquet:"name=commitment_discount_type, type=BYTE_ARRAY, convertedtype=UTF8, repetitiontype=OPTIONAL"`
	ConsumedQuantity       *float64 `parquet:"name=consumed_quantity, type=DOUBLE, repetitiontype=OPTIONAL"`
	ConsumedUnit           *string  `parquet:"name=consumed_unit, type=BYTE_ARRAY, convertedtype=UTF8, repetitiontype=OPTIONAL"`
	ContractedCost         *float64 `parquet:"name=contracted_cost, type=DOUBLE, repetitiontype=OPTIONAL"`
	ContractedUnitPrice    *float64 `parquet:"name=contracted_unit_price, type=DOUBLE, repetitiontype=OPTIONAL"`
	InstanceId             *string  `parquet:"name=instance_id, type=BYTE_ARRAY, convertedtype=UTF8, repetitiontype=OPTIONAL"`
	InstanceName           *string  `parquet:"name=instance_name, type=BYTE_ARRAY, convertedtype=UTF8, repetitiontype=OPTIONAL"`
	InstanceType           *string  `parquet:"name=instance_type, type=BYTE_ARRAY, convertedtype=UTF8, repetitiontype=OPTIONAL"`
	InvoiceId              *string  `parquet:"name=invoice_id, type=BYTE_ARRAY, convertedtype=UTF8, repetitiontype=OPTIONAL"`
	Region                 *string  `parquet:"name=region, type=BYTE_ARRAY, convertedtype=UTF8, repetitiontype=OPTIONAL"`
	TagsJSON               *string  `parquet:"name=tags, type=BYTE_ARRAY, convertedtype=UTF8, repetitiontype=OPTIONAL"`

	SourceProvider      string `parquet:"name=source_provider, type=BYTE_ARRAY, convertedtype=UTF8"`
	SourceFileName      string `parquet:"name=source_file_name, type=BYTE_ARRAY, convertedtype=UTF8"`
	ConversionTimestamp int64  `parquet:"name=conversion_timestamp, type=INT64, convertedtype=TIMESTAMP_MILLIS"`
}

func toParquetRecord(fr *types.FocusRecord) *focusRecordParquet {
	var tagsJSON *string
	if len(fr.Tags) > 0 {
		b, _ := json.Marshal(fr.Tags)
		s := string(b)
		tagsJSON = &s
	}
	toMillis := func(t time.Time) int64 { return t.UnixNano() / int64(time.Millisecond) }
	return &focusRecordParquet{
		BillingAccountId:       fr.BillingAccountId,
		BillingAccountName:     fr.BillingAccountName,
		BillingCurrency:        fr.BillingCurrency,
		BillingPeriodEnd:       toMillis(fr.BillingPeriodEnd),
		BillingPeriodStart:     toMillis(fr.BillingPeriodStart),
		ChargeCategory:         fr.ChargeCategory,
		ChargeClass:            fr.ChargeClass,
		ChargeDescription:      fr.ChargeDescription,
		ChargeFrequency:        fr.ChargeFrequency,
		ChargePeriodEnd:        toMillis(fr.ChargePeriodEnd),
		ChargePeriodStart:      toMillis(fr.ChargePeriodStart),
		ChargeSubcategory:      fr.ChargeSubcategory,
		EffectiveCost:          fr.EffectiveCost,
		InvoiceIssuerName:      fr.InvoiceIssuerName,
		ListCost:               fr.ListCost,
		ListUnitPrice:          fr.ListUnitPrice,
		PricingCategory:        fr.PricingCategory,
		PricingQuantity:        fr.PricingQuantity,
		PricingUnit:            fr.PricingUnit,
		ProviderName:           fr.ProviderName,
		PublisherName:          fr.PublisherName,
		ResourceId:             fr.ResourceId,
		ResourceName:           fr.ResourceName,
		ResourceType:           fr.ResourceType,
		ServiceCategory:        fr.ServiceCategory,
		ServiceName:            fr.ServiceName,
		SkuId:                  fr.SkuId,
		SkuPriceId:             fr.SkuPriceId,
		SubAccountId:           fr.SubAccountId,
		SubAccountName:         fr.SubAccountName,
		UsageQuantity:          fr.UsageQuantity,
		UsageUnit:              fr.UsageUnit,
		AvailabilityZone:       fr.AvailabilityZone,
		BilledCost:             fr.BilledCost,
		CommitmentDiscountId:   fr.CommitmentDiscountId,
		CommitmentDiscountName: fr.CommitmentDiscountName,
		CommitmentDiscountType: fr.CommitmentDiscountType,
		ConsumedQuantity:       fr.ConsumedQuantity,
		ConsumedUnit:           fr.ConsumedUnit,
		ContractedCost:         fr.ContractedCost,
		ContractedUnitPrice:    fr.ContractedUnitPrice,
		InstanceId:             fr.InstanceId,
		InstanceName:           fr.InstanceName,
		InstanceType:           fr.InstanceType,
		InvoiceId:              fr.InvoiceId,
		Region:                 fr.Region,
		TagsJSON:               tagsJSON,
		SourceProvider:         fr.SourceProvider,
		SourceFileName:         fr.SourceFileName,
		ConversionTimestamp:    toMillis(fr.ConversionTimestamp),
	}
}
