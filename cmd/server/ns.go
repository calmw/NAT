package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
)

func main() {
	udpAddr, err := net.ResolveUDPAddr("udp4", ":53771") // 转换地址，作为服务器使用时需要监听本机的一个端口
	// 端口号写 0 可以由系统随机分配可用的端口号
	if err != nil {
		log.Printf("server addr error: %v", err)
		return
	}
	conn, err := net.ListenUDP("udp", udpAddr) // 启动UDP监听本机端口
	if err != nil {
		log.Printf("server Listen UDP error: %v", err)
		return
	}

	for {
		var buf [128]byte
		dataLen, addr, err := conn.ReadFromUDP(buf[:]) // 读取数据，返回值依次为读取数据长度、远端地址、错误信息 // 读取操作会阻塞直至有数据可读取
		if err != nil {
			log.Printf("json.Marshal error:%v", err)
			continue
		}

		fmt.Println(string(buf[:dataLen])) // 向终端打印收到的消息
		log.Printf("remote addr:%s, port:%d", addr.IP.String(), addr.Port)

		//type UDPAddr struct {
		//	IP   IP
		//	Port int
		//	Zone string // IPv6 scoped addressing zone
		//}

		addrMarshal, err := json.Marshal(*addr)
		if err != nil {
			log.Printf("json.Marshal error:%v", err)
			continue
		}
		var exportAddr net.UDPAddr
		err = json.Unmarshal(addrMarshal, &exportAddr)
		fmt.Println(err, exportAddr, string(addrMarshal))
		_, err = conn.WriteToUDP(addrMarshal, addr) // 写数据，返回值依次为写入数据长度、错误信息 // WriteToUDP()并非只能用于应答的，只要有个远程地址可以随时发消息
		if err != nil {
			log.Printf("json.Marshal error:%v", err)
			continue
		}
	}
}

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error %s", err.Error())
		os.Exit(1)
	}
}
