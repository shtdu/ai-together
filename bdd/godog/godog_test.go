// Package godog provides the main test runner for BDD tests
package godog

// This file is intentionally minimal. The main test runner is defined in godog_suite_test.go
// which uses the TestGodog function to run the godog test suite.
//
// To run tests:
//   ./bdd-test.sh                    # Run all BDD tests
//   cd godog && go test -v           # Run with verbose output
//   ./bdd-test.sh --tags "@smoke"    # Run specific tags
