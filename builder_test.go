package modbus

import (
	"context"
	"errors"
	"github.com/aldas/go-modbus-client/modbustest"
	"github.com/aldas/go-modbus-client/packet"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestBuilder_ReadCoilsTCP(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	receivedChan := make(chan []byte, 1)
	handler := func(received []byte, bytesRead int) (response []byte, closeConnection bool) {
		receivedChan <- received
		resp := packet.ReadCoilsResponseTCP{
			MBAPHeader: packet.MBAPHeader{TransactionID: 123, ProtocolID: 0},
			ReadCoilsResponse: packet.ReadCoilsResponse{
				UnitID:          0,
				CoilsByteLength: 1,
				Data:            []byte{0xff},
			},
		}
		return resp.Bytes(), true
	}
	addr, err := modbustest.RunServerOnRandomPort(ctx, handler)
	if err != nil {
		t.Fatal(err)
	}

	b := NewRequestBuilder(addr, 1)

	reqs, err := b.Add(b.Coil(10).UnitID(0)).ReadCoilsTCP()
	assert.NoError(t, err)
	assert.Len(t, reqs, 1)

	client := NewTCPClient()
	err = client.Connect(context.Background(), addr)
	assert.NoError(t, err)

	request := reqs[0]
	resp, err := client.Do(context.Background(), request)
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	received := <-receivedChan
	assert.Equal(t, []byte{0, 0, 0, 6, 0, 1, 0, 0xa, 0, 1}, received[2:]) // trim transaction ID
}

func TestBuilder_ReadCoilsRTU(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	receivedChan := make(chan []byte, 1)
	handler := func(received []byte, bytesRead int) (response []byte, closeConnection bool) {
		receivedChan <- received
		resp := packet.ReadCoilsResponseRTU{
			ReadCoilsResponse: packet.ReadCoilsResponse{
				UnitID:          0,
				CoilsByteLength: 1,
				Data:            []byte{0xff},
			},
		}
		return resp.Bytes(), true
	}
	addr, err := modbustest.RunServerOnRandomPort(ctx, handler)
	if err != nil {
		t.Fatal(err)
	}

	b := NewRequestBuilder(addr, 1)

	reqs, err := b.Add(b.Coil(10).UnitID(0)).ReadCoilsRTU()
	assert.NoError(t, err)
	assert.Len(t, reqs, 1)

	client := NewRTUClient()
	err = client.Connect(context.Background(), addr)
	assert.NoError(t, err)

	request := reqs[0]
	resp, err := client.Do(context.Background(), request)
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	received := <-receivedChan
	assert.Equal(t, []byte{0x0, 0x1, 0x0, 0xa, 0x0, 0x1, 0xdc, 0x19}, received)
}

func TestBuilder_ReadDiscreteInputsTCP(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	receivedChan := make(chan []byte, 1)
	handler := func(received []byte, bytesRead int) (response []byte, closeConnection bool) {
		receivedChan <- received
		resp := packet.ReadCoilsResponseTCP{
			MBAPHeader: packet.MBAPHeader{TransactionID: 123, ProtocolID: 0},
			ReadCoilsResponse: packet.ReadCoilsResponse{
				UnitID:          0,
				CoilsByteLength: 1,
				Data:            []byte{0xff},
			},
		}
		return resp.Bytes(), true
	}
	addr, err := modbustest.RunServerOnRandomPort(ctx, handler)
	if err != nil {
		t.Fatal(err)
	}

	b := NewRequestBuilder(addr, 1)

	reqs, err := b.Add(b.Coil(10).UnitID(0)).ReadDiscreteInputsTCP()
	assert.NoError(t, err)
	assert.Len(t, reqs, 1)

	client := NewTCPClient()
	err = client.Connect(context.Background(), addr)
	assert.NoError(t, err)

	request := reqs[0]
	resp, err := client.Do(context.Background(), request)
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	received := <-receivedChan
	assert.Equal(t, []byte{0, 0, 0, 6, 0, 2, 0, 0xa, 0, 1}, received[2:]) // trim transaction ID
}

func TestBuilder_ReadDiscreteInputsRTU(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	receivedChan := make(chan []byte, 1)
	handler := func(received []byte, bytesRead int) (response []byte, closeConnection bool) {
		receivedChan <- received
		resp := packet.ReadCoilsResponseRTU{
			ReadCoilsResponse: packet.ReadCoilsResponse{
				UnitID:          0,
				CoilsByteLength: 1,
				Data:            []byte{0xff},
			},
		}
		return resp.Bytes(), true
	}
	addr, err := modbustest.RunServerOnRandomPort(ctx, handler)
	if err != nil {
		t.Fatal(err)
	}

	b := NewRequestBuilder(addr, 1)

	reqs, err := b.Add(b.Coil(10).UnitID(0)).ReadDiscreteInputsRTU()
	assert.NoError(t, err)
	assert.Len(t, reqs, 1)

	client := NewRTUClient()
	err = client.Connect(context.Background(), addr)
	assert.NoError(t, err)

	request := reqs[0]
	resp, err := client.Do(context.Background(), request)
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	received := <-receivedChan
	assert.Equal(t, []byte{0x0, 0x2, 0x0, 0xa, 0x0, 0x1, 0x98, 0x19}, received)
}

func TestBuilder_ReadHoldingRegistersTCP(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	receivedChan := make(chan []byte, 1)
	handler := func(received []byte, bytesRead int) (response []byte, closeConnection bool) {
		receivedChan <- received
		resp := packet.ReadHoldingRegistersResponseTCP{
			MBAPHeader: packet.MBAPHeader{TransactionID: 123, ProtocolID: 0},
			ReadHoldingRegistersResponse: packet.ReadHoldingRegistersResponse{
				UnitID:          0,
				RegisterByteLen: 2,
				Data:            []byte{0xca, 0xfe},
			},
		}
		return resp.Bytes(), true
	}
	addr, err := modbustest.RunServerOnRandomPort(ctx, handler)
	if err != nil {
		t.Fatal(err)
	}

	b := NewRequestBuilder(addr, 1)

	reqs, err := b.Add(b.Int64(18).UnitID(0)).ReadHoldingRegistersTCP()
	assert.NoError(t, err)
	assert.Len(t, reqs, 1)

	ctxReq, cancelReq := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancelReq()

	client := NewTCPClient()
	err = client.Connect(ctxReq, addr)
	assert.NoError(t, err)

	request := reqs[0]
	resp, err := client.Do(ctxReq, request)
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	select {
	case received := <-receivedChan:
		assert.Equal(t, []byte{0, 0, 0, 6, 0, 3, 0, 18, 0, 4}, received[2:]) // trim transaction ID
	default:
		t.Errorf("nothing received")
	}

}

func TestBuilder_ReadHoldingRegistersRTU(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	receivedChan := make(chan []byte, 1)
	handler := func(received []byte, bytesRead int) (response []byte, closeConnection bool) {
		receivedChan <- received
		resp := packet.ReadHoldingRegistersResponseRTU{
			ReadHoldingRegistersResponse: packet.ReadHoldingRegistersResponse{
				UnitID:          0,
				RegisterByteLen: 2,
				Data:            []byte{0xca, 0xfe},
			},
		}
		return resp.Bytes(), true
	}
	addr, err := modbustest.RunServerOnRandomPort(ctx, handler)
	if err != nil {
		t.Fatal(err)
	}

	b := NewRequestBuilder(addr, 1)

	reqs, err := b.Add(b.Int64(18).UnitID(0)).ReadHoldingRegistersRTU()
	assert.NoError(t, err)
	assert.Len(t, reqs, 1)

	client := NewRTUClient()
	err = client.Connect(context.Background(), addr)
	assert.NoError(t, err)

	request := reqs[0]
	resp, err := client.Do(context.Background(), request)
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	received := <-receivedChan
	assert.Equal(t, []byte{0x0, 0x3, 0x0, 0x12, 0x0, 0x4, 0xe5, 0xdd}, received)
}

func TestBuilder_ReadInputRegistersTCP(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	receivedChan := make(chan []byte, 1)
	handler := func(received []byte, bytesRead int) (response []byte, closeConnection bool) {
		receivedChan <- received
		resp := packet.ReadInputRegistersResponseTCP{
			MBAPHeader: packet.MBAPHeader{TransactionID: 123, ProtocolID: 0},
			ReadInputRegistersResponse: packet.ReadInputRegistersResponse{
				UnitID:          0,
				RegisterByteLen: 2,
				Data:            []byte{0xca, 0xfe},
			},
		}
		return resp.Bytes(), true
	}
	addr, err := modbustest.RunServerOnRandomPort(ctx, handler)
	if err != nil {
		t.Fatal(err)
	}

	b := NewRequestBuilder(addr, 1)

	reqs, err := b.Add(b.Int64(18).UnitID(0)).ReadInputRegistersTCP()
	assert.NoError(t, err)
	assert.Len(t, reqs, 1)

	client := NewTCPClient()
	err = client.Connect(context.Background(), addr)
	assert.NoError(t, err)

	request := reqs[0]
	resp, err := client.Do(context.Background(), request)
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	received := <-receivedChan
	assert.Equal(t, []byte{0, 0, 0, 6, 0, 4, 0, 18, 0, 4}, received[2:]) // trim transaction ID
}

func TestBuilder_ReadInputRegistersRTU(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	receivedChan := make(chan []byte, 1)
	handler := func(received []byte, bytesRead int) (response []byte, closeConnection bool) {
		receivedChan <- received
		resp := packet.ReadInputRegistersResponseRTU{
			ReadInputRegistersResponse: packet.ReadInputRegistersResponse{
				UnitID:          0,
				RegisterByteLen: 2,
				Data:            []byte{0xca, 0xfe},
			},
		}
		return resp.Bytes(), true
	}
	addr, err := modbustest.RunServerOnRandomPort(ctx, handler)
	if err != nil {
		t.Fatal(err)
	}

	b := NewRequestBuilder(addr, 1)

	reqs, err := b.Add(b.Int64(18).UnitID(0)).ReadInputRegistersRTU()
	assert.NoError(t, err)
	assert.Len(t, reqs, 1)

	client := NewRTUClient()
	err = client.Connect(context.Background(), addr)
	assert.NoError(t, err)

	request := reqs[0]
	resp, err := client.Do(context.Background(), request)
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	received := <-receivedChan
	assert.Equal(t, []byte{0x0, 0x4, 0x0, 0x12, 0x0, 0x4, 0x50, 0x1d}, received)
}

func TestField_ModbusAddress(t *testing.T) {
	given := &BField{}

	given.ServerAddress(":502")

	assert.Equal(t, ":502", given.Field.ServerAddress)
}

func TestField_UnitID(t *testing.T) {
	given := &BField{}

	given.UnitID(1)

	assert.Equal(t, uint8(1), given.Field.UnitID)
}

func TestField_ByteOrder(t *testing.T) {
	given := &BField{}

	given.ByteOrder(packet.BigEndian)

	assert.Equal(t, packet.BigEndian, given.Field.ByteOrder)
}

func TestField_Name(t *testing.T) {
	given := &BField{}

	given.Name("fire_alarm_do")

	assert.Equal(t, "fire_alarm_do", given.Field.Name)
}

func TestNewBuilder(t *testing.T) {
	b := NewRequestBuilder(":5020", 2)

	assert.Equal(t, ":5020", b.serverAddress)
	assert.Equal(t, uint8(2), b.unitID)
}

func TestBuilder_Add(t *testing.T) {
	b := NewRequestBuilder(":5020", 2)
	b.Add(&BField{Field{ServerAddress: "test", UnitID: 1}})

	assert.Equal(t, "test", b.fields[0].ServerAddress)
	assert.Equal(t, uint8(1), b.fields[0].UnitID)
}

func TestBuilder_Bit(t *testing.T) {
	b := NewRequestBuilder(":5020", 2)

	b.Add(b.Bit(256, 4).Name("fire_alarm_di"))

	expect := Field{
		ServerAddress: ":5020",
		UnitID:        2,
		Type:          FieldTypeBit,
		Address:       256,
		Bit:           4,
		Name:          "fire_alarm_di",
	}
	assert.Equal(t, expect, b.fields[0])
}

func TestBuilder_Byte(t *testing.T) {
	b := NewRequestBuilder(":5020", 2)

	b.Add(b.Byte(256, true).Name("fire_alarm_di"))

	expect := Field{
		ServerAddress: ":5020",
		UnitID:        2,
		Type:          FieldTypeByte,
		Address:       256,
		FromHighByte:  true,
		Name:          "fire_alarm_di",
	}
	assert.Equal(t, expect, b.fields[0])
}

func TestBuilder_Uint8(t *testing.T) {
	b := NewRequestBuilder(":5020", 2)

	b.Add(b.Uint8(256, true).Name("fire_alarm_di"))

	expect := Field{
		ServerAddress: ":5020",
		UnitID:        2,
		Type:          FieldTypeUint8,
		Address:       256,
		FromHighByte:  true,
		Name:          "fire_alarm_di",
	}
	assert.Equal(t, expect, b.fields[0])
}

func TestBuilder_Int8(t *testing.T) {
	b := NewRequestBuilder(":5020", 2)

	b.Add(b.Int8(256, true).Name("fire_alarm_di"))

	expect := Field{
		ServerAddress: ":5020",
		UnitID:        2,
		Type:          FieldTypeInt8,
		Address:       256,
		FromHighByte:  true,
		Name:          "fire_alarm_di",
	}
	assert.Equal(t, expect, b.fields[0])
}

func TestBuilder_Uint16(t *testing.T) {
	b := NewRequestBuilder(":5020", 2)

	b.Add(b.Uint16(256).Name("fire_alarm_di"))

	expect := Field{
		ServerAddress: ":5020",
		UnitID:        2,
		Type:          FieldTypeUint16,
		Address:       256,
		Name:          "fire_alarm_di",
	}
	assert.Equal(t, expect, b.fields[0])
}

func TestBuilder_Int16(t *testing.T) {
	b := NewRequestBuilder(":5020", 2)

	b.Add(b.Int16(256).Name("fire_alarm_di"))

	expect := Field{
		ServerAddress: ":5020",
		UnitID:        2,
		Type:          FieldTypeInt16,
		Address:       256,
		Name:          "fire_alarm_di",
	}
	assert.Equal(t, expect, b.fields[0])
}

func TestBuilder_Uint32(t *testing.T) {
	b := NewRequestBuilder(":5020", 2)

	b.Add(b.Uint32(256).Name("fire_alarm_di"))

	expect := Field{
		ServerAddress: ":5020",
		UnitID:        2,
		Type:          FieldTypeUint32,
		Address:       256,
		Name:          "fire_alarm_di",
	}
	assert.Equal(t, expect, b.fields[0])
}

func TestBuilder_Int32(t *testing.T) {
	b := NewRequestBuilder(":5020", 2)

	b.Add(b.Int32(256).Name("fire_alarm_di"))

	expect := Field{
		ServerAddress: ":5020",
		UnitID:        2,
		Type:          FieldTypeInt32,
		Address:       256,
		Name:          "fire_alarm_di",
	}
	assert.Equal(t, expect, b.fields[0])
}

func TestBuilder_Uint64(t *testing.T) {
	b := NewRequestBuilder(":5020", 2)

	b.Add(b.Uint64(256).Name("fire_alarm_di"))

	expect := Field{
		ServerAddress: ":5020",
		UnitID:        2,
		Type:          FieldTypeUint64,
		Address:       256,
		Name:          "fire_alarm_di",
	}
	assert.Equal(t, expect, b.fields[0])
}

func TestBuilder_Int64(t *testing.T) {
	b := NewRequestBuilder(":5020", 2)

	b.Add(b.Int64(256).Name("fire_alarm_di"))

	expect := Field{
		ServerAddress: ":5020",
		UnitID:        2,
		Type:          FieldTypeInt64,
		Address:       256,
		Name:          "fire_alarm_di",
	}
	assert.Equal(t, expect, b.fields[0])
}

func TestBuilder_Float32(t *testing.T) {
	b := NewRequestBuilder(":5020", 2)

	b.Add(b.Float32(256).Name("fire_alarm_di"))

	expect := Field{
		ServerAddress: ":5020",
		UnitID:        2,
		Type:          FieldTypeFloat32,
		Address:       256,
		Name:          "fire_alarm_di",
	}
	assert.Equal(t, expect, b.fields[0])
}

func TestBuilder_Float64(t *testing.T) {
	b := NewRequestBuilder(":5020", 2)

	b.Add(b.Float64(256).Name("fire_alarm_di"))

	expect := Field{
		ServerAddress: ":5020",
		UnitID:        2,
		Type:          FieldTypeFloat64,
		Address:       256,
		Name:          "fire_alarm_di",
	}
	assert.Equal(t, expect, b.fields[0])
}

func TestBuilder_String(t *testing.T) {
	b := NewRequestBuilder(":5020", 2)

	b.Add(b.String(256, 10).Name("fire_alarm_di"))

	expect := Field{
		ServerAddress: ":5020",
		UnitID:        2,
		Type:          FieldTypeString,
		Address:       256,
		Length:        10,
		Name:          "fire_alarm_di",
	}
	assert.Equal(t, expect, b.fields[0])
}

func TestBuilder_Coil(t *testing.T) {
	b := NewRequestBuilder(":5020", 2)

	b.Add(b.Coil(256).Name("fire_alarm_di"))

	expect := Field{
		ServerAddress: ":5020",
		UnitID:        2,
		Type:          FieldTypeCoil,
		Address:       256,
		Name:          "fire_alarm_di",
	}
	assert.Equal(t, expect, b.fields[0])
}

func TestBuilder_AddAll(t *testing.T) {
	var testCases = []struct {
		name   string
		when   Fields
		expect Fields
	}{
		{
			name: "ok",
			when: Fields{
				{
					ServerAddress: ":502",
					UnitID:        1,
					Address:       100,
					Type:          FieldTypeString,
					Bit:           1,
					FromHighByte:  true,
					Length:        10,
					ByteOrder:     packet.BigEndian,
					Name:          "added",
				},
			},
			expect: Fields{
				{
					ServerAddress: ":502",
					UnitID:        1,
					Address:       100,
					Type:          FieldTypeString,
					Bit:           1,
					FromHighByte:  true,
					Length:        10,
					ByteOrder:     packet.BigEndian,
					Name:          "added",
				},
			},
		},
		{
			name: "ok, add multiple",
			when: Fields{
				{ServerAddress: ":502", UnitID: 1},
				{ServerAddress: ":5020", UnitID: 2},
			},
			expect: Fields{
				{ServerAddress: ":502", UnitID: 1},
				{ServerAddress: ":5020", UnitID: 2},
			},
		},
		{
			name:   "ok, add nil",
			when:   nil,
			expect: Fields{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			b := NewRequestBuilder(":5020", 2)

			b.AddAll(tc.when)

			assert.Equal(t, tc.expect, b.fields)
		})
	}
}

func TestRegisterRequest_Fields(t *testing.T) {
	given := BuilderRequest{
		Request:       nil,
		ServerAddress: ":502",
		UnitID:        1,
		StartAddress:  100,
		Fields: Fields{
			{
				ServerAddress: ":502",
				UnitID:        1,
				Name:          "test1",
			},
			{
				ServerAddress: ":502",
				UnitID:        1,
				Name:          "test2",
			},
		},
	}

	expect := Fields{
		{
			ServerAddress: ":502",
			UnitID:        1,
			Name:          "test1",
		},
		{
			ServerAddress: ":502",
			UnitID:        1,
			Name:          "test2",
		},
	}
	assert.Equal(t, expect, given.Fields)
}

func TestRegisterRequest_AsRegisters(t *testing.T) {
	rr := BuilderRequest{
		Request:       nil,
		ServerAddress: ":502",
		UnitID:        1,
		StartAddress:  100,
		Fields:        nil,
	}

	resp := packet.ReadHoldingRegistersResponseTCP{
		MBAPHeader: packet.MBAPHeader{},
		ReadHoldingRegistersResponse: packet.ReadHoldingRegistersResponse{
			UnitID:          1,
			RegisterByteLen: 6,
			Data:            []byte{0xff, 0xff, 0x7f, 0xff, 0x0, 0x1},
		},
	}

	registers, err := rr.AsRegisters(resp)
	assert.NoError(t, err)

	value, err := registers.Uint16(102)
	assert.NoError(t, err)
	assert.Equal(t, uint16(1), value)
}

func TestRegisterRequest_ExtractFields(t *testing.T) {
	var testCases = []struct {
		name                           string
		givenFields                    Fields
		givenResponseData              []byte
		givenResponseFC                uint8
		whenContinueOnExtractionErrors bool
		expect                         []FieldValue
		expectErr                      string
	}{
		{
			name: "ok, extract registers",
			givenFields: Fields{
				{
					UnitID:  1,
					Address: 21,
					Type:    FieldTypeInt16,
					Name:    "f1",
				},
				{
					UnitID:  1,
					Address: 22,
					Type:    FieldTypeBit,
					Bit:     8,
					Name:    "f2",
				},
			},
			givenResponseData: []byte{0x0, 0x0, 0x0, 0x1, 0b00010001, 0x0},
			expect: []FieldValue{
				{
					Field: Field{
						UnitID:  1,
						Address: 21,
						Type:    FieldTypeInt16,
						Name:    "f1",
					},
					Value: int16(1),
					Error: nil,
				},
				{
					Field: Field{
						UnitID:  1,
						Address: 22,
						Type:    FieldTypeBit,
						Bit:     8,
						Name:    "f2",
					},
					Value: true,
					Error: nil,
				},
			},
		},
		{
			name: "ok, extract coils",
			givenFields: Fields{
				{
					UnitID:  1,
					Address: 20,
					Type:    FieldTypeCoil,
					Name:    "f1",
				},
				{
					UnitID:  1,
					Address: 21,
					Type:    FieldTypeCoil,
					Name:    "f2",
				},
			},
			givenResponseData: []byte{0b0000_0101},
			givenResponseFC:   packet.FunctionReadCoils,
			expect: []FieldValue{
				{
					Field: Field{
						UnitID:  1,
						Address: 20,
						Type:    FieldTypeCoil,
						Name:    "f1",
					},
					Value: true,
					Error: nil,
				},
				{
					Field: Field{
						UnitID:  1,
						Address: 21,
						Type:    FieldTypeCoil,
						Name:    "f2",
					},
					Value: false,
					Error: nil,
				},
			},
		},
		{
			name: "nok, register packet had errors, ContinueOnExtractionErrors=true",
			givenFields: Fields{
				{
					UnitID:  1,
					Address: 21,
					Type:    FieldTypeInt16,
					Name:    "f1",
				},
				{
					UnitID:  1,
					Address: 22,
					Type:    FieldTypeFloat64,
					Name:    "f2",
				},
			},
			givenResponseData:              []byte{0x0, 0x0, 0x0, 0x1, 0b00010001, 0x0},
			whenContinueOnExtractionErrors: true,
			expect: []FieldValue{
				{
					Field: Field{
						UnitID:  1,
						Address: 21,
						Type:    FieldTypeInt16,
						Name:    "f1",
					},
					Value: int16(1),
					Error: nil,
				},
				{
					Field: Field{
						UnitID:  1,
						Address: 22,
						Type:    FieldTypeFloat64,
						Name:    "f2",
					},
					Value: float64(0),
					Error: errors.New("address over startAddress+quantity bounds"),
				},
			},
			expectErr: ErrorFieldExtractHadError.Error(),
		},
		{
			name: "nok, coils packet had errors, ContinueOnExtractionErrors=true",
			givenFields: Fields{
				{
					UnitID:  1,
					Address: 20,
					Type:    FieldTypeCoil,
					Name:    "f1",
				},
				{
					UnitID:  1,
					Address: 0,
					Type:    FieldTypeCoil,
					Name:    "f2",
				},
			},
			givenResponseData:              []byte{0b0000_0101},
			givenResponseFC:                packet.FunctionReadCoils,
			whenContinueOnExtractionErrors: true,
			expect: []FieldValue{
				{
					Field: Field{
						UnitID:  1,
						Address: 20,
						Type:    FieldTypeCoil,
						Name:    "f1",
					},
					Value: true,
					Error: nil,
				},
				{
					Field: Field{
						UnitID:  1,
						Address: 0,
						Type:    FieldTypeCoil,
						Name:    "f2",
					},
					Value: false,
					Error: errors.New("bit can not be before startBit"),
				},
			},
			expectErr: ErrorFieldExtractHadError.Error(),
		},
		{
			name: "nok, had errors, ContinueOnExtractionErrors=false",
			givenFields: Fields{
				{
					UnitID:  1,
					Address: 21,
					Type:    FieldTypeInt16,
					Name:    "f1",
				},
				{
					UnitID:  1,
					Address: 22,
					Type:    FieldTypeFloat64,
					Name:    "f2",
				},
			},
			givenResponseData:              []byte{0x0, 0x0, 0x0, 0x1, 0b00010001, 0x0},
			whenContinueOnExtractionErrors: false,
			expect:                         nil,
			expectErr:                      "field extraction failed. name: f2 err: address over startAddress+quantity bounds",
		},
		{
			name: "nok, error creating registers",
			givenFields: Fields{
				{
					UnitID:  1,
					Address: 21,
					Type:    FieldTypeInt16,
					Name:    "f1",
				},
				{
					UnitID:  1,
					Address: 22,
					Type:    FieldTypeFloat64,
					Name:    "f2",
				},
			},
			givenResponseData:              []byte{0x0, 0x0, 0x0, 0x1, 0b00010001},
			whenContinueOnExtractionErrors: false,
			expect:                         nil,
			expectErr:                      "data length must be odd number of bytes as 1 register is 2 bytes",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := BuilderRequest{
				Request:       nil,
				ServerAddress: ":502",
				UnitID:        1,
				StartAddress:  20,
				Fields:        tc.givenFields,
			}
			var response packet.Response
			switch tc.givenResponseFC {
			case packet.FunctionReadCoils:
				response = packet.ReadCoilsResponseTCP{
					MBAPHeader: packet.MBAPHeader{},
					ReadCoilsResponse: packet.ReadCoilsResponse{
						UnitID:          1,
						CoilsByteLength: uint8(len(tc.givenResponseData)),
						Data:            tc.givenResponseData,
					},
				}
			default:
				response = packet.ReadHoldingRegistersResponseTCP{
					MBAPHeader: packet.MBAPHeader{},
					ReadHoldingRegistersResponse: packet.ReadHoldingRegistersResponse{
						UnitID:          1,
						RegisterByteLen: uint8(len(tc.givenResponseData)),
						Data:            tc.givenResponseData,
					},
				}
			}

			fields, err := req.ExtractFields(response, tc.whenContinueOnExtractionErrors)
			if tc.expectErr != "" {
				assert.EqualError(t, err, tc.expectErr)
			} else {
				assert.NoError(t, err)
			}

			assert.Len(t, fields, len(tc.expect))
			assert.Equal(t, tc.expect, fields)
		})
	}
}

