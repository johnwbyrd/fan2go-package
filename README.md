# fan2go Debian Packaging

This repository contains the Debian packaging files for [fan2go](https://github.com/markusressel/fan2go), a daemon to control system fans based on temperature readings. The packaging process is fully automated using GitHub Actions.

## Features
- **Automatic builds**: Packages built on every push to `main`
- **Release management**: Automatic release creation with version tags
- **Clean builds**: Uses pristine Debian bookworm environment
- **Go version management**: Ensures Go 1.23.1 is used for builds

## Installation
1. Download the latest `.deb` package from the [Releases page](https://github.com/johnwbyrd/fan2go-package/releases)
2. Install using:
   ```bash
   sudo apt install ./fan2go_*_amd64.deb
   ```

## Building Locally
### Prerequisites
```bash
sudo apt install git-buildpackage golang-go
```

### Build Process
```bash
# Clone repository
git clone https://github.com/johnwbyrd/fan2go-package.git
cd fan2go-package

# Build package
gbp buildpackage \
  --git-upstream-branch=upstream \
  --git-debian-branch=debian \
  -us -uc
```

## CI/CD Pipeline
The GitHub Actions workflow:
1. Checks out the repository
2. Installs Go 1.23.1 from Debian backports
3. Imports upstream source using `gbp import-orig`
4. Builds the package in a clean environment
5. Creates a GitHub release with artifacts:
   - Binary package (.deb)
   - Source package (.dsc)
   - Build information (.buildinfo)
   - Source tarball (.tar.xz)

## Repository Structure
- `upstream` branch: Pristine source code
- `debian` branch: Packaging files only
- `main` branch: Combined source and packaging (buildable state)

## Contributing
Contributions are welcome! Please submit pull requests against the `debian` branch.
