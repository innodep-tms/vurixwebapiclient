package vurixwebapiclient

import (
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/parjom/goutil"
	"github.com/spf13/cast"
)

type OptVurixEventReceiver struct {
	DeviceEvent      bool // event
	SystemEvent      bool // system
	MonitoringEvent  bool // monitoring
	UserLogEvent     bool // userlog
	UserEvent        bool // user
	VfsEvent         bool // vfs
	MasEvent         bool // mas
	EventPeriodByDev int  // 장비 이벤트(DEVICE_EVENT) 수신 시 장비별 수신주기 값(초)
}

type VurixEventReceiver struct {
	opt    OptVurixEventReceiver
	vc     *VurixWebApiClient
	logger Logger
	conn   net.Conn
}

func NewVurixEventReceiver(option OptVurixEventReceiver) *VurixEventReceiver {
	return &VurixEventReceiver{
		opt: option,
	}
}

func (ver *VurixEventReceiver) SetVurixWebApiClient(vc *VurixWebApiClient) {
	ver.vc = vc
	if ver.conn != nil {
		ver.conn.Close()
		ver.conn = nil
	}
}
func (ver *VurixEventReceiver) Run(logger Logger) {
	ver.logger = logger
	if ver.opt.DeviceEvent || ver.opt.SystemEvent || ver.opt.MonitoringEvent || ver.opt.UserLogEvent || ver.opt.UserEvent || ver.opt.VfsEvent || ver.opt.MasEvent {
		go ver.runLoop()
	} else {
		ver.logger.Infof("Invaild value for OptVurixEventReceiver : DeviceEvent = %v, SystemEvent = %v, MonitoringEvent = %v, UserLogEvent = %v, UserEvent = %v, VfsEvent = %v, MasEvent = %v ", ver.opt.DeviceEvent, ver.opt.SystemEvent, ver.opt.MonitoringEvent, ver.opt.UserLogEvent, ver.opt.UserEvent, ver.opt.VfsEvent, ver.opt.MasEvent)
	}
}

func (ver *VurixEventReceiver) Stop() {
	if ver.conn != nil {
		ver.conn.Close()
		ver.conn = nil
	}
}

func (ver *VurixEventReceiver) runLoop() {
	for ver.vc.isRetryLogin {
		if ver.vc.isLogin {
			ver.receviceEvent()
		}
		ver.vc.SleepWithContext(60 * time.Second)
	}
	ver.logger.Infof("End Loop VurixEventReceiver !!")
}

func (ver *VurixEventReceiver) receviceEvent() {
	host := fmt.Sprintf("%s:%d", ver.vc.opt.Host, ver.vc.opt.Port)
	rawPath := ver.makeQueryPath()
	authToken := ver.vc.loginInfo.AuthToken
	apiSerial := ver.vc.loginInfo.ApiSerial

	parser := NewMultiPartParser()

	var err error = nil
	ver.conn, err = net.Dial("tcp", host)
	if err != nil {
		ver.logger.Infof("Connection fail to vurix Webapi : %s, %s", host, err.Error())
		return
	}
	defer ver.conn.Close()

	sendData := make([]byte, 0)
	sendData = append(sendData, []byte("GET "+rawPath+" HTTP/1.1\r\n")...)
	sendData = append(sendData, []byte("Host: "+host+"\r\n")...)
	sendData = append(sendData, []byte("x-auth-token: "+authToken+"\r\n")...)
	sendData = append(sendData, []byte("x-api-serial: "+cast.ToString(apiSerial)+"\r\n")...)
	sendData = append(sendData, []byte("\r\n")...)

	ver.logger.Debugf("Send Data : \n%s", string(sendData))
	ver.conn.Write(sendData)

	recv := make([]byte, 4096)
	for {
		n, err := ver.conn.Read(recv)
		if err != nil {
			ver.logger.Warn("Failed to Read data : ", err)
			break
		}

	PARSINGLABEL:
		msg, err2 := parser.Parse(recv, n)
		if err2 != nil {
			ver.logger.Warn("Failed to Parsing data : ", err2)
			break
		}
		if msg != nil {
			// 메세지 파싱 완료 샌드 큐에 메세지 입력
			ver.SendMessage(msg)
			n = 0
			goto PARSINGLABEL
		}

	}
}

func (ver *VurixEventReceiver) makeQueryPath() (queryPath string) {
	queryPath = "/api/event/receive"
	queryPath = queryPath + "?event=" + cast.ToString(ver.opt.DeviceEvent)
	queryPath = queryPath + "&system=" + cast.ToString(ver.opt.DeviceEvent)
	queryPath = queryPath + "&monitoring=" + cast.ToString(ver.opt.MonitoringEvent)
	queryPath = queryPath + "&userlog=" + cast.ToString(ver.opt.UserLogEvent)
	queryPath = queryPath + "&user=" + cast.ToString(ver.opt.UserEvent)
	queryPath = queryPath + "&vfs=" + cast.ToString(ver.opt.VfsEvent)
	queryPath = queryPath + "&mas=" + cast.ToString(ver.opt.MasEvent)
	queryPath = queryPath + "&event_period_by_dev=" + cast.ToString(ver.opt.EventPeriodByDev)

	return
}

func (ver *VurixEventReceiver) SendMessage(msg []byte) {
	var jsonMsg interface{}
	err := json.Unmarshal(msg, &jsonMsg)
	if err == nil {
		if cast.ToBool(goutil.JsonGetValue(jsonMsg, "success")) && cast.ToInt(goutil.JsonGetValue(jsonMsg, "code")) == 200 {
			// 이경우는 접속 성공 메세지 이므로 폐기
			// 예) {"success":true,"code":200,"message":"Success","detail":"success to connection"}
			ver.logger.Debugf("Discard Msg : %s", string(msg))
		} else {
			ver.vc.eventCallback(jsonMsg)
		}
	} else {
		ver.logger.Debugf("Discard Msg : %s", string(msg))
	}
}

