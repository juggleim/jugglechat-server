package dbs

import (
	"bytes"
	"fmt"
	"time"

	"github.com/juggleim/jugglechat-server/commons/dbcommons"
	"github.com/juggleim/jugglechat-server/storages/models"
)

type GroupMemberDao struct {
	ID             int64     `gorm:"primary_key"`
	GroupId        string    `gorm:"group_id"`
	MemberId       string    `gorm:"member_id"`
	MemberType     int       `gorm:"member_type"`
	CreatedTime    time.Time `gorm:"created_time"`
	AppKey         string    `gorm:"app_key"`
	IsMute         int       `gorm:"is_mute"`
	IsAllow        int       `gorm:"is_allow"`
	MuteEndAt      int64     `gorm:"mute_end_at"`
	GrpDisplayName string    `gorm:"grp_display_name"`
}

func (msg GroupMemberDao) TableName() string {
	return "groupmembers"
}

func (msg GroupMemberDao) Create(item models.GroupMember) error {
	err := dbcommons.GetDb().Create(&GroupMemberDao{
		GroupId:        item.GroupId,
		MemberId:       item.MemberId,
		MemberType:     item.MemberType,
		CreatedTime:    time.Now(),
		AppKey:         item.AppKey,
		IsMute:         item.IsMute,
		IsAllow:        item.IsAllow,
		MuteEndAt:      item.MuteEndAt,
		GrpDisplayName: item.GrpDisplayName,
	}).Error
	return err
}

func (member GroupMemberDao) Find(appkey, groupId, memberId string) (*models.GroupMember, error) {
	var item GroupMemberDao
	err := dbcommons.GetDb().Where("app_key=? and group_id=? and member_id=?", appkey, groupId, memberId).Take(&item).Error
	if err != nil {
		return nil, err
	}
	return &models.GroupMember{
		ID:             item.ID,
		GroupId:        item.GroupId,
		MemberId:       item.MemberId,
		MemberType:     item.MemberType,
		CreatedTime:    item.CreatedTime,
		AppKey:         item.AppKey,
		IsMute:         item.IsMute,
		IsAllow:        item.IsAllow,
		MuteEndAt:      item.MuteEndAt,
		GrpDisplayName: item.GrpDisplayName,
	}, nil
}

func (member GroupMemberDao) FindByMemberIds(appkey, groupId string, memberIds []string) ([]*models.GroupMember, error) {
	var items []*GroupMemberDao
	err := dbcommons.GetDb().Where("app_key=? and group_id=? and member_id in (?)", appkey, groupId, memberIds).Find(&items).Error
	ret := []*models.GroupMember{}
	for _, item := range items {
		ret = append(ret, &models.GroupMember{
			ID:             item.ID,
			GroupId:        item.GroupId,
			MemberId:       item.MemberId,
			MemberType:     item.MemberType,
			CreatedTime:    item.CreatedTime,
			AppKey:         item.AppKey,
			IsMute:         item.IsMute,
			IsAllow:        item.IsAllow,
			MuteEndAt:      item.MuteEndAt,
			GrpDisplayName: item.GrpDisplayName,
		})
	}
	return ret, err
}

func (member GroupMemberDao) BatchCreate(items []models.GroupMember) error {
	var buffer bytes.Buffer
	sql := fmt.Sprintf("insert into %s (`group_id`,`member_id`,`app_key`)values", member.TableName())

	buffer.WriteString(sql)
	for i, item := range items {
		if i == len(items)-1 {
			buffer.WriteString(fmt.Sprintf("('%s','%s','%s');", item.GroupId, item.MemberId, item.AppKey))
		} else {
			buffer.WriteString(fmt.Sprintf("('%s','%s','%s'),", item.GroupId, item.MemberId, item.AppKey))
		}
	}
	err := dbcommons.GetDb().Exec(buffer.String()).Error
	return err
}

type GroupMemberWithUser struct {
	GroupMemberDao
	Nickname     string `gorm:"nickname"`
	UserPortrait string `gorm:"user_portrait"`

	FriendDisplayName string `gorm:"friend_display_name"`
	FriendId          string `gorm:"friend_id"`
}

