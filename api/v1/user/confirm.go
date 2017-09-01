package user

import (
	"net/http"
	"github.com/BoxLinker/boxlinker-api/auth"
	"github.com/BoxLinker/boxlinker-api/controller/models"
	"github.com/Sirupsen/logrus"
)

func (a *Api) ConfirmEmail(w http.ResponseWriter, r *http.Request) {
	confirmToken := r.URL.Query().Get("confirm_token")
	if len(confirmToken) == 0 {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	ok, result, err := auth.AuthToken(confirmToken)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if !ok || result == nil {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	uid := result["uid"].(string)
	username := result["username"].(string)

	u, err := a.manager.GetUserToBeConfirmed(uid, username)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if u == nil {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	u1 := &models.User{
		Name: u.Name,
		Email: u.Email,
		Password: u.Password,
	}

	if err := a.manager.SaveUser(u1); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := a.manager.DeleteUsersToBeConfirmedByName(u.Name); err != nil {
		// to be continue
		logrus.Warnf("DeleteUsersToBeConfirmedByName err: %v, after save user", err)
	}

	w.Write([]byte("confirm user success: "+u1.Id+" "+ u1.Name))

}

