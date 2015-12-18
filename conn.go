package mysqldriver

import (
	"net"

	"github.com/pubnative/mysqlproto-go"
)

var capabilityFlags = mysqlproto.CLIENT_LONG_PASSWORD |
	mysqlproto.CLIENT_FOUND_ROWS |
	mysqlproto.CLIENT_LONG_FLAG |
	mysqlproto.CLIENT_CONNECT_WITH_DB |
	mysqlproto.CLIENT_PLUGIN_AUTH |
	mysqlproto.CLIENT_TRANSACTIONS |
	mysqlproto.CLIENT_PROTOCOL_41

type Conn struct {
	conn mysqlproto.Conn
}

type Stats struct {
	Syscalls int
}

func NewConn(username, password, protocol, address, database string) (Conn, error) {
	conn, err := net.Dial(protocol, address)
	if err != nil {
		return Conn{}, err
	}

	stream, err := mysqlproto.Handshake(
		conn, capabilityFlags,
		username, password, database, nil,
	)

	if err != nil {
		return Conn{}, err
	}

	if err = setUTF8Charset(stream); err != nil {
		return Conn{}, err
	}

	return Conn{stream}, nil
}

func (c Conn) Close() error {
	return c.conn.Close()
}

func (c Conn) Stats() Stats {
	return Stats{
		Syscalls: c.conn.Syscalls(),
	}
}

func (s Stats) Add(stats Stats) Stats {
	return Stats{
		Syscalls: s.Syscalls + stats.Syscalls,
	}
}

func setUTF8Charset(conn mysqlproto.Conn) error {
	data := mysqlproto.ComQueryRequest([]byte("SET NAMES utf8"))
	if _, err := conn.Write(data); err != nil {
		return err
	}

	packet, err := conn.NextPacket()
	if err != nil {
		return err
	}

	return handleOK(packet.Payload, conn.CapabilityFlags)
}
