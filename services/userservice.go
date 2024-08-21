package services

import (
	"appserver/dbs"
	"appserver/utils"
	"fmt"
	"time"

	imsdk "github.com/juggleim/imserver-sdk-go"

	"github.com/golang-jwt/jwt/v4"
)

func QryUserInfo(curUid, userId string) (ErrorCode, *User) {
	retUser := &User{}
	userDao := dbs.UserDao{}
	userIdInt, err := utils.Decode(userId)
	if err != nil || userIdInt <= 0 {
		return ErrorCode_IdDecodeFail, nil
	}
	userdb := userDao.FindByUserId(userIdInt)
	if userdb != nil {
		var isFriend bool = false
		curUidInt, err := utils.Decode(curUid)
		if err == nil && curUidInt > 0 {
			friendDao := dbs.FriendDao{}
			isFriend = friendDao.CheckFriend(curUidInt, userdb.ID)
		}
		retUser.UserId = userId
		retUser.Nickname = userdb.Nickname
		retUser.Avatar = userdb.Avatar
		retUser.Phone = userdb.Phone
		retUser.IsFriend = isFriend
	}
	return ErrorCode_Success, retUser
}

func GetUserInfo(userId string) *User {
	retUser := &User{
		UserId: userId,
	}
	userIdInt, err := utils.Decode(userId)
	if err != nil || userIdInt <= 0 {
		return retUser
	}
	dbUser := dbs.UserDao{}.FindByUserId(userIdInt)
	if dbUser != nil {
		retUser.Nickname = dbUser.Nickname
		retUser.Avatar = dbUser.Avatar
	}
	return retUser
}

func SearchByPhone(curUid, phone string) (ErrorCode, *Users) {
	userDao := dbs.UserDao{}
	userdb, err := userDao.FindByPhone(phone)
	if err != nil {
		return ErrorCode_UserDbReadFail, nil
	}
	var isFriend bool = false
	curUidInt, err := utils.Decode(curUid)
	if err == nil && curUidInt > 0 {
		friendDao := dbs.FriendDao{}
		isFriend = friendDao.CheckFriend(curUidInt, userdb.ID)

	}
	idStr, _ := utils.Encode(userdb.ID)
	users := &Users{
		Items: []*User{},
	}
	users.Items = append(users.Items, &User{
		UserId:   idStr,
		Nickname: userdb.Nickname,
		Avatar:   userdb.Avatar,
		Phone:    userdb.Phone,
		Status:   userdb.Status,
		IsFriend: isFriend,
	})
	return ErrorCode_Success, users
}

func UpdateUser(user User) ErrorCode {
	userDao := dbs.UserDao{}
	id, err := utils.Decode(user.UserId)
	if err != nil || id == 0 {
		return ErrorCode_ParseIntFail
	}
	err = userDao.Update(dbs.UserDao{
		ID:       id,
		Phone:    user.Phone,
		Nickname: user.Nickname,
		Avatar:   user.Avatar,
	})
	if err != nil {
		return ErrorCode_UserDbUpdateFail
	}
	//sync to im
	RegisterImToken(imsdk.User{
		UserId:       user.UserId,
		Nickname:     user.Nickname,
		UserPortrait: user.Avatar,
	})
	return ErrorCode_Success
}
func RegisterOrLoginBySms(user User) (string, *User, error) {
	userDao := dbs.UserDao{}
	retUser := &User{}
	var dbId int64
	userdb, err := userDao.FindByPhone(user.Phone)
	if err != nil { //入库
		dbId, err = userDao.Create(dbs.UserDao{
			Phone:    user.Phone,
			Nickname: user.Nickname,
			Avatar:   user.Avatar,
			Status:   0,
		})
		if err != nil {
			return "", nil, GetError(ErrorCode_UserDbInsertFail)
		}
		retUser.Nickname = user.Nickname
		retUser.Avatar = user.Avatar
		retUser.Status = 0
	} else {
		dbId = userdb.ID
		retUser.Nickname = userdb.Nickname
		retUser.Avatar = userdb.Avatar
		retUser.Status = userdb.Status
		retUser.ImToken = userdb.ImToken
	}
	if dbId > 0 {
		idStr, _ := utils.Encode(dbId)
		retUser.UserId = idStr
		auth, _ := generateAuthorization(idStr)
		imToken := RegisterImToken(imsdk.User{
			UserId:       idStr,
			Nickname:     retUser.Nickname,
			UserPortrait: retUser.Avatar,
		})
		retUser.ImToken = imToken
		return auth, retUser, nil
	} else {
		return "", nil, GetError(ErrorCode_UserIdIs0)
	}
}

type Users struct {
	Items []*User `json:"items"`
}
type User struct {
	UserId   string `json:"user_id"`
	Nickname string `json:"nickname"`
	Phone    string `json:"phone,omitempty"`
	Avatar   string `json:"avatar,omitempty"`
	Status   int    `json:"status"`
	City     string `json:"city,omitempty"`
	Country  string `json:"country,omitempty"`
	Language string `json:"language,omitempty"`
	Province string `json:"province,omitempty"`
	ImToken  string `json:"im_token,omitempty"`
	IsFriend bool   `json:"is_friend"`
}
type LoginUserResp struct {
	UserId        string `json:"user_id"`
	Authorization string `json:"authorization"`
	NickName      string `json:"nickname"`
	Avatar        string `json:"avatar"`
	Status        int    `json:"status"`
	ImToken       string `json:"im_token,omitempty"`
}

var jwtkey = []byte("appserve")

type Claims struct {
	Account string
	jwt.RegisteredClaims
}

func generateAuthorization(account string) (string, error) {
	expireTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		Account: account,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: &jwt.NumericDate{
				Time: expireTime,
			},
			IssuedAt: &jwt.NumericDate{
				Time: time.Now(),
			},
			Issuer:  "aabbcc",
			Subject: "user token",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtkey)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func ValidateAuthorization(authorization string) (string, error) {
	token, claims, err := parseToken(authorization)
	if err != nil || !token.Valid {
		return "", fmt.Errorf("auth fail")
	}
	return claims.Account, nil
}

func parseToken(tokenString string) (*jwt.Token, *Claims, error) {
	Claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, Claims, func(token *jwt.Token) (i interface{}, err error) {
		return jwtkey, nil
	})
	return token, Claims, err
}
