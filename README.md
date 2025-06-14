# fan2go Debian Packaging

This repository provides automated Debian packaging for [fan2go](https://github.com/markusressel/fan2go), a daemon to control system fans based on temperature readings. The packaging uses the `git-buildpackage` (gbp) workflow with full automation for upstream tracking, changelog management, and multi-distribution builds.

## Features

### Automated Workflow
- **Daily upstream monitoring**: Automatically checks for new fan2go releases
- **Smart version management**: Automatically updates `debian/changelog` with proper version handling
- **Multi-distribution builds**: Builds packages for Debian bookworm, bullseye, and trixie
- **Automated releases**: Creates GitHub releases with `.deb` packages and source files
- **DEP-14 compliant**: Follows Debian packaging standards with proper branch structure

### Build System
- **Clean build environment**: Uses `git-buildpackage` with proper export directories
- **Go modules support**: Handles complex Go dependency management
- **Pristine source handling**: Maintains reproducible builds with `pristine-tar`
- **Quality assurance**: Runs `lintian` checks on all packages

## Installation

### From GitHub Releases

1. Download the latest `.deb` package for your distribution from the [Releases page](https://github.com/johnwbyrd/fan2go-package/releases)
2. Install the package:
   ```bash
   sudo apt install ./fan2go_*_amd64.deb
   ```
3. Enable and start the service:
   ```bash
   sudo systemctl enable --now fan2go
   ```

### Package Contents

The package installs:
- **Binary**: `/usr/bin/fan2go`
- **Configuration**: `/etc/fan2go/fan2go.yaml` (default template)
- **Service**: `fan2go.service` (systemd unit)
- **Documentation**: Manual pages and examples

## Repository Structure

This repository follows the DEP-14 standard for Debian packaging:

### Branch Layout
- **`main`**: Repository automation (GitHub Actions) and documentation
- **`upstream`**: Pristine upstream source code from markusressel/fan2go
- **`debian/unstable`**: Debian packaging files + upstream source (buildable)
- **`pristine-tar`**: Metadata for reproducible source packages

### Key Files
- **`.github/workflows/updater.yml`**: Daily upstream monitoring and import
- **`.github/workflows/build-test.yml`**: Multi-distribution package building
- **`debian/`**: Complete Debian packaging configuration
- **`CLAUDE.md`**: Detailed technical documentation and workflow guide

## Automation Workflows

### Upstream Updater (`updater.yml`)
Runs daily at midnight UTC and:

1. **Monitors upstream**: Uses `uscan` to check for new fan2go releases
2. **Imports sources**: Uses `gbp import-orig` to import new upstream versions
3. **Updates changelog**: Automatically updates `debian/changelog` with smart version management
4. **Pushes changes**: Updates all relevant branches and tags
5. **Triggers builds**: Starts automated package building for new versions

### Build & Test (`build-test.yml`)
Triggered by upstream updates and:

1. **Multi-distribution builds**: Creates packages for bookworm, bullseye, and trixie
2. **Clean environments**: Uses Docker containers for reproducible builds
3. **Quality checks**: Runs `lintian` on all generated packages
4. **Artifact management**: Collects `.deb`, `.dsc`, and build metadata
5. **GitHub releases**: Creates releases with proper version naming

### Smart Version Management

The automation handles two scenarios:

**New Upstream Version** (e.g., 0.9.0 → 0.10.0):
- Creates new changelog entry: `fan2go (0.10.0-1) unstable`
- Uses `gbp dch --new-version=0.10.0-1`

**Packaging Updates** (same upstream):
- Increments Debian revision: `0.10.0-1` → `0.10.0-2`
- Uses `dch -i` to increment properly

## Local Development

### Prerequisites

```bash
# Install build dependencies
sudo apt install devscripts git-buildpackage pristine-tar golang-go libsensors-dev help2man

# For clean chroot builds (recommended)
sudo apt install pbuilder
```

### Building Packages

#### Quick Build
```bash
git clone https://github.com/johnwbyrd/fan2go-package.git
cd fan2go-package
git checkout debian/unstable

# Build in current environment
gbp buildpackage --git-upstream-branch=upstream --git-debian-branch=debian/unstable -us -uc
```

#### Clean Build (Recommended)
```bash
# Set up pbuilder (one-time setup)
sudo pbuilder create --distribution bookworm

# Build in clean chroot
gbp buildpackage --git-pbuilder --git-upstream-branch=upstream --git-debian-branch=debian/unstable
```

### Development Workflow

1. **Make packaging changes** on the `debian/unstable` branch
2. **Update changelog**: `dch -i` to add new entry
3. **Test build**: `gbp buildpackage --git-ignore-new -us -uc`
4. **Quality check**: `lintian ../*.deb`
5. **Commit changes**: Standard git workflow

### Importing New Upstream Versions

```bash
# Check for new versions
uscan --verbose --report

# Import new upstream version
gbp import-orig --uscan --upstream-branch=upstream --debian-branch=debian/unstable --pristine-tar

# Update changelog (done automatically in CI)
gbp dch --new-version=NEW_VERSION-1 --distribution=unstable --release --spawn-editor=never

# Push changes
git push origin upstream debian/unstable pristine-tar --tags
```

## Technical Challenges Solved

### Go Modules + dh-golang Compatibility

This package overcomes significant challenges with packaging Go modules using `dh-golang`:

**Issues Encountered**:
- `dh-golang` was designed for GOPATH, not Go modules
- Module cache dependency analysis breaks builds
- Test environment conflicts with dependency resolution

**Solutions Implemented**:
- **Environment variables**: `GO111MODULE=on`, `GOPROXY=https://proxy.golang.org,direct`
- **Limited scope**: `DH_GOLANG_BUILDPKG := ./cmd/...` to avoid module cache analysis
- **Custom rules**: Override `dh_auto_build` and `dh_auto_test` for proper Go modules handling
- **Build process**: Uses `make build` explicitly instead of relying on dh defaults

### Build Environment Isolation

**Challenge**: GitHub Actions workspace contamination affecting `dpkg-source`

**Solution**: 
- Export builds to `/build-area` outside mounted workspace
- Move artifacts back to mounted volume after build completion
- Use `--git-ignore-new` for GitHub Actions temporary files

## Configuration

### Default Configuration

The package installs a template configuration at `/etc/fan2go/fan2go.yaml`. Key sections:

```yaml
# Database location (writable by fan2go service)
dbPath: /var/lib/fan2go/fan2go.db

# Sensor definitions (detect with 'fan2go detect')
sensors:
  - id: cpu_temp
    hwmon:
      platform: "coretemp"
      index: 1

# Fan definitions
fans:
  - id: cpu_fan
    hwmon:
      platform: "nct6798"
      rpmChannel: 1
      pwmChannel: 1
    curve: cpu_curve

# Temperature curves
curves:
  - id: cpu_curve
    linear:
      sensor: cpu_temp
      min: 40    # 40°C = minimum fan speed
      max: 80    # 80°C = maximum fan speed
```

### Service Management

```bash
# Enable automatic startup
sudo systemctl enable fan2go

# Start service
sudo systemctl start fan2go

# Check status
sudo systemctl status fan2go

# View logs
journalctl -u fan2go -f

# Validate configuration
sudo fan2go config validate
```

## Troubleshooting

### Build Issues

**Problem**: `gbp buildpackage` fails with git repository errors
```bash
# Solution: Use export directory
gbp buildpackage --git-export-dir=../build-area --git-ignore-new
```

**Problem**: Go module dependency errors
```bash
# Solution: Ensure proper environment
export GO111MODULE=on
export GOPROXY=https://proxy.golang.org,direct
```

### Runtime Issues

**Problem**: Permission denied accessing sensors
```bash
# Solution: Ensure user is in appropriate groups
sudo usermod -a -G sensors $USER
# Or run as root (recommended for systemd service)
```

**Problem**: No sensors detected
```bash
# Solution: Set up lm-sensors first
sudo sensors-detect
sudo modprobe <detected_modules>
```

### Package Issues

**Problem**: Dependency conflicts
```bash
# Solution: Fix dependencies
sudo apt install -f
```

**Problem**: Configuration file conflicts
```bash
# Solution: Back up custom config before upgrade
sudo cp /etc/fan2go/fan2go.yaml /etc/fan2go/fan2go.yaml.backup
```

## Contributing

### Reporting Issues

- **Packaging issues**: Report to this repository
- **Software bugs**: Report to [upstream fan2go](https://github.com/markusressel/fan2go/issues)

### Submitting Changes

1. **Fork** this repository
2. **Create branch** from `debian/unstable`
3. **Make changes** to packaging files in `debian/`
4. **Test build** locally
5. **Submit pull request** against `debian/unstable` branch

### Workflow Guidelines

- **Packaging changes**: Target `debian/unstable` branch
- **Automation changes**: Target `main` branch
- **Version updates**: Handled automatically, no manual PRs needed
- **Testing**: Always test builds before submitting

## Resources

### Documentation
- [Upstream fan2go](https://github.com/markusressel/fan2go)
- [Debian New Maintainer's Guide](https://www.debian.org/doc/manuals/maint-guide/)
- [git-buildpackage Manual](https://honk.sigxcpu.org/projects/git-buildpackage/manual-html/)
- [DEP-14 Specification](https://dep-team.pages.debian.net/deps/dep14/)

### Tools
- [lm-sensors Setup](https://wiki.archlinux.org/title/Lm_sensors)
- [fan2go-tui](https://github.com/markusressel/fan2go-tui) - Terminal UI for fan2go

### Support
- **GitHub Issues**: Package-specific problems
- **Debian Packaging**: [debian-mentors mailing list](https://lists.debian.org/debian-mentors/)
- **fan2go Usage**: [Upstream documentation](https://github.com/markusressel/fan2go)

## License

- **Packaging files** (`debian/`): Licensed under the same terms as fan2go (AGPL-3.0)
- **Automation scripts**: Licensed under AGPL-3.0
- **fan2go software**: Licensed under AGPL-3.0 by Markus Ressel

See `debian/copyright` for detailed licensing information.