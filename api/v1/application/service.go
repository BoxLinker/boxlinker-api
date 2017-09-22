package application

import (
	"net/http"
	"github.com/BoxLinker/boxlinker-api"
	appsv1beta1 "k8s.io/api/apps/v1beta1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"fmt"
	"github.com/Sirupsen/logrus"
)

type DeploymentForm struct {
	Name string `json:"name"`
	Image string `json:"image"`
	HardwareConfigure int `json:"hardwareConfigure"`
}

func (a *Api) CreateService(w http.ResponseWriter, r *http.Request){
	user := a.getUserInfo(r)
	form := &DeploymentForm{}
	if err := boxlinker.ReadRequestBody(r, form); err != nil {
		boxlinker.Resp(w, boxlinker.STATUS_FORM_VALIDATE_ERR, nil, err.Error())
		return
	}

	deploymentsClient := a.clientSet.AppsV1beta1().Deployments(user.Name)

	deployment := &appsv1beta1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: form.Name,
		},
		Spec: appsv1beta1.DeploymentSpec{
			Replicas: int32Ptr(1),
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": form.Name,
					},
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:  fmt.Sprintf("%s-%s", form.Name, "container"),
							Image: form.Image,
							Ports: []apiv1.ContainerPort{
								{
									Name:          "http",
									Protocol:      apiv1.ProtocolTCP,
									ContainerPort: 80,
								},
							},
						},
					},
				},
			},
		},
	}

	logrus.Debugf("Create Deployment %s/%s (%+v)", user.Name, form)

	result, err := deploymentsClient.Create(deployment)
	if err != nil {
		boxlinker.Resp(w, boxlinker.STATUS_INTERNAL_SERVER_ERR, nil, err.Error())
		return
	}

	logrus.Debugf("Created deployment %q.\n", result.GetObjectMeta().GetName())

	boxlinker.Resp(w, boxlinker.STATUS_OK, nil)
}

