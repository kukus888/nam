# NAM - New Appliccation Management
This tool is targeted at Level 2 Application Support Specialists. 

## Current features
 - Application, server, and healthcheck template management
 - Secret management
 - Health check functionality 
 - Dashboard

## Planned features
 - Remote control (servers via SSH) - Actions
 - Tag feature (tag servers, apps, and instances for easier filtering)
 - Load balancer integration (nginx, F5)
 - IBM MQ integration

# Installation

## From source
Download or clone the latest repo from github. Build the program with `src/build.sh`. You may want to use [CGO_ENABLE](https://github.com/go101/go101/wiki/CGO-Environment-Setup). This will output the `nam.static.linux64.bin` and dynamic binary. If running on separate system, you need to manually copy the `config.yaml` resource. In the end, run whichever binary you want to run (static is there due to compatibility reasons when deploying to unknown or incompatible system). For correct installation, you need the correct structure:
```bash
nam.static.linux64.bin # Binary of the NAM itself
config.yaml # Configuration file
```
The NAM will output the logs to STDOUT. It supports structured logging via [log/slog](https://pkg.go.dev/log/slog) package.

## From release
Exact same as from source, except we build the binary for you. Check out the releases section of this repo.

# Setup
## Configuration
NAM is configured via `src/config.yaml` file. This file can be inputted with the `-config <path>` argument. See `config.example.yaml` for detailed description.

## Database
NAM expects a postgres database. Provide the connection string to the config file mentioned above. Migrations of schemas are handled automatically on startup via [migrate](https://pkg.go.dev/github.com/golang-migrate/migrate/v4) library. Migration files themselves are embedded in the binary. You can view these migrations in `src/layers/data/migrations` folder. 

## First run
NAM uses Role-Based Access Control (RBAC) by default. On the first run, there is no admin account. To create one, NAM provides a **one-time-use endpoint**. This endpoint is **automatically disabled** as soon as the first user is created (the first user will be the admin). After that, it cannot be accessed again.

Adjust the following command for your environment:

```bash
curl http://localhost:8080/login/setup -d '{"username":"admin","password":"admin","email":"admin@nam.local"}' -X POST
```

A successful response will look like:

```json
{"message":"Admin user created successfully","email":"admin@nam.local","user_id":1}
```

Admin's password can be changed later in `settings/users`, or `/profile`.

## Podman development environment
To setup podman environment, simply enter `dev-env-docker` folder, and run:
```bash
podman compose up -d
```
This will create a two test subjects (simple http web server and openssh server).

`web` - Returns `200 OK` on `GET /` and errors on other endpoints. For healthcheck testing.

`openssh-server` - Server for playing with ssh. Default login is `admin:password`; `ssh admin@127.0.0.1 -p 2222`. For actions testing.
