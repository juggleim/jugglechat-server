package services

import "github.com/juggleim/jugglechat-server/commons/appinfos"

type PostMod string

const (
	PostMod_Global PostMod = "global"
	PostMod_Friend PostMod = "friend"
)

func GetPostMod(appkey string) PostMod {
	appinfo, exist := appinfos.GetAppInfo(appkey)
	if exist && appinfo != nil {
		exist, val := appinfo.GetExtByCreator("jchat_post_mod", func(val string) interface{} {
			if val != "" {
				switch val {
				case "global":
					return PostMod_Global
				case "friend":
					return PostMod_Friend
				}
			}
			return PostMod_Global
		})
		if exist && val != nil {
			return PostMod(val.(string))
		}
	}
	return PostMod_Global
}
