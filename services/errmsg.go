package services

import "fmt"

type PageInfo struct {
	Page  int `json:"page"`
	Count int `json:"count"`
}

type ErrorCode int

const (
	ErrorCode_Success = 0
	ErrorCode_Unknown = 10000

	//100xx 基本参数校验
	ErrorCode_ParseIntFail       = 10001
	ErrorCode_IdDecodeFail       = 10002
	ErrorCode_ParamRequired      = 10003
	ErrorCode_ParamErr           = 10004
	ErrorCode_PlatformNotSupport = 10005
	ErrorCode_ButtonNotExist     = 10006
	ErrorCode_TxtAuditFail       = 10007

	//102xx  支付相关
	ErrorCode_RemainNotEnough = 10200 //余额不足
	ErrorCode_AddOrderFail    = 10201 //创建订单失败
	ErrorCode_PayReqFail      = 10202 //转账接口调用失败
	ErrorCode_PayFail         = 10203 //支付接口返回了错误码
	ErrorCode_PrepayCallErr   = 10204

	//103xx 消息相关
	ErrorCode_MsgIdDecodeFail = 10300 //Id解码失败
	ErrorCode_MsgSaveFail     = 10301 //消息入库失败
	ErrorCode_MsgDecodeFail   = 10302

	//104xx 模型相关
	ErrorCode_ModelCreateFail = 10400
	ErrorCode_ModelNotFound   = 10401
	ErrorCode_ModelUpdateFail = 10402
	ErrorCode_ModelDeleteFail = 10403
	ErrorCode_EmbeddingsFail  = 10404

	//105xx  用户相关
	ErrorCode_NoWxJsCode         = 10501
	ErrorCode_WxLoginFail        = 10502
	ErrorCode_WxLoginRespErr     = 10503
	ErrorCode_TokenErr           = 10504
	ErrorCode_NotLogin           = 10505
	ErrorCode_NoUid              = 10506
	ErrorCode_UserIdIs0          = 10507
	ErrorCode_UidStrError        = 10508
	ErrorCode_UserDbReadFail     = 10509
	ErrorCode_UserDbInsertFail   = 10510
	ErrorCode_UserDbUpdateFail   = 10511
	ErrorCode_UserDbDelFail      = 10512
	ErrorCode_InteractionExhaust = 10513 // 交互次数超限
	ErrorCode_SmsSendFail        = 10514 //短信发送失败
	ErrorCode_UserHaveNoRight    = 10515 //用户没有操作权限
	ErrorCode_GenImgExhaust      = 10516 //图片生成次数超限

	//106xx 群组相关
	ErrorCode_GrpDbInsertFail = 10600
	ErrorCode_GrpDbQryFail    = 10601
	ErrorCode_Sync2ImFail     = 10602
)

var errMsgMap map[ErrorCode]string = map[ErrorCode]string{
	ErrorCode_Success: "success",
	ErrorCode_Unknown: "unknown error",

	ErrorCode_NoWxJsCode: "js_code is required",
}

type CommonError struct {
	Code     int    `json:"code"`
	ErrorMsg string `json:"msg"`
}
type CommonResp struct {
	CommonError
	Data interface{} `json:"data,omitempty"`
}

func (err *CommonError) Error() string {
	return fmt.Sprintf("%d:%s", err.Code, err.ErrorMsg)
}

func GetError(code ErrorCode) error {
	if msg, ok := errMsgMap[code]; ok {
		return &CommonError{
			Code:     int(code),
			ErrorMsg: msg,
		}
	}
	return &CommonError{
		Code:     int(code),
		ErrorMsg: "unknown error",
	}
}

func GetCoustomErr(errMsg string) error {
	return &CommonError{
		Code:     ErrorCode_Unknown,
		ErrorMsg: errMsg,
	}
}

func GetSuccess() interface{} {
	return &CommonError{
		Code:     int(ErrorCode_Success),
		ErrorMsg: "success",
	}
}

func SuccessResp(obj interface{}) *CommonResp {
	return &CommonResp{
		CommonError: CommonError{
			Code:     ErrorCode_Success,
			ErrorMsg: "success",
		},
		Data: obj,
	}
}
