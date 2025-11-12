package services

import (
	"context"
	"sort"
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

func PostCommentAdd(ctx context.Context, req *apimodels.PostComment) (errs.IMErrorCode, *apimodels.PostComment) {
	appkey := ctxs.GetAppKeyFromCtx(ctx)
	userId := ctxs.GetRequesterIdFromCtx(ctx)
	postId := req.PostId
	storage := storages.NewPostCommentStorage()
	commentId := utils.GenerateUUIDString()
	currTime := time.Now()
	parentUserId := req.ParentUserId
	if parentUserId == "" && req.ParentUserInfo != nil && req.ParentUserInfo.UserId != "" {
		parentUserId = req.ParentUserInfo.UserId
	}
	comment := models.PostComment{
		CommentId:       commentId,
		PostId:          postId,
		ParentCommentId: req.ParentCommentId,
		ParentUserId:    parentUserId,
		Text:            req.Text,
		UserId:          userId,
		CreatedTime:     currTime.UnixMilli(),
		UpdatedTime:     currTime,
		AppKey:          appkey,
	}
	err := storage.Create(comment)
	if err == nil {
		AppendCommentFeed(ctx, comment)
	}
	return errs.IMErrorCode_SUCCESS, &apimodels.PostComment{
		CommentId:   commentId,
		CreatedTime: currTime.UnixMilli(),
	}
}

func PostCommentUpdate(ctx context.Context, req *apimodels.PostComment) errs.IMErrorCode {
	appkey := ctxs.GetAccountFromCtx(ctx)
	userId := ctxs.GetRequesterIdFromCtx(ctx)
	storage := storages.NewPostCommentStorage()
	dbComment, err := storage.FindById(appkey, req.CommentId)
	if err != nil || dbComment == nil {
		return errs.IMErrorCode_APP_POST_NOTEXISTED
	}
	if dbComment.UserId != userId {
		return errs.IMErrorCode_APP_POST_NORIGHT
	}
	storage.UpdateComment(appkey, req.CommentId, req.Text)
	return errs.IMErrorCode_SUCCESS
}

func PostCommentDel(ctx context.Context, req *apimodels.PostCommentIds) errs.IMErrorCode {
	appkey := ctxs.GetAppKeyFromCtx(ctx)
	storage := storages.NewPostCommentStorage()
	for _, commentId := range req.CommentIds {
		storage.DelComment(appkey, commentId)
	}
	return errs.IMErrorCode_SUCCESS
}

func QryPostComments(ctx context.Context, postId string, startTime int64, limit int64, isPositive bool) (errs.IMErrorCode, *apimodels.PostComments) {
	appkey := ctxs.GetAppKeyFromCtx(ctx)
	userId := ctxs.GetRequesterIdFromCtx(ctx)
	ret := &apimodels.PostComments{
		Items:      []*apimodels.PostComment{},
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
	comments := []*models.PostComment{}
	if isFriendMod {
		storage := storages.NewPostCommentFeedStorage()
		dbcomments, err := storage.QryPostComments(appkey, userId, postId, startTime, limit+1, isPositive)
		if err == nil && len(dbcomments) > 0 {
			comments = append(comments, dbcomments...)
		}
	} else {
		storage := storages.NewPostCommentStorage()
		dbcomments, err := storage.QryPostComments(appkey, postId, startTime, limit+1, isPositive)
		if err == nil && len(dbcomments) > 0 {
			comments = append(comments, dbcomments...)
		}
	}

	for _, c := range comments {
		ret.Items = append(ret.Items, &apimodels.PostComment{
			PostId:          c.PostId,
			CommentId:       c.CommentId,
			Text:            c.Text,
			ParentCommentId: c.ParentCommentId,
			ParentUserInfo:  jimservices.GetUser(ctx, c.ParentUserId),
			UserInfo:        jimservices.GetUser(ctx, c.UserId),
			CreatedTime:     c.CreatedTime,
			UpdatedTime:     c.UpdatedTime.UnixMilli(),
		})
	}
	if len(ret.Items) > int(limit) {
		ret.Items = ret.Items[:limit]
		ret.IsFinished = false
	}
	return errs.IMErrorCode_SUCCESS, ret
}

func TopComments(ctx context.Context, postId string, limit int64) []*apimodels.PostComment {
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
	comments := []*models.PostComment{}
	if isFriendMod {
		storage := storages.NewPostCommentFeedStorage()
		dbcomments, err := storage.QryPostComments(appkey, userId, postId, 0, limit, false)
		if err == nil && len(dbcomments) > 0 {
			comments = append(comments, dbcomments...)
		}
	} else {
		commentStorage := storages.NewPostCommentStorage()
		dbcomments, err := commentStorage.QryPostComments(appkey, postId, 0, limit, false)
		if err == nil && len(dbcomments) > 0 {
			comments = append(comments, dbcomments...)
		}
	}

	ret := []*apimodels.PostComment{}
	for _, comment := range comments {
		ret = append(ret, &apimodels.PostComment{
			CommentId:       comment.CommentId,
			PostId:          comment.PostId,
			ParentCommentId: comment.ParentCommentId,
			Text:            comment.Text,
			ParentUserId:    comment.ParentUserId,
			ParentUserInfo:  jimservices.GetUser(ctx, comment.ParentUserId),
			UserInfo:        jimservices.GetUser(ctx, comment.UserId),
			CreatedTime:     comment.CreatedTime,
			UpdatedTime:     comment.UpdatedTime.UnixMilli(),
		})
	}
	sort.Slice(ret, func(i, j int) bool {
		return i < j
	})
	return ret
}
