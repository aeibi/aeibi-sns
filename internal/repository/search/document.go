package search

type PostDocument struct {
	UID             string   `json:"uid"`
	AuthorUID       string   `json:"author_uid"`
	AuthorNickname  string   `json:"author_nickname"`
	Text            string   `json:"text"`
	TagNames        []string `json:"tag_names"`
	Images          []string `json:"images,omitempty"`
	Attachments     []string `json:"attachments,omitempty"`
	ImageCount      int      `json:"image_count"`
	AttachmentCount int      `json:"attachment_count"`
	CommentCount    int      `json:"comment_count"`
	CollectionCount int      `json:"collection_count"`
	LikeCount       int      `json:"like_count"`
	Pinned          bool     `json:"pinned"`
	Visibility      string   `json:"visibility"` // PUBLIC / PRIVATE
	Status          string   `json:"status"`     // NORMAL / ARCHIVED
	LatestRepliedOn int64    `json:"latest_replied_on"`
	CreatedAt       int64    `json:"created_at"`
	UpdatedAt       int64    `json:"updated_at"`
}

type UserDocument struct {
	UID         string `json:"uid"`
	Nickname    string `json:"nickname"`
	AvatarUrl   string `json:"avatar_url"`
	Description string `json:"description"`
	Status      string `json:"status"`
}

type TagDocument struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
