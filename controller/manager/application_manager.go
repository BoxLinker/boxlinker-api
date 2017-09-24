package manager

import (
	"github.com/go-xorm/xorm"
	appModels "github.com/BoxLinker/boxlinker-api/controller/models/application"
	"github.com/Sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	appsv1beta1 "k8s.io/api/apps/v1beta1"
	extv1beta1 "k8s.io/api/extensions/v1beta1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"fmt"
)

type ApplicationManager interface {
	Manager
	SyncPodConfigure(pcs []*appModels.PodConfigure) (int, error)
	GetServiceByName(namespace, svcName string) (bool, error, *apiv1.Service, *extv1beta1.Ingress, *appsv1beta1.Deployment)
}

type DefaultApplicationManager struct {
	DefaultManager
	engine *xorm.Engine
	clientSet *kubernetes.Clientset
}

func (m *DefaultApplicationManager) GetServiceByName(namespace, svcName string) (bool, error, *apiv1.Service, *extv1beta1.Ingress, *appsv1beta1.Deployment) {
	var (
		err error
		svc *apiv1.Service
		ing *extv1beta1.Ingress
		deploy *appsv1beta1.Deployment
	)
	if svc, err = m.clientSet.CoreV1().Services(namespace).Get(svcName, metav1.GetOptions{}); err != nil {
		return false, fmt.Errorf("Service %s/%s not found: %v", namespace, svcName, err), nil, nil, nil
	}
	if ing, err = m.clientSet.ExtensionsV1beta1().Ingresses(namespace).Get(svcName, metav1.GetOptions{}); err != nil {
		return false, fmt.Errorf("Ingress %s/%s not found: %v", namespace, svcName, err), nil, nil, nil
	}
	if deploy, err = m.clientSet.AppsV1beta1().Deployments(namespace).Get(svcName, metav1.GetOptions{}); err != nil {
		return false, fmt.Errorf("Deployment %s/%s not found: %v", namespace, svcName, err), nil, nil, nil
	}
	return true, nil, svc, ing, deploy
}
func (m *DefaultApplicationManager) SyncPodConfigure(pcs []*appModels.PodConfigure) (int, error) {
	sess := m.engine.NewSession()
	defer sess.Close()
	i := 0
	for _, pc := range pcs {
		if _, err := sess.Insert(pc); err != nil {
			logrus.Warnf("Sync PodConfigure (%+v) failed (%v)", pc, err)
		} else {
			i++
			logrus.Debugf("Sync PodConfigure (%+v)", pc)
		}
	}
	return i, sess.Commit()
}

func NewApplicationManager(engine *xorm.Engine, clientSet *kubernetes.Clientset) (ApplicationManager, error) {
	return &DefaultApplicationManager{
		engine: engine,
		clientSet: clientSet,
	}, nil
}
