package packet

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"strings"
)

// Data types with Double Word/Register (4 bytes) length can have different byte order when sent over wire depending of PLC vendor
// Usually data is sent in true big endian format, Big-Endian with Low Word first.
//
// Background info: http://www.digi.com/wiki/developer/index.php/Modbus_Floating_Points (about floats but 32bit int is also double word)
//
// Example:
// Our PLC (modbus serving) controller/computer is using little endian
//
// 32bit (4 byte) integer 67305985 is in hex 0x01020304 (little endian), most significant byte is 01 and the
// lowest byte contain hex value 04.
// Source: http://unixpapa.com/incnote/byteorder.html
//
// 32bit (dword) integer is in:
//      Little Endian (ABCD) = 0x01020304  (0x04 + (0x03 << 8) + (0x02 << 16) + (0x01 << 24))
//
// May be sent over tcp/udp as:
//      Big Endian (DCBA) = 0x04030201
//      Big Endian Low Word First (BADC) = 0x02010403 <-- used by WAGO 750-XXX to send modbus packets over tcp/udp
//
const (
	// BigEndian system stores the most significant byte of a word at the smallest memory address and the least
	// significant byte at the largest
	BigEndian = 1
	// LittleEndian - little-endian system, in contrast, stores the least-significant byte at the smallest address.
	LittleEndian = 2

	// Double words (word=register) (32bit types) consist of two 16bit words. Different PLCs send double words
	// differently over wire. So 0xDCBA can be sent low word (0xBA) first 0xBADC or high word (0xDC) first 0xDCBA.
	// High word first on true big/little endian and does not have separate flag.
	LowWordFirst = 4

	// When bytes for little endian are in 'ABCD' order then Big Endian Low Word First is in 'BADC' order
	// This mean that high word (BA) is first and low word (DC) for double word is last and bytes in words are in big endian order.
	BigEndianLowWordFirst = BigEndian | LowWordFirst // this is default endian+word order we use

	// BigEndianHighWordFirst is big-endian with high word first
	BigEndianHighWordFirst = BigEndian

	// LittleEndianLowWordFirst is little-endian with low word first
	LittleEndianLowWordFirst = LittleEndian | LowWordFirst
	// LittleEndianHighWordFirst is little-endian with high word first
	LittleEndianHighWordFirst = LittleEndian
)

// Registers provides more convenient access to data returned by register response
type Registers struct {
	startAddress uint16
	endAddress   uint16 // end address is not addressable. endAddress-1 is last addressable register (2 bytes)
	data         []byte
}

// NewRegisters creates new instance of Registers
func NewRegisters(data []byte, startAddress uint16) (*Registers, error) {
	dataLen := len(data)
	if dataLen < 2 {
		return nil, errors.New("data length at least 2 bytes as 1 register is 2 bytes")
	}
	if dataLen%2 != 0 {
		return nil, errors.New("data length must be odd number of bytes as 1 register is 2 bytes")
	}
	return &Registers{
		startAddress: startAddress,
		endAddress:   startAddress + uint16(dataLen/2),
		data:         data,
	}, nil
}

func (r Registers) register(address uint16) ([]byte, error) {
	if address < r.startAddress {
		return nil, errors.New("address under startAddress bounds")
	}
	if address >= r.endAddress {
		return nil, errors.New("address over startAddress+quantity bounds")
	}
	startIndex := (address - r.startAddress) * 2
	endIndex := startIndex + 2
	return r.data[startIndex:endIndex], nil
}

func (r Registers) doubleRegister(address uint16) ([]byte, error) {
	if address < r.startAddress {
		return nil, errors.New("address under startAddress bounds")
	}
	if address > (r.endAddress - 2) {
		return nil, errors.New("address over startAddress+quantity bounds")
	}
	startIndex := (address - r.startAddress) * 2
	endIndex := startIndex + 4
	return r.data[startIndex:endIndex], nil
}

func (r Registers) quadRegister(address uint16) ([]byte, error) {
	if address < r.startAddress {
		return nil, errors.New("address under startAddress bounds")
	}
	if address > (r.endAddress - 4) {
		return nil, errors.New("address over startAddress+quantity bounds")
	}
	startIndex := (address - r.startAddress) * 2
	endIndex := startIndex + 8
	return r.data[startIndex:endIndex], nil
}

// Bit checks if N-th bit is set in register. NB: Bits are counted from 0 and right to left.
func (r Registers) Bit(address uint16, bit uint8) (bool, error) {
	if bit > 15 {
		return false, errors.New("bit value more than register (16bit) contains")
	}
	register, err := r.register(address)
	if err != nil {
		return false, err
	}
	nThByte := 1 // low byte of register
	if bit > 7 {
		bit -= 8
		nThByte = 0 // high byte of register
	}
	b := register[nThByte]
	return b&(1<<bit) != 0, nil
}

