package services

import (
	"appserver/dbs"
	"appserver/serversdk"
	"appserver/utils"
	"time"
)

type Group struct {
	GroupId       string  `json:"group_id"`
	GroupName     string  `json:"group_name"`
	GroupPortrait string  `json:"group_portrait"`
	GrpMembers    []*User `json:"members"`
}

type Groups struct {
	Items []*Group `json:"items"`
}

func UpdateGroup(grp Group) ErrorCode {
	grpId, err := utils.Decode(grp.GroupId)
	if err != nil || grpId == 0 {
		return ErrorCode_ParseIntFail
	}
	groupDao := dbs.GroupDao{}
	err = groupDao.UpdateGrpInfo(grpId, grp.GroupName, grp.GroupPortrait)
	if err != nil {
		return ErrorCode_UserDbUpdateFail
	}
	//sync to im
	UpdateGroupInfo2Im(serversdk.GroupInfo{
		GroupId:       grp.GroupId,
		GroupName:     grp.GroupName,
		GroupPortrait: grp.GroupPortrait,
	})
	return ErrorCode_Success
}

func CreateGroup(curUid string, grp Group) (ErrorCode, *Group) {
	grpDao := dbs.GroupDao{}
	grpId, err := grpDao.Create(dbs.GroupDao{
		GroupName:     grp.GroupName,
		GroupPortrait: grp.GroupPortrait,
		CreatedTime:   time.Now(),
		UpdatedTime:   time.Now(),
	})
	if err != nil {
		return ErrorCode_GrpDbInsertFail, nil
	}
	curUserIdInt, _ := utils.Decode(curUid)
	memberIdStrs := []string{}
	if len(grp.GrpMembers) > 0 {
		memberDao := dbs.GroupMemberDao{}
		dbMembers := []dbs.GroupMemberDao{}
		needAddSelf := true
		for _, member := range grp.GrpMembers {
			if member.UserId == curUid {
				needAddSelf = false
			}
			memberIdStrs = append(memberIdStrs, member.UserId)
			userIdInt, err := utils.Decode(member.UserId)
			if err == nil {
				dbMembers = append(dbMembers, dbs.GroupMemberDao{
					GroupId:  grpId,
					MemberId: userIdInt,
				})
			}
		}
		if needAddSelf && curUserIdInt > 0 {
			dbMembers = append(dbMembers, dbs.GroupMemberDao{
				GroupId:  grpId,
				MemberId: curUserIdInt,
			})
			memberIdStrs = append(memberIdStrs, curUid)
		}
		if len(dbMembers) > 0 {
			memberDao.BatchCreate(dbMembers)
		}
	}
	// add to im
	grpIdStr, _ := utils.Encode(grpId)

	code := CreateGroup2Im(serversdk.GroupMembersReq{
		GroupId:       grpIdStr,
		GroupName:     grp.GroupName,
		GroupPortrait: grp.GroupPortrait,
		MemberIds:     memberIdStrs,
	})
	if code != ErrorCode_Success {
		return code, nil
	}
	return ErrorCode_Success, &Group{
		GroupId:       grpIdStr,
		GroupName:     grp.GroupName,
		GroupPortrait: grp.GroupPortrait,
	}
}

func DelGroupMembers(grp Group) ErrorCode {
	grpId, err := utils.Decode(grp.GroupId)
	if err != nil {
		return ErrorCode_IdDecodeFail
	}
	memberIds := []string{}
	if len(grp.GrpMembers) > 0 {
		memberDao := dbs.GroupMemberDao{}
		delMemberIds := []int64{}
		for _, member := range grp.GrpMembers {
			memberIds = append(memberIds, member.UserId)
			userIdInt, err := utils.Decode(member.UserId)
			if err == nil {
				delMemberIds = append(delMemberIds, userIdInt)
			}
		}
		if len(delMemberIds) > 0 {
			memberDao.BatchDelete(grpId, delMemberIds)
		}
	}
	if len(memberIds) > 0 {
		code := DelGroupMembers2Im(serversdk.GroupMembersReq{
			GroupId:   grp.GroupId,
			MemberIds: memberIds,
		})
		if code != ErrorCode_Success {
			return code
		}
	}
	return ErrorCode_Success
}