func (member GroupMemberDao) QueryMembers(appkey, userId, groupId string, startId, limit int64) ([]*models.GroupMember, error) {
	sql := fmt.Sprintf("select m.*,u.nickname,u.user_portrait,u.user_type,f.friend_id,f.display_name as friend_display_name from %s as m left join %s as u on m.app_key=u.app_key and m.member_id=u.user_id left join %s as f on m.app_key=f.app_key and f.user_id=? and f.friend_id=m.member_id where m.app_key=? and m.group_id=? and m.id>? order by m.id asc limit ?", member.TableName(), UserDao{}.TableName(), FriendRelDao{}.TableName())
	var items []*GroupMemberWithUser
	err := dbcommons.GetDb().Raw(sql, userId, appkey, groupId, startId, limit).Scan(&items).Error
	ret := []*models.GroupMember{}
	for _, item := range items {
		friendInfo := &models.FriendInfo{}
		if item.FriendId != "" {
			friendInfo.IsFriend = true
			friendInfo.DisplayName = item.FriendDisplayName
		}
		ret = append(ret, &models.GroupMember{
			ID:               item.ID,
			GroupId:          item.GroupId,
			MemberId:         item.MemberId,
			MemberType:       item.MemberType,
			CreatedTime:      item.CreatedTime,
			AppKey:           item.AppKey,
			IsMute:           item.IsMute,
			IsAllow:          item.IsAllow,
			MuteEndAt:        item.MuteEndAt,
			GrpDisplayName:   item.GrpDisplayName,
			Nickname:         item.Nickname,
			UserPortrait:     item.UserPortrait,
			MemberFriendInfo: friendInfo,
		})
	}
	return ret, err
}

func (member GroupMemberDao) SearchMembersByName(appkey, userId, groupId, nickname string, startId, limit int64) ([]*models.GroupMember, error) {
	sql := fmt.Sprintf("select m.*,u.nickname,u.user_portrait,u.user_type,f.friend_id,f.display_name as friend_display_name from %s as m left join %s as u on m.app_key=u.app_key and m.member_id=u.user_id left join %s as f on m.app_key=f.app_key and f.user_id=? and f.friend_id=m.member_id where m.app_key=? and m.group_id=? and m.id>? and u.nickname like ? order by m.id asc limit ?", member.TableName(), UserDao{}.TableName(), FriendRelDao{}.TableName())
	var items []*GroupMemberWithUser
	err := dbcommons.GetDb().Raw(sql, userId, appkey, groupId, startId, "%"+nickname+"%", limit).Scan(&items).Error
	ret := []*models.GroupMember{}
	for _, item := range items {
		friendInfo := &models.FriendInfo{}
		if item.FriendId != "" {
			friendInfo.IsFriend = true
			friendInfo.DisplayName = item.FriendDisplayName
		}
		ret = append(ret, &models.GroupMember{
			ID:               item.ID,
			GroupId:          item.GroupId,
			MemberId:         item.MemberId,
			MemberType:       item.MemberType,
			CreatedTime:      item.CreatedTime,
			AppKey:           item.AppKey,
			IsMute:           item.IsMute,
			IsAllow:          item.IsAllow,
			MuteEndAt:        item.MuteEndAt,
			GrpDisplayName:   item.GrpDisplayName,
			Nickname:         item.Nickname,
			UserPortrait:     item.UserPortrait,
			MemberFriendInfo: friendInfo,
		})
	}
	return ret, err
}

type GroupMemberWithGroup struct {
	GroupMemberDao
	GroupName     string `gorm:"group_name"`
	GroupPortrait string `gorm:"group_portrait"`
}

