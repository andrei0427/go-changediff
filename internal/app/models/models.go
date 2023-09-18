package models

type ProjectModel struct {
	ID          *int32  `form:"id"`
	Name        string  `form:"name"`
	Description string  `form:"description"`
	AccentColor string  `form:"accent_color"`
	AppKey      *string `form:"appkey"`
}

type LabelModel struct {
	ID    *int32 `form:"id"`
	Label string `form:"label"`
	Color string `form:"color"`
}

type PostModel struct {
	Id          *int64  `form:"id"`
	Title       string  `form:"title"`
	Content     string  `form:"content"`
	PublishedOn *string `form:"published_on"`
	LabelId     *int    `form:"label_id"`
	First       *bool   `form:"first"`
}
