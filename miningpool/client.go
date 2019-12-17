package miningpool

import (
	"github.com/hacash/core/interfaces"
	"net"
)

type Client struct {
	belongAccount *Account

	conn *net.TCPConn

	workblock interfaces.Block

	coinbaseMsgNum uint32 // > 0

	successNonce uint32 // > 0

}
