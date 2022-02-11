package vurixwebapiclient

type OptVurixWebApiClient struct {
    Host string
    Port int
    User string
    Pass string
    Group string
    License string
}

func NewOptVurixWebApiClient() OptVurixWebApiClient {
    return OptVurixWebApiClient{
        Host: "",
        Port: 8080,
        User: "",
        Pass: "",
        Group: "group1",
        License: "licNormalClient",
    }
}