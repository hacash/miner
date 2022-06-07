package message

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
)

// Receive message mustlen = message must be long
func MsgReadFromTcpConn(conn *net.TCPConn, mustlen uint32) (msgty uint16, msgbody []byte, err error) {
	var rn int = 0
	err = nil
	msgty = 0
	msgbody = nil

	msgitemsizebts := make([]byte, 4)
	rn, err = io.ReadFull(conn, msgitemsizebts)
	if rn != 4 || err != nil {
		if err != nil {
			err = fmt.Errorf("read msg size error:", err)
		}
		return 0, nil, err
	}

	// Message body length
	msgsize := binary.BigEndian.Uint32(msgitemsizebts)
	if msgsize < 2 || (mustlen > 0 && msgsize != mustlen+2) {
		return 0, nil, fmt.Errorf("msg len error, msgsize = %d, mustlen = %d", msgsize, mustlen)
	}

	//fmt.Println("read msgsize=", msgsize)

	// Read message body
	msgbodysizebts := make([]byte, msgsize)
	rn, err = io.ReadFull(conn, msgbodysizebts)
	if uint32(rn) != msgsize || err != nil {
		if err != nil {
			err = fmt.Errorf("read msg body error")
		}
		return 0, nil, err
	}

	//fmt.Println("msgbodysizebts=", msgbodysizebts)

	// body
	msgty = binary.BigEndian.Uint16(msgbodysizebts[0:2])
	if msgsize > 2 {
		msgbody = msgbodysizebts[2:]
	}
	return
}

// send message
func MsgSendToTcpConn(conn *net.TCPConn, msgty uint16, msgbody []byte) (err error) {
	err = nil
	if conn == nil {
		return fmt.Errorf("conn is nil")
	}

	sendbts := make([]byte, 4+2+len(msgbody))
	binary.BigEndian.PutUint32(sendbts[0:4], 2+uint32(len(msgbody))) // len
	binary.BigEndian.PutUint16(sendbts[4:6], msgty)                  // type
	copy(sendbts[6:], msgbody)

	//fmt.Println("MsgSendToTcpConn: ", sendbts)

	// send
	_, err = conn.Write(sendbts)

	return
}
