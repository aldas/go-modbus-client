# Modbus TCP and RTU protocol client

Modbus client (TCP/RTU) over TCP for Golang.

* Modbus TCP/IP specification: http://www.modbus.org/specs.php
* Modbus TCP/IP and RTU simpler description: http://www.simplymodbus.ca/TCP.htm

## Installation

```bash
go install github.com/aldas/go-modbus-client
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
* FC23 - Read / Write Multiple Registers ([req](packet/readwritemultipleregistersrequest.go)/[resp](packet/readwritemultipleregistersresponse.go))

## Goals

* Packets separate from Client implementation
* Client (TCP/RTU) separated from Modbus packets
* Convenience methods to convert register data to/from different data types (with endianess/word order)
* Builders to group multiple fields into request batches

## Examples

Higher level API allows you to compose register requests out of arbitrary number of fields and extract those
field values from response registers with convenience methods

```go
b := modbus.NewRequestBuilder("localhost:5020", 1)

requests, _ := b.Add(b.Int64(18).UnitID(0).Name("test_do")).
    Add(b.Int64(18).Name("alarm_do_1").UnitID(0)).
    ReadHoldingRegistersTCP() // split added fields into multiple requests with suitable quantity size

client := modbus.NewTCPClient()
if err := client.Connect(context.Background(), "localhost:5020"); err != nil {
    return err
}
for _, req := range requests {
    resp, err := client.Do(context.Background(), req)
    if err != nil {
        return err
    }
    // extract response as packet.Registers instance to have access to convenience methods to 
    // extracting registers as different data types
    registers, _ := resp.(*packet.ReadHoldingRegistersResponseTCP).AsRegisters(req.StartAddress())
    alarmDo1, _ := registers.Int64(18)
    fmt.Printf("int64 @ address 18: %v", alarmDo1)
}
```

### Low level packets

```go
client := modbus.NewTCPClient(modbus.WithTimeouts(10*time.Second, 10*time.Second))
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
	req, err := packet.NewReadCoilsRequestTCP(0, 10, 9)
	req, err := packet.NewReadDiscreteInputsRequestTCP(0, 10, 9)
	req, err := packet.NewReadHoldingRegistersRequestTCP(0, 10, 9)
	req, err := packet.NewReadInputRegistersRequestTCP(0, 10, 9)
	req, err := packet.NewWriteSingleCoilRequestTCP(0, 10, true)
	req, err := packet.NewWriteSingleRegisterRequestTCP(0, 10, []byte{0xCA, 0xFE})
	req, err := packet.NewWriteMultipleCoilsRequestTCP(0, 10, []bool{true, false, true})
	req, err := packet.NewWriteMultipleRegistersRequestTCP(0, 10, []byte{0xCA, 0xFE, 0xBA, 0xBE})
```

### Builder to group fields to packets

```go
b := modbus.NewRequestBuilder("localhost:5020", 1)

requests, _ := b.Add(b.Int64(18).UnitID(0).Name("test_do")).
   Add(b.Int64(18).Name("alarm_do_1").UnitID(0)).
   ReadHoldingRegistersTCP() // split added fields into multiple requests with suitable quantity size
```

## Changelog

See [CHANGELOG.md](CHANGELOG.md)

## Tests

```bash
make check
```
