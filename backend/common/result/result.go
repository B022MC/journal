package result

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
)

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func Success(w http.ResponseWriter, data interface{}) {
	httpx.OkJsonCtx(nil, w, &Response{
		Code:    0,
		Message: "success",
		Data:    data,
	})
}

func Error(w http.ResponseWriter, code int, message string) {
	httpx.OkJsonCtx(nil, w, &Response{
		Code:    code,
		Message: message,
	})
}

func ParamError(w http.ResponseWriter, err error) {
	httpx.OkJsonCtx(nil, w, &Response{
		Code:    10002,
		Message: "参数错误: " + err.Error(),
	})
}
