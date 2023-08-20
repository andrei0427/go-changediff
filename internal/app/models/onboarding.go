package models

type OnboardingModel struct {
	Name        string `form:"name"`
	Description string `form:"description"`
	AccentColor string `form:"accent_color"`
}
