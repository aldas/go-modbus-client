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
// For example, if the number 2923517522 (hex: AE 41 56 52) was to be sent as a 32 bit unsigned integen then bytes that
// are send over the wire depend on 2 factors - byte order and/or register/word order.
//
// Some devices store the 32bits in 2 registers/words in following order:
// a) AE41 5652 - higher (leftmost) 16 bits (high word) in the first register and the remaining low word in the second (AE41 before 5652)
// b) 5652 AE41 - low word in the first register and high word in the second (5652 before AE41)
//
// Ordered in memory (vertical table):
// | Memory | Big E | Little E | BE Low Word First | LE Low Word First |
// | byte 0 | AE    | 52       | 56                | 41                |
// | byte 1 | 41    | 56       | 52                | AE                |
// | byte 2 | 56    | 41       | AE                | 52                |
// | byte 3 | 52    | AE       | 41                | 56                |
//
// Ordered in memory (horizontal table):
// |  0 1  2 3 | Byte order      | Word order      | Name                            |
// | AE41 5652 | high byte first | high word first | big endian (high word first)    |
// | 5652 AE41 | high byte first | low word first  | big endian (low word first)     |
// | 41AE 5256 | low byte first  | high word first | little endian (low word first)  |
// | 5256 41AE | low byte first  | low word first  | little endian (high word first) |
//
// Example:
// Our PLC (modbus serving) controller/computer is using little endian
//
// 32bit (4 byte) integer 67305985 is in hex 0x01020304 (little endian), most significant byte is 01 and the
// lowest byte contain hex value 04.
// Source: http://unixpapa.com/incnote/byteorder.html
//
// 32bit (dword) integer is in:
//
//	Little Endian (ABCD) = 0x01020304  (0x04 + (0x03 << 8) + (0x02 << 16) + (0x01 << 24))
//
// May be sent over tcp/udp as:
//
//	Big Endian (DCBA) = 0x04030201
//	Big Endian Low Word First (BADC) = 0x02010403 <-- used by WAGO 750-XXX to send modbus packets over tcp/udp
const (
	useDefaultByteOrder ByteOrder = 0
	// BigEndian system stores the most significant byte of a word at the smallest memory address and the least
	// significant byte at the largest. By Modbus spec BigEndian is the order how bytes are transferred over the wire.
	BigEndian ByteOrder = 1
	// LittleEndian - little-endian system stores the least-significant byte at the smallest address.
	LittleEndian ByteOrder = 2

	// Double words (word=register) (32bit types) consist of two 16bit words. Different PLCs send double words
	// differently over the wire. So 0xDCBA can be sent low word (0xBA) first 0xBADC or high word (0xDC) first 0xDCBA.
	LowWordFirst ByteOrder = 4

	// HighWordFirst reads data as words/register are ordered from left to right. High word (0xDC) is sent first.
	// Meaning PLCs little endian value 0xABCD is sent as each byte swapped and each 2 byte pair (word/register) is swapped 0xDCBA
	HighWordFirst ByteOrder = 8

	// When bytes for little endian are in 'ABCD' order then Big Endian Low Word First is in 'BADC' order
	// This mean that high word (BA) is first and low word (DC) for double word is last and bytes in words are in big endian order.
	BigEndianLowWordFirst = BigEndian | LowWordFirst // this is default endian+word order we use

	// BigEndianHighWordFirst is big-endian with high word first
	BigEndianHighWordFirst = BigEndian | HighWordFirst

	// LittleEndianLowWordFirst is little-endian with low word first
	LittleEndianLowWordFirst = LittleEndian | LowWordFirst
	// LittleEndianHighWordFirst is little-endian with high word first
	LittleEndianHighWordFirst = LittleEndian | HighWordFirst
)

// ByteOrder determines how bytes are ordered in data
type ByteOrder uint8

// Registers provides more convenient access to data returned by register response
type Registers struct {
	defaultByteOrder ByteOrder
	startAddress     uint16
	endAddress       uint16 // end address is not addressable. endAddress-1 is last addressable register (2 bytes)
	data             []byte
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
		defaultByteOrder: BigEndianHighWordFirst,
		startAddress:     startAddress,
		endAddress:       startAddress + uint16(dataLen/2),
		data:             data,
	}, nil
}

// WithByteOrder sets byte order as default byte order in Registers
func (r *Registers) WithByteOrder(byteOrder ByteOrder) *Registers {
	r.defaultByteOrder = byteOrder
	return r
}

