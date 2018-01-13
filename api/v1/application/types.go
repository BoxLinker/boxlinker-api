package application

type PortResult struct {
	Port int `json:"port"`
	Protocol string `json:"protocol"`
	Path string `json:"path"`
}

type ServiceResult struct {
	Name string `json:"name"`
	Image string `json:"image"`
	Memory string `json:"memory"`
	Host string `json:"host"`
	Ports []*PortResult `json:"ports"`
}