func TestField_registerSize(t *testing.T) {
	var testCases = []struct {
		name   string
		when   Field
		expect uint16
	}{
		{
			name:   "bit",
			when:   Field{Type: FieldTypeBit, Bit: 1},
			expect: 1,
		},
		{
			name:   "byte",
			when:   Field{Type: FieldTypeByte, FromHighByte: true},
			expect: 1,
		},
		{
			name:   "uint8",
			when:   Field{Type: FieldTypeUint8, FromHighByte: false},
			expect: 1,
		},
		{
			name:   "int8",
			when:   Field{Type: FieldTypeInt8, FromHighByte: true},
			expect: 1,
		},
		{
			name:   "uint16",
			when:   Field{Type: FieldTypeUint16},
			expect: 1,
		},
		{
			name:   "int16",
			when:   Field{Type: FieldTypeInt16},
			expect: 1,
		},
		{
			name:   "uint32",
			when:   Field{Type: FieldTypeUint32},
			expect: 2,
		},
		{
			name:   "int32",
			when:   Field{Type: FieldTypeInt32},
			expect: 2,
		},
		{
			name:   "uint64",
			when:   Field{Type: FieldTypeUint64},
			expect: 4,
		},
		{
			name:   "int64",
			when:   Field{Type: FieldTypeInt64},
			expect: 4,
		},
		{
			name:   "float32",
			when:   Field{Type: FieldTypeFloat32},
			expect: 2,
		},
		{
			name:   "float64",
			when:   Field{Type: FieldTypeFloat64},
			expect: 4,
		},
		{
			name:   "string odd size",
			when:   Field{Type: FieldTypeString, Length: 5},
			expect: 3,
		},
		{
			name:   "string even size",
			when:   Field{Type: FieldTypeString, Length: 6},
			expect: 3,
		},
		{
			name:   "string even size2",
			when:   Field{Type: FieldTypeString, Length: 4},
			expect: 2,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expect, tc.when.registerSize())
		})
	}
}

