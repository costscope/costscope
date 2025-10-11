#!/bin/bash

# CostScope Framework Demo Script
echo " CostScope Framework Architecture Demo"
echo "========================================"
echo

# Build the application
echo " Building CostScope with Framework..."
go build -o bin/costscope-framework
echo " Build completed"
echo

# Test framework commands
echo " Testing Framework Commands:"
echo

echo "1. Plugin Management:"
./bin/costscope-framework plugin list
echo

echo "2. Analytics Framework:"
./bin/costscope-framework analyze costs --help | head -10
echo

echo "3. Reporting Framework:"
./bin/costscope-framework report summary --help | head -10
echo

echo "4. Configuration Management:"
./bin/costscope-framework config list --help | head -10
echo

echo "5. Server Management:"
./bin/costscope-framework server status --help | head -10
echo

# Test existing commands (should still work)
echo " Testing Existing Commands Compatibility:"
echo

echo "1. Original Analytics:"
./bin/costscope-framework analytics --help | head -5
echo

echo "2. Original Reports:"
./bin/costscope-framework reports --help | head -5
echo

echo "3. FOCUS Conversion:"
./bin/costscope-framework convert --help | head -5
echo

# Framework health and stats
echo " Framework Status:"
echo

echo "All commands available:"
./bin/costscope-framework --help | grep -E "^  [a-z]" | wc -l
echo " total commands registered"
echo

echo "Framework-specific commands:"
./bin/costscope-framework --help | grep -E "(analyze|report|config|plugin|server)$" | wc -l
echo " framework commands added"
echo

# Performance test
echo " Performance Test:"
echo
echo "Framework startup time:"
time ./bin/costscope-framework --help > /dev/null 2>&1
echo

echo " Framework Demo Complete!"
echo
echo " Key Features Demonstrated:"
echo "    Extensible Plugin Architecture"
echo "    Enhanced CLI Framework"
echo "    Dependency Injection"
echo "    Event-Driven Communication"
echo "    Backward Compatibility"
echo "    Auto-Discovery Commands"
echo
echo " Next Steps:"
echo "   - Add custom plugins"
echo "   - Extend command providers"
echo "   - Configure event handlers"
echo "   - Deploy to production"
