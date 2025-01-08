# Incident Checker

A watcher for service incidents.
A Go-based monitoring application that checks for service incidents and sends notifications via ntfy.sh.

![image](image.jpg)

## Purpose

This application monitors a status API for service incidents and:
- Sends notifications when new incidents are detected in "outage" or "degraded" state
- Prevents duplicate notifications for the same incident
- Monitors internet connectivity
- Runs continuously with configurable polling intervals

## Features

- Startup notification with node identification
- Periodic incident polling (every 60 seconds)
- Internet connectivity monitoring
- Incident state filtering (outage/degraded)
- Duplicate notification prevention
- Multi-architecture support

## Requirements

- Go 1.x
- Make (for building)
- Internet connectivity
- Environment Variables:
  - `NODE_NAME`: Custom node identifier (optional)
  - `HOSTNAME`: Fallback node identifier (optional)

## Building

The project includes a Makefile with several build targets:

```bash
# Build for current architecture
make build

# Build for ARM (e.g., Raspberry Pi)
make build-arm

# Build for 32-bit x86
make build-amd32

# Build all architectures
make all
```

Output binaries:
- `my-incident-checker` (native)
- `my-incident-checker-arm` (ARM)
- `my-incident-checker-386` (32-bit x86)

## Running

1. Build the application for your architecture:
```bash
make build
```

2. Run with a node name:
```bash
NODE_NAME=my-node ./my-incident-checker
```

The application will:
1. Check internet connectivity
2. Send a startup notification
3. Begin polling for incidents
4. Send notifications for new outages or degraded services

## Monitoring

The application logs:
- Startup status
- Connectivity issues
- Polling errors
- Notification delivery status

## Architecture

- Uses standard Go libraries
- HTTP-based communication
- In-memory incident tracking
- Configurable endpoints and intervals