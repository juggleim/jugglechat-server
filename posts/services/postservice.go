package services

import (
	"context"
	"time"

	"github.com/juggleim/jugglechat-server/commons/appinfos"
	"github.com/juggleim/jugglechat-server/commons/ctxs"
	"github.com/juggleim/jugglechat-server/commons/errs"
	utils "github.com/juggleim/jugglechat-server/commons/tools"
	apimodels "github.com/juggleim/jugglechat-server/posts/apis/models"
	"github.com/juggleim/jugglechat-server/posts/storages"
	"github.com/juggleim/jugglechat-server/posts/storages/models"
	jimservices "github.com/juggleim/jugglechat-server/services"
)

func PostAdd(ctx context.Context, req *apimodels.Post) (errs.IMErrorCode, *apimodels.Post) {
	appkey := ctxs.GetAppKeyFromCtx(ctx)
	userId := ctxs.GetRequesterIdFromCtx(ctx)
	storage := storages.NewPostStorage()
	postId := utils.GenerateUUIDString()
	createdTime := time.Now().UnixMilli()
	post := models.Post{
		PostId:      postId,
		Content:     []byte(utils.ToJson(req.Content)),
		UserId:      userId,
		CreatedTime: createdTime,
		UpdatedTime: time.Now(),
		AppKey:      appkey,
	}
	err := storage.Create(post)
	if err == nil {
		AppendPostFeed(ctx, post)
	}
	return errs.IMErrorCode_SUCCESS, &apimodels.Post{
		PostId:      postId,
		CreatedTime: createdTime,
	}
}

func PostUpdate(ctx context.Context, req *apimodels.Post) errs.IMErrorCode {
	appkey := ctxs.GetAppKeyFromCtx(ctx)
	userId := ctxs.GetRequesterIdFromCtx(ctx)
	storage := storages.NewPostStorage()
	dbPost, err := storage.FindById(appkey, req.PostId)
	if err != nil || dbPost == nil {
		return errs.IMErrorCode_APP_POST_NOTEXISTED
	}
	if dbPost.UserId != userId {
		return errs.IMErrorCode_APP_POST_NORIGHT
	}
	storage.UpdatePost(appkey, req.PostId, "", []byte(utils.ToJson(req.Content)), "")
	return errs.IMErrorCode_SUCCESS
}

func PostDel(ctx context.Context, req *apimodels.PostIds) errs.IMErrorCode {
	appkey := ctxs.GetAppKeyFromCtx(ctx)
	storage := storages.NewPostStorage()
	for _, postId := range req.PostIds {
		storage.DelPost(appkey, postId)
	}
	return errs.IMErrorCode_SUCCESS
}

func QryPosts(ctx context.Context, startTime int64, limit int64, isPositive bool) (errs.IMErrorCode, *apimodels.Posts) {
	appkey := ctxs.GetAppKeyFromCtx(ctx)
	userId := ctxs.GetRequesterIdFromCtx(ctx)
	ret := &apimodels.Posts{
		Items:      []*apimodels.Post{},
		IsFinished: true,
	}

	isFriendMod := false
	appinfo, exist := appinfos.GetAppInfo(appkey)
	if exist && appinfo != nil {
		if exist, val := appinfo.GetExt("post_mode"); exist {
			if v, ok := val.(string); ok && v == "1" {
				isFriendMod = true
			}
		}
	}
	posts := []*models.Post{}
	if isFriendMod {
		storage := storages.NewPostFeedStorage()
		dbposts, err := storage.QryPosts(appkey, userId, startTime, limit+1, isPositive)
		if err == nil {
			posts = append(posts, dbposts...)
		}
	} else {
		storage := storages.NewPostStorage()
		dbposts, err := storage.QryPosts(appkey, startTime, limit+1, isPositive)
		if err == nil {
			posts = append(posts, dbposts...)
		}
	}
	for _, p := range posts {
		postContent := &apimodels.PostContent{}
		utils.JsonUnMarshal(p.Content, postContent)
		post := &apimodels.Post{
			PostId:      p.PostId,
			Content:     postContent,
			UserInfo:    jimservices.GetUser(ctx, p.UserId),
			CreatedTime: p.CreatedTime,
			UpdatedTime: p.UpdatedTime.UnixMilli(),
		}
		//top comments
		post.TopComments = TopComments(ctx, p.PostId, 10)
		//top reactions
		post.Reactions = TopReactions(ctx, p.PostId, 100)
		ret.Items = append(ret.Items, post)
	}
	if len(ret.Items) > int(limit) {
		ret.Items = ret.Items[:limit]
		ret.IsFinished = false
	}
	return errs.IMErrorCode_SUCCESS, ret
}

func QryPostInfo(ctx context.Context, postId string) (errs.IMErrorCode, *apimodels.Post) {
	appkey := ctxs.GetAppKeyFromCtx(ctx)
	ret := &apimodels.Post{}
	storage := storages.NewPostStorage()
	post, err := storage.FindById(appkey, postId)
	if err == nil {
		postContent := &apimodels.PostContent{}
		err = utils.JsonUnMarshal(post.Content, postContent)
		if err == nil {
			ret.Content = postContent
		}
		ret.PostId = post.PostId
		ret.CreatedTime = post.CreatedTime
		ret.UpdatedTime = post.UpdatedTime.UnixMilli()
		ret.UserInfo = jimservices.GetUser(ctx, post.UserId)
		//comments
		ret.TopComments = TopComments(ctx, postId, 10)
		//reactions
		ret.Reactions = TopReactions(ctx, postId, 100)
	}
	return errs.IMErrorCode_SUCCESS, ret
}
