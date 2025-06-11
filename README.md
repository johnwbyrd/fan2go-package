# fan2go Debian Packaging

This repository contains the Debian packaging files for [fan2go](https://github.com/markusressel/fan2go), a daemon to control system fans based on temperature readings.

The goal of this project is to provide an easy way to install `fan2go` on Debian-based systems. Packages are automatically built using GitHub Actions.

## Installation from GitHub Releases

You can install `fan2go` by downloading the `.deb` package from the [**Releases Page**](https://github.com/johnwbyrd/fan2go-package/releases) of this repository.

1.  **Download the `.deb` package:**
    *   Go to the [Releases Page](https://github.com/johnwbyrd/fan2go-package/releases).
    *   Find the latest release compatible with your system (e.g., for Debian Bookworm amd64, look for a file like `fan2go_x.y.z-1~bookworm1_amd64.deb`).
    *   Download the `.deb` file.

2.  **Install the package:**
    Open a terminal and navigate to the directory where you downloaded the file. Then run the following commands:

    ```bash
    # Replace fan2go_x.y.z-1~bookworm1_amd64.deb with the actual filename you downloaded
    sudo dpkg -i fan2go_x.y.z-1~bookworm1_amd64.deb

    # If dpkg reports missing dependencies, install them with:
    sudo apt-get update
    sudo apt-get install -f
    ```

3.  **Verify Installation:**
    Once installed, the `fan2go` service should be running. You can check its status with:
    ```bash
    systemctl status fan2go.service
    ```

4.  **Configuration:**
    The main configuration file is located at `/etc/fan2go/fan2go.yaml`. You will need to edit this file to configure your sensors, fans, and curves according to your hardware. The `fan2go` daemon uses `/var/lib/fan2go/fan2go.db` for its internal database, as specified in the default configuration.

    Use the command `fan2go detect` to help identify your system's hardware for configuration.

## Building from Source (using these packaging files)

If you wish to build the package yourself:

1.  Clone this repository:
    ```bash
    git clone https://github.com/johnwbyrd/fan2go-package.git
    cd fan2go-package
    ```
2.  Download the corresponding `fan2go` upstream source code and place it in a directory named `_upstream_src` at the root of this `fan2go-package` repository. For example, if building version `0.10.0`:
    ```bash
    git clone --branch v0.10.0 https://github.com/markusressel/fan2go.git _upstream_src
    ```
3.  Install build dependencies:
    ```bash
    sudo apt-get update
    sudo apt-get install build-essential debhelper dh-golang golang-go devscripts
    ```
4.  Build the package:
    ```bash
    dpkg-buildpackage -us -uc -b
    ```
    The `.deb` file will be created in the parent directory.

## License

*   The upstream `fan2go` code is licensed under the AGPL-3.0 license.
*   The Debian packaging files in this repository are also licensed under the AGPL-3.0 license.
