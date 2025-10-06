package domain

type UserES struct {
	Id       int64
	Email    string
	Nickname string
	Phone    string
}
type ArticleES struct {
	Id      int64
	Title   string
	Status  int32
	Content string
	Tags    []string
}

type SearchResult struct {
	Users    []UserES
	Articles []ArticleES
}
