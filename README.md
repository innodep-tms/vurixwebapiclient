# vurix-webapi-client
Innodep Vurix Webapi Client

Golang을 이용한 이노뎁(주)사의 WebAPI 클라이언트를 구현함
## 현재 처리된 내용

	1. 로그인
	2. KeepAlive (세션유지)
	3. 이벤트 수신

## 사용예
```go
	opt := NewOptVurixWebApiClient()
	opt.Host = "172.16.31.202"
	opt.Port = 8080
	opt.User = "admin"
	opt.Pass = "admin"
	opt.License = "licNormalClient"

	vc := NewVurixWebApiClient(opt)
	vc.Login()   // 로그인
	vc.Logout()  // 로그아웃
```
로그인이 완료되면 기본적으로 KeepAlive는 자동으로 처리된다

## 외부 logger 사용
```go
	opt := NewOptVurixWebApiClient()
	opt.Host = "172.16.31.202"
	opt.Port = 8080
	opt.User = "admin"
	opt.Pass = "admin"
	opt.License = "licNormalClient"

	vc := NewVurixWebApiClient(opt)

	logger, _ := zap.NewDevelopment()   // Uber 로거를 만들어서 붙임
	vc.SetLogger(logger.Sugar())        // logger
	vc.SetDebug(true)                   // HTTP호출 상세로그를 기록하도록, 설정하는 플래그
	vc.Login()   // 로그인

	vc.Logout()  // 로그아웃
```

## 이벤트 수신
```go
	opt := NewOptVurixWebApiClient()
	opt.Host = "172.16.31.202"
	opt.Port = 8080
	opt.User = "admin"
	opt.Pass = "admin"
	opt.License = "licNormalClient"

	vc := NewVurixWebApiClient(opt)
	vc.Login()

	// 이벤트를 붙이는 작업을 해야 한다.
	optVER := OptVurixEventReceiver{
		DeviceEvent:     true,
		MonitoringEvent: true,
		SystemEvent:     true,
		UserEvent:       true,
	}
	// 콜백함수와 함께 이벤트 옵션을 추가
	vc.SetEventHandler(CallbackFunc, optVER)

// 수신받을 Callback 함수를 선언
func CallbackFunc(msg interface{}) {
	jsonmsg, _ := json.Marshal(msg)
	fmt.Println("recv : ", time.Now(), string(jsonmsg))
}
```