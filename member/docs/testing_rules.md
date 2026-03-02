# Testing Rules for Member App

## Critical Rule: NO Live Data in Tests

**ALL tests MUST use temporary directories instead of `~/.code-together` live data.**

### Why This Matters

Tests that use live databases:
- Modify/delete user data during development
- Cause unreliable test results (dependent on user's actual data)
- Create race conditions when app is running while tests execute
- Risk corrupting production databases

### Required Pattern

Every test that uses a database or config files MUST:

```go
func setupMyTest(t *testing.T) {
    t.Helper()

    // Create temp directory for test database
    tmpDir := t.TempDir()

    // Set HOME to temp dir to avoid using real database
    oldHome := os.Getenv("HOME")
    os.Setenv("HOME", tmpDir)
    t.Cleanup(func() {
        os.Setenv("HOME", oldHome)
    })

    // Initialize database in temp directory
    if err := db.Init(); err != nil {
        t.Fatalf("failed to initialize test database: %v", err)
    }

    t.Cleanup(func() {
        db.Close()
    })
}
```

### Files Already Following This Rule

✅ `hookservice_test.go` - Uses `setupHookServiceTest(t)` with `t.TempDir()`
✅ `collect_test.go` - Uses `setupTestDB(t)` with `t.TempDir()`
✅ `collectorutil_test.go` - Pure unit tests (no database)
✅ `configservice_test.go` - Uses `t.TempDir()` for config files
✅ `opencodesettings_test.go` - Uses `t.TempDir()` with `HOME` override
✅ `providerservice_test.go` - Pure unit tests (no database)
✅ `providerrelay_test.go` - Pure unit tests (no database)
✅ `usagesyncservice_test.go` - Uses `setupUsageSyncTest(t)` with `t.TempDir()`

### Database Close Pattern

Both `db.Close()` and `hookdb.Close()` MUST reset the global database variable to `nil` to allow re-initialization:

```go
func Close() error {
    mu.Lock()
    defer mu.Unlock()

    if DB != nil {
        err := DB.Close()
        DB = nil // Reset to allow re-initialization
        return err
    }
    return nil
}
```

This allows tests to run sequentially without interference.

### Verification

To verify tests don't use live data:

```bash
# Check timestamps before
ls -la ~/.code-together/*.db

# Run tests
go test ./member/...

# Check timestamps after (should be unchanged)
ls -la ~/.code-together/*.db
```

### What NOT to Do

❌ **NEVER use `TestMain` to initialize a shared database for all tests:**
```go
// WRONG - Uses live database!
func TestMain(m *testing.M) {
    if err := db.Init(); err != nil {
        panic(err)
    }
    defer db.Close()
    m.Run()
}
```

❌ **NEVER initialize databases without redirecting HOME:**
```go
// WRONG - Uses ~/.code-together/app.db!
func TestMyService(t *testing.T) {
    db.Init()  // Creates real database!
    // ...
}
```

❌ **NEVER write to real config files:**
```go
// WRONG - Modifies user's real config!
func TestConfig(t *testing.T) {
    cs := NewConfigService()  // Uses ~/.code-together/config.json
    cs.Save(config)
}
```

### Pure Unit Tests (No Database Needed)

For tests that don't need database access, prefer pure unit tests:

```go
func TestModelReplacement(t *testing.T) {
    // Test pure functions without database
    result := ReplaceModelInRequestBody(input, "new-model")
    // No database needed!
}
```

### Checklist for New Tests

When adding a new test file, verify:

- [ ] Test helper creates `t.TempDir()`
- [ ] Sets `HOME` environment variable to temp directory
- [ ] Uses `t.Cleanup()` to restore `HOME`
- [ ] Uses `t.Cleanup()` to close database
- [ ] Database files are created in temp directory only
- [ ] Test passes when run multiple times
- [ ] Test passes when run in parallel with other tests
- [ ] No files created in `~/.code-together/` during test execution
