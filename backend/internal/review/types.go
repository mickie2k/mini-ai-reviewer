package review

type Language string

const (
	Go         Language = "golang"
	Python     Language = "python"
	Typescript Language = "typescript"
	JavaScript Language = "javascript"
	Java       Language = "java"
)

type ReviewRequest struct {
	Code     string   `json:"code" binding:"required"`
	Language Language `json:"language" binding:"required,oneof=golang python typescript javascript java"`
}

type ReviewResponse struct {
	Comments []Comment `json:"comments"`
}

type Comment struct {
	StartLine   int    `json:"startLine"`
	EndLine     int    `json:"endLine"`
	Explanation string `json:"explanation"`
	Fix         string `json:"fix"`
}
