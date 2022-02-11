package vurixwebapiclient

type OptVurixEventReceiver struct {
	DeviceEvent     bool    // event
	SystemEvent     bool    // system
	MonitoringEvent bool    // monitoring
	UserLogEvent    bool    // userlog
	UserEvent       bool    // user
	VfsEvent        bool    // vfs
	MasEvent        bool    // mas
}

type VurixEventReceiver struct {
	opt OptVurixEventReceiver
}

func NewVurixEventReceiver(option OptVurixEventReceiver) VurixEventReceiver {
	return VurixEventReceiver{
		opt: option,
	}
}
