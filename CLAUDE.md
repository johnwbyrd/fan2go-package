# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a **Debian packaging repository** for fan2go using the `git-buildpackage` (gbp) workflow. The repository creates and maintains Debian packages for the upstream fan2go project (https://github.com/markusressel/fan2go) with automated upstream tracking and multi-distribution builds.

**Repository URL**: https://github.com/johnwbyrd/fan2go-package

## Git-Buildpackage Architecture

This repository follows the standard Debian `git-buildpackage` branch structure:

### Branch Layout
- **`main`**: Repository automation (.github/workflows/) and documentation
- **`upstream`**: Contains pristine upstream source code from markusressel/fan2go
- **`debian/unstable`**: Upstream source + debian/ packaging directory (DEP-14 compliant)
- **`pristine-tar`**: Stores compressed tarballs for reproducible builds

### Automated Workflows

**Upstream Updater (`.github/workflows/updater.yml`)**
- Runs daily at midnight UTC and on workflow dispatch
- Uses `gbp import-orig` to check for new upstream releases
- Automatically imports new versions to the `upstream` branch
- Merges upstream into `debian/unstable` branch when new versions are found
- Leverages `debian/watch` file to monitor https://github.com/markusressel/fan2go/tags
- Triggers automated builds via `build-test.yml` on successful import

**Build & Test (`.github/workflows/build-test.yml`)**
- Triggered on pushes to `main` or `debian/unstable`, or workflow dispatch
- Builds packages for multiple Debian distributions: bookworm, bullseye, trixie
- Checks out `debian/unstable` branch for building (contains upstream source + packaging)
- Runs tests and captures output in release artifacts
- **Includes lintian checks** for package quality assurance
- Creates pre-releases with .deb packages and build artifacts for each distribution

## Packaging Commands

### Local Development
```bash
# Build package locally
dpkg-buildpackage -us -uc

# Build with git-buildpackage
gbp buildpackage --git-upstream-branch=upstream --git-debian-branch=debian/unstable

# Import new upstream version manually
gbp import-orig --uscan --upstream-branch=upstream --debian-branch=debian/unstable --pristine-tar

# Check for upstream updates
uscan --verbose --report
```

### Testing Package Installation
```bash
# After building, test the .deb package
sudo dpkg -i ../fan2go_*.deb
sudo apt-get install -f  # Fix any dependency issues
```

## Debian Package Configuration

### Key Files in `/debian/`
- **`control`**: Package metadata, dependencies, maintainer info
- **`rules`**: Build process automation (uses dh with golang support)
- **`changelog`**: Version history and release notes
- **`watch`**: Upstream version monitoring configuration
- **`gbp.conf`**: Git-buildpackage configuration
- **`fan2go.install`**: File installation mapping
- **`fan2go.service`**: Systemd service definition
- **`fan2go.yaml`**: Default configuration template

### Build Dependencies
- `debhelper-compat (= 13)`
- `dh-golang`
- `golang-1.23 (>= 1.23.1)`
- `libsensors-dev`
- `help2man`

### Runtime Dependencies
- `lm-sensors`
- `systemd`

## Workflow Maintenance

### Updating Packaging
1. Modify files in `/debian/` directory on `debian/unstable` branch
2. Update `debian/changelog` with new entry using `dch -i`
3. Commit changes to `debian/unstable` branch
4. Push to trigger automated builds

### Manual Upstream Update
If the automatic updater fails, manually import upstream using `gbp import-orig --uscan` on the `debian/unstable` branch.

### Release Management
- Pre-releases are created automatically for each distribution build
- Full releases should be created manually after testing
- Each distribution gets its own build artifact

## Repository Structure Understanding

This repository contains:
- **`main` branch**: Automation workflows (.github/workflows/) and documentation (README.debian.md, CLAUDE.md)
- **`debian/unstable` branch**: Current upstream source code + debian/ packaging directory
- **`upstream` branch**: Pristine upstream source imports

The actual fan2go source code is maintained separately at https://github.com/markusressel/fan2go and is automatically imported and merged into the `debian/unstable` branch for packaging.

## DEP-14 Compliance

This repository follows DEP-14 (Debian Enhancement Proposal 14) standards:

- **Branch naming**: Uses `debian/unstable` instead of non-standard `debian/sid`
- **Tag format**: Follows `upstream/<version>` and `debian/<version>` patterns  
- **Automated workflows**: Updated to work with DEP-14 branch structure
- **Lintian integration**: Package quality checks included in build process

The previous `debian/sid` structure has been preserved in the `backup-pre-dep14` branch for reference.

## Go Modules + Debian Packaging Challenges

This project encountered several challenges packaging a Go modules project with dh-golang:

### Key Issues and Solutions

**1. dh-golang vs Go Modules Compatibility**
- `dh-golang` was designed for GOPATH, not Go modules
- Required `GO111MODULE=on` and `GOPROXY=https://proxy.golang.org,direct`
- Needed `DH_GOLANG_BUILDPKG := ./cmd/...` to avoid analyzing module cache

**2. Test Environment Problems**
- `dh_auto_test` with Go modules tests ALL dependencies, not just project code
- Created ANSI spam from dependency tests (github.com/mgutz/ansi)
- Solution: Override `dh_auto_test` and run manual tests before dpkg-buildpackage

**3. Build Target Issues**
- `dh_auto_build` runs `make` without target, doesn't build binary by default
- Required `override_dh_auto_build: $(MAKE) build` to create executable
- Upstream clean rule uses `rm` without `-f`, fails on fresh checkout

**4. Critical Environment Variables**
```bash
export GO111MODULE := on                           # Enable Go modules  
export GOPROXY := https://proxy.golang.org,direct  # Allow dependency downloads
export DH_GOPKG := github.com/markusressel/fan2go  # Go import path
export DH_GOLANG_BUILDPKG := ./cmd/...             # Limit analysis scope
```

### Working debian/rules Pattern

For Go modules projects, use these overrides:
```bash
override_dh_auto_build:
	$(MAKE) build

override_dh_auto_test:
	# Skip or run diagnostics - dh environment breaks module resolution
```

### Future Improvements

- Replace GOPROXY workaround with proper Debian-packaged Go dependencies
- Consider using debian:trixie for all builds (has golang-1.24 natively)
- May need per-distribution branches when more complex packaging required

## Repository Restructuring (2025-06-14)

The repository was restructured to solve gbp merge conflicts:

### Problem Solved
- `gbp import-orig --merge` was overwriting automation files (.github/workflows/) during upstream merges
- This destroyed the automated build infrastructure when importing new upstream versions

### Solution Implemented
- **Moved automation to `main` branch**: .github/workflows/, README.debian.md, CLAUDE.md
- **Cleaned `debian/unstable` branch**: Now contains only upstream source + debian/ directory
- **Updated workflows**: Check out `debian/unstable` from `main` for building
- **Prevented workflow conflicts**: Removed redundant build-release.yml workflow
- **Fixed workflow triggers**: Only trigger builds on relevant changes

### Current Architecture
```
main branch:           .github/workflows/, README.debian.md, CLAUDE.md
debian/unstable:       upstream source + debian/ (can be safely merged by gbp)
upstream:              pristine upstream source
pristine-tar:          upstream tarball metadata
```

### gbp import-orig Behavior
- Successfully imports upstream source to `upstream` branch
- Creates pristine-tar metadata for new versions  
- Merges upstream source into `debian/unstable` branch
- **Does NOT automatically update debian/changelog**

### Outstanding Issue: Changelog Management

**Problem**: `gbp import-orig` does not update `debian/changelog` automatically, causing version mismatches:
- debian/changelog shows old version (e.g., 0.1.0-1)
- pristine-tar has new version files (e.g., 0.10.0)
- `gbp buildpackage` fails looking for wrong pristine-tar files

**Build Process**: Reads version from first line of `debian/changelog`

**Potential Solutions Being Evaluated**:
1. Add `postimport` hook to gbp.conf to run `gbp dch` after import
2. Use manual `gbp dch` commands in automation
3. Implement custom versioning logic that handles Debian revision increments properly

**Versioning Complexity**: Cannot simply append `-1` to upstream versions because:
- May need multiple Debian revisions (0.10.0-1, 0.10.0-2, etc.)
- Must handle existing versions properly
- Debian versioning schemes are more complex than simple increments

### Workflow Triggers Fixed
- **Upstream Updater**: Schedule + manual dispatch + updater.yml changes (for testing)
- **Build & Test**: Only on debian/unstable changes to debian/** files + manual dispatch + workflow_call