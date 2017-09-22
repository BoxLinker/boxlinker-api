package application

type HardwareConfigure struct {
	Id string `xorm:"pk"`
	Name string `xorm:"VARCHAR(255) NOT NULL"`
	Memory int64 `xorm:"NOT NULL"`
	CPU int `xorm:"NOT NULL"`
}
