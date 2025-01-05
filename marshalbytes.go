package modbus

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
)

// to help determine how long int is: 32 or 64bits
const bytesPerWord = int((32 << (^uint(0) >> 63)) / 8)

type valueType uint8

const (
	valueUnknown valueType = iota
	valueInteger
	valueFloat
	valueStringOrByte
)

func size(v any) (sizeBytes int, vType valueType, err error) {
	switch t := v.(type) {
	case bool, uint8, int8: // byte == uint8
		return 1, valueInteger, nil
	case uint16, int16:
		return 2, valueInteger, nil
	case uint32, int32:
		return 4, valueInteger, nil
	case uint64, int64:
		return 8, valueInteger, nil
	case int, uint: // could be 32bit or 64bit
		return bytesPerWord, valueInteger, nil
	case float32:
		return 4, valueFloat, nil
	case float64:
		return 8, valueFloat, nil
	case string:
		return len(t), valueStringOrByte, nil
	case []byte:
		return len(t), valueStringOrByte, nil
	default:
		return 0, valueUnknown, errors.New("can not marshal unsupported type")
	}
}

func marshalFieldTypeByte(dst []byte, value any, toHighByte bool) error {
	if len(dst) < 2 {
		return errors.New("field type byte or uint8 requires at least 2 bytes")
	}
	// dst[1] because for big endian first byte is 1 in little endian
	idx := 1
	if toHighByte {
		idx = 0
	}
	switch v := value.(type) {
	case bool:
		if v {
			dst[idx] = 1
		}
	case uint8: // byte == uint8
		dst[idx] = v
	case int8:
		dst[idx] = byte(v)
	case uint16:
		dst[idx] = byte(limitUnsigned(v, math.MaxUint8))
	case int16:
		dst[idx] = byte(limitSigned(v, math.MaxUint8, 0))
	case uint32:
		dst[idx] = byte(limitUnsigned(v, math.MaxUint8))
	case int32:
		dst[idx] = byte(limitSigned(v, math.MaxUint8, 0))
	case uint64:
		dst[idx] = byte(limitUnsigned(v, math.MaxUint8))
	case int64:
		dst[idx] = byte(limitSigned(v, math.MaxUint8, 0))
	case int: // could be 32bit or 64bit
		dst[idx] = byte(limitSigned(v, math.MaxUint8, 0))
	case uint: // could be 32bit or 64bit
		dst[idx] = byte(limitUnsigned(v, math.MaxUint8))
	case float32:
		dst[idx] = byte(limitFloat32(v, math.MaxUint8, 0))
	case float64:
		dst[idx] = byte(limitFloat64(v, math.MaxUint8, 0))
	case []byte:
		dst[idx] = v[0]
	default: // string
		return errors.New("can not marshal unsupported type")
	}
	return nil
}

func marshalFieldTypeInt8(dst []byte, value any, toHighByte bool) error {
	if len(dst) < 2 {
		return errors.New("field type byte or int8 requires at least 2 bytes")
	}
	// dst[1] because for big endian first byte is 1 in little endian
	idx := 1
	if toHighByte {
		idx = 0
	}
	switch v := value.(type) {
	case bool:
		if v {
			dst[idx] = 1
		}
	case uint8:
		dst[idx] = limitUnsigned(v, math.MaxInt8)
	case int8:
		dst[idx] = byte(v)
	case uint16:
		dst[idx] = byte(limitUnsigned(v, math.MaxInt8))
	case int16:
		dst[idx] = byte(limitSigned(v, math.MaxInt8, math.MinInt8))
	case uint32:
		dst[idx] = byte(limitUnsigned(v, math.MaxInt8))
	case int32:
		dst[idx] = byte(limitSigned(v, math.MaxInt8, math.MinInt8))
	case uint64:
		dst[idx] = byte(limitUnsigned(v, math.MaxInt8))
	case int64:
		dst[idx] = byte(limitSigned(v, math.MaxInt8, math.MinInt8))
	case int: // could be 32bit or 64bit
		dst[idx] = byte(limitSigned(v, math.MaxInt8, math.MinInt8))
	case uint: // could be 32bit or 64bit
		dst[idx] = byte(limitUnsigned(v, math.MaxInt8))
	case float32:
		dst[idx] = byte(limitFloat32(v, math.MaxInt8, math.MinInt8))
	case float64:
		dst[idx] = byte(limitFloat64(v, math.MaxInt8, math.MinInt8))
	default: // including []byte, string
		return errors.New("can not marshal unsupported type")
	}
	return nil
}

