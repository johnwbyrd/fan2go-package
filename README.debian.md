# fan2go Debian Packaging

This repository contains Debian packaging for the [fan2go](https://github.com/markusressel/fan2go) fan control daemon. The packaging follows DEP-14 standards and uses git-buildpackage (gbp) workflow.

## Repository Structure

This is a DEP-14 compliant packaging repository with the following branch structure:

- `upstream` - Pristine upstream source code
- `debian/unstable` - Debian packaging for unstable/sid (main development branch)
- `pristine-tar` - Pristine tar metadata for reproducible source packages

## Building

### Prerequisites

```bash
sudo apt-get install devscripts git-buildpackage pristine-tar golang-go libsensors-dev help2man
```

### Build Process

1. Clone the repository:
```bash
git clone https://github.com/johnwbyrd/fan2go-package.git
cd fan2go-package
```

2. Switch to the debian/unstable branch:
```bash
git checkout debian/unstable
```

3. Build the package:
```bash
# Use -nc flag to skip clean phase (upstream Makefile clean rule has issues)
dpkg-buildpackage -us -uc -nc
```

The resulting .deb packages will be created in the parent directory.

## Go Modules + Debian Packaging Challenges

This package faces significant challenges due to incompatibilities between Go modules and dh-golang:

### Issues Encountered

1. **dh-golang GOPATH vs Go modules**: dh-golang was designed for GOPATH-based Go packages but fan2go uses Go modules
2. **Module cache dependencies**: dh_golang tries to analyze and test all dependencies in the module cache
3. **Build environment conflicts**: dh-golang's build environment breaks Go module resolution

### Current Workarounds

Several environment variables and debian/rules overrides are required:

```makefile
export DH_GOLANG_GO_GENERATE := 1
export GO111MODULE := on
export DH_GOPKG := github.com/markusressel/fan2go
export DH_GOLANG_BUILDPKG := ./cmd/...
export GOPROXY := https://proxy.golang.org,direct
```

Key overrides in `debian/rules`:
- `override_dh_auto_build`: Explicitly runs `make build` since dh_auto_build doesn't build the binary
- `override_dh_auto_test`: Runs diagnostics instead of tests to avoid dependency testing issues

### Temporary Solutions

- **GOPROXY enabled**: Downloads dependencies during build (not ideal for Debian packaging)
- **Limited scope**: `DH_GOLANG_BUILDPKG := ./cmd/...` limits dh-golang to only analyze the main package
- **Manual testing**: Tests are run manually before dpkg-buildpackage to capture output properly

### Future Improvements

1. Package all 39 Go dependencies for Debian to eliminate GOPROXY dependency
2. Investigate dh-golang alternatives or improvements for Go modules support
3. Work with upstream to improve Makefile clean rule (remove `-nc` flag requirement)

## Dependencies

fan2go requires 39 Go dependencies that are not currently packaged for Debian:

- github.com/spf13/cobra (and 38 others)

See `go.mod` in the upstream source for the complete list.

## Automated Workflows

### Upstream Tracking (.github/workflows/updater.yml)

- Monitors upstream releases using `uscan`
- Automatically imports new upstream versions using `gbp import-orig`
- Updates the upstream and debian/unstable branches
- Can trigger automated builds for new releases

### Build Testing (.github/workflows/build-test.yml)

- Tests package building for multiple Debian releases (bookworm, bullseye, trixie)
- Uses debian:trixie container for consistent build environment
- Runs lintian checks on generated packages
- Captures test output for debugging
- Can optionally create GitHub releases

## Installation

After building, install the package:

```bash
sudo dpkg -i ../fan2go_*.deb
sudo apt-get install -f  # Fix any dependency issues
```

## Configuration

The package installs:
- Binary: `/usr/bin/fan2go`
- Default config: `/etc/fan2go/fan2go.yaml`
- Systemd service: `fan2go.service`

Enable the service:
```bash
sudo systemctl enable --now fan2go
```

## Troubleshooting

### Build Issues

1. **"go: not found"**: Install `golang-go` package, not `golang-1.24`
2. **Clean phase fails**: Use `-nc` flag with dpkg-buildpackage
3. **Module resolution errors**: Check Go environment variables in debian/rules
4. **Test failures**: Tests run manually before dpkg-buildpackage; check test-output.log

### Runtime Issues

1. **Permission errors**: fan2go requires root privileges to access hardware sensors
2. **Missing sensors**: Ensure lm-sensors is properly configured (`sensors-detect`)
3. **Service startup**: Check `journalctl -u fan2go` for errors

## Development

To modify packaging:

1. Make changes on debian/unstable branch
2. Test build with `dpkg-buildpackage -us -uc -nc`
3. Run lintian checks: `lintian ../*.deb`
4. Update changelog: `dch -i`

For upstream updates:
1. Use `gbp import-orig --uscan` to import new versions
2. Resolve any packaging conflicts
3. Test build and update packaging as needed

## Resources

- [DEP-14 Specification](https://dep-team.pages.debian.net/deps/dep14/)
- [git-buildpackage Manual](https://honk.sigxcpu.org/piki/projects/git-buildpackage/)
- [dh-golang Documentation](https://pkg-go-maintainers.alioth.debian.org/packaging.html)
- [Upstream fan2go Repository](https://github.com/markusressel/fan2go)