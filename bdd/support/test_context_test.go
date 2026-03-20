package support

import (
	"log/slog"
	"os"
	"testing"
)

func TestInitializeClients(t *testing.T) {
	ctx := &BDDTestContext{
		ServerURL: "http://localhost:8088",
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	err := ctx.InitializeClients(ctx.ServerURL, logger)

	if err != nil {
		t.Fatalf("InitializeClients failed: %v", err)
	}

	if ctx.AnonymousClient == nil {
		t.Error("AnonymousClient should not be nil after initialization")
	}

	if ctx.Logger == nil {
		t.Error("Logger should not be nil after initialization")
	}
}
