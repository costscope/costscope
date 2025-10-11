package response

// Centralized error code constants for 4xx client errors.
// Keep list small & stable; prefer adding only when clients need to branch.
// Document any new code in API docs (Response helpers & error codes section).
const (
	ErrCodeBadRequest         = "bad_request"         // generic malformed input / validation failure
	ErrCodeMissingInput       = "missing_input"       // required parameter absent
	ErrCodeLoadParquet        = "load_parquet"        // parquet load/open failure
	ErrCodeInvalidGranularity = "invalid_granularity" // unsupported time granularity value
)