// Register returns single register data (16bit) from given address
func (r Registers) Register(address uint16) ([]byte, error) {
	b, err := r.register(address)
	if err != nil {
		return nil, err
	}
	return []byte{b[0], b[1]}, nil
}

func (r Registers) register(address uint16) ([]byte, error) {
	if address < r.startAddress {
		return nil, errors.New("address under startAddress bounds")
	}
	if address >= r.endAddress {
		return nil, errors.New("address over startAddress+quantity bounds")
	}
	startIndex := (address - r.startAddress) * 2
	return r.data[startIndex : startIndex+2], nil
}

// DoubleRegister returns two registers data (32bit) from starting from given address using word/register order
func (r Registers) DoubleRegister(address uint16, byteOrder ByteOrder) ([]byte, error) {
	b, err := r.doubleRegister(address, byteOrder)
	if err != nil {
		return nil, err
	}
	return []byte{b[0], b[1], b[2], b[3]}, nil
}

func (r Registers) doubleRegister(address uint16, byteOrder ByteOrder) ([]byte, error) {
	if address < r.startAddress {
		return nil, errors.New("address under startAddress bounds")
	}
	if address > (r.endAddress - 2) {
		return nil, errors.New("address over startAddress+quantity bounds")
	}
	startIndex := (address - r.startAddress) * 2
	if byteOrder&LowWordFirst != 0 {
		// reverse words/registers order (low word first)
		return []byte{
			r.data[startIndex+2],
			r.data[startIndex+3],

			r.data[startIndex],
			r.data[startIndex+1],
		}, nil
	}
	return r.data[startIndex : startIndex+4], nil
}

// QuadRegister returns four registers data (64bit) from starting from given address using word/register order
func (r Registers) QuadRegister(address uint16, byteOrder ByteOrder) ([]byte, error) {
	b, err := r.quadRegister(address, byteOrder)
	if err != nil {
		return nil, err
	}
	return []byte{b[0], b[1], b[2], b[3], b[4], b[5], b[6], b[7]}, nil
}

func (r Registers) quadRegister(address uint16, byteOrder ByteOrder) ([]byte, error) {
	if address < r.startAddress {
		return nil, errors.New("address under startAddress bounds")
	}
	if address > (r.endAddress - 4) {
		return nil, errors.New("address over startAddress+quantity bounds")
	}
	startIndex := (address - r.startAddress) * 2
	if byteOrder&LowWordFirst != 0 {
		// reverse words/registers order (low word first)
		return []byte{
			r.data[startIndex+6],
			r.data[startIndex+7],

			r.data[startIndex+4],
			r.data[startIndex+5],

			r.data[startIndex+2],
			r.data[startIndex+3],

			r.data[startIndex],
			r.data[startIndex+1],
		}, nil
	}
	return r.data[startIndex : startIndex+8], nil
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
	if r.defaultByteOrder&LittleEndian != 0 {
		return binary.LittleEndian.Uint16(b), nil
	}
	return binary.BigEndian.Uint16(b), nil
}

// Int16 returns register data as int16 from given address. NB: Int16 size is 1 register (16bits, 2 bytes).
func (r Registers) Int16(address uint16) (int16, error) {
	b, err := r.register(address)
	if err != nil {
		return 0, err
	}
	if r.defaultByteOrder&LittleEndian != 0 {
		return int16(binary.LittleEndian.Uint16(b)), nil
	}
	return int16(binary.BigEndian.Uint16(b)), nil
}

// Uint32 returns register data as uint32 from given address. NB: Uint32 size is 2 registers (32bits, 4 bytes).
func (r Registers) Uint32(address uint16) (uint32, error) {
	b, err := r.doubleRegister(address, r.defaultByteOrder)
	if err != nil {
		return 0, err
	}
	if r.defaultByteOrder&LittleEndian != 0 {
		return binary.LittleEndian.Uint32(b), nil
	}
	return binary.BigEndian.Uint32(b), nil
}

// Uint32WithByteOrder returns register data as uint32 from given address with given byte order. NB: uint32 size is 2 registers (32bits, 4 bytes).
func (r Registers) Uint32WithByteOrder(address uint16, byteOrder ByteOrder) (uint32, error) {
	if byteOrder == useDefaultByteOrder {
		byteOrder = r.defaultByteOrder
	}
	b, err := r.doubleRegister(address, byteOrder)
	if err != nil {
		return 0, err
	}
	if byteOrder&LittleEndian != 0 {
		return binary.LittleEndian.Uint32(b), nil
	}
	return binary.BigEndian.Uint32(b), nil
}

