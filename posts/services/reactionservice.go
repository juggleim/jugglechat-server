package services

import (
	"context"
	"fmt"
	"time"

	"github.com/juggleim/jugglechat-server/commons/ctxs"
	"github.com/juggleim/jugglechat-server/commons/errs"
	apimodels "github.com/juggleim/jugglechat-server/posts/apis/models"
	"github.com/juggleim/jugglechat-server/posts/storages"
	"github.com/juggleim/jugglechat-server/posts/storages/models"
	jimservices "github.com/juggleim/jugglechat-server/services"
)

func AddPostReaction(ctx context.Context, req *apimodels.PostReaction) errs.IMErrorCode {
	appkey := ctxs.GetAppKeyFromCtx(ctx)
	userId := ctxs.GetRequesterIdFromCtx(ctx)
	storage := storages.NewPostReactionStorage()
	err := storage.Create(models.PostReaction{
		BusId:       req.PostId,
		BusType:     models.PostBusType_Post,
		ItemKey:     req.Key,
		ItemValue:   req.Value,
		CreatedTime: time.Now().UnixMilli(),
		UserId:      userId,
		AppKey:      appkey,
	})
	if err != nil {
		return errs.IMErrorCode_APP_POST_DEFAULT
	}
	return errs.IMErrorCode_SUCCESS
}

func DelPostReaction(ctx context.Context, req *apimodels.PostReaction) errs.IMErrorCode {
	appkey := ctxs.GetAppKeyFromCtx(ctx)
	storage := storages.NewPostReactionStorage()
	err := storage.DelReaction(appkey, req.PostId, models.PostBusType_Post, req.Key)
	if err != nil {
		fmt.Println(err)
		return errs.IMErrorCode_APP_POST_DEFAULT
	}
	return errs.IMErrorCode_SUCCESS
}

func QryPostReactions(ctx context.Context, postId string) (errs.IMErrorCode, *apimodels.PostReactions) {
	appkey := ctxs.GetAppKeyFromCtx(ctx)
	storage := storages.NewPostReactionStorage()
	reactions, err := storage.QryReactions(appkey, postId, models.PostBusType_Post, 0, 100)
	if err != nil {
		return errs.IMErrorCode_APP_POST_DEFAULT, nil
	}
	ret := &apimodels.PostReactions{
		Reactions: []*apimodels.PostReaction{},
	}
	for _, reaction := range reactions {
		ret.Reactions = append(ret.Reactions, &apimodels.PostReaction{
			PostId:    reaction.BusId,
			Key:       reaction.ItemKey,
			Value:     reaction.ItemValue,
			Timestamp: reaction.CreatedTime,
			UserInfo:  jimservices.GetUser(ctx, reaction.UserId),
		})
	}
	return errs.IMErrorCode_SUCCESS, ret
}

func TopReactions(ctx context.Context, postId string, limit int64) map[string][]*apimodels.PostReaction {
	appkey := ctxs.GetAppKeyFromCtx(ctx)
	ret := map[string][]*apimodels.PostReaction{}
	storage := storages.NewPostReactionStorage()
	reactions, err := storage.QryReactions(appkey, postId, models.PostBusType_Post, 0, limit)
	if err == nil {
		for _, reaction := range reactions {
			var newArr []*apimodels.PostReaction
			if arr, exist := ret[reaction.ItemKey]; exist {
				newArr = arr
			} else {
				newArr = []*apimodels.PostReaction{}
			}
			newArr = append(newArr, &apimodels.PostReaction{
				PostId:    reaction.BusId,
				Key:       reaction.ItemKey,
				Value:     reaction.ItemValue,
				Timestamp: reaction.CreatedTime,
				UserInfo:  jimservices.GetUser(ctx, reaction.UserId),
			})
			ret[reaction.ItemKey] = newArr
		}
	}
	return ret
}
