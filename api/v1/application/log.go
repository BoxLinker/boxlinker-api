package application

import (
	"context"
	"github.com/olivere/elastic"
	"net/http"
)

func (a *Api) Log(w http.ResponseWriter, r *http.Request){

	ctx := context.Background()

	elastic.NewClient()

}
