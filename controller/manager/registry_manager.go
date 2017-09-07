package manager

import (
	"github.com/go-xorm/xorm"
	registryModels "github.com/BoxLinker/boxlinker-api/controller/models/registry"
)

type RegistryManager interface {
	QueryAllACL() ([]*registryModels.ACL, error)
	SaveACL(acl *registryModels.ACL) error
}

type DefaultRegistryManager struct {
	engine *xorm.Engine
}

func NewRegistryManager(engine *xorm.Engine) (RegistryManager, error) {
	return &DefaultRegistryManager{
		engine: engine,
	}, nil
}

func (dm *DefaultRegistryManager) SaveACL(acl *registryModels.ACL) error {
	sess := dm.engine.NewSession()
	defer sess.Close()
	if _, err := sess.Insert(acl); err != nil {
		return err
	}
	return sess.Commit()
}

func (dm *DefaultRegistryManager) QueryAllACL() ([]*registryModels.ACL, error) {
	var acls []*registryModels.ACL
	if err := dm.engine.Find(&acls); err != nil {
		return nil, err
	}
	return acls, nil
}