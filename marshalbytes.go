package modbus

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
)

// to help determine how long int is: 32bits = 4 bytes or 64bits = 8 bytes
const bytesPerWord = int((32 << (^uint(0) >> 63)) / 8)

func size(v any) (sizeBytes int, isNumber bool, err error) {
	switch t := v.(type) {
	case bool, uint8, int8: // byte == uint8
		return 1, true, nil
	case uint16, int16:
		return 2, true, nil
	case uint32, int32:
		return 4, true, nil
	case uint64, int64:
		return 8, true, nil
	case int, uint: // could be 32bit or 64bit
		return bytesPerWord, true, nil
	case float32:
		return 4, true, nil
	case float64:
		return 8, true, nil
	case string:
		return len(t), false, nil
	case []byte:
		return len(t), false, nil
	default:
		return 0, false, errors.New("can not marshal unsupported type")
	}
}

func marshalFieldTypeUint8(dst []byte, value any, toHighByte bool) error {
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
		return errors.New("marshalFieldTypeUint8: can not marshal unsupported type")
	}
	return nil
}

func marshalFieldTypeInt8(dst []byte, value any, toHighByte bool) error {
	if len(dst) < 2 {
		return errors.New("field type int8 requires at least 2 bytes")
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
		return errors.New("marshalFieldTypeInt8: can not marshal unsupported type")
	}
	return nil
}

func marshalFieldTypeUint16(dst []byte, value any) error {
	if len(dst) < 2 {
		return errors.New("field type uint16 requires at least 2 bytes")
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
		tmp = uint16(limitToPositive(v))
	case uint16:
		tmp = v
	case int16:
		tmp = uint16(limitToPositive(v))
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
		return errors.New("marshalFieldTypeUint16: can not marshal unsupported type")
	}
	binary.BigEndian.PutUint16(dst, tmp)
	return nil
}

func marshalFieldTypeInt16(dst []byte, value any) error {
	if len(dst) < 2 {
		return errors.New("field type int16 requires at least 2 bytes")
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
		tmp = int16(limitSigned(v, math.MaxInt16, math.MinInt16)) // #nosec G115
	case uint64:
		tmp = int16(limitUnsigned(v, math.MaxInt16))
	case int64:
		tmp = int16(limitSigned(v, math.MaxInt16, math.MinInt16)) // #nosec G115
	case int: // could be 32bit or 64bit
		tmp = int16(limitSigned(v, math.MaxInt16, math.MinInt16)) // #nosec G115
	case uint: // could be 32bit or 64bit
		tmp = int16(limitUnsigned(v, math.MaxInt16))
	case float32:
		tmp = int16(limitFloat32(v, math.MaxInt16, math.MinInt16)) // #nosec G115
	case float64:
		tmp = int16(limitFloat64(v, math.MaxInt16, math.MinInt16)) // #nosec G115
	default: // including []byte, string
		return errors.New("marshalFieldTypeInt16: can not marshal unsupported type")
	}
	binary.BigEndian.PutUint16(dst, uint16(tmp))
	return nil
}

func marshalFieldTypeUint32(dst []byte, value any) error {
	if len(dst) < 4 {
		return errors.New("field type byte or uint32 requires at least 4 bytes")
	}
	var tmp uint32
	switch v := value.(type) {
	case bool:
		if v {
			tmp = 1
		}
	case uint8:
		tmp = uint32(v)
	case int8:
		tmp = uint32(limitToPositive(v))
	case uint16:
		tmp = uint32(v)
	case int16:
		tmp = uint32(limitToPositive(v))
	case uint32:
		tmp = v
	case int32:
		tmp = uint32(limitSigned(v, math.MaxInt32, 0))
	case uint64:
		tmp = uint32(limitUnsigned(v, math.MaxUint32))
	case int64:
		tmp = uint32(limitSigned(v, math.MaxInt32, 0))
	case int: // could be 32bit or 64bit
		tmp = uint32(limitSigned(v, math.MaxInt32, 0))
	case uint: // could be 32bit or 64bit
		tmp = uint32(limitUnsigned(v, math.MaxUint32))
	case float32:
		tmp = uint32(limitFloat32(v, math.MaxInt32, 0))
	case float64:
		tmp = uint32(limitFloat64(v, math.MaxInt32, 0))
	default: // including []byte, string
		return errors.New("marshalFieldTypeUint32: can not marshal unsupported type")
	}
	binary.BigEndian.PutUint32(dst, tmp)
	return nil
}

