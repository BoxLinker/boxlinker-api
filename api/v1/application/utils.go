package application

import (
	appv1beta1 "k8s.io/api/extensions/v1beta1"
	"errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/api/core/v1"
)

var ErrNotFound = errors.New("resource not found")

func (a *Api) getIngresses(namespace string) ([]*appv1beta1.Ingress) {
	ings, err := a.clientSet.ExtensionsV1beta1().Ingresses(namespace).List(metav1.ListOptions{})
	if err != nil {
		return nil
	}
	results := make([]*appv1beta1.Ingress, 0, len(ings.Items))
	for _, item := range ings.Items {
		results = append(results, &item)
	}
	return results
}

func (a *Api) getIngressByName(namespace string, ingressName string) (*appv1beta1.Ingress, error) {
	ings := a.getIngresses(namespace)
	if ings == nil {
		return nil, ErrNotFound
	}
	for _, item := range ings {
		if item.Name == ingressName {
			return item, nil
		}
	}
	return nil, ErrNotFound
}

func (a *Api) findPathByPortAndSvcName(svcName string, port v1.ServicePort, ing *appv1beta1.Ingress) string {
	if ing == nil {
		return ""
	}
	rules := ing.Spec.Rules
	if len(rules) > 0 {
		paths := rules[0].HTTP.Paths
		for _, path := range paths {
			if path.Backend.ServiceName == svcName && path.Backend.ServicePort.IntVal == port.Port {
				return path.Path
			}
		}
	}
	return ""
}