func marshalFieldTypeUint16(dst []byte, value any) error {
	if len(dst) < 2 {
		return errors.New("field type byte or uint16 requires at least 2 bytes")
	}
	var tmp uint16
	switch v := value.(type) {
	case bool:
		if v {
			tmp = 1
		}
	case uint8:
		tmp = uint16(v)
	case int8:
		tmp = uint16(limitSigned(v, math.MaxInt16, 0))
	case uint16:
		tmp = v
	case int16:
		tmp = uint16(limitSigned(v, math.MaxInt16, 0))
	case uint32:
		tmp = uint16(limitUnsigned(v, math.MaxUint16))
	case int32:
		tmp = uint16(limitSigned(v, math.MaxInt16, 0))
	case uint64:
		tmp = uint16(limitUnsigned(v, math.MaxUint16))
	case int64:
		tmp = uint16(limitSigned(v, math.MaxInt16, 0))
	case int: // could be 32bit or 64bit
		tmp = uint16(limitSigned(v, math.MaxInt16, 0))
	case uint: // could be 32bit or 64bit
		tmp = uint16(limitUnsigned(v, math.MaxUint16))
	case float32:
		tmp = uint16(limitFloat32(v, math.MaxInt16, 0))
	case float64:
		tmp = uint16(limitFloat64(v, math.MaxInt16, 0))
	default: // including []byte, string
		return errors.New("can not marshal unsupported type")
	}
	binary.BigEndian.PutUint16(dst, tmp)
	return nil
}

func marshalFieldTypeInt16(dst []byte, value any) error {
	if len(dst) < 2 {
		return errors.New("field type byte or uint16 requires at least 2 bytes")
	}
	var tmp int16
	switch v := value.(type) {
	case bool:
		if v {
			tmp = 1
		}
	case uint8:
		tmp = int16(v)
	case int8:
		tmp = int16(v)
	case uint16:
		tmp = int16(limitUnsigned(v, math.MaxInt16))
	case int16:
		tmp = v
	case uint32:
		tmp = int16(limitUnsigned(v, math.MaxInt16))
	case int32:
		tmp = int16(limitSigned(v, math.MaxInt16, math.MinInt16))
	case uint64:
		tmp = int16(limitUnsigned(v, math.MaxInt16))
	case int64:
		tmp = int16(limitSigned(v, math.MaxInt16, math.MinInt16))
	case int: // could be 32bit or 64bit
		tmp = int16(limitSigned(v, math.MaxInt16, math.MinInt16))
	case uint: // could be 32bit or 64bit
		tmp = int16(limitUnsigned(v, math.MaxInt16))
	case float32:
		tmp = int16(limitFloat32(v, math.MaxInt16, math.MinInt16))
	case float64:
		tmp = int16(limitFloat64(v, math.MaxInt16, math.MinInt16))
	default: // including []byte, string
		return errors.New("can not marshal unsupported type")
	}
	binary.BigEndian.PutUint16(dst, uint16(tmp))
	return nil
}

func limitSigned[T ~int8 | ~int16 | ~int32 | ~int64 | ~int](v T, max T, min T) T {
	if v > max {
		return max
	} else if v <= min {
		return min
	} else {
		return v
	}
}

func limitUnsigned[T ~byte | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uint](v T, max T) T {
	if v > max {
		return max
	}
	return v
}

func limitFloat32[T ~float32](v T, max T, min T) T {
	v = T(math.Round(float64(v)))
	if v > max {
		return max
	} else if v <= min {
		return min
	} else {
		return v
	}
}

func limitFloat64[T ~float64](v T, max T, min T) T {
	v = T(math.Round(float64(v)))
	if v > max {
		return max
	} else if v <= min {
		return min
	} else {
		return v
	}
}

func roundFloatToInteger(value any, signed bool) (any, error) {
	result := value
	switch v := value.(type) {
	case float32:
		round := math.Round(float64(v))
		if signed {
			result = int32(round)
		} else {
			result = uint32(round)
		}
	case float64:
		round := math.Round(v)
		if signed {
			result = int64(round)
		} else {
			result = uint64(round)
		}
	default:
		return nil, errors.New("unsupported type passed to roundFloatToInteger")
	}
	return result, nil
}

