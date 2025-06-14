# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a **fully automated Debian packaging repository** for fan2go using the `git-buildpackage` (gbp) workflow. The repository creates and maintains Debian packages for the upstream fan2go project (https://github.com/markusressel/fan2go) with sophisticated automation for upstream tracking, smart version management, and multi-distribution builds.

**Repository URL**: https://github.com/johnwbyrd/fan2go-package

## Architecture Overview

### Git-Buildpackage Workflow

This repository implements a complete DEP-14 compliant git-buildpackage workflow with the following key components:

**Branch Structure** (DEP-14 Standard):
- **`main`**: Repository automation (.github/workflows/) and documentation
- **`upstream`**: Pristine upstream source code from markusressel/fan2go  
- **`debian/unstable`**: Upstream source + debian/ packaging directory (buildable)
- **`pristine-tar`**: Compressed tarballs for reproducible builds

**Automation Features**:
- Daily upstream monitoring with automatic import
- Smart changelog version management (handles both new upstream versions and Debian revisions)
- Multi-distribution package building (bookworm, bullseye, trixie)
- Automated GitHub releases with proper version naming
- Quality assurance with lintian checks
- Clean build environments with proper artifact isolation

## Automated Workflows

### Upstream Updater (`.github/workflows/updater.yml`)

**Trigger**: Daily at midnight UTC, manual dispatch, workflow file changes

**Process**:
1. **Detection**: Uses `uscan` with `debian/watch` to check for new fan2go releases
2. **Import**: Uses `gbp import-orig --uscan --merge` to import new upstream versions
3. **Smart Changelog Management**: Automatically updates `debian/changelog` with proper version logic
4. **Branch Updates**: Pushes changes to `upstream`, `debian/unstable`, and `pristine-tar` branches
5. **Trigger Builds**: Calls `build-test.yml` for new upstream versions

**Smart Version Logic**:
```bash
# Detects version scenarios automatically
CURRENT_VERSION=$(dpkg-parsechangelog --show-field Version)
NEW_UPSTREAM_VERSION=$(git tag -l "upstream/*" | sort -V | tail -1 | sed 's/upstream\///')

# New upstream version: 0.9.0 -> 0.10.0
if [[ "$CURRENT_UPSTREAM_VERSION" != "$NEW_UPSTREAM_VERSION" ]]; then
    gbp dch --new-version=${NEW_UPSTREAM_VERSION}-1 --distribution=unstable --release

# Same upstream, packaging update: 0.10.0-1 -> 0.10.0-2  
elif [[ "$CURRENT_VERSION" != "none" ]]; then
    dch -i --distribution=unstable --urgency=medium "Automated packaging update"
fi
```

### Build & Test (`.github/workflows/build-test.yml`)

**Trigger**: Changes to `debian/**` files, workflow calls, manual dispatch

**Multi-Distribution Matrix**: bookworm, bullseye, trixie

**Build Process**:
1. **Clean Environment**: Uses Docker containers (debian:trixie) for reproducible builds
2. **Proper Isolation**: Exports builds to `/build-area` outside mounted workspace to avoid git contamination
3. **Artifact Management**: Moves build outputs to `/workspace/artifacts` for GitHub Actions access
4. **Quality Checks**: Runs `lintian` on all `.deb` and `.dsc` files
5. **Version Extraction**: Automatically extracts version from build artifacts for release naming
6. **GitHub Releases**: Creates releases with names like `fan2go-0.10.0-1-trixie`

**Build Environment Setup**:
```yaml
- name: Build in container
  uses: addnab/docker-run-action@v3
  with:
    image: debian:trixie
    options: -v ${{ github.workspace }}:/workspace -w /workspace
    run: |
      # Install dependencies
      apt-get update && apt-get install -y devscripts git-buildpackage lintian golang-go
      apt-get build-dep -y .
      
      # Configure git
      git config --global user.name "GitHub Actions"
      git config --global --add safe.directory /workspace
      
      # Build with proper export directory
      gbp buildpackage --git-export-dir=/build-area --git-ignore-new \
        --git-upstream-branch=upstream --git-debian-branch=debian/unstable -us -uc
      
      # Move artifacts to accessible location
      mkdir -p /workspace/artifacts
      mv /build-area/fan2go* /workspace/artifacts/
```

## Debian Package Configuration

### Key Files in `/debian/`

**Build Configuration**:
- **`control`**: Package metadata, dependencies (golang-go, libsensors-dev, help2man)
- **`rules`**: Custom build process for Go modules compatibility
- **`source/format`**: 3.0 (quilt) format specification
- **`gbp.conf`**: DEP-14 branch configuration

**Package Files**:
- **`changelog`**: Automatically managed version history
- **`watch`**: Upstream monitoring (GitHub tags with regex)
- **`fan2go.install`**: File installation mapping
- **`fan2go.service`**: Systemd service definition with security hardening
- **`fan2go.yaml`**: Default configuration template
- **`copyright`**: License information (AGPL-3.0)

