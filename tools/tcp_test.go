package tools

import (
	"fmt"
	"net"
	"testing"
	"time"

	"gotest.tools/assert"
)

type MockConn struct {
	MockBanner   string
	MockContent  string
	ForceTimeout bool
	Expiration   time.Time
}

func (conn *MockConn) Read(b []byte) (n int, err error) {
	if conn.ForceTimeout {
		return 0, fmt.Errorf("timeout")
	}
	copy(b, []byte(conn.MockBanner))
	return len(b), nil
}

func (conn *MockConn) Write(b []byte) (n int, err error) {
	if conn.ForceTimeout {
		return 0, fmt.Errorf("timeout")
	}
	conn.MockContent = string(b)
	return len(b), nil
}

func (conn *MockConn) Close() error {
	return nil
}

func (conn *MockConn) LocalAddr() net.Addr {
	return nil
}

func (conn *MockConn) RemoteAddr() net.Addr {
	return nil
}

func (conn *MockConn) SetDeadline(t time.Time) error {
	conn.Expiration = t
	return nil
}

func (conn *MockConn) SetReadDeadline(t time.Time) error {
	conn.Expiration = t
	return nil
}

func (conn *MockConn) SetWriteDeadline(t time.Time) error {
	conn.Expiration = t
	return nil
}

func TestNetMessageContains(t *testing.T) {
	conn := &MockConn{MockBanner: "hello"}
	timeout := 1 * time.Second

	ok, err := NetMessageContains(conn, timeout, "")
	assert.NilError(t, err)
	assert.Assert(t, ok)

	ok, err = NetMessageContains(conn, timeout, "*")
	assert.NilError(t, err)
	assert.Assert(t, ok)

	ok, err = NetMessageContains(conn, timeout, "hello")
	assert.NilError(t, err)
	assert.Assert(t, ok)

	ok, err = NetMessageContains(conn, timeout, "~^h.*")
	assert.NilError(t, err)
	assert.Assert(t, ok)

	ok, err = NetMessageContains(conn, timeout, "hi")
	assert.Assert(t, !ok)
	assert.Assert(t, err != nil)

	ok, err = NetMessageContains(conn, timeout, "~^UNK")
	assert.Assert(t, !ok)
	assert.Assert(t, err != nil)
}
