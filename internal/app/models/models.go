package models

type ProjectModel struct {
	Name        string `form:"name"`
	Description string `form:"description"`
	AccentColor string `form:"accent_color"`
}

type PostModel struct {
	Id          *int64  `form:"id"`
	Title       string  `form:"title"`
	Content     string  `form:"content"`
	PublishedOn *string `form:"published_on"`
	LabelId     *int    `form:"label_id"`
	First       *bool   `form:"first"`
}