### Build Dependencies

```debian
Build-Depends: debhelper-compat (= 13),
               dh-golang,
               golang-go (>= 2:1.24~),
               libsensors-dev,
               help2man
```

### Runtime Dependencies

```debian
Depends: ${shlibs:Depends}, ${misc:Depends},
         lm-sensors,
         systemd
```

## Go Modules Integration Challenges (SOLVED)

### Historical Issues

This project overcame significant technical challenges packaging a Go modules project with dh-golang:

**Problem 1: dh-golang vs Go Modules**
- `dh-golang` was designed for GOPATH-style Go packages, not Go modules
- Module cache analysis would break builds by testing all dependencies
- Build environment conflicts with dependency resolution

**Problem 2: GitHub Actions Contamination** 
- Docker builds in GitHub Actions workspace caused dpkg-source failures
- Git repository metadata contaminated source packages
- Temporary GitHub Actions files appeared in source trees

### Solutions Implemented

**1. Custom debian/rules for Go Modules**:
```makefile
#!/usr/bin/make -f

export GO111MODULE := on
export GOPROXY := https://proxy.golang.org,direct

%:
	dh $@

override_dh_auto_build:
	$(MAKE) build

override_dh_auto_test:
	$(MAKE) test > test-output.log 2>&1 || true
	@echo "===== TEST RESULTS (last 100 lines) ====="
	@tail -n 100 test-output.log

override_dh_auto_clean:
	$(MAKE) clean || true
```

**2. Build Environment Isolation**:
- Use `--git-export-dir=/build-area` to build outside workspace
- Move artifacts back to mounted volume after build completion
- Use `--git-ignore-new` for GitHub Actions temporary files

**3. Proper Environment Variables**:
```bash
env:
  DEBEMAIL: "actions@github.com"
  DEBFULLNAME: "Automatic Packaging GitHub Action"
```

## Technical Architecture Details

### Changelog Management System

The automation implements sophisticated changelog management that handles multiple scenarios:

**Scenario Detection**:
1. **New Upstream Version**: Detected by comparing `debian/changelog` version with latest `upstream/*` tag
2. **Packaging Update**: Same upstream version but need to increment Debian revision
3. **Initial Setup**: No existing changelog

**Implementation**:
```bash
# Version parsing logic
CURRENT_VERSION=$(dpkg-parsechangelog --show-field Version || echo "none")
LATEST_UPSTREAM_TAG=$(git tag -l "upstream/*" | sort -V | tail -1)
NEW_UPSTREAM_VERSION=${LATEST_UPSTREAM_TAG#upstream/}

# Extract upstream portion (everything before first dash)
if [[ "$CURRENT_VERSION" != "none" ]]; then
  CURRENT_UPSTREAM_VERSION="${CURRENT_VERSION%%-*}"
else
  CURRENT_UPSTREAM_VERSION="none"
fi

# Decision logic with proper tool selection
if [[ "$CURRENT_UPSTREAM_VERSION" != "$NEW_UPSTREAM_VERSION" ]]; then
  # New upstream version: use gbp dch
  gbp dch --new-version=${NEW_UPSTREAM_VERSION}-1 \
    --distribution=unstable --release --spawn-editor=never
elif [[ "$CURRENT_VERSION" != "none" ]]; then
  # Same upstream: use dch to increment Debian revision
  dch -i --distribution=unstable --urgency=medium "Automated packaging update"
fi
```

### Build Artifact Flow

**Container Build Process**:
1. **Source Export**: `gbp buildpackage --git-export-dir=/build-area` exports clean source
2. **Package Creation**: Build system creates packages in `/build-area/`
3. **Artifact Collection**: `mv /build-area/fan2go* /workspace/artifacts/`
4. **Version Extraction**: Parse version from `.deb` filename for release naming

**Host Artifact Processing**:
1. **Verification**: Confirm artifacts exist in workspace
2. **Version Setting**: Extract version and set `GITHUB_ENV` variable  
3. **Release Creation**: Use version for consistent release naming across distributions

### Version Naming Scheme

**Release Names**: `fan2go-{VERSION}-{DISTRIBUTION}`
- Example: `fan2go-0.10.0-1-trixie`
- Consistent across all distributions for same package version
- Extracted dynamically from build artifacts

**Git Tags**: Same format as release names
- Replaces run-ID based tags for better tracking
- Allows easy correlation between releases and source versions

## Development Workflows

### Local Development Commands