// Int32 returns register data as int32 from given address. NB: Int32 size is 2 registers (32bits, 4 bytes).
func (r Registers) Int32(address uint16) (int32, error) {
	b, err := r.doubleRegister(address, r.defaultByteOrder)
	if err != nil {
		return 0, err
	}
	if r.defaultByteOrder&LittleEndian != 0 {
		return int32(binary.LittleEndian.Uint32(b)), nil
	}
	return int32(binary.BigEndian.Uint32(b)), nil
}

// Int32WithByteOrder returns register data as int32 from given address with given byte order. NB: int32 size is 2 registers (32bits, 4 bytes).
func (r Registers) Int32WithByteOrder(address uint16, byteOrder ByteOrder) (int32, error) {
	if byteOrder == useDefaultByteOrder {
		byteOrder = r.defaultByteOrder
	}
	b, err := r.doubleRegister(address, byteOrder)
	if err != nil {
		return 0, err
	}
	if byteOrder&LittleEndian != 0 {
		return int32(binary.LittleEndian.Uint32(b)), nil
	}
	return int32(binary.BigEndian.Uint32(b)), nil
}

// Uint64 returns register data as uint64 from given address. NB: Uint64 size is 4 registers (64bits, 8 bytes).
func (r Registers) Uint64(address uint16) (uint64, error) {
	b, err := r.quadRegister(address, r.defaultByteOrder)
	if err != nil {
		return 0, err
	}
	if r.defaultByteOrder&LittleEndian != 0 {
		return binary.LittleEndian.Uint64(b), nil
	}
	return binary.BigEndian.Uint64(b), nil
}

// Uint64WithByteOrder returns register data as uint64 from given address with given byte order. NB: uint64 size is 4 registers (64bits, 8 bytes).
func (r Registers) Uint64WithByteOrder(address uint16, byteOrder ByteOrder) (uint64, error) {
	if byteOrder == useDefaultByteOrder {
		byteOrder = r.defaultByteOrder
	}
	b, err := r.quadRegister(address, byteOrder)
	if err != nil {
		return 0, err
	}
	if byteOrder&LittleEndian != 0 {
		return binary.LittleEndian.Uint64(b), nil
	}
	return binary.BigEndian.Uint64(b), nil
}

// Int64 returns register data as int64 from given address. NB: Int64 size is 4 registers (64bits, 8 bytes).
func (r Registers) Int64(address uint16) (int64, error) {
	b, err := r.quadRegister(address, r.defaultByteOrder)
	if err != nil {
		return 0, err
	}
	if r.defaultByteOrder&LittleEndian != 0 {
		return int64(binary.LittleEndian.Uint64(b)), nil
	}
	return int64(binary.BigEndian.Uint64(b)), nil
}

// Int64WithByteOrder returns register data as int64 from given address with given byte order. NB: int64 size is 4 registers (64bits, 8 bytes).
func (r Registers) Int64WithByteOrder(address uint16, byteOrder ByteOrder) (int64, error) {
	if byteOrder == useDefaultByteOrder {
		byteOrder = r.defaultByteOrder
	}
	b, err := r.quadRegister(address, byteOrder)
	if err != nil {
		return 0, err
	}
	if byteOrder&LittleEndian != 0 {
		return int64(binary.LittleEndian.Uint64(b)), nil
	}
	return int64(binary.BigEndian.Uint64(b)), nil
}

// Float32 returns register data as float32 from given address. NB: Float32 size is 2 registers (32bits, 4 bytes).
func (r Registers) Float32(address uint16) (float32, error) {
	b, err := r.doubleRegister(address, r.defaultByteOrder)
	if err != nil {
		return 0, err
	}
	var u uint32
	if r.defaultByteOrder&LittleEndian != 0 {
		u = binary.LittleEndian.Uint32(b)
	} else {
		u = binary.BigEndian.Uint32(b)
	}
	return math.Float32frombits(u), nil
}

// Float32WithByteOrder returns register data as float32 from given address with given byte order. NB: float32 size is 2 registers (32bits, 4 bytes).
func (r Registers) Float32WithByteOrder(address uint16, byteOrder ByteOrder) (float32, error) {
	if byteOrder == useDefaultByteOrder {
		byteOrder = r.defaultByteOrder
	}
	b, err := r.doubleRegister(address, byteOrder)
	if err != nil {
		return 0, err
	}
	var u uint32
	if byteOrder&LittleEndian != 0 {
		u = binary.LittleEndian.Uint32(b)
	} else {
		u = binary.BigEndian.Uint32(b)
	}
	return math.Float32frombits(u), nil
}

