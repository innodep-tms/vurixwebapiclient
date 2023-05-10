package vurixwebapiclient

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/parjom/goutil"
	"github.com/spf13/cast"
)

type EventCallbackFunc func(interface{})

type LoginInfo struct {
	AuthToken  string
	ApiSerial  int
	VmsID      int
	GrpSerial  int
	UserSerial int
	UserID     string
	UserName   string
	Utc        bool
}

type VurixWebApiClient struct {
	client             *resty.Client
	opt                OptVurixWebApiClient // 로그인 계정 정보
	webApiURL          string               // API호출 URL
	isRetryLogin       bool                 // 로그인 리트라이를 할지 저장하는 변수
	isLogin            bool                 // 현재 로그인 상태인지 저장하는 변수
	loginInfo          LoginInfo            // 로그인 완료 상태라면, 로그인 정보를 담고 있는 변수
	logger             Logger              // 로그 객체 정보
	ctx                context.Context      //
	cancel             context.CancelFunc   //
	eventReceiveClient *VurixEventReceiver  // 이벤트 수신 객체
	eventCallback      EventCallbackFunc    // 이벤트 수신 콜백
	isDebug            bool                 // 디버그 모드
}

func (vc *VurixWebApiClient) GetToken() (token string, apiserial int) {
	token = vc.loginInfo.AuthToken
	apiserial = vc.loginInfo.ApiSerial
	return
}

func NewVurixWebApiClient(opt OptVurixWebApiClient) *VurixWebApiClient {

	logger := createLogger()
	vc := VurixWebApiClient{
		client:    resty.New(),
		opt:       opt,
		webApiURL: fmt.Sprintf("http://%s:%d", opt.Host, opt.Port),
		isLogin:   false,
		loginInfo: LoginInfo{},
		logger:    logger,
	}

	// 고루틴을 종료하기 위한 Cancel 객체 생성
	vc.ctx, vc.cancel = context.WithCancel(context.Background())
	vc.client.SetLogger(logger)

	return &vc
}

func (vc *VurixWebApiClient) SetLogger(logger Logger) {
	vc.logger = logger
	vc.client.SetLogger(logger)
}


func (vc *VurixWebApiClient) SetDebug(debug bool) {
	vc.isDebug = debug
	vc.client.SetDebug(debug)
}

func (vc *VurixWebApiClient) GetDebug() bool {
	return vc.isDebug
}

func (vc *VurixWebApiClient) Login() {
	vc.isRetryLogin = true
	vc.loginAction()
	// 1분에 한번씩 로그인 리트라이 루프를 돌린다.
	go vc.retryLoginLoop(60 * time.Second)
}

// http://host:port/api/logout
func (vc *VurixWebApiClient) Logout() {
	vc.isRetryLogin = false
	vc.isLogin = false
	if vc.eventReceiveClient != nil {
		vc.eventReceiveClient.Stop()
		vc.eventReceiveClient = nil
	}
	vc.cancel() // 모든 루프가 종료될 수 있도록, Sleep을 Cancel한다.
	// 응답여부에 관계없이 로그아웃 요청을 보낸다.
	vc.client.R().
		SetHeader("x-auth-token", vc.loginInfo.AuthToken).
		Delete(vc.webApiURL + "/api/logout")
	time.Sleep(1 * time.Second) // 재접속 고루틴과, 킵얼라이브 고루틴이 모두 종료될수 있도록 명시적으로 1초 쉬어준다
}

// http://host:port/api/login
func (vc *VurixWebApiClient) loginAction() {
	resp, err := vc.client.R().
		SetHeader("x-account-id", vc.opt.User).
		SetHeader("x-account-pass", vc.opt.Pass).
		SetHeader("x-account-group", vc.opt.Group).
		SetHeader("x-license", vc.opt.License).
		Get(vc.webApiURL + "/api/login")

	if err == nil {
		if resp.StatusCode() == 200 {
			var bodyJson interface{}
			body := resp.Body()
			err := json.Unmarshal(body, &bodyJson)
			if err == nil {
				vc.loginInfo.AuthToken = cast.ToString(goutil.JsonGetValue(bodyJson, "results.auth_token"))
				vc.loginInfo.ApiSerial = cast.ToInt(goutil.JsonGetValue(bodyJson, "results.api_serial"))
				vc.loginInfo.VmsID = cast.ToInt(goutil.JsonGetValue(bodyJson, "results.vms_id"))
				vc.loginInfo.GrpSerial = cast.ToInt(goutil.JsonGetValue(bodyJson, "results.grp_serial"))
				vc.loginInfo.UserSerial = cast.ToInt(goutil.JsonGetValue(bodyJson, "results.user_serial"))
				vc.loginInfo.UserID = cast.ToString(goutil.JsonGetValue(bodyJson, "results.user_id"))
				vc.loginInfo.UserName = cast.ToString(goutil.JsonGetValue(bodyJson, "results.user_name"))
				vc.loginInfo.Utc = cast.ToBool(goutil.JsonGetValue(bodyJson, "results.utc"))
				vc.isLogin = true
				// KeepAlive를 10분에 한번씩 전송하는 루프 동작
				go vc.sendKeepAliveLoop(600 * time.Second)
			} else {
				vc.isLogin = false
			}
		} else {
			vc.isLogin = false
		}
	} else {
		vc.isLogin = false
	}
}

func (vc *VurixWebApiClient) retryLoginLoop(dTime time.Duration) {
	for vc.isRetryLogin {
		if !vc.isLogin {
			vc.logger.Infof("Retry Login !!!")
			vc.loginAction()
		}
		vc.SleepWithContext(dTime)
	}
}

func (vc *VurixWebApiClient) sendKeepAliveLoop(dTime time.Duration) {
	for vc.isLogin {
		if vc.KeepAlive() {
			// KeepAlive에 성공하였으므로 10분간 대기
			vc.SleepWithContext(dTime)
		} else {
			// KeepAlive에 실패한 것이므로 루프를 빠져나옴
			break
		}
	}
}

// http://host:port/api/keep-alive
// keep Alive에 실패할 경우, 로그인이 튕긴것이므로 다시 로그인이 되도록 처리한다
func (vc *VurixWebApiClient) KeepAlive() bool {
	if vc.isLogin {
		resp, err := vc.client.R().
			SetHeader("x-auth-token", vc.loginInfo.AuthToken).
			Get(vc.webApiURL + "/api/keep-alive")

		if err != nil {
			// 에러가 있으므로 인증실패
			vc.logger.Warnf("Keep-Alive send error : %v", err.Error())
			vc.isLogin = false
			return false
		} else {
			if resp.StatusCode() != 200 {
				// 200 OK 가 아니므로 인증 실패
				vc.logger.Warnf("Keep-Alive response error : response code : %v", resp.StatusCode())
				vc.isLogin = false
				return false
			}
		}
		return true
	}
	return false
}

func (vc *VurixWebApiClient) SleepWithContext(d time.Duration) {
	select {
	case <-vc.ctx.Done(): //context cancelled
	case <-time.After(d): //timeout
	}
}

func (vc *VurixWebApiClient) SetEventHandler(callback EventCallbackFunc, opt OptVurixEventReceiver) {
	vc.eventCallback = callback
	if vc.eventReceiveClient != nil {
		vc.eventReceiveClient.Stop()
		vc.eventReceiveClient = nil
	}
	vc.eventReceiveClient = NewVurixEventReceiver(opt)
	vc.eventReceiveClient.SetVurixWebApiClient(vc)
	vc.eventReceiveClient.Run(vc.logger)
}
