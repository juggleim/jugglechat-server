package services

import (
	"context"
	"fmt"

	"github.com/juggleim/jugglechat-server/commons/appinfos"
	"github.com/juggleim/jugglechat-server/commons/ctxs"
	"github.com/juggleim/jugglechat-server/posts/storages"
	"github.com/juggleim/jugglechat-server/posts/storages/models"
	jimStorages "github.com/juggleim/jugglechat-server/storages"
)

func AppendPostFeed(ctx context.Context, post models.Post) {
	appkey := ctxs.GetAppKeyFromCtx(ctx)
	userId := ctxs.GetRequesterIdFromCtx(ctx)

	isFriendMod := false
	appinfo, exist := appinfos.GetAppInfo(appkey)
	if exist && appinfo != nil {
		if exist, val := appinfo.GetExt("post_mode"); exist {
			if v, ok := val.(string); ok && v == "1" {
				isFriendMod = true
			}
		}
	}
	storage := storages.NewPostFeedStorage()
	storage.Create(models.PostFeed{
		UserId:   userId,
		PostId:   post.PostId,
		FeedTime: post.CreatedTime,
		AppKey:   appkey,
	})
	ForeachFriends(appkey, userId, func(friendIds []string) {
		postFeeds := []models.PostFeed{}
		for _, friendId := range friendIds {
			postFeeds = append(postFeeds, models.PostFeed{
				UserId:   friendId,
				PostId:   post.PostId,
				FeedTime: post.CreatedTime,
				AppKey:   appkey,
			})
		}
		if len(postFeeds) > 0 {
			err := storage.BatchCreate(postFeeds)
			if err != nil {
				fmt.Println(err)
			}
			if isFriendMod {
				SendPostNotify(ctx, friendIds, &PostNtfMsg{
					SponsorId:   userId,
					PostBusType: models.PostBusType_Post,
				})
			}
		}
	})
}

func AppendCommentFeed(ctx context.Context, comment models.PostComment) {
	appkey := ctxs.GetAppKeyFromCtx(ctx)
	userId := ctxs.GetRequesterIdFromCtx(ctx)

	isFriendMod := false
	appinfo, exist := appinfos.GetAppInfo(appkey)
	if exist && appinfo != nil {
		if exist, val := appinfo.GetExt("post_mode"); exist {
			if v, ok := val.(string); ok && v == "1" {
				isFriendMod = true
			}
		}
	}
	storage := storages.NewPostCommentFeedStorage()
	storage.Create(models.PostCommentFeed{
		UserId:    userId,
		CommentId: comment.CommentId,
		PostId:    comment.PostId,
		FeedTime:  comment.CreatedTime,
		AppKey:    appkey,
	})
	ForeachFriends(appkey, userId, func(friendIds []string) {
		commentFeeds := []models.PostCommentFeed{}
		for _, friendId := range friendIds {
			commentFeeds = append(commentFeeds, models.PostCommentFeed{
				UserId:    friendId,
				CommentId: comment.CommentId,
				PostId:    comment.PostId,
				FeedTime:  comment.CreatedTime,
				AppKey:    appkey,
			})
		}
		if len(commentFeeds) > 0 {
			storage.BatchCreate(commentFeeds)
			if isFriendMod {
				SendPostNotify(ctx, friendIds, &PostNtfMsg{
					SponsorId:   userId,
					PostBusType: models.PostBusType_Comment,
				})
			}
		}
	})
}

func ForeachFriends(appkey, userId string, f func(friendIds []string)) {
	var startId int64 = 0
	var limit int64 = 100
	storage := jimStorages.NewFriendRelStorage()
	for {
		rels, err := storage.QueryFriendRels(appkey, userId, startId, limit)
		if err != nil {
			fmt.Println("err:", err)
			break
		}
		friendIds := []string{}
		for _, rel := range rels {
			friendIds = append(friendIds, rel.FriendId)
		}
		f(friendIds)
		if len(rels) < int(limit) {
			break
		}
	}
}
