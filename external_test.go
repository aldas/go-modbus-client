package modbus_test

import (
	"github.com/aldas/go-modbus-client"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestHello(t *testing.T) {
	assert.Equal(t, "Hello World!", modbus.Hello())
}
