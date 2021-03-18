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
* FC15 - Write Multiple Coils
* FC16 - Write Multiple Registers
* FC23 - Read / Write Multiple Registers

## Goals

* Packets separate from Client implementation
* Client (TCP/RTU) separated from Modbus packets
* Convenience methods to convert register data to different data types (with endianess/word order)
* Builders to group multiple fields into request batches

## Examples

Higher level API allows you to compose register requests out of arbritary number of fields and extract those
field values from response registers with convenience methods

```go
b := modbus.NewRequestBuilder("localhost:5020", 1)

requests, _ := b.Add(b.Int64(18).UnitID(0).Name("test_do")).
    Add(b.Int64(18).Name("alarm_do_1").UnitID(0)).
    ReadHoldingRegistersTCP() // split added fields into multiple requests with suitable quantity size

client := modbus.NewClient()
if err := client.Connect(context.Background(), "localhost:5020"); err != nil {
    return err
}
for _, req := range requests {
    resp, err := client.Do(context.Background(), req)
    if err != nil {
        return err
    }
    // extract response as packet.Registers instance to have access to convenience methods to extracting registers
    // as different data types
    registers, _ := resp.(*packet.ReadHoldingRegistersResponseTCP).AsRegisters(req.StartAddress())
    alarmDo1, _ := registers.Int64(18)
    fmt.Printf("int64 @ address 18: %v", alarmDo1)
}
```

### Low level packets

```go
client := modbus.NewTCPClient()
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

### Builder to group fields to packets

## Changelog

See [CHANGELOG.md](CHANGELOG.md)

## Tests

```bash
make check
```
