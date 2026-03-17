# Release Process

This document describes how to create a new release of the AI Together platform.

## Quick Reference

```bash
# 1. Prepare release
git checkout main && git pull origin main
git checkout -b release/vX.Y.Z

# 2. Update versions & notes (see Version Locations below)
# 3. Run tests
make test

# 4. Commit and tag
git add . && git commit -m "chore: bump version to vX.Y.Z"
git tag -a vX.Y.Z -m "Release vX.Y.Z"

# 5. Push and trigger CI
git push origin release/vX.Y.Z
git push origin vX.Y.Z

# 6. Verify release
gh release view vX.Y.Z
```

---

## Version Locations

Update these files for every release:

| File | Current Value | What to Update |
|------|---------------|----------------|
| `manager/package.json` | `"version": "0.1.0"` | `version` field |
| `RELEASE_NOTES.md` | `## Version 0.2.0` | Add new version section |

**Note**: Go modules (`server/go.mod`, `member/go.mod`) do not embed version numbers. Version is determined by Git tag only.

### Version Naming Convention

- **Format**: `vX.Y.Z` (semantic versioning)
- **X** (Major): Breaking changes
- **Y** (Minor): New features, backwards compatible
- **Z** (Patch): Bug fixes, backwards compatible

**Examples**:
- `v0.3.0` - New features (minor bump)
- `v0.2.1` - Bug fixes (patch bump)
- `v1.0.0` - First stable release (major bump)

---

## Pre-Release Checklist

Before starting a release:

- [ ] All tests pass: `make test`
- [ ] Code is formatted: `make fmt` (if available)
- [ ] `RELEASE_NOTES.md` updated with changes for this version
- [ ] Version number updated in `manager/package.json`
- [ ] No uncommitted changes: `git status`

---

## Release Process

### Step 1: Preparation

```bash
# Ensure clean main branch
git checkout main
git pull origin main

# Create release branch
git checkout -b release/vX.Y.Z
```

### Step 2: Update Version & Notes

1. **Update `manager/package.json`**:
   ```json
   {
     "name": "code-together-ui",
     "version": "X.Y.Z"
   }
   ```

2. **Update `RELEASE_NOTES.md`** - Add new section at the top:
   ```markdown
   ## Version X.Y.Z (YYYY-MM-DD)

   ### Highlights
   - Brief description of main changes

   ### New Features
   - Feature 1
   - Feature 2

   ### Bug Fixes
   - Fix 1

   ---

   ## Version 0.2.0 (当前版本)
   ...
   ```

### Step 3: Verify Build

```bash
# Run all tests
make test

# Verify build works
make build
```

### Step 4: Commit and Tag

```bash
# Commit version updates
git add .
git commit -m "chore: bump version to vX.Y.Z"

# Create annotated tag
git tag -a vX.Y.Z -m "Release vX.Y.Z"
```

### Step 5: Push and Trigger CI

```bash
# Push release branch
git push origin release/vX.Y.Z

# Push tag (triggers release workflow)
git push origin vX.Y.Z
```

The `.github/workflows/release.yml` will automatically:
1. Build macOS binaries (arm64, amd64)
2. Build Windows installer
3. Create GitHub Release with all assets

### Step 6: Verify Release

```bash
# Check release was created
gh release view vX.Y.Z

# View release in browser
gh release view vX.Y.Z --web
```

### Step 7: Post-Release

```bash
# Merge release branch back to main
git checkout main
git merge release/vX.Y.Z
git push origin main

# Clean up
git branch -d release/vX.Y.Z
```

---

## Automated Release Workflow

The `.github/workflows/release.yml` workflow handles:

### Triggers
- Push of tag matching `v*` (e.g., `v0.3.0`)

### Build Matrix

| Platform | Architecture | Output |
|----------|--------------|--------|
| macOS | arm64 (Apple Silicon) | `codeswitch-macos-arm64.zip` |
| macOS | amd64 (Intel) | `codeswitch-macos-amd64.zip` |
| Windows | amd64 | `CodeSwitch-amd64-installer.exe`, `CodeSwitch.exe` |

