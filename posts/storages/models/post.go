package models

import "time"

type Post struct {
	ID           int64
	PostId       string
	Title        string
	Content      []byte
	ContentBrief string
	PostExset    []byte
	IsDelete     int
	UserId       string
	CreatedTime  int64
	UpdatedTime  time.Time
	Status       int
	AppKey       string
}

type IPostStorage interface {
	Create(item Post) error
	FindById(appkey, postId string) (*Post, error)
	FindByIds(appkey string, postIds []string) (map[string]*Post, error)
	UpdatePost(appkey, postId string, title string, content []byte, contentBrief string) error
	DelPost(appkey, postId string) error
	QryPosts(appkey string, startTime, limit int64, isPositive bool) ([]*Post, error)
}

type PostComment struct {
	ID              int64
	CommentId       string
	PostId          string
	ParentCommentId string
	ParentUserId    string
	Text            string
	IsDelete        int
	UserId          string
	CreatedTime     int64
	UpdatedTime     time.Time
	Status          int
	AppKey          string
}

type IPostCommentStorage interface {
	Create(item PostComment) error
	FindById(appkey, commentId string) (*PostComment, error)
	FindByIds(appkey string, commentIds []string) (map[string]*PostComment, error)
	UpdateComment(appkey, commentId string, text string) error
	DelComment(appkey, commentId string) error
	QryPostComments(appkey, postId string, startTime, limit int64, isPosttive bool) ([]*PostComment, error)
}

type PostBusType int

const (
	PostBusType_Post     PostBusType = 0
	PostBusType_Comment  PostBusType = 1
	PostBusType_Reaction PostBusType = 2
)

type PostReaction struct {
	BusId       string
	BusType     PostBusType
	ItemKey     string
	ItemValue   string
	CreatedTime int64
	UserId      string
	AppKey      string
}

type IPostReactionStorage interface {
	Create(item PostReaction) error
	Upsert(item PostReaction) error
	DelReaction(appkey, busId string, busType PostBusType, key string) error
	QryReactions(appkey, busId string, busType PostBusType, startId, limit int64) ([]*PostReaction, error)
}

type PostFeed struct {
	UserId   string
	PostId   string
	FeedTime int64
	AppKey   string
}

type IPostFeedStorage interface {
	Create(item PostFeed) error
	BatchCreate(items []PostFeed) error
	QryPosts(appkey, userId string, startTime, limit int64, isPositive bool) ([]*Post, error)
}

type PostCommentFeed struct {
	UserId    string
	CommentId string
	PostId    string
	FeedTime  int64
	AppKey    string
}

type IPostCommentFeedStorage interface {
	Create(item PostCommentFeed) error
	BatchCreate(items []PostCommentFeed) error
	QryPostComments(appkey, userId, postId string, startTime, limit int64, isPositive bool) ([]*PostComment, error)
}
