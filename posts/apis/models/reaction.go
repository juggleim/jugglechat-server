package models

import (
	jimApiModels "github.com/juggleim/jugglechat-server/apis/models"
)

type PostReaction struct {
	PostId    string                `json:"post_id,omitempty"`
	Key       string                `json:"key"`
	Value     string                `json:"value"`
	Timestamp int64                 `json:"timestamp"`
	UserInfo  *jimApiModels.UserObj `json:"user_info,omitempty"`
}

type PostReactions struct {
	Reactions []*PostReaction `json:"reactions"`
}
