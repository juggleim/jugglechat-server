package models

type Group struct {
	GroupId       string `json:"group_id"`
	GroupName     string `json:"group_name"`
	GroupPortrait string `json:"group_portrait"`
	MemberCount   int    `json:"member_count"`
	Owner         *User  `json:"owner,omitempty"`

	CreatedTime int64 `json:"created_time"`
}

type Groups struct {
	Items  []*Group `json:"items"`
	Offset string   `json:"offset,omitempty"`
}

type GroupIds struct {
	AppKey   string   `json:"app_key"`
	GroupIds []string `json:"group_ids"`
}
