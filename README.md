# Entropy

Entropy is extensible infrastructure orchestration and application deployment service

### Installation

#### Compiling from source

It requires the following dependencies:

* Docker
* Golang (version 1.16 or above)
* Git

Run the application dependecies using Docker:

```
$ docker-compose up
```

Update the configs(db credentials etc.) as per your dev machine and docker configs.

Run the following commands to compile from source

```
$ git clone git@github.com:odpf/entropy.git
$ cd entropy
$ go build main.go
```

To run tests locally

```
$ make test
```

To run tests locally with coverage

```
$ make test-coverage
```

To run server locally

```
$ go run main.go serve
```

To view swagger docs of HTTP APIs visit `/documentation` route on the server.
e.g. [http://localhost:3000/documentation](http://localhost:3000/documentation)

#### Config

The config file used by application is `config.yaml` which should be present at the root of this directory.

For any variable the order of precedence is:

1. Env variable
2. Config file
3. Default in Struct defined in the application code

For list of all available configuration keys check the [configuration](docs/reference/configuration.md) reference.

### List of available commands

1. Serve
    - Runs the Server  `$ go run main.go serve`

2. Migrate
    - Runs the DB Migrations `$ go run main.go migrate`
