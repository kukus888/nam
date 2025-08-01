# NAM - New Appliccation Management
This tool is targeted at Level 2 Application Support Specialists. It provides Application management (start, stop, kill), as well as monitoring (health checks).

# Installation

## Database migration
Currently, there are no mechanisms implemented to automatically migrate DB schemas. So, you need to create a new postgres database yourself. To migrate the schemas, just run `Postgre.sql` on your database, and it will create all tables and triggers.

## From source
Download or clone the latest repo from github. Build the program with `src/build.sh`. This will output the `nam.static.linux64.bin` and dynamic binary. If running on separate system, you need to manually copy the `config.yaml` and `web/*` resources. In the end, run whichever binary you want to run (static is there due to compatibility reasons when deploying to unknown system). For correct installation, you need the correct structure:
```bash
nam.static.linux64.bin # Binary of the NAM itself
config.yaml # Configuration file
web/ # Resources for web interface
```
The NAM will output the logs to STDOUT.

## From release
Exact same as from source, except we build the binary for you. Check out the releases section of this repo.

# Setup

NAM uses Role-Based Access Control (RBAC) by default. On the first run, there is no admin account. To create one, NAM provides a **one-time-use endpoint**. This endpoint is **automatically disabled** as soon as the first user is created (the first user will be the admin). After that, it cannot be accessed again.

Adjust the following command for your environment:

```bash
curl http://localhost:8080/login/setup -d '{"username":"admin","password":"admin"}' -X POST
```

A successful response will look like:

```json
{"message":"Admin user created successfully","user_id":1}
```

# Usage
## Start/Stop/Restart
To start/stop/restart a service, use the following endpoint:
```bash
curl http://localhost:8080/service/start -d '{"service":"httpd"}' -X POST
```
You 

# Podman development environment
To setup podman environment, simply enter `dev-env` folder, and run:
```bash
podman compose up -d
```
This will create a rundeck community cluster, as well as nginx for balancing, and two test subjects (simple http web server and openssh server).

`web` - Returns `200 OK` on `GET /` and errors on other endpoints.
`openssh-server` - Server for playing with ssh. Default login is `admin:password`; `ssh admin@127.0.0.1 -p 2222`

[CGO_ENABLE](https://github.com/go101/go101/wiki/CGO-Environment-Setup)

# TODO

### IBM MQ Integration
Watch IBM MQ Queue Managers for health status.

### F5 Integration
Watch F5 Load Balancer for health status.

### Arrange servers into farms (groups)
Group servers into farms (groups) for easier management.

Workaround: Add tagging system to tag components

### Cron job to delete old healthcheckrecords
### better build pipeline

### Rewrite the healthcheck service to meet following criteria:
- Space out checks, so that they do not all happen at the same time.
- Better TLS settings

### Better dashboard, with filtering

## Issues
- Sometimes the automagic sync between healthcheck tempaltes does not work thourhg the database (INSERT)
- Sometimes broken CSS
    - Dashboard when theres a lot of instances
    - health check history, weird amount of padding
- Unable to edit application definitions
- No cron to delete old healthcheck results
- Unable to call healthcheck directly (have to wait for interval)
- Not working /api/rest/v1/servers API

- No unified 5xx 4xx pages (make templated ones)
- No role UI
- Favicon based on what is on the page