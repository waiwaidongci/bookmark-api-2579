package model

type Bookmark struct {
	ID         int64  `json:"id"`
	Title      string `json:"title"`
	URL        string `json:"url"`
	Tags       string `json:"tags"`
	Note       string `json:"note"`
	ClickCount int64  `json:"click_count"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
}

type CreateBookmarkInput struct {
	Title string `json:"title"`
	URL   string `json:"url"`
	Tags  string `json:"tags"`
	Note  string `json:"note"`
}

type UpdateBookmarkInput struct {
	Title *string `json:"title"`
	URL   *string `json:"url"`
	Tags  *string `json:"tags"`
	Note  *string `json:"note"`
}

type ListFilter struct {
	Tag      string
	Keyword  string
	Page     int
	PageSize int
}

type ListResult struct {
	Items    []Bookmark `json:"items"`
	Total    int64      `json:"total"`
	Page     int        `json:"page"`
	PageSize int        `json:"page_size"`
}
