package models

type OnboardingModel struct {
	Name        string `form:"name"`
	Description string `form:"description"`
	AccentColor string `form:"accent_color"`
}

type PostModel struct {
	Id      *int64 `form:"id"`
	Title   string `form:"title"`
	Content string `form:"content"`
}
