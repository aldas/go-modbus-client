# Modbus TCP and RTU protocol client

Modbus client (TCP/RTU) for Golang.

## Installation

```bash
go install github.com/aldas/go-modbus-client
```

## Supported functions

* FC1 - Read Coils
* FC2 - Read Discrete Inputs
* FC3 - Read Holding Registers
* FC4 - Read Input Registers
* FC5 - Write Single Coil
* FC6 - Write Single Register

## Goal

* Packets separate from Client
* Client (TCP/RTU) separated from Modbus packets
* Builders to group fields into request batches

## Examples

### Low level packets

### Builder to group fields to packets

## Changelog

See [CHANGELOG.md](CHANGELOG.md)

## Tests

```bash
make check
```