func TestField_ExtractFrom(t *testing.T) {
	var testCases = []struct {
		name              string
		givenRegisterData []byte
		whenType          FieldType
		whenByteOrder     packet.ByteOrder
		expect            interface{}
		expectErr         string
	}{
		{
			name:              "bit",
			givenRegisterData: []byte{0x0, 0x0, 0b00010001, 0x0},
			whenType:          FieldTypeBit,
			expect:            true,
		},
		{
			name:              "byte",
			givenRegisterData: []byte{0x0, 0x0, 0x1, 0x0},
			whenType:          FieldTypeByte,
			expect:            byte(1),
		},
		{
			name:              "uint8",
			whenType:          FieldTypeUint8,
			givenRegisterData: []byte{0x0, 0x0, 0xFF, 0x0},
			expect:            uint8(255),
		},
		{
			name:              "int8",
			whenType:          FieldTypeInt8,
			givenRegisterData: []byte{0x0, 0x0, 0xFF, 0x0},
			expect:            int8(-1),
		},
		{
			name:              "uint16",
			whenType:          FieldTypeUint16,
			givenRegisterData: []byte{0x0, 0x0, 0x0, 0xFF},
			expect:            uint16(255),
		},
		{
			name:              "int16",
			whenType:          FieldTypeInt16,
			givenRegisterData: []byte{0x0, 0x0, 0xFF, 0xFF},
			expect:            int16(-1),
		},
		{
			name:              "uint32",
			whenType:          FieldTypeUint32,
			givenRegisterData: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x1},
			expect:            uint32(1),
		},
		{
			name:              "int32",
			whenType:          FieldTypeInt32,
			givenRegisterData: []byte{0x0, 0x0, 0xFF, 0xFF, 0xFF, 0xFF},
			expect:            int32(-1),
		},
		{
			name:              "uint64",
			whenType:          FieldTypeUint64,
			givenRegisterData: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1},
			expect:            uint64(1),
		},
		{
			name:              "int64",
			whenType:          FieldTypeInt64,
			givenRegisterData: []byte{0x0, 0x0, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
			expect:            int64(-1),
		},
		{
			name:              "float32",
			whenType:          FieldTypeFloat32,
			whenByteOrder:     packet.BigEndianLowWordFirst,
			givenRegisterData: []byte{0x0, 0x0, 0xcc, 0xcd, 0x3f, 0xec},
			expect:            float32(1.85),
		},
		{
			name:              "float64",
			whenType:          FieldTypeFloat64,
			whenByteOrder:     packet.BigEndianLowWordFirst,
			givenRegisterData: []byte{0x0, 0x0, 0x99, 0x9a, 0x99, 0x99, 0x99, 0x99, 0x3f, 0xfd},
			expect:            float64(1.85),
		},
		{
			name:              "string odd size",
			whenType:          FieldTypeString,
			whenByteOrder:     packet.LittleEndian,
			givenRegisterData: []byte{0x0, 0x0, 0x53, 0x56, 0x43, 0x83},
			expect:            "SVC",
		},
		{
			name:              "nok, coil can not be extracted from registers",
			whenType:          FieldTypeCoil,
			givenRegisterData: []byte{0x0, 0x0, 0x53, 0x56, 0x43, 0x83},
			expectErr:         "extraction failure due unknown field type",
		},
		{
			name:      "nok, unknown type",
			whenType:  0,
			expect:    nil,
			expectErr: "extraction failure due unknown field type",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			f := Field{
				ServerAddress: ":502",
				UnitID:        1,
				Address:       1,
				Type:          tc.whenType,
				Bit:           8,
				FromHighByte:  true,
				Length:        3,
				ByteOrder:     tc.whenByteOrder,
				Name:          "test",
			}

			registers, _ := packet.NewRegisters(tc.givenRegisterData, 0)

			result, err := f.ExtractFrom(registers)

			assert.Equal(t, tc.expect, result)
			if tc.expectErr != "" {
				assert.EqualError(t, err, tc.expectErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestField_Validate(t *testing.T) {
	example := Field{
		ServerAddress: ":502",
		UnitID:        1,
		Address:       100,
		Type:          FieldTypeString,
		Bit:           0,
		FromHighByte:  false,
		Length:        10,
		ByteOrder:     0,
		Name:          "fire_alarm_di",
	}
	var testCases = []struct {
		name      string
		given     func(f *Field)
		expectErr string
	}{
		{
			name:  "ok",
			given: func(f *Field) {},
		},
		{
			name:      "nok, server address is empty",
			given:     func(f *Field) { f.ServerAddress = "" },
			expectErr: "field server address can not be empty",
		},
		{
			name:      "nok, type is not set",
			given:     func(f *Field) { f.Type = 0 },
			expectErr: "field type must be set",
		},
		{
			name:      "nok, type is invalid value",
			given:     func(f *Field) { f.Type = 15 },
			expectErr: "field type has invalid value",
		},
		{
			name:      "nok, bit out of range",
			given:     func(f *Field) { f.Bit = 16 },
			expectErr: "field bit value must be in range (0-15)",
		},
		{
			name: "nok, string type must have length",
			given: func(f *Field) {
				f.Type = FieldTypeString
				f.Length = 0
			},
			expectErr: "field with type string must have length set",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			f := example

			tc.given(&f)

			err := f.Validate()
			if tc.expectErr != "" {
				assert.EqualError(t, err, tc.expectErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
