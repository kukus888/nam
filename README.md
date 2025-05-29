# NAM - New Appliccation Management
This tool is targeted at Level 2 Application Support Specialists. It provides Application management (start, stop, kill), as well as monitoring (health checks).

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

### DB migration
Automagic DB migration on startup.

### Configuration file
Use configuration file, and perhaps environment variables to configure application.

### IBM MQ Integration
Watch IBM MQ Queue Managers for health status.

### F5 Integration
Watch F5 Load Balancer for health status.

### Arrange servers into farms (groups)
Group servers into farms (groups) for easier management.

## Issues

This section intentionally left blank.