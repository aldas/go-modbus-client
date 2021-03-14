package modbus

import "github.com/aldas/go-modbus-client/packet"

// EXPERIMENTAL EXPERIMENTAL EXPERIMENTAL EXPERIMENTAL EXPERIMENTAL EXPERIMENTAL EXPERIMENTAL
// EXPERIMENTAL EXPERIMENTAL EXPERIMENTAL EXPERIMENTAL EXPERIMENTAL EXPERIMENTAL EXPERIMENTAL
// EXPERIMENTAL EXPERIMENTAL EXPERIMENTAL EXPERIMENTAL EXPERIMENTAL EXPERIMENTAL EXPERIMENTAL
// EXPERIMENTAL EXPERIMENTAL EXPERIMENTAL EXPERIMENTAL EXPERIMENTAL EXPERIMENTAL EXPERIMENTAL
// EXPERIMENTAL EXPERIMENTAL EXPERIMENTAL EXPERIMENTAL EXPERIMENTAL EXPERIMENTAL EXPERIMENTAL

// Builder helps to group extractable field values of different types into modbus requests with minimal amount of separate requests produced
type Builder struct {
}

// Build combines fields into modbus requests
func (b *Builder) Build() ([]packet.Request, error) {
	// FIXME: sort and group (host:port+unitID+ optimized max amount of fields for max quantity) fields into packets
	return nil, nil
}

// Add adds field into Builder
func (b *Builder) Add(field Field) *Builder {
	return b
}

// Field creates Field with defaults passed from Builder
func (b *Builder) Field(address int64) Field {
	return Field{
		Address: address,
	}
}

// Int64 add int64 field to Builder to be requested and extracted
func (b *Builder) Int64(address int64) Field {
	return Field{
		Address: address,
	}
}

// Field is distinct field be requested and extracted from response
type Field struct {
	UnitID  uint8
	Address int64

	Type      uint8
	WordOrder uint8

	Name string
}

// SetName sets custom name for Field
func (f Field) SetName(name string) Field {
	f.Name = name
	return f
}
