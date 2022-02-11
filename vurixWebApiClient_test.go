package vurixwebapiclient

import (
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/spf13/cast"
	"go.uber.org/zap"
)

func TestVurixWebClient(t *testing.T) {

	opt := NewOptVurixWebApiClient()
	opt.Host = "172.16.31.202"
	opt.Port = 8080
	opt.User = "admin"
	opt.Pass = "admin"
	opt.License = "licNormalClient"

	vc := NewVurixWebApiClient(opt, zap.S())
	vc.Login()

	auth_token, api_serial := vc.GetToken()

	host := "172.16.31.202:8080"
	path := "/api/event/receive"
	query := "event=true&system=true&monitoring=true&user=true"

	conn, err := net.Dial("tcp", host)
	if err != nil {
		fmt.Println("Faield to Dial : ", err)
	}
	defer conn.Close()

	rawPath := path
	if query != "" {
		rawPath += "?" + query
	}

	sendData := make([]byte, 0)
	sendData = append(sendData, []byte("GET "+rawPath+" HTTP/1.1\r\n")...)
	sendData = append(sendData, []byte("Host: "+host+"\r\n")...)
	sendData = append(sendData, []byte("x-auth-token: "+auth_token+"\r\n")...)
	sendData = append(sendData, []byte("x-api-serial: "+cast.ToString(api_serial)+"\r\n")...)
	sendData = append(sendData, []byte("\r\n")...)

	fmt.Println("send : ", time.Now(), string(sendData))

	conn.Write(sendData)

	go func(c net.Conn) {
		recv := make([]byte, 4096)

		for {
			n, err := c.Read(recv)
			if err != nil {
				fmt.Println(time.Now(), "Failed to Read data : ", err)
				break
			}

			fmt.Println("recv : ", time.Now(), string(recv[:n]))
		}
	}(conn)

	time.Sleep(1* time.Hour)

}
