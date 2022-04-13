# Entropy

![test workflow](https://github.com/odpf/entropy/actions/workflows/test.yml/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/odpf/entropy)](https://goreportcard.com/report/github.com/odpf/entropy)
[![Version](https://img.shields.io/github/v/release/odpf/entropy?logo=semantic-release)](Version)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg?logo=apache)](LICENSE)

Entropy is an extensible infrastructure orchestration and application deployment tool. Entropy provides features
required for deploying and managing complex applications like resource versioning, config schema versioning, rollbacks
dry runs etc.

## Key Features

- **No Dependency:** Written in Go. It compiles into a single binary with no external dependency.
- **Extensible:** Entropy provides framework to easily write and deploy applications to your choice of cloud
- **Runtime:** Entropy can run inside VMs or containers with minimal memory footprint.

## Documentation

Explore the following resources to get started with Entropy:

- [Configuration](./docs/reference/configuration.md) to get familar with configuring Entropy.

## Installation

Install Entropy on macOS, Windows, Linux, OpenBSD, FreeBSD, and on any machine.

#### Binary (Cross-platform)

Download the appropriate version for your platform from [releases](https://github.com/odpf/entropy/releases) page. Once
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

## Usage

Entropy typically runs as a service and requires a MongoDB to store its state.

## Running locally

```sh
# Clone the repo
$ git clone https://github.com/odpf/entropy.git

# Build entropy binary file
$ make build

# Start a MongoDB instance
$ docker-compose up

# Run entropy on a recipe file
$ ./entropy serve

```

## Running tests

```sh
# Running all unit tests, excluding extractors
$ make test
```

## Contribute

Development of Entropy happens in the open on GitHub, and we are grateful to the community for contributing bugfixes and
improvements. Read below to learn how you can take part in improving Entropy.

Read our [contributing guide](https://odpf.github.io/entropy/docs/contribute/contributing) to learn about our
development process, how to propose bugfixes and improvements, and how to build and test your changes to Entropy.

To help you get your feet wet and get you familiar with our contribution process, we have a list
of [good first issues](https://github.com/odpf/entropy/labels/good%20first%20issue) that contain bugs which have a
relatively limited scope. This is a great place to get started.

This project exists thanks to all the [contributors](https://github.com/odpf/entropy/graphs/contributors).

## License

Entropy is [Apache 2.0](LICENSE) licensed.
