# Bitcoin Node Handshake Project

This project demonstrates a handshake with a Bitcoin node using the Bitcoin P2P protocol. The handshake involves sending a `version` message and waiting for a `verack` message from the node.

## Getting Started

### Prerequisites

- Docker
- Docker Compose
- Go

### Setup Bitcoin Node

To run a Bitcoin Core node using Docker, use the provided `docker-compose.yml` file.

The bitcoin node data will be stored in a directory called `bitcoin-data` at the root of the project.


Run the following command to start the Bitcoin node:

```bash
docker-compose up
```

**Important:** Wait for the Bitcoin node to start fully before launching the handshake project.


### Running the Project

To build and run the project, use the provided `Makefile`. The `Makefile` includes commands for building, running, and testing the project.

- To build the project: `make build`
- To run the project: `make run`
- To test the project: `make test`

### Project Overview

This project implements a handshake with a Bitcoin node by following the [Bitcoin P2P protocol documentation](https://en.bitcoin.it/wiki/Protocol_documentation#version).

#### Go Routines

The project uses two main goroutines for handling network communication:

1. **Read Goroutine**: Continuously reads messages from the Bitcoin node.
2. **Send Goroutine**: Handles sending messages to the Bitcoin node.

#### Connection Management

- The connection is closed once the `verack` message is sent and received.
- If an unknown command is received before `verack`, the connection will be closed.

#### Version Checking

The project includes version checking to ensure the received `version` message is valid. It verifies the magic bytes, command, checksum, and payload length.

### Configuration

The configuration is directly put in the Go code, but ideally, it should be in a `config.yaml` file with more validation. This would make it easier to manage and validate configurations.

### References

This project uses the Bitcoin P2P protocol documentation as a reference:

- [Bitcoin P2P Protocol Documentation](https://en.bitcoin.it/wiki/Protocol_documentation#version)

## Example Configuration

Here is an example of the configuration used in the project:

```go
const (
    ProtocolVersion = 70016
    Services        = 1
    UserAgent       = "/Satoshi:0.21.0/"
    StartHeight     = 0
    NodeID          = 12345
    BTCNodeHost     = "127.0.0.1"
    BTCNodePort     = 8333
    Host            = "0.0.0.0"
    Port            = 8333
)
```

This setup allows the project to connect to a local Bitcoin node running on `127.0.0.1:8333`.

## Conclusion

This project demonstrates a basic handshake with a Bitcoin node using Go. It includes reading and sending messages using goroutines, managing the connection lifecycle, and validating messages according to the Bitcoin P2P protocol.



