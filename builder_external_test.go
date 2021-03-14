package modbus_test

import (
	"context"
	"github.com/aldas/go-modbus-client"
	"github.com/aldas/go-modbus-client/packet"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func xTestHello(t *testing.T) {
	// try out builder API

	b := modbus.Builder{}

	reqs, err := b.Add(b.Int64(18)).
		Add(b.Int64(18).SetName("alarm_do_1")).
		Add(modbus.Field{Address: 18, Name: "alarm_do_1"}).
		Build()
	assert.NoError(t, err)

	client := modbus.NewClient()
	if err := client.Connect(context.Background(), ":502"); err != nil {
		return
	}
	for _, req := range reqs {
		resp, err := client.Do(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
	}

	c := http.Client{
		Transport:     nil,
		CheckRedirect: nil,
		Jar:           nil,
		Timeout:       0,
	}

	c.Do(nil)

	resp, err := packet.ParseTCPResponse([]byte{})
	assert.NoError(t, err)
	assert.NotNil(t, resp)

}
