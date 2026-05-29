package dto

type OEMResponse struct {
	ID        *int     `json:"ID"`
	ArticleID *int     `json:"articleId"`
	OEMs      []OemDTO `json:"oems"`
}