### Build Steps (per platform)
1. Checkout code
2. Setup Go 1.24
3. Setup Node.js 24
4. Install Wails 3
5. Install frontend dependencies
6. Generate bindings
7. Build application
8. Create archive/installer
9. Upload artifacts

### Release Creation
- Downloads all artifacts
- Creates GitHub Release with all binaries
- Includes release notes (body template)

---

## Docker Releases (Optional)

For Docker image releases:

```bash
# Build for current architecture
make docker DOCKER_TAG=vX.Y.Z

# Build and push multi-arch images
make docker-multi DOCKER_TAG=vX.Y.Z

# Or individually
make docker-server-multi DOCKER_TAG=vX.Y.Z
make docker-manager-multi DOCKER_TAG=vX.Y.Z
```

### Docker Images

| Component | Image Name |
|-----------|------------|
| Server | `genewoo/ai-together-server:vX.Y.Z` |
| Manager | `genewoo/ai-together-manager:vX.Y.Z` |

---

## Rollback Procedure

If a release has critical issues:

### Delete GitHub Release

```bash
# Delete release
gh release delete vX.Y.Z --yes

# Delete tag locally and remotely
git tag -d vX.Y.Z
git push origin :refs/tags/vX.Y.Z
```

### Hotfix Release

1. Create hotfix branch from the problematic tag:
   ```bash
   git checkout vX.Y.Z
   git checkout -b hotfix/vX.Y.Z+1
   ```

2. Apply fixes, bump version to `vX.Y.Z+1`

3. Follow normal release process

---

## Release Candidate (RC) Process

For pre-release testing:

```bash
# Tag as RC
git tag -a vX.Y.Z-rc.1 -m "Release Candidate vX.Y.Z-rc.1"
git push origin vX.Y.Z-rc.1
```

Create GitHub Release manually with `--prerelease` flag:
```bash
gh release create vX.Y.Z-rc.1 --prerelease --title "vX.Y.Z-rc.1" --notes "Release candidate for testing"
```

---

## Component-Specific Notes

### Server (switch-server)

- PostgreSQL migrations should be backwards compatible
- Check API endpoint changes for breaking changes
- Verify configuration file compatibility

### Member (codeswitch)

- Wails 3 requires Go 1.24+
- Cross-platform builds require respective OS runners
- Test installer packages on clean machines

### Manager (code-together-ui)

- React build should complete without warnings
- Verify API client compatibility with server
- Check browser compatibility (Chrome, Firefox, Safari)

---

## Common Issues

### Tag Already Exists

```bash
# Delete and recreate
git tag -d vX.Y.Z
git push origin :refs/tags/vX.Y.Z
git tag -a vX.Y.Z -m "Release vX.Y.Z"
git push origin vX.Y.Z
```

### CI Build Failed

1. Check workflow logs: `gh run list --workflow=release.yml`
2. Fix issues in release branch
3. Delete tag, amend commit, recreate tag
4. Push again

### Missing Artifacts

- Check workflow completed successfully
- Verify artifacts were uploaded
- May need to re-run the workflow

---

## Summary Checklist

```
[ ] Create release branch: release/vX.Y.Z
[ ] Update manager/package.json version
[ ] Update RELEASE_NOTES.md
[ ] Run tests: make test
[ ] Commit: chore: bump version to vX.Y.Z
[ ] Create tag: git tag -a vX.Y.Z -m "Release vX.Y.Z"
[ ] Push branch: git push origin release/vX.Y.Z
[ ] Push tag: git push origin vX.Y.Z
[ ] Verify GitHub Release created
[ ] Merge to main
[ ] Clean up release branch
```

---

## Related Files

- `RELEASE_NOTES.md` - Version history
- `.github/workflows/release.yml` - Automated release workflow
- `CONTRIBUTING.md` - Contribution guidelines
- `Makefile` - Build commands