func AddGroupMembers(grp Group) ErrorCode {
	grpId, err := utils.Decode(grp.GroupId)
	if err != nil {
		return ErrorCode_IdDecodeFail
	}
	memberIds := []string{}
	if len(grp.GrpMembers) > 0 {
		memberDao := dbs.GroupMemberDao{}
		dbMembers := []dbs.GroupMemberDao{}
		for _, member := range grp.GrpMembers {
			memberIds = append(memberIds, member.UserId)
			userIdInt, err := utils.Decode(member.UserId)
			if err == nil {
				dbMembers = append(dbMembers, dbs.GroupMemberDao{
					GroupId:  grpId,
					MemberId: userIdInt,
				})
			} else {
				break
			}
		}
		if len(grp.GrpMembers) != len(dbMembers) {
			return ErrorCode_IdDecodeFail
		}
		if len(dbMembers) > 0 {
			memberDao.BatchCreate(dbMembers)
		}
	}
	if len(memberIds) > 0 {
		code := AddGroupMembers2Im(serversdk.GroupMembersReq{
			GroupId:   grp.GroupId,
			MemberIds: memberIds,
		})
		if code != ErrorCode_Success {
			return code
		}
	}
	return ErrorCode_Success
}

func QryGroup(groupId string) (ErrorCode, *Group) {
	//groupInfo
	grpId, err := utils.Decode(groupId)
	if err != nil {
		return ErrorCode_IdDecodeFail, nil
	}
	grpDao := dbs.GroupDao{}
	grpDb, err := grpDao.FindById(grpId)
	if err != nil {
		return ErrorCode_GrpDbQryFail, nil
	}
	grpIdStr, _ := utils.Encode(grpDb.ID)
	grp := &Group{
		GroupId:       grpIdStr,
		GroupName:     grpDb.GroupName,
		GroupPortrait: grpDb.GroupPortrait,
		GrpMembers:    []*User{},
	}
	//groupMembers
	grpMemberDao := dbs.GroupMemberDao{}
	grpMembers, err := grpMemberDao.QueryMembers(grpId, 0, 1000)
	if err != nil {
		return ErrorCode_GrpDbQryFail, nil
	}
	userDao := dbs.UserDao{}
	for _, member := range grpMembers {
		userIdStr, _ := utils.Encode(member.MemberId)
		userDb := userDao.FindByUserId(member.MemberId)
		if userDb != nil {
			grp.GrpMembers = append(grp.GrpMembers, &User{
				UserId:   userIdStr,
				Nickname: userDb.Nickname,
				Avatar:   userDb.Avatar,
			})
		}
	}
	return ErrorCode_Success, grp
}

func QryMyGroups(curUid string, startId string, count int64) (ErrorCode, *Groups) {
	retGrps := &Groups{
		Items: []*Group{},
	}
	userIdInt, err := utils.Decode(curUid)
	if err != nil || userIdInt <= 0 {
		return ErrorCode_Success, retGrps
	}
	grpMemberDao := dbs.GroupMemberDao{}
	startIdInt, err := utils.Decode(startId)
	if err != nil || startIdInt <= 0 {
		startIdInt = 0
	}
	grpMemRels, err := grpMemberDao.QueryGroupsByMemberId(userIdInt, startIdInt, count)
	if err == nil {
		grpIds := []int64{}
		for _, grpRel := range grpMemRels {
			grpIds = append(grpIds, grpRel.GroupId)
		}
		if len(grpIds) > 0 {
			grpDao := dbs.GroupDao{}
			grpInfos, err := grpDao.FindByIds(grpIds)
			if err == nil {
				for _, grpInfo := range grpInfos {
					grpIdStr, _ := utils.Encode(grpInfo.ID)
					retGrps.Items = append(retGrps.Items, &Group{
						GroupId:       grpIdStr,
						GroupName:     grpInfo.GroupName,
						GroupPortrait: grpInfo.GroupPortrait,
					})
				}
			}
		}
	}
	return ErrorCode_Success, retGrps
}