func marshalFieldTypeInt32(dst []byte, value any) error {
	if len(dst) < 4 {
		return errors.New("field type int32 requires at least 4 bytes")
	}
	var tmp int32
	switch v := value.(type) {
	case bool:
		if v {
			tmp = 1
		}
	case uint8:
		tmp = int32(v)
	case int8:
		tmp = int32(v)
	case uint16:
		tmp = int32(v)
	case int16:
		tmp = int32(v)
	case uint32:
		tmp = int32(limitUnsigned(v, math.MaxInt32))
	case int32:
		tmp = v
	case uint64:
		tmp = int32(limitUnsigned(v, math.MaxInt32))
	case int64:
		tmp = int32(limitSigned(v, math.MaxInt32, math.MinInt32))
	case int: // could be 32bit or 64bit
		tmp = int32(limitSigned(v, math.MaxInt32, math.MinInt32))
	case uint: // could be 32bit or 64bit
		tmp = int32(limitUnsigned(v, math.MaxInt32))
	case float32:
		tmp = int32(v)
	case float64:
		tmp = int32(limitFloat64(v, math.MaxInt32, math.MinInt32))
	default: // including []byte, string
		return errors.New("marshalFieldTypeInt32: can not marshal unsupported type")
	}
	binary.BigEndian.PutUint32(dst, uint32(tmp))
	return nil
}

func marshalFieldTypeUint64(dst []byte, value any) error {
	if len(dst) < 8 {
		return errors.New("field type byte or uint64 requires at least 8 bytes")
	}
	var tmp uint64
	switch v := value.(type) {
	case bool:
		if v {
			tmp = 1
		}
	case uint8:
		tmp = uint64(v)
	case int8:
		tmp = uint64(limitToPositive(v))
	case uint16:
		tmp = uint64(v)
	case int16:
		tmp = uint64(limitToPositive(v))
	case uint32:
		tmp = uint64(v)
	case int32:
		tmp = uint64(limitToPositive(v))
	case uint64:
		tmp = v
	case int64:
		tmp = uint64(limitToPositive(v))
	case int: // could be 32bit or 64bit
		tmp = uint64(limitToPositive(v))
	case uint: // could be 32bit or 64bit
		tmp = uint64(v)
	case float32:
		tmp = uint64(limitFloat32(v, math.MaxInt32, 0))
	case float64:
		tmp = uint64(limitFloat64(v, math.MaxInt64, 0))
	default: // including []byte, string
		return errors.New("marshalFieldTypeUint64: can not marshal unsupported type")
	}
	binary.BigEndian.PutUint64(dst, tmp)
	return nil
}

func marshalFieldTypeInt64(dst []byte, value any) error {
	if len(dst) < 8 {
		return errors.New("field type int64 requires at least 8 bytes")
	}
	var tmp int64
	switch v := value.(type) {
	case bool:
		if v {
			tmp = 1
		}
	case uint8:
		tmp = int64(v)
	case int8:
		tmp = int64(v)
	case uint16:
		tmp = int64(v)
	case int16:
		tmp = int64(v)
	case uint32:
		tmp = int64(v)
	case int32:
		tmp = int64(v)
	case uint64:
		tmp = int64(limitUnsigned(v, math.MaxInt64))
	case int64:
		tmp = v
	case int: // could be 32bit or 64bit
		tmp = int64(v)
	case uint: // could be 32bit or 64bit
		tmp = int64(v)
	case float32:
		tmp = int64(math.Round(float64(v)))
	case float64:
		tmp = int64(math.Round(v))
	default: // including []byte, string
		return errors.New("marshalFieldTypeInt64: can not marshal unsupported type")
	}
	binary.BigEndian.PutUint64(dst, uint64(tmp))
	return nil
}

