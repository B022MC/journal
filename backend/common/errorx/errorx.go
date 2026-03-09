package errorx

const (
	OK              = 0
	ErrInternal     = 10001
	ErrInvalidParam = 10002
	ErrUnauthorized = 10003
	ErrForbidden    = 10004
	ErrNotFound     = 10005
	ErrDuplicate    = 10006

	// User errors
	ErrUserNotFound  = 20001
	ErrPasswordWrong = 20002
	ErrUsernameTaken = 20003
	ErrEmailTaken    = 20004
	ErrUserBanned    = 20005

	// Paper errors
	ErrPaperNotFound     = 30001
	ErrInvalidDiscipline = 30002
	ErrInvalidZone       = 30003

	// Rating errors
	ErrAlreadyRated = 40001
	ErrSelfRating   = 40002
	ErrInvalidScore = 40003

	// Flag errors
	ErrAlreadyFlagged  = 50001
	ErrSelfFlag        = 50002
	ErrFlagNotFound    = 50003
	ErrDegradedContent = 50004
	ErrRateLimited     = 50005
)

var codeMsg = map[int]string{
	OK:                   "success",
	ErrInternal:          "内部错误 / Internal error",
	ErrInvalidParam:      "参数错误 / Invalid parameter",
	ErrUnauthorized:      "未授权 / Unauthorized",
	ErrForbidden:         "权限不足 / Forbidden",
	ErrNotFound:          "资源不存在 / Not found",
	ErrDuplicate:         "重复操作 / Duplicate",
	ErrUserNotFound:      "用户不存在 / User not found",
	ErrPasswordWrong:     "密码错误 / Wrong password",
	ErrUsernameTaken:     "用户名已被占用 / Username taken",
	ErrEmailTaken:        "邮箱已被占用 / Email taken",
	ErrUserBanned:        "用户已被封禁 / User banned",
	ErrPaperNotFound:     "论文不存在 / Paper not found",
	ErrInvalidDiscipline: "无效学科分类 / Invalid discipline",
	ErrInvalidZone:       "无效分区 / Invalid zone",
	ErrAlreadyRated:      "已评分 / Already rated",
	ErrSelfRating:        "不能给自己的论文评分 / Cannot rate own paper",
	ErrInvalidScore:      "评分无效(1-10) / Invalid score",
	ErrAlreadyFlagged:    "已举报 / Already flagged",
	ErrSelfFlag:          "不能举报自己 / Cannot flag yourself",
	ErrFlagNotFound:      "举报不存在 / Flag not found",
	ErrDegradedContent:   "内容已被降解 / Content degraded",
	ErrRateLimited:       "请求频率过高 / Rate limited",
}

func CodeMsg(code int) string {
	if msg, ok := codeMsg[code]; ok {
		return msg
	}
	return "unknown error"
}
