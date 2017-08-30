package boxlinker

import (
	"net/http"
	"encoding/json"
)

func Resp(w http.ResponseWriter, status int, results interface{}, msg ...string){
	var (
		err error
		outS []byte
	)

	msgS := ""
	if len(msg) == 1 {
		msgS = msg[0]
	}

	outS, err = json.Marshal(map[string]interface{}{
		"status": status,
		"msg": msgS,
		"results": results,
	})
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.WriteHeader(200)
	w.Write(outS)
}
