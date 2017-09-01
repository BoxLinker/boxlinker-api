package authz

import (
	"github.com/BoxLinker/boxlinker-api/controller/manager"
	"sync"
)

type ACLMysqlAuthorizer struct {
	manager manager.RegistryManager
	lock sync.RWMutex
}

func NewACLMysqlAuthorizer(manager manager.RegistryManager) Authorizer {
	return &ACLMysqlAuthorizer{
		manager: manager,
	}
}

func (acl *ACLMysqlAuthorizer) Authorize(ai *AuthRequestInfo) ([]string, error) {

}

func (acl *ACLMysqlAuthorizer) Stop(){}
func (acl *ACLMysqlAuthorizer) Name() string {
	return "ACLMysqlAuthorizer"
}