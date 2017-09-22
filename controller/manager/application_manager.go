package manager

import "github.com/go-xorm/xorm"

type ApplicationManager interface {
	Manager

}

type DefaultApplicationManager struct {
	DefaultManager
	engine *xorm.Engine
}

func NewApplicationManager(engine *xorm.Engine) (ApplicationManager, error) {
	return &DefaultApplicationManager{
		engine: engine,
	}, nil
}
