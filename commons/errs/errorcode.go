package errs

/*
0 : success
*/

type IMErrorCode int32

var IMErrorCode_SUCCESS IMErrorCode = 0
var IMErrorCode_PBILLEGAL IMErrorCode = 1 //pb解析失败，内部错误码
var IMErrorCode_DEFAULT IMErrorCode = 2

// app errorcode
var (
	IMErrorCode_APP_DEFAULT             IMErrorCode = 17000
	IMErrorCode_APP_APPKEY_REQUIRED     IMErrorCode = 17001
	IMErrorCode_APP_NOT_EXISTED         IMErrorCode = 17002
	IMErrorCode_APP_REQ_BODY_ILLEGAL    IMErrorCode = 17003
	IMErrorCode_APP_INTERNAL_TIMEOUT    IMErrorCode = 17004
	IMErrorCode_APP_NOT_LOGIN           IMErrorCode = 17005
	IMErrorCode_APP_CONTINUE            IMErrorCode = 17006
	IMErrorCode_APP_QRCODE_EXPIRED      IMErrorCode = 17007
	IMErrorCode_APP_SMS_SEND_FAILED     IMErrorCode = 17008
	IMErrorCode_APP_SMS_CODE_EXPIRED    IMErrorCode = 17009
	IMErrorCode_APP_TRANS_NOTRANSENGINE IMErrorCode = 17010
	IMErrorCode_APP_USER_EXISTED        IMErrorCode = 17011
	IMErrorCode_APP_USER_NOT_EXIST      IMErrorCode = 17012
	IMErrorCode_APP_LOGIN_ERR_PASS      IMErrorCode = 17013
	IMErrorCode_APP_PHONE_EXISTED       IMErrorCode = 17014
	IMErrorCode_APP_EMAIL_EXIST         IMErrorCode = 17015

	//friends
	IMErrorCode_APP_FRIEND_DEFAULT         IMErrorCode = 17100
	IMErrorCode_APP_FRIEND_APPLY_DECLINE   IMErrorCode = 17101
	IMErrorCode_APP_FRIEND_APPLY_REPEATED  IMErrorCode = 17102
	IMErrorCode_APP_FRIEND_CONFIRM_EXPIRED IMErrorCode = 17103

	//group
	IMErrorCode_APP_GROUP_DEFAULT       IMErrorCode = 17200
	IMErrorCode_APP_GROUP_MEMBEREXISTED IMErrorCode = 17201
	IMErrorCode_APP_GROUP_NORIGHT       IMErrorCode = 17202

	//assistant
	IMErrorCode_APP_ASSISTANT_PROMPT_DBERROR IMErrorCode = 17300

	//file
	IMErrorCode_APP_FILE_NOOSS   IMErrorCode = 17401
	IMErrorCode_APP_FILE_SIGNERR IMErrorCode = 17402

	//post
	IMErrorCode_APP_POST_DEFAULT    IMErrorCode = 17500
	IMErrorCode_APP_POST_NOTEXISTED IMErrorCode = 17501
	IMErrorCode_APP_POST_NORIGHT    IMErrorCode = 17502
)

var imCode2ApiErrorMap map[IMErrorCode]*ApiErrorMsg = map[IMErrorCode]*ApiErrorMsg{
	//api
	IMErrorCode_SUCCESS: newApiErrorMsg(200, IMErrorCode_SUCCESS, "success"),
}

func GetApiErrorByCode(code IMErrorCode) *ApiErrorMsg {
	if err, ok := imCode2ApiErrorMap[code]; ok {
		return err
	}
	return newApiErrorMsg(200, code, "")
}

type ApiErrorMsg struct {
	HttpCode int         `json:"-"`
	Code     IMErrorCode `json:"code"`
	Msg      string      `json:"msg"`
}

func newApiErrorMsg(httpCode int, code IMErrorCode, msg string) *ApiErrorMsg {
	return &ApiErrorMsg{
		HttpCode: httpCode,
		Code:     code,
		Msg:      msg,
	}
}

type SuccHttpResp struct {
	ApiErrorMsg
	Data interface{} `json:"data"`
}
