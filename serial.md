# RTU over serial connection

These are examples for reading [SHT20 Temp sensor with Modbus RS2485](https://www.aliexpress.com/item/32923628973.html) sensor
using cheap [USB To RS485 422](https://www.aliexpress.com/item/32888122294.html) dongle with different libraries.

## github.com/jacobsa/go-serial/serial

Example for: [github.com/jacobsa/go-serial/serial](https://github.com/jacobsa/go-serial/)
```go
// import "github.com/jacobsa/go-serial/serial"
serialPort, err := serial.Open(serial.OpenOptions{
    // cheap Aliexpress USB-serial dongle (Bus 001 Device 005: ID 1a86:7523 QinHeng Electronics HL-340 USB-Serial adapter)
    // and connected "SHT20 Temperature Humidity Sensor Modbus RS485" 
    PortName:        "/dev/ttyUSB0",
    BaudRate:        9600,
    DataBits:        8,
    StopBits:        1,
    MinimumReadSize: 1,
})
if err != nil {
    return
}

client := modbus.NewSerialClient(serialPort)
req, _ := packet.NewReadInputRegistersRequestRTU(1, 1, 2)

resp, err := client.Do(context.Background(), req)
if err != nil {
    return
}
fmt.Printf("response function code: %v\n", resp.FunctionCode())
rtu := resp.(*packet.ReadInputRegistersResponseRTU)
fmt.Printf("response data as hex: %x\n", rtu.Data)

registers, _ := rtu.AsRegisters(1)
temp, _ := registers.Int16(1) // convert register 1 (temperature) bytes as int16 value
fmt.Printf("temperature: %v\n", float32(temp) / 10)
```

## github.com/tarm/serial

Example for: [github.com/tarm/serial](https://github.com/tarm/serial)
```go
serialPort, err := serial.OpenPort(&serial.Config{Name: "/dev/ttyUSB0", Baud: 9600, ReadTimeout: 2 * time.Second})
if err != nil {
    return
}

client := modbus.NewSerialClient(serialPort)
req, _ := packet.NewReadInputRegistersRequestRTU(1, 1, 2)

resp, err := client.Do(context.Background(), req)
if err != nil {
    return
}
fmt.Printf("response function code: %v\n", resp.FunctionCode())
rtu := resp.(*packet.ReadInputRegistersResponseRTU)
fmt.Printf("response data as hex: %x\n", rtu.Data)

registers, _ := rtu.AsRegisters(1)
temp, _ := registers.Int16(1) // convert register 1 (temperature) bytes as int16 value
fmt.Printf("temperature: %v\n", float32(temp) / 10)
```

## Raw syscall

Example for with raw syscalls (only on unix/linux systems)
```go
serialPort, _ := os.OpenFile("/dev/ttyUSB0", syscall.O_RDWR|syscall.O_NOCTTY|syscall.O_NONBLOCK, 0666)

// create timeout on the file handle
// Read constant descriptions here: https://man7.org/linux/man-pages/man3/termios.3.html
tios := syscall.Termios{
    // input modes
    Iflag:  syscall.IGNPAR, // Ignore framing errors and parity errors
    // control modes
    Cflag:  syscall.CS8 | // character size 8 bits
        syscall.CREAD | // Enable receiver.
        syscall.CLOCAL | // Ignore modem status lines.
        syscall.B9600, // 9600 baud rate
    // special characters
    Cc:     [32]uint8{
        syscall.VMIN: 0, // Minimum number of characters for noncanonical read (MIN)
        syscall.VTIME: uint8(20) // Timeout in deciseconds for noncanonical read (TIME)
    }, // 2.0s timeout for serialPort.Read() calls
    Ispeed: syscall.B9600, // 9600 baud rate
    Ospeed: syscall.B9600, // 9600 baud rate
}
// syscall
syscall.Syscall6(syscall.SYS_IOCTL, uintptr(serialPort.Fd()),
    uintptr(syscall.TCSETS), uintptr(unsafe.Pointer(&tios)),
    0, 0, 0)

client := modbus.NewSerialClient(serialPort)
req, _ := packet.NewReadInputRegistersRequestRTU(1, 1, 2)

resp, err := client.Do(context.Background(), req)
if err != nil {
    return
}
fmt.Printf("response function code: %v\n", resp.FunctionCode())
rtu := resp.(*packet.ReadInputRegistersResponseRTU)
fmt.Printf("response function code: %x\n", rtu.Data)

registers, _ := rtu.AsRegisters(1)
temp, _ := registers.Int16(1) // convert register 1 (temperature) bytes as int16 value
fmt.Printf("temperature: %v\n", float32(temp) / 10)
```
