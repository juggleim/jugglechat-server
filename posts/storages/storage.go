package storages

import (
	"github.com/juggleim/jugglechat-server/posts/storages/dbs"
	"github.com/juggleim/jugglechat-server/posts/storages/models"
)

func NewPostStorage() models.IPostStorage {
	return &dbs.PostDao{}
}

func NewPostCommentStorage() models.IPostCommentStorage {
	return &dbs.PostCommentDao{}
}

func NewPostReactionStorage() models.IPostReactionStorage {
	return &dbs.PostReactionDao{}
}

func NewPostFeedStorage() models.IPostFeedStorage {
	return &dbs.PostFeedDao{}
}

func NewPostCommentFeedStorage() models.IPostCommentFeedStorage {
	return &dbs.PostCommentFeedDao{}
}
