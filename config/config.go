package config

//var (
//	ApplicationVersion string = "development"
//)

type Endpoint struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}
type Service struct {
	Local    Endpoint `json:"local"`
	Remote   Endpoint `json:"remote"`
	Server   Endpoint `json:"server"`
	Name     string   `json:"name"`
	Username string   `json:"username"`
}

type Group struct {
	Name     string    `json:"name"`
	Services []Service `json:"services"`
}

type Config struct {
	Groups []Group `json:"groups"`
}
