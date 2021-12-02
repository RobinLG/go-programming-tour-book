package model

import "github.com/go-programming-tour-book/blog-service/pkg/app"

type ArticleTag struct {
	*Model
	TagID     uint32 `json:"tag_id"`
	ArticleID uint32 `json:"article_id"`
}

type ArticleSwagger struct {
	List  []*Article
	Pager *app.Pager
}

func (a ArticleTag) TableName() string {
	return "blog_article_tag"
}