// Float64 returns register data as float64 from given address. NB: Float64 size is 4 registers (64bits, 8 bytes).
func (r Registers) Float64(address uint16) (float64, error) {
	b, err := r.quadRegister(address, r.defaultByteOrder)
	if err != nil {
		return 0, err
	}
	var u uint64
	if r.defaultByteOrder&LittleEndian != 0 {
		u = binary.LittleEndian.Uint64(b)
	} else {
		u = binary.BigEndian.Uint64(b)
	}
	return math.Float64frombits(u), nil
}

// Float64WithByteOrder returns register data as float64 from given address with given byte order. NB: Float64 size is 4 registers (64bits, 8 bytes).
func (r Registers) Float64WithByteOrder(address uint16, byteOrder ByteOrder) (float64, error) {
	if byteOrder == useDefaultByteOrder {
		byteOrder = r.defaultByteOrder
	}
	b, err := r.quadRegister(address, byteOrder)
	if err != nil {
		return 0, err
	}
	var u uint64
	if byteOrder&LittleEndian != 0 {
		u = binary.LittleEndian.Uint64(b)
	} else {
		u = binary.BigEndian.Uint64(b)
	}
	return math.Float64frombits(u), nil
}

// String returns register data as string starting from given address to given length.
// Data is interpreted as ASCII 0x0 (null) terminated string.
func (r Registers) String(address uint16, length uint8) (string, error) {
	return r.StringWithByteOrder(address, length, useDefaultByteOrder)
}

// StringWithByteOrder returns register data as string starting from given address to given length and byte order.
// Data is interpreted as ASCII 0x0 (null) terminated string.
func (r Registers) StringWithByteOrder(address uint16, length uint8, byteOrder ByteOrder) (string, error) {
	rawBytes, err := r.BytesWithByteOrder(address, length, byteOrder)
	if err != nil {
		return "", err
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

// Bytes returns register data as byte slice starting from given address to given length in bytes.
func (r Registers) Bytes(address uint16, length uint8) ([]byte, error) {
	return r.BytesWithByteOrder(address, length, useDefaultByteOrder)
}

// BytesWithByteOrder returns register data as byte slice starting from given address to given length in bytes and byte order.
func (r Registers) BytesWithByteOrder(address uint16, length uint8, wantByteOrder ByteOrder) ([]byte, error) {
	if wantByteOrder == useDefaultByteOrder {
		wantByteOrder = r.defaultByteOrder
	}
	if address < r.startAddress {
		return nil, errors.New("address under startAddress bounds")
	}
	startIndex := (address - r.startAddress) * 2
	endIndex := startIndex + uint16(length)
	// length is bytes. but data is sent in registers (2 bytes) and in big endian format. so last character for odd size
	// needs 1 more byte (it needs to be swapped)
	isOddSize := length%2 != 0
	neededLength := length
	if isOddSize {
		neededLength++
		endIndex++
	}
	if int(endIndex) > len(r.data) {
		return nil, errors.New("address over data bounds")
	}

	// TODO: clean these loops up to single for loop

	rawBytes := make([]byte, neededLength)
	copy(rawBytes, r.data[startIndex:endIndex])

	// on the wire, modbus data is considered assumed to be in big endian order
	// when we want to interpret dat as Little endian we need to switch bytes in each register
	if wantByteOrder&LittleEndian != 0 {
		for i := 1; i < len(rawBytes); i++ {
			// data is in BIG ENDIAN format in register (register is 2 bytes). so every 2 bytes needs to have their bytes swapped
			// to get little endian order
			if i%2 != 0 {
				previous := rawBytes[i-1]
				rawBytes[i-1] = rawBytes[i]
				rawBytes[i] = previous
			}
		}
	}
	if isOddSize {
		return rawBytes[0:length], nil
	}
	return rawBytes, nil
}

// IsEqualBytes checks if data at given address, to given length, is equal to given bytes
// Equality check is done against raw data from request which is in Big Endian format
func (r Registers) IsEqualBytes(registerAddress uint16, addressLengthInBytes uint8, bytes []byte) (bool, error) {
	if registerAddress < r.startAddress {
		return false, errors.New("address under startAddress bounds")
	}
	startIndex := (registerAddress - r.startAddress) * 2

	l := int(addressLengthInBytes)
	if len(bytes) < l {
		l = len(bytes)
	}
	endIndex := startIndex + uint16(l)
	if int(endIndex) > len(r.data) {
		return false, errors.New("address+length over data bounds")
	}
	data := r.data[startIndex:endIndex]
	for i := 0; i < l; i++ {
		if bytes[i] != data[i] {
			return false, nil
		}
	}
	return true, nil
}
