package registry

type ACL struct {
	Id string `json:"id" xorm:"pk UNIQUE NOT NULL"`
	Seq int `xorm:"INDEX UNIQUE NOT NULL"`
	Account string `json:"account,omitempty" xorm:"account"`
	Type string `json:"type,omitempty" xorm:"type"`
	Name string `json:"name,omitempty" xorm:"name"`
	IP string `json:"ip,omitempty" xorm:"ip"`
	Service string `json:"service,omitempty" xorm:"service"`
	Actions string `json:"actions,omitempty" xorm:"actions"`
	ActionsArray []string `xorm:"-"`
	Comment string `json:"comment,omitempty" xorm:"comment"`
}
