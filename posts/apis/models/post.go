package models

import (
	jimApiModels "github.com/juggleim/jugglechat-server/apis/models"
)

type Posts struct {
	Items      []*Post `json:"items"`
	IsFinished bool    `json:"is_finished"`
}

type Post struct {
	PostId      string                     `json:"post_id"`
	Content     *PostContent               `json:"content"`
	UserInfo    *jimApiModels.UserObj      `json:"user_info"`
	CreatedTime int64                      `json:"created_time"`
	UpdatedTime int64                      `json:"updated_time"`
	Reactions   map[string][]*PostReaction `json:"reactions"`
	TopComments []*PostComment             `json:"top_comments"`
}

type PostIds struct {
	PostIds []string `json:"post_ids"`
}

type PostCommentIds struct {
	CommentIds []string `json:"comment_ids"`
}

// type Reaction struct {
// 	Value    string                `json:"value"`
// 	UserInfo *jimApiModels.UserObj `json:"user_info"`
// }

type PostContent struct {
	Text   string              `json:"text"`
	Images []*PostContentImage `json:"images"`
	Video  *PostContentVideo   `json:"video"`
}

type PostContentImage struct {
	Url         string `json:"url"`
	TumbnailUrl string `json:"thumbnail_url"`
	Height      int    `json:"height"`
	Width       int    `json:"width"`
}

type PostContentVideo struct {
	Url         string `json:"url"`
	SnapshotUrl string `json:"snapshot_url"`
	Duration    int    `json:"duration"`
	Height      int    `json:"height"`
	Width       int    `json:"width"`
}

type PostComment struct {
	CommentId       string                `json:"comment_id"`
	PostId          string                `json:"post_id"`
	ParentCommentId string                `json:"parent_comment_id"`
	Text            string                `json:"text"`
	ParentUserId    string                `json:"parent_user_id,omitempty"`
	ParentUserInfo  *jimApiModels.UserObj `json:"parent_user_info"`
	UserInfo        *jimApiModels.UserObj `json:"user_info"`
	CreatedTime     int64                 `json:"created_time"`
	UpdatedTime     int64                 `json:"updated_time"`
}

type PostComments struct {
	Items      []*PostComment `json:"items"`
	IsFinished bool           `json:"is_finished"`
}