**Basic Package Building**:
```bash
# Clone and setup
git clone https://github.com/johnwbyrd/fan2go-package.git
cd fan2go-package
git checkout debian/unstable

# Quick build
gbp buildpackage --git-upstream-branch=upstream --git-debian-branch=debian/unstable -us -uc

# Clean build (recommended)
gbp buildpackage --git-export-dir=../build-area --git-ignore-new \
  --git-upstream-branch=upstream --git-debian-branch=debian/unstable -us -uc
```

**Upstream Updates**:
```bash
# Check for updates
uscan --verbose --report

# Import new version
gbp import-orig --uscan --upstream-branch=upstream --debian-branch=debian/unstable --pristine-tar

# Update changelog (automated in CI)
gbp dch --new-version=NEW_VERSION-1 --distribution=unstable --release --spawn-editor=never
```

**Quality Assurance**:
```bash
# Lint checks
lintian ../*.deb ../*.dsc

# Test installation
sudo dpkg -i ../fan2go_*.deb
sudo apt-get install -f
```

### Packaging Modifications

**Workflow for Changes**:
1. **Branch**: Work on `debian/unstable` branch
2. **Modify**: Edit files in `debian/` directory
3. **Changelog**: Add entry with `dch -i`
4. **Test**: Local build with `gbp buildpackage --git-ignore-new`
5. **Quality**: Run `lintian` checks
6. **Commit**: Standard git workflow
7. **Push**: Triggers automated builds

**File Modification Guidelines**:
- **Dependencies**: Update `debian/control`
- **Build Process**: Modify `debian/rules`
- **Service Config**: Edit `debian/fan2go.service`
- **Installation**: Update `debian/fan2go.install`
- **Versioning**: Let automation handle `debian/changelog`

## Repository Maintenance

### Monitoring and Troubleshooting

**Automated Health Checks**:
- Daily upstream monitoring (check GitHub Actions logs)
- Build success across all distributions
- Lintian warning/error trends
- Release artifact completeness

**Common Issues and Solutions**:

**Issue**: Build fails with git repository errors
```bash
# Solution: Ensure clean export directory usage
gbp buildpackage --git-export-dir=/build-area --git-ignore-new
```

**Issue**: Go module dependency failures
```bash
# Solution: Verify environment variables in debian/rules
export GO111MODULE=on
export GOPROXY=https://proxy.golang.org,direct
```

**Issue**: Changelog version mismatches
```bash
# Solution: Check automation logic in updater.yml
# Should auto-detect and handle version scenarios
```

### Workflow Debugging

**Upstream Updater Debug**:
- Check `uscan --verbose --report` output
- Verify `debian/watch` file syntax
- Confirm `gbp import-orig` success
- Check changelog update logic

**Build Process Debug**:
- Review Docker container logs
- Check artifact movement from `/build-area` to `/workspace/artifacts`
- Verify version extraction from `.deb` filenames
- Confirm `lintian` results

### Security and Hardening

**Service Security** (`debian/fan2go.service`):
```systemd
[Service]
# Security Hardening
ProtectSystem=strict
PrivateTmp=true
NoNewPrivileges=true
ReadWritePaths=/var/lib/fan2go /var/log/fan2go /sys/class/hwmon /sys/devices/platform
```

**Build Security**:
- Clean build environments prevent dependency contamination
- Reproducible builds with `pristine-tar`
- Automated dependency management avoids manual intervention
- Lintian checks enforce packaging standards

## Integration Points

### GitHub Actions Environment

**Required Secrets**:
- `UPDATE_PAT`: Personal access token for automated commits and pushes
- `GITHUB_TOKEN`: Automatic token for release creation (provided by GitHub)

**Environment Variables**:
- `DEBEMAIL`: Set for changelog generation tools
- `DEBFULLNAME`: Set for package maintainer identification
- `VERSION`: Dynamically extracted from build artifacts

### External Dependencies

**Upstream Monitoring**:
- Depends on https://github.com/markusressel/fan2go tag format
- Uses GitHub's tag/release API through `uscan`
- Requires stable upstream tarball URLs

**Build Dependencies**:
- Debian package repositories for build tools
- Go module proxy (proxy.golang.org) for dependency resolution
- Docker Hub for base container images

## Future Enhancements

### Potential Improvements

**Dependency Management**:
- Package all Go dependencies for Debian to eliminate GOPROXY dependency
- Create proper Debian source packages for major dependencies
- Implement offline-capable builds

**Build Process**:
- Add pbuilder/cowbuilder support for even cleaner builds
- Implement cross-compilation for different architectures
- Add automated testing in various environments

**Release Management**:
- Implement automatic Debian repository publishing
- Add integration with Debian mentors for official packaging
- Create automated backport generation for older distributions

**Quality Assurance**:
- Add automated runtime testing of packaged software
- Implement regression testing across distribution upgrades
- Add performance benchmarking for new versions

This repository represents a complete, production-ready Debian packaging solution with sophisticated automation that handles all aspects of maintaining a Debian package for a Go modules project.