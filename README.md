# Modbus TCP and RTU protocol client

[![License](https://img.shields.io/github/license/aldas/go-modbus-client)](LICENSE)
[![GoDoc](http://img.shields.io/badge/go-documentation-blue.svg)](https://pkg.go.dev/github.com/aldas/go-modbus-client)
[![Go Report Card](https://goreportcard.com/badge/github.com/aldas/go-modbus-client)](https://goreportcard.com/report/github.com/aldas/go-modbus-client)
[![Codecov](https://codecov.io/gh/aldas/go-modbus-client/branch/main/graph/badge.svg)](https://codecov.io/gh/aldas/go-modbus-client)

Modbus client (TCP/RTU) over TCP/UDP/Serial for Golang.

* Modbus TCP/IP specification: http://www.modbus.org/specs.php
* Modbus TCP/IP and RTU simpler description: http://www.simplymodbus.ca/TCP.htm

For questions use Github [Discussions](https://github.com/aldas/go-modbus-client/discussions)

## Installation

```bash
go get github.com/aldas/go-modbus-client
```

## Supported functions

* FC1 - Read Coils ([req](packet/readcoilsrequest.go)/[resp](packet/readcoilsresponse.go))
* FC2 - Read Discrete Inputs ([req](packet/readdiscreteinputsrequest.go)/[resp](packet/readdiscreteinputsresponse.go))
* FC3 - Read Holding Registers ([req](packet/readholdingregistersrequest.go)/[resp](packet/readholdingregistersresponse.go))
* FC4 - Read Input Registers ([req](packet/readinputregistersrequest.go)/[resp](packet/readinputregistersresponse.go))
* FC5 - Write Single Coil ([req](packet/writesinglecoilrequest.go)/[resp](packet/writesinglecoilresponse.go))
* FC6 - Write Single Register ([req](packet/writesingleregisterrequest.go)/[resp](packet/writesingleregisterresponse.go))
* FC15 - Write Multiple Coils ([req](packet/writemultiplecoilsrequest.go)/[resp](packet/writemultiplecoilsresponse.go))
* FC16 - Write Multiple Registers ([req](packet/writemultipleregistersrequest.go)/[resp](packet/writemultipleregistersresponse.go))
* FC17 - Read Server ID ([req](packet/readserveridrequest.go)/[resp](packet/readserveridresponse.go))
* FC23 - Read / Write Multiple Registers ([req](packet/readwritemultipleregistersrequest.go)/[resp](packet/readwritemultipleregistersresponse.go))

## Goals

* Packets separate from Client implementation
* Client (TCP/UDP +RTU) separated from Modbus packets
* Convenience methods to convert register data to/from different data types (with endianess/word order)
* Builders to group multiple fields into request batches
* Poller to request batches request and parse response to field values with long-running process.

## Examples

Higher level API allows you to compose register requests out of arbitrary number of fields and extract those
field values from response registers with convenience methods

Addresses without scheme (i.e. `localhost:5020`) are considered as TCP addresses. For UDP unicast use `udp://localhost:5020`.

```go
b := modbus.NewRequestBuilder("tcp://localhost:5020", 1)

requests, _ := b.
    AddField(modbus.Field{Name: "test_do", Type: modbus.FieldTypeUint16, Address: 18}).
    AddField(modbus.Field{Name: "alarm_do_1", Type: modbus.FieldTypeInt64, Address: 19}).
    ReadHoldingRegistersTCP() // split added fields into multiple requests with suitable quantity size

client := modbus.NewTCPClient()
if err := client.Connect(context.Background(), "tcp://localhost:5020"); err != nil {
    return err
}
for _, req := range requests {
    resp, err := client.Do(context.Background(), req)
    if err != nil {
        return err
    }
    // extract response as packet.Registers instance to have access to convenience methods to 
    // extracting registers as different data types
    registers, _ := resp.(*packet.ReadHoldingRegistersResponseTCP).AsRegisters(req.StartAddress)
    alarmDo1, _ := registers.Int64(19)
    fmt.Printf("int64 @ address 19: %v", alarmDo1)
    
    // or extract values to FieldValue struct
    fields, _ := req.ExtractFields(resp, true)
    assert.Equal(t, uint16(1), fields[0].Value)
    assert.Equal(t, "alarm_do_1", fields[1].Field.Name)
}
```

### Polling values with long-running process

See simple poller implementation [cmd/modbus-poller/main.go](cmd/modbus-poller/main.go).

### RTU over serial port

RTU examples to interact with serial port can be found from [serial.md](serial.md)

Addresses without scheme (i.e. `localhost:5020`) are considered as TCP addresses. For UDP unicast use `udp://localhost:5020`.

### Low level packets

```go
client := modbus.NewTCPClientWithConfig(modbus.ClientConfig{
    WriteTimeout: 2 * time.Second,
    ReadTimeout:  2 * time.Second,
})
if err := client.Connect(context.Background(), "localhost:5020"); err != nil {
    return err
}
defer client.Close()
startAddress := uint16(10)
req, err := packet.NewReadHoldingRegistersRequestTCP(0, startAddress, 9)
if err != nil {
    return err
}

ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
resp, err := client.Do(ctx, req)
if err != nil {
    return err
}

registers, err := resp.(*packet.ReadHoldingRegistersResponseTCP).AsRegisters(startAddress)
if err != nil {
    return err
}
uint32Var, err := registers.Uint32(17) // extract uint32 value from register 17
```

To create single TCP packet use following methods. Use `RTU` suffix to create RTU packets.

```go
import "github.com/aldas/go-modbus-client/packet"

req, err := packet.NewReadCoilsRequestTCP(0, 10, 9)
req, err := packet.NewReadDiscreteInputsRequestTCP(0, 10, 9)
req, err := packet.NewReadHoldingRegistersRequestTCP(0, 10, 9)
req, err := packet.NewReadInputRegistersRequestTCP(0, 10, 9)
req, err := packet.NewWriteSingleCoilRequestTCP(0, 10, true)
req, err := packet.NewWriteSingleRegisterRequestTCP(0, 10, []byte{0xCA, 0xFE})
req, err := packet.NewWriteMultipleCoilsRequestTCP(0, 10, []bool{true, false, true})
req, err := packet.NewReadServerIDRequestTCP(0)
req, err := packet.NewWriteMultipleRegistersRequestTCP(0, 10, []byte{0xCA, 0xFE, 0xBA, 0xBE})
```

### Builder to group fields to packets

```go
b := modbus.NewRequestBuilder("localhost:5020", 1)

requests, _ := b.
   AddField(modbus.Field{Name: "test_do", Type: modbus.FieldTypeUint16, Address: 18}).
   AddField(modbus.Field{Name: "alarm_do_1", Type: modbus.FieldTypeInt64, Address: 19}).
   ReadHoldingRegistersTCP() // split added fields into multiple requests with suitable quantity size
```

## Changelog

See [CHANGELOG.md](CHANGELOG.md)

## Tests

```bash
make check
```
