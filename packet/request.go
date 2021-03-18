package packet

// Request is common interface of modbus request packets
type Request interface {
	// FunctionCode returns function code of this request
	FunctionCode() uint8
	// Bytes returns packet as bytes form
	Bytes() []byte
	// ExpectedResponseLength returns length of bytes that valid response to this request would be
	ExpectedResponseLength() int
}
