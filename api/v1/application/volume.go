package application

import (
	"net/http"
	"github.com/BoxLinker/boxlinker-api"
	"github.com/BoxLinker/boxlinker-api/controller/models"
	"github.com/gorilla/mux"
)

type VolumeForm struct {
	Name string `json:"name"`
	Size string `json:"size"`
}

func (a *Api) CreateVolume(w http.ResponseWriter, r *http.Request){
	user := a.getUserInfo(r)
	form := &VolumeForm{}
	if err := boxlinker.ReadRequestBody(r, form); err != nil {
		boxlinker.Resp(w, boxlinker.STATUS_FORM_VALIDATE_ERR, nil, err.Error())
		return
	}
	claim, err := a.manager.CreateVolume(user.Name, &models.Volume{
		Name: form.Name,
		Size: form.Size,
	})
	if err != nil {
		boxlinker.Resp(w, boxlinker.STATUS_INTERNAL_SERVER_ERR, nil, err.Error())
		return
	}
	boxlinker.Resp(w, boxlinker.STATUS_OK, claim)
}
func (a *Api) DeleteVolume(w http.ResponseWriter, r *http.Request){
	user := a.getUserInfo(r)
	name := mux.Vars(r)["name"]
	if err := a.manager.DeleteVolume(user.Name, name); err != nil {
		boxlinker.Resp(w, boxlinker.STATUS_NOT_FOUND, nil, err.Error())
		return
	}
	boxlinker.Resp(w, boxlinker.STATUS_OK, nil)
}
func (a *Api) QueryVolume(w http.ResponseWriter, r *http.Request){
	user := a.getUserInfo(r)
	claims, err := a.manager.QueryVolume(user.Name, boxlinker.ParsePageConfig(r))
	if err != nil {
		boxlinker.Resp(w, boxlinker.STATUS_NOT_FOUND, nil, err.Error())
		return
	}
	boxlinker.Resp(w, boxlinker.STATUS_OK, claims)
}