func integerToFloat(value any) (any, error) {
	result := value
	switch v := value.(type) {
	case bool:
		if v {
			result = float32(1)
		} else {
			result = float32(0)
		}
	case uint8:
		result = float32(v)
	case int8:
		result = float32(v)
	case uint16:
		result = float32(v)
	case int16:
		result = float32(v)
	case uint32:
		result = float32(v)
	case int32:
		result = float32(v)
	case uint64:
		result = float64(v)
	case int64:
		result = float64(v)
	case uint:
		if bytesPerWord == 4 {
			result = float32(v)
		} else {
			result = float64(v)
		}
	case int:
		if bytesPerWord == 4 {
			result = float32(v)
		} else {
			result = float64(v)
		}
	default:
		return nil, errors.New("unsupported type passed to integerToFloat")
	}
	return result, nil
}

func truncateNumber(value any, requiredByteSize int, signed bool) (any, error) {
	result := value
	switch v := value.(type) {
	case int:
		if bytesPerWord == 4 {
			result = int32(v)
		} else {
			result = int64(v)
		}
	case uint:
		if bytesPerWord == 4 {
			result = uint32(v)
		} else {
			result = uint64(v)
		}
	}

	switch v := value.(type) {
	case uint32:
		switch requiredByteSize {
		case 1:
			if v > math.MaxUint8 {
				return uint8(math.MaxUint8), nil
			}
			return uint8(v), nil
		case 2:
			if v > math.MaxUint16 {
				return uint16(math.MaxUint16), nil
			}
			return uint16(v), nil
		}
	case int32:
		switch requiredByteSize {
		case 1:
			if v > math.MaxInt8 {
				return int8(math.MaxInt8), nil
			} else if v < math.MinInt8 {
				return int8(math.MinInt8), nil
			}
			return int8(v), nil
		case 2:
			if v > math.MaxInt16 {
				return int16(math.MaxInt16), nil
			} else if v < math.MinInt16 {
				return int16(math.MinInt16), nil
			}
			return int16(v), nil
		}
	case uint64:
		switch requiredByteSize {
		case 1:
			if v > math.MaxUint8 {
				return uint8(math.MaxUint8), nil
			}
			return uint8(v), nil
		case 2:
			if v > math.MaxUint16 {
				return uint16(math.MaxUint16), nil
			}
			return uint16(v), nil
		case 4:
			if v > math.MaxUint32 {
				return uint32(math.MaxUint32), nil
			}
			return uint32(v), nil
		}
	case int64:
		switch requiredByteSize {
		case 1:
			if signed {
				if v > math.MaxInt8 {
					return int8(math.MaxInt8), nil
				} else if v < math.MinInt8 {
					return int8(math.MinInt8), nil
				}
				return int8(v), nil
			} else {
				if v > math.MaxUint8 {
					return uint8(math.MaxUint8), nil
				} else if v < 0 {
					return uint8(0), nil
				}
				return uint8(v), nil
			}
		case 2:
			if v > math.MaxInt16 {
				return int16(math.MaxInt16), nil
			} else if v < math.MinInt16 {
				return int16(math.MinInt16), nil
			}
			return uint16(v), nil
		case 4:
			if v > math.MaxInt32 {
				return int32(math.MaxInt32), nil
			} else if v < math.MinInt32 {
				return int32(math.MinInt32), nil
			}
			return uint32(v), nil
		}
	case float64:
		switch requiredByteSize {
		case 1, 2:
			return nil, errors.New("can not truncate float64 to single register")
		case 4:
			if v > math.MaxFloat32 {
				return float32(math.MaxFloat32), nil
			} else if v < -math.MaxFloat32 {
				return float32(-math.MaxFloat32), nil
			}
			return float32(v), nil
		}
	default:
		return nil, errors.New("can not truncate unsupported type")
	}
	return result, nil
}

// marshalNumberBytes marshal value into byte slice as big endian byte order
func marshalNumberBytes(dst []byte, value any) error {
	valSize, valType, err := size(value)
	if err != nil {
		return err
	}
	if valType != valueInteger && valType != valueFloat {
		return errors.New("given value is not a number")
	}
	switch valSize {
	case 1, 2:
		return marshal16BitRegisters(dst, value)
	case 4:
		return marshal32BitRegisters(dst, value)
	case 8:
		return marshal64BitRegisters(dst, value)
	}
	// should not be possible
	return fmt.Errorf("unsupported size number seen, size: %d", valSize)
}

func marshal16BitRegisters(dst []byte, value any) error {
	var tmp uint16
	switch v := value.(type) {
	case bool:
		if v {
			tmp = 1
		}
	case uint8:
		tmp = uint16(v)
	case int8:
		tmp = uint16(v)
	case uint16:
		tmp = v
	case int16:
		tmp = uint16(v)
	default:
		return fmt.Errorf("unknown type given to marshal16BitRegisters")
	}
	binary.BigEndian.PutUint16(dst, tmp)
	return nil
}

func marshal32BitRegisters(dst []byte, value any) error {
	var tmp uint32
	switch v := value.(type) {
	case uint32:
		tmp = v
	case int32:
		tmp = uint32(v)
	case int:
		tmp = uint32(v)
	case float32:
		tmp = math.Float32bits(v)
	default:
		return fmt.Errorf("unknown type given to marshal32BitRegisters")
	}
	binary.BigEndian.PutUint32(dst, tmp)
	return nil
}

func marshal64BitRegisters(dst []byte, value any) error {
	var tmp uint64
	switch v := value.(type) {
	case uint64:
		tmp = v
	case int64:
		tmp = uint64(v)
	case int:
		tmp = uint64(v)
	case float64:
		tmp = math.Float64bits(v)
	default:
		return fmt.Errorf("unknown type given to marshal64BitRegisters")
	}
	binary.BigEndian.PutUint64(dst, tmp)
	return nil
}

func (f *Field) marshalBitToRegister(dst []byte, value any) error {
	endIdx := len(dst) - 1
	if endIdx < 1 {
		return errors.New("can not marshal bit to registers: too few bytes")
	}
	switch val := value.(type) {
	case bool:
		if val == false {
			return nil
		}
	case uint8:
		if val == 0 {
			return nil
		}
	case int8:
		if val == 0 {
			return nil
		}
	case uint16:
		if val == 0 {
			return nil
		}
	case int16:
		if val == 0 {
			return nil
		}
	case uint32:
		if val == 0 {
			return nil
		}
	case int32:
		if val == 0 {
			return nil
		}
	case uint64:
		if val == 0 {
			return nil
		}
	case int64:
		if val == 0 {
			return nil
		}
	case int: // 32bit and 64bits are fine for FieldTypeBit
		if val == 0 {
			return nil
		}
	case uint: // 32bit and 64bits are fine for FieldTypeBit
		if val == 0 {
			return nil
		}
	case float32:
		if val == 0 {
			return nil
		}
	case float64:
		if val == 0 {
			return nil
		}
	default:
		return errors.New("unsupported value type for field with bit type")
	}

	bit := f.Bit
	if bit > 7 { // high byte of register (big endian register)
		dst[endIdx-1] = 1 << (bit - 8)
		dst[endIdx] = 0
		return nil
	}
	// low byte of register (big endian register)
	dst[endIdx-1] = 0
	dst[endIdx] = 1 << bit
	return nil
}

func marshalStringOrByteRegisters(dst []byte, value any) error {
	var tmp []byte
	switch v := value.(type) {
	case string:
		if v == "" {
			return nil
		}
		tmp = []byte(v)
	case []byte:
		if len(v) == 0 {
			return nil
		}
		tmp = v
	default:
		return fmt.Errorf("unknown type given to marshalStringOrByteRegisters")
	}

	if len(tmp) > len(dst) {
		tmp = tmp[:len(dst)]
	} else if len(tmp) < len(dst) {
		dst = dst[len(dst)-len(tmp):]
	}

	// reverse (little endian) bytes to get big endian order
	for i := 0; i < len(dst); i++ {
		dst[i] = tmp[len(tmp)-1-i]
	}
	return nil
}

func registersToLowWordFirst(target []byte) error {
	sizeBytes := len(target)
	if sizeBytes%2 != 0 {
		return fmt.Errorf("registersToLowWordFirst: target size must be even bytes")
	}
	if sizeBytes == 2 {
		return nil // do nothing
	}

	for i, j := 0, sizeBytes-2; i < j; {
		target[i], target[j] = target[j], target[i]
		target[i+1], target[j+1] = target[j+1], target[i+1]

		i += 2
		j -= 2
	}
	return nil
}
