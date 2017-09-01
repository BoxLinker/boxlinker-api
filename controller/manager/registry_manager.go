package manager

import "github.com/go-xorm/xorm"

type RegistryManager interface {

}

type DefaultRegistryManager struct {
	engine *xorm.Engine
}

func NewRegistryManager(engine *xorm.Engine) (RegistryManager, error) {
	return &DefaultRegistryManager{
		engine: engine,
	}, nil
}