// Byte returns register data as byte from given address high/low byte. By default High byte is 0th and Low byte is 1th byte.
func (r Registers) Byte(address uint16, fromHighByte bool) (byte, error) {
	return r.Uint8(address, fromHighByte)
}

// Uint8 returns register data as uint8 from given address high/low byte. By default High byte is 0th and Low byte is 1th byte.
func (r Registers) Uint8(address uint16, fromHighByte bool) (uint8, error) {
	b, err := r.register(address)
	if err != nil {
		return 0, err
	}
	if fromHighByte {
		return b[0], nil
	}
	return b[1], nil
}

// Int8 returns register data as int8 from given address high/low byte. By default High byte is 0th and Low byte is 1th byte.
func (r Registers) Int8(address uint16, fromHighByte bool) (int8, error) {
	b, err := r.register(address)
	if err != nil {
		return 0, err
	}
	if fromHighByte {
		return int8(b[0]), nil
	}
	return int8(b[1]), nil
}

// Uint16 returns register data as uint16 from given address. NB: Uint16 size is 1 register (16bits, 2 bytes).
func (r Registers) Uint16(address uint16) (uint16, error) {
	b, err := r.register(address)
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint16(b), nil
}

// Int16 returns register data as int16 from given address. NB: Int16 size is 1 register (16bits, 2 bytes).
func (r Registers) Int16(address uint16) (int16, error) {
	b, err := r.register(address)
	if err != nil {
		return 0, err
	}
	return int16(binary.BigEndian.Uint16(b)), nil
}

// Uint32 returns register data as uint32 from given address. NB: Uint32 size is 2 registers (32bits, 4 bytes).
func (r Registers) Uint32(address uint16) (uint32, error) {
	b, err := r.doubleRegister(address)
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint32(b), nil
}

// Int32 returns register data as int32 from given address. NB: Int32 size is 2 registers (32bits, 4 bytes).
func (r Registers) Int32(address uint16) (int32, error) {
	b, err := r.doubleRegister(address)
	if err != nil {
		return 0, err
	}
	return int32(binary.BigEndian.Uint32(b)), nil
}

// Uint64 returns register data as uint64 from given address. NB: Uint64 size is 4 registers (64bits, 8 bytes).
func (r Registers) Uint64(address uint16) (uint64, error) {
	b, err := r.quadRegister(address)
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint64(b), nil
}

// Int64 returns register data as int64 from given address. NB: Int64 size is 4 registers (64bits, 8 bytes).
func (r Registers) Int64(address uint16) (int64, error) {
	b, err := r.quadRegister(address)
	if err != nil {
		return 0, err
	}
	return int64(binary.BigEndian.Uint64(b)), nil
}

// Float32 returns register data as float32 from given address. NB: Float32 size is 2 registers (32bits, 4 bytes).
func (r Registers) Float32(address uint16) (float32, error) {
	b, err := r.doubleRegister(address)
	if err != nil {
		return 0, err
	}
	u := binary.BigEndian.Uint32(b)
	return math.Float32frombits(u), nil
}

// Float64 returns register data as float64 from given address. NB: Float64 size is 4 registers (64bits, 8 bytes).
func (r Registers) Float64(address uint16) (float64, error) {
	b, err := r.quadRegister(address)
	if err != nil {
		return 0, err
	}
	u := binary.BigEndian.Uint64(b)
	return math.Float64frombits(u), nil
}

// String returns register data as string starting from given address to given length.
// Data is interpreted as ASCII 0x0 (null) terminated string.
func (r Registers) String(address uint16, length uint16) (string, error) {
	if address < r.startAddress {
		return "", errors.New("address under startAddress bounds")
	}
	startIndex := (address - r.startAddress) * 2
	endIndex := startIndex + length
	// length is bytes. but data is sent in registers (2 bytes) and in big endian format. so last character for odd size
	// needs 1 more byte (it needs to be swapped)
	if length%2 != 0 {
		endIndex++
	}
	if int(endIndex) > len(r.data) {
		return "", errors.New("address over data bounds")
	}

	// TODO: clean these loops up to single for loop

	rawBytes := r.data[startIndex:endIndex]
	for i := 1; i < len(rawBytes); i++ {
		// data is in BIG ENDIAN format in register (register is 2 bytes). so every to 2 bytes needs to be switched
		if i%2 != 0 {
			previous := rawBytes[i-1]
			rawBytes[i-1] = rawBytes[i]
			rawBytes[i] = previous
		}
	}

	builder := new(strings.Builder)
	builder.Grow(int(length))
	for _, b := range rawBytes[0:length] {
		if b == 0 { // strings are terminated by first null
			break
		}
		// what we create here is ASCII string
		_, _ = fmt.Fprintf(builder, "%c", rune(b))
	}

	return builder.String(), nil
}
