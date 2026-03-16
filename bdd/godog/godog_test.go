// Package godog provides the main test runner for BDD tests
package godog

import (
	"flag"
	"fmt"
	"os"
	"testing"

	"github.com/cucumber/godog"
)

// GODOG flags
var (
	godogFormat = flag.String("godog.format", "pretty", "godog output format (pretty, junit, etc.)")
	godogTags   = flag.String("godog.tags", "", "godog tags to filter scenarios")
	godogPaths  = flag.String("godog.paths", "../features", "godog feature file paths")
)

func init() {
	// Parse flags for godog
	flag.Parse()
}

// Main function for godog test runner (optional - can be run via go test)
// To run tests directly: go test -v ./godog
// To run with flags: go test -v -godog.format=junit ./godog
func TestMain(m *testing.M) {
	flag.Parse()

	status := godog.TestSuite{
		Name: "bdd",
		TestSuiteInitializer: InitializeScenario,
		Options: &godog.Options{
			Format:   *godogFormat,
			Paths:    []string{*godogPaths},
			Tags:     *godogTags,
			Strict:   true,
			NoColors: false,
		},
	}.Run()

	if status > 0 {
		fmt.Fprintf(os.Stderr, "BDD tests failed with status: %d\n", status)
		os.Exit(1)
	}

	os.Exit(0)
}
