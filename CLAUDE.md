# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a **Debian packaging repository** for fan2go using the `git-buildpackage` (gbp) workflow. The repository creates and maintains Debian packages for the upstream fan2go project (https://github.com/markusressel/fan2go) with automated upstream tracking and multi-distribution builds.

**Repository URL**: https://github.com/johnwbyrd/fan2go-package

## Git-Buildpackage Architecture

This repository follows the standard Debian `git-buildpackage` branch structure:

### Branch Layout
- **`upstream`**: Contains pristine upstream source code from markusressel/fan2go
- **`debian/sid`**: Main packaging branch with Debian-specific files
- **`pristine-tar`**: Stores compressed tarballs for reproducible builds
- **`main`**: Repository metadata and documentation
- **`gbp`**: Git-buildpackage configuration branch

### Automated Workflows

**Upstream Updater (`.github/workflows/updater.yml`)**
- Runs daily at midnight UTC and on workflow dispatch
- Uses `gbp import-orig --uscan` to check for new upstream releases
- Automatically imports new versions to the `upstream` branch
- Updates `debian/sid` branch when new versions are found
- Leverages `debian/watch` file to monitor https://github.com/markusressel/fan2go/tags

**Build & Release (`.github/workflows/build-release.yml`)**
- Triggered on pushes to `debian/sid` or workflow dispatch
- Builds packages for multiple Debian distributions: bookworm, bullseye, trixie
- Uses appropriate Golang containers for each distribution
- Merges `upstream` branch content before building
- Runs tests and captures output in release artifacts
- Creates pre-releases with .deb packages for each distribution

## Packaging Commands

### Local Development
```bash
# Build package locally
dpkg-buildpackage -us -uc

# Build with git-buildpackage
gbp buildpackage --git-upstream-branch=upstream --git-debian-branch=debian/sid

# Import new upstream version manually
gbp import-orig --uscan --upstream-branch=upstream --debian-branch=debian/sid --pristine-tar

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
1. Modify files in `/debian/` directory
2. Update `debian/changelog` with new entry using `dch -i`
3. Commit changes to `debian/sid` branch
4. Push to trigger automated builds

### Manual Upstream Update
```bash
# If automatic updater fails, manually import upstream
git checkout debian/sid
gbp import-orig --uscan --upstream-branch=upstream --debian-branch=debian/sid --pristine-tar --no-interactive
```

### Release Management
- Pre-releases are created automatically for each distribution build
- Full releases should be created manually after testing
- Each distribution gets its own build artifact

## Repository Structure Understanding

This is **NOT** the upstream fan2go source code repository. This repository only contains:
- Debian packaging metadata (`/debian/` directory)
- Automated workflows for upstream tracking and package building  
- Git-buildpackage branch structure for maintaining pristine upstream sources

The actual fan2go source code is maintained separately at https://github.com/markusressel/fan2go and is automatically imported into the `upstream` branch.