var maxFloat32 = math.MaxFloat32
var maxFloat64 = math.MaxFloat64

func marshalFieldTypeFloat32(dst []byte, value any) error {
	// float32 can exactly represent whole uint32 values up to 16777216.
	// For uint32 values larger than 16777216, precision issues arise, and not all integers can be
	// represented exactly.
	if len(dst) < 4 {
		return errors.New("field type float32 requires at least 4 bytes")
	}
	var tmp float32
	switch v := value.(type) {
	case bool:
		if v {
			tmp = 1
		}
	case uint8:
		tmp = float32(v)
	case int8:
		tmp = float32(v)
	case uint16:
		tmp = float32(v)
	case int16:
		tmp = float32(v)
	case uint32:
		tmp = float32(v)
	case int32:
		tmp = float32(v)
	case uint64:
		tmp = float32(limitUnsigned(v, uint64(maxFloat32)))
	case int64:
		tmp = float32(limitSigned(v, int64(maxFloat32), int64(-maxFloat32)))
	case int: // could be 32bit or 64bit
		tmp = float32(limitSigned(v, int(maxFloat32), int(-maxFloat32)))
	case uint: // could be 32bit or 64bit
		tmp = float32(limitUnsigned(v, uint(maxFloat32)))
	case float32:
		tmp = v
	case float64:
		tmp = float32(limitFloat64(v, maxFloat32, -maxFloat32))
	default: // including []byte, string
		return errors.New("marshalFieldTypeFloat32: can not marshal unsupported type")
	}
	binary.BigEndian.PutUint32(dst, math.Float32bits(tmp))
	return nil
}

func marshalFieldTypeFloat64(dst []byte, value any) error {
	if len(dst) < 8 {
		return errors.New("field type float64 requires at least 8 bytes")
	}
	var tmp float64
	switch v := value.(type) {
	case bool:
		if v {
			tmp = 1
		}
	case uint8:
		tmp = float64(v)
	case int8:
		tmp = float64(v)
	case uint16:
		tmp = float64(v)
	case int16:
		tmp = float64(v)
	case uint32:
		tmp = float64(v)
	case int32:
		tmp = float64(v)
	case uint64:
		tmp = float64(limitUnsigned(v, uint64(maxFloat64)))
	case int64:
		tmp = float64(limitSigned(v, int64(maxFloat64), int64(-maxFloat64)))
	case int: // could be 32bit or 64bit
		tmp = float64(limitSigned(v, int(maxFloat64), int(-maxFloat64)))
	case uint: // could be 32bit or 64bit
		tmp = float64(limitUnsigned(v, uint(maxFloat64)))
	case float32:
		tmp = float64(v)
	case float64:
		tmp = v
	default: // including []byte, string
		return errors.New("marshalFieldTypeFloat64: can not marshal unsupported type")
	}
	binary.BigEndian.PutUint64(dst, math.Float64bits(tmp))
	return nil
}

func limitSigned[T ~int8 | ~int16 | ~int32 | ~int64 | ~int](v T, max T, min T) T {
	if v > max {
		return max
	} else if v <= min {
		return min
	}
	return v
}

func limitToPositive[T ~int8 | ~int16 | ~int32 | ~int64 | ~int](v T) T {
	if v <= 0 {
		return 0
	}
	return v
}

func limitUnsigned[T ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uint](v T, max T) T {
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
	}
	return v
}

func limitFloat64[T ~float64](v T, max T, min T) T {
	v = T(math.Round(float64(v)))
	if v > max {
		return max
	} else if v <= min {
		return min
	}
	return v
}

func marshalFieldTypeBit(dst []byte, value any, NthBit uint8) error {
	endIdx := len(dst) - 1
	if endIdx < 1 {
		return errors.New("field type bit requires at least 2 bytes")
	}
	switch val := value.(type) {
	case bool:
		if !val {
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
		return errors.New("marshalFieldTypeFloat64: can not marshal unsupported type")
	}

	bit := NthBit
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

func marshalFieldTypeStringOrBytes(dst []byte, value any) error {
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
		return fmt.Errorf("can not marshal number type to field with string or bytes type")
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
