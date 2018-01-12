package application

import (
	"net/http"
	"github.com/gorilla/mux"
	"github.com/BoxLinker/boxlinker-api"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiv1 "k8s.io/api/core/v1"
	"fmt"
)
// 这个接口一般是由管理员调用的，验证统一使用 user token，但是指定 username = boxlinker
// 创建 namespace 的同时，还必须创建 registry-key
func (a *Api) Event(w http.ResponseWriter, r *http.Request) {
	u := a.getUserInfo(r)
	if u.Name != "boxlinker" {
		boxlinker.Resp(w, boxlinker.STATUS_UNAUTHORIZED, nil, "only admin can operate")
		return
	}
	tType := mux.Vars(r)["type"]
	if tType == "reg_message" {
		var regMsg struct{
			Username string `json:"username"`
		}
		if err := boxlinker.ReadRequestBody(r, &regMsg); err != nil {
			boxlinker.Resp(w, boxlinker.STATUS_PARAM_ERR, nil, fmt.Sprintf("获取 ns 参数错误: %v", err))
			return
		}
		ns := regMsg.Username
		if len(ns) == 0 {
			boxlinker.Resp(w, boxlinker.STATUS_PARAM_ERR, nil, "需要 ns 参数")
			return
		}
		nsClient := a.clientSet.Namespaces()
		if _, err := nsClient.Get(ns, metav1.GetOptions{}); err == nil { // err == nil 说明找到了
			boxlinker.Resp(w, boxlinker.STATUS_FAILED, nil, fmt.Sprintf("%s 已经存在", ns))
			return
		}

		if _, err := nsClient.Create(&apiv1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: ns,
			},
		}); err != nil {
			boxlinker.Resp(w, boxlinker.STATUS_INTERNAL_SERVER_ERR, nil, fmt.Sprintf("create namespace %s err: %v", ns, err))
			return
		}
		// TODO 这里需要拿到用户的明文用户名和密码，得有个安全的解决方式
		// 但毕竟 registry-key 的 secret 文件因为是 base64 编码 也不够安全

		// 创建 registry-key
		a.clientSet.Secrets(ns).Create(&apiv1.Secret{
			Type: apiv1.SecretTypeDockerConfigJson,
			ObjectMeta: metav1.ObjectMeta{
				Name: "registry-key",
				Namespace: ns,
			},
			Data: map[string][]byte{
				".dockerconfigjson": []byte(""),
			},
		})
		boxlinker.Resp(w, boxlinker.STATUS_OK, map[string]string{
			"namespace": ns,
		})
		return
	}
	boxlinker.Resp(w, boxlinker.STATUS_FAILED, nil, fmt.Sprintf("unknow type %s", tType))
}