func (member GroupMemberDao) QueryGroupsByMemberId(appkey, memberId string, startId, limit int64) ([]*models.GroupMember, error) {
	sql := fmt.Sprintf("select m.*,g.group_name,g.group_portrait from %s as m left join %s as g on m.app_key=g.app_key and m.group_id=g.group_id where m.app_key=? and m.member_id=? and m.id>? order by m.id asc limit ?", member.TableName(), GroupDao{}.TableName())
	var items []*GroupMemberWithGroup
	err := dbcommons.GetDb().Raw(sql, appkey, memberId, startId, limit).Scan(&items).Error
	ret := []*models.GroupMember{}
	for _, item := range items {
		ret = append(ret, &models.GroupMember{
			ID:             item.ID,
			GroupId:        item.GroupId,
			GroupName:      item.GroupName,
			GroupPortrait:  item.GroupPortrait,
			MemberId:       item.MemberId,
			MemberType:     item.MemberType,
			CreatedTime:    item.CreatedTime,
			AppKey:         item.AppKey,
			IsMute:         item.IsMute,
			IsAllow:        item.IsAllow,
			MuteEndAt:      item.MuteEndAt,
			GrpDisplayName: item.GrpDisplayName,
		})
	}
	return ret, err
}

func (member GroupMemberDao) SearchGroupsByMemberId(appkey, memberId string, keyword string, startId, limit int64) ([]*models.GroupMember, error) {
	sql := fmt.Sprintf("select m.*,g.group_name,g.group_portrait from %s as m left join %s as g on m.app_key=g.app_key and m.group_id=g.group_id where m.app_key=? and m.member_id=? and m.id>? and g.group_name like ? order by m.id asc limit ?", member.TableName(), GroupDao{}.TableName())
	var items []*GroupMemberWithGroup
	err := dbcommons.GetDb().Raw(sql, appkey, memberId, startId, "%"+keyword+"%", limit).Scan(&items).Error
	ret := []*models.GroupMember{}
	for _, item := range items {
		ret = append(ret, &models.GroupMember{
			ID:             item.ID,
			GroupId:        item.GroupId,
			GroupName:      item.GroupName,
			GroupPortrait:  item.GroupPortrait,
			MemberId:       item.MemberId,
			MemberType:     item.MemberType,
			CreatedTime:    item.CreatedTime,
			AppKey:         item.AppKey,
			IsMute:         item.IsMute,
			IsAllow:        item.IsAllow,
			MuteEndAt:      item.MuteEndAt,
			GrpDisplayName: item.GrpDisplayName,
		})
	}
	return ret, err
}

func (member GroupMemberDao) BatchDelete(appkey, groupId string, memberIds []string) error {
	return dbcommons.GetDb().Where("app_key=? and group_id=? and member_id in (?)", appkey, groupId, memberIds).Delete(&GroupMemberDao{}).Error
}

func (member GroupMemberDao) DeleteByGroupId(appkey, groupId string) error {
	return dbcommons.GetDb().Where("app_key=? and group_id=?", appkey, groupId).Delete(&GroupMemberDao{}).Error
}

func (member GroupMemberDao) UpdateMute(appkey, groupId string, isMute int, memberIds []string, muteEndAt int64) error {
	upd := map[string]interface{}{}
	upd["is_mute"] = isMute
	if isMute == 0 {
		upd["mute_end_at"] = 0
	} else {
		upd["mute_end_at"] = muteEndAt
	}
	return dbcommons.GetDb().Model(&GroupMemberDao{}).Where("app_key=? and group_id=? and member_id in (?)", appkey, groupId, memberIds).Updates(upd).Error
}

func (member GroupMemberDao) UpdateAllow(appkey, groupId string, isAllow int, memberIds []string) error {
	return dbcommons.GetDb().Model(&member).Where("app_key=? and group_id=? and member_id in (?)", appkey, groupId, memberIds).Update("is_allow", isAllow).Error
}

func (member GroupMemberDao) CountByGroup(appkey, groupId string) int {
	var count int64
	err := dbcommons.GetDb().Model(&GroupMemberDao{}).Where("app_key=? and group_id=?", appkey, groupId).Count(&count).Error
	if err != nil {
		return 0
	}
	return int(count)
}

func (member GroupMemberDao) UpdateGrpDisplayName(appkey, groupId, memberId string, displayName string) error {
	return dbcommons.GetDb().Model(&member).Where("app_key=? and group_id=? and member_id=?", appkey, groupId, memberId).Update("grp_display_name", displayName).Error
}
