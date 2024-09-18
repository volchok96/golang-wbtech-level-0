# Demo NATS-streaming-service. Task Level-0 by WB Techno School.

## Description
This NATS-streaming demo application is a Go-based service designed to handle order data. It connects to a NATS-streaming channel, subscribes to it, and stores the received data in a PostgreSQL database while simultaneously caching it in memory. If the service fails, the cache is automatically restored from the database. Additionally, the service includes an HTTP server that retrieves cached data by order ID.

## Installation and Launch
The project uses Docker and Makefile to simplify the The project leverages Docker and Makefile to streamline the setup and deployment process.

### Main Commands

```bash
# Build Docker images
make build

# Start Docker containers in detached mode
make up

# Run the Go service
make run

# Stop Docker containers
make down

# Initialize the PostgreSQL database by executing the init.sql script
make initdb

# Launch NATS server locally
make nats

# Launch local development environment (includes DB init, NATS server, and the Go service)
make local

# Combine build and start processes in Docker
make docker

# Clean up the environment (stops and removes containers)
make clean

# Clean up everything, including images, volumes, and orphaned containers
make clean-all
```

## Dependencies
The project uses a number of Go packages to handle various aspects of its functionality, including database interaction, NATS-streaming, logging, and configuration management. Below is a list of the primary packages used in this project:

- **github.com/jackc/pgx/v4**: A PostgreSQL driver and toolkit for Go, used for database operations.
- **github.com/nats-io/nats.go**: NATS client for Go, enabling the service to subscribe to and interact with NATS-streaming channels.
- **github.com/rs/zerolog**: A high-performance logging library used for structured and leveled logging in the service.
- **gopkg.in/yaml.v3**: A YAML parser and serializer for Go, used for reading and processing configuration files.

### Indirect Dependencies
The project also includes several indirect dependencies, primarily related to PostgreSQL interaction and other low-level utilities. Some of these are:

- **github.com/bradfitz/gomemcache**: A memcached client for Go.
- **github.com/jackc/pgconn**: PostgreSQL connection library used by pgx.
- **github.com/jackc/pgproto3/v2**: Protocol handling for PostgreSQL communication.
- **github.com/nats-io/nkeys**: Public-key cryptography for use with NATS.
- **golang.org/x/crypto**: Cryptographic libraries used for various security-related operations.
These packages are automatically fetched when you build the project using Go modules.


## Usage
After launching the service, you can get order data using its id. Just go to the following URL in your browser:

```
http://localhost:8082/orders/{id}
```
where `{id}` is the order id.