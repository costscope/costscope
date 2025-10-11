# CostScope Tests

This directory contains validation tests for current implementation.

**️ IMPORTANT**: Test files with `main` functions should ONLY be placed in this `tests/` directory to avoid conflicts with the main application's `main` function in `/main.go`.

## Implementation Validation

**File**: `validation.go`

**Purpose**: Comprehensive validation test for FOCUS analysis and comparison implementation featuring:
- FOCUS dataset comparison engine validation
- Enhanced analysis engine validation  
- CLI diff command creation testing
- CLI analyze command creation testing  
- Package compilation verification
- Type system alignment verification

**How to Run**:
```bash
cd tests
go run validation.go
```

**Expected Output**:
-  Diff command created successfully
-  Analyze command created successfully  
-  All packages compile successfully
-  Implementation status: Ready

## Test Organization

Each major implementation should have its corresponding validation test in this directory:
- `validation.go` - General validation test for current implementation
- Future implementations will have their own validation files

**️ WARNING**: Do NOT place test files with `main` functions in the root directory as this will cause "main redeclared" compilation errors.

## Integration with Main Application

These tests are separate from the main application to avoid function conflicts and maintain clean separation of concerns. The main application entry point remains in `/main.go`.
