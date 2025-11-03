package services

import (
	"context"
	"fmt"

	"github.com/juggleim/commons/errs"
	"github.com/juggleim/commons/imsdk"
	"github.com/juggleim/commons/tools"
	apimodels "github.com/juggleim/jugglechat-server/admins/apis/models"
	"github.com/juggleim/jugglechat-server/storages"
)

func QryGroups(ctx context.Context, appkey, groupId, name string, offset string, limit int64, isPositive bool) (errs.AdminErrorCode, *apimodels.Groups) {
	var startId int64 = 0
	var err error
	if offset != "" {
		startId, err = tools.DecodeInt(offset)
		if err != nil {
			startId = 0
		}
	}
	ret := &apimodels.Groups{
		Items: []*apimodels.Group{},
	}
	storage := storages.NewGroupStorage()
	memberStorage := storages.NewGroupMemberStorage()
	if groupId != "" {
		grp, err := storage.FindById(appkey, groupId)
		if err == nil && grp != nil {
			apiGrp := &apimodels.Group{
				GroupId:       grp.GroupId,
				GroupName:     grp.GroupName,
				GroupPortrait: grp.GroupPortrait,
				Owner:         QryUserInfo(appkey, grp.CreatorId),
				CreatedTime:   grp.CreatedTime.UnixMilli(),
			}
			apiGrp.MemberCount = memberStorage.CountByGroup(appkey, groupId)
			ret.Items = append(ret.Items, apiGrp)
		}
	} else {
		grps, err := storage.QryGroups(appkey, name, startId, limit, isPositive)
		if err == nil {
			for _, grp := range grps {
				ret.Offset, _ = tools.EncodeInt(grp.ID)
				apiGrp := &apimodels.Group{
					GroupId:       grp.GroupId,
					GroupName:     grp.GroupName,
					GroupPortrait: grp.GroupPortrait,
					Owner:         QryUserInfo(appkey, grp.CreatorId),
					CreatedTime:   grp.CreatedTime.UnixMilli(),
				}
				apiGrp.MemberCount = memberStorage.CountByGroup(appkey, grp.GroupId)
				ret.Items = append(ret.Items, apiGrp)
			}
		}
	}
	return errs.AdminErrorCode_Success, ret
}

func QryGroupInfo(appkey, groupId string) *apimodels.Group {
	storage := storages.NewGroupStorage()
	grp, err := storage.FindById(appkey, groupId)
	if err != nil || grp == nil {
		return &apimodels.Group{
			GroupId: groupId,
		}
	}
	return &apimodels.Group{
		GroupId:       grp.GroupId,
		GroupName:     grp.GroupName,
		GroupPortrait: grp.GroupPortrait,
	}
}

func DissolveGroups(ctx context.Context, req *apimodels.GroupIds) errs.AdminErrorCode {
	fmt.Println(tools.ToJson(req))
	appkey := req.AppKey
	sdk := imsdk.GetImSdk(appkey)
	storage := storages.NewGroupStorage()
	memberStorage := storages.NewGroupMemberStorage()
	for _, groupId := range req.GroupIds {
		err := storage.Delete(appkey, groupId)
		fmt.Println("xxx:", err)
		err = memberStorage.DeleteByGroupId(appkey, groupId)
		if sdk != nil {
			sdk.DissolveGroup(groupId)
		}
		fmt.Println("yyy:", err)
	}
	return errs.AdminErrorCode_Success
}
