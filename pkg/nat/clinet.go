package nat

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"nat/pkg/log"
	"nat/pkg/nat/internal/model"
	"net"
	"time"
)

const (
	NatTypeError    = iota // 未知类型
	NatTypeTaper           // 完全锥型
	NatTypeIp              // IP限制型
	NatTypePort            // 端口限制型
	NatTypeSymmetry        // 对称型
)

type Client struct {
	Conn       *net.UDPConn // UDP连接
	ServerAddr net.UDPAddr  // UDP连接
	ExportAddr net.UDPAddr  // UDP连接
	NatType    int          // NAT 类型
}

func NewClient() *Client {
	return &Client{}
}

func (c *Client) DialUDP(serverIp string, serverPort int) error {
	serverAddr := fmt.Sprintf("%s:%d", serverIp, serverPort)
	udpAddr, err := net.ResolveUDPAddr("udp4", serverAddr) // 转换地址，作为客户端使用要向远程发送消息，这里用远程地址与端口号
	if err != nil {
		log.Logger.Error(err.Error())
		return err
	}
	conn, err := net.DialUDP("udp", nil, udpAddr) // 建立连接，第二个参数为nil时通过默认本地地址（猜测可能是第一个可用的地址，未进行测试）发送且端口号自动分配，第三个参数为远程端地址与端口号
	if err != nil {
		log.Logger.Error(err.Error())
		return err
	}
	c.Conn = conn
	c.ServerAddr = net.UDPAddr{
		IP:   net.IP(serverIp),
		Port: serverPort,
	}
	return nil
}

func (c *Client) CheckNatType() error {
	log.Logger.Sugar().Info(c.isTaperType())
	return nil
}

// 第一步，判断是否有 NAT 防护主机向服务器
func (c *Client) HasNatProtection() (bool, error) {
	ctx, cancle := context.WithTimeout(context.Background(), time.Second)
	defer cancle()
	localIp, err := getLocalIp()

	fmt.Println(localIp, err, 7777)
	exportAddr, serverAddr, err := c.getExportAndServerAddr(ctx)
	fmt.Println(serverAddr, 8888)
	if err != nil {
		// 如果主机收不到服务器 #1 返回的消息，则说明用户的网络限制了 UDP 协议，直接退出。
		log.Logger.Sugar().Error(err)
		return false, err
	}
	c.ExportAddr = exportAddr
	// 拥有公网地址的主机. 如果能收到包，则判断返回的主机的外网 IP 地址是否与主机自身的 IP 地址一样。如果一样，说明主机就是一台拥有公网地址的主机；
	fmt.Println(localIp, c.ExportAddr.IP.String(), localIp == c.ExportAddr.IP.String(), 9999)
	if localIp == c.ExportAddr.IP.String() {

	} else { // 主机是处于 NAT 的防护之下

	}
	return false, err
}

func (c *Client) isTaperType() (bool, error) {
	ctx, cancle := context.WithTimeout(context.Background(), time.Second)
	defer cancle()
	exportAddr, serverAddr, err := c.getExportAndServerAddr(ctx)
	fmt.Println(exportAddr, serverAddr, err, 123)
	if err != nil {
		log.Logger.Sugar().Errorf("getExportAddr error:%v", err)
		return false, err
	}

	return false, nil
}

// 获取出口和服务器的Addr
func (c *Client) getExportAndServerAddr(ctx context.Context) (net.UDPAddr, net.UDPAddr, error) {
	applyChan := make(chan model.ExportAddr)
	go func() {
		var exportAddr net.UDPAddr
		var serverAddr net.UDPAddr
		_, err := c.Conn.Write([]byte("ping")) // 向远程端发送消息
		if err != nil {
			log.Logger.Sugar().Errorf("send to ServerAddr error: %v", err)
			applyChan <- model.ExportAddr{
				ExportAddr: exportAddr,
				ServerAddr: serverAddr,
				Err:        err,
			}
		}
		var buf [128]byte
		length, server, err := c.Conn.ReadFrom(buf[0:])
		if err != nil {
			log.Logger.Sugar().Errorf("read from ServerAddr error: %v", err)
			applyChan <- model.ExportAddr{
				ExportAddr: exportAddr,
				ServerAddr: serverAddr,
				Err:        err,
			}
		} // 读取数据 // 读取操作会阻塞直至有数据可读取
		serverAddr = *server.(*net.UDPAddr)
		err = json.Unmarshal(buf[:length], &exportAddr)
		fmt.Println(string(buf[:length]))

		applyChan <- model.ExportAddr{
			ExportAddr: exportAddr,
			ServerAddr: serverAddr,
			Err:        err,
		}
	}()

	for {
		select {
		case apply := <-applyChan:
			return apply.ExportAddr, apply.ServerAddr, apply.Err
		case <-ctx.Done():
			return net.UDPAddr{}, net.UDPAddr{}, ctx.Err()
		}
	}
}

func getLocalIp() (string, error) {
	addrs, err := net.InterfaceAddrs()

	if err != nil {
		return "", err
	}

	for _, address := range addrs {
		// 检查ip地址判断是否回环地址
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}
		}
	}

	return "", errors.New("Can not find the client ip address!")
}
