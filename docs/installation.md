# Installation

Entropy installation is simple. You can install Entropy on macOS, Windows, Linux, OpenBSD, FreeBSD, and on any machine. There are several approaches to installing Entropy.

1. Using a [pre-compiled binary](#binary-cross-platform)
2. Installing with [package manager](#homebrew)
3. Installing from [source](#building-from-source)
4. Installing with [Docker](#using-docker-image)

#### Binary (Cross-platform)

Download the appropriate version for your platform from [releases](https://github.com/goto/entropy/releases) page. Once
downloaded, the binary can be run from anywhere. You don’t need to install it into a global location. This works well
for shared hosts and other systems where you don’t have a privileged account. Ideally, you should install it somewhere
in your PATH for easy use. `/usr/local/bin` is the most probable location.

#### Homebrew

```sh
# Install entropy (requires homebrew installed)
$ brew install odpf/taps/entropy

# Upgrade entropy (requires homebrew installed)
$ brew upgrade entropy

# Check for installed entropy version
$ entropy version
```

### Building from source

To compile from source, you will need [Go](https://golang.org/) installed in your `PATH`.

```bash
# Clone the repo
$ https://github.com/goto/entropy.git

# Build entropy binary file
$ make build

# Check for installed entropy version
$ ./entropy version
```

### Using Docker image

Entropy ships a Docker image [odpf/entropy](https://hub.docker.com/r/goto/entropy) that enables you to use `entropy` as part of your Docker workflow.

For example, you can run `entropy version` with this command:

```bash
$ docker run odpf/entropy version
```

### Verifying the installation

To verify Entropy is properly installed, run `entropy version` on your system.

```bash
$ entropy version
```

### Running tests

```sh
# Running all unit tests, excluding extractors
$ make test
```