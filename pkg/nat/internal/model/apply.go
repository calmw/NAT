package model

import "net"

type ExportAddr struct {
	ExportAddr net.UDPAddr // 出口Addr
	ServerAddr net.UDPAddr // 服务器Addr
	Err        error
}
