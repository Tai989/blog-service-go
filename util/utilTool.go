package util

import (
	"blog-service-go/entity"
	"blog-service-go/repository"
	"database/sql"
	"strconv"
	"time"
)

func ParseStr(s string) (int32, bool) {
	if "" != s {
		i, err := strconv.Atoi(s)
		if nil != err {
			return 0, false
		} else {
			return int32(i), true
		}
	} else {
		return 0, false
	}
}
func CopyPost(post *repository.Post) entity.PostDTO {
	dto := entity.PostDTO{
		ID:         post.ID,
		Title:      post.Title,
		Content:    post.Content.String,
		CreatedAt:  entity.JsonTime(post.CreatedAt.Time),
		UpdatedAt:  entity.JsonTime(post.UpdatedAt.Time),
		Deleted:    post.Deleted,
		PubDate:    entity.JsonTime(post.PubDate.Time),
		CoverImg:   post.CoverImg.String,
		Categories: nil,
		Tags:       nil,
	}
	return dto
}
func CopyPostArray(from *[]repository.Post) []entity.PostDTO {
	var postDTOS []entity.PostDTO
	for _, post := range *from {
		postEntity := entity.PostDTO{
			ID:         post.ID,
			Title:      post.Title,
			Content:    post.Content.String,
			CreatedAt:  entity.JsonTime(post.CreatedAt.Time),
			UpdatedAt:  entity.JsonTime(post.UpdatedAt.Time),
			Deleted:    post.Deleted,
			PubDate:    entity.JsonTime(post.PubDate.Time),
			CoverImg:   post.CoverImg.String,
			Categories: nil,
			Tags:       nil,
		}
		postDTOS = append(postDTOS, postEntity)
	}
	if nil != postDTOS && len(postDTOS) > 0 {
		return postDTOS
	} else {
		return make([]entity.PostDTO, 0)
	}
}

func ToSqlNullString(s string) sql.NullString {
	var flag bool
	if "" != s {
		flag = true
	} else {
		flag = false
	}
	return sql.NullString{String: s, Valid: flag}
}
func ToSqlNullTime(s string) sql.NullTime {
	if "" != s {
		t, _ := time.Parse("2006-01-02", s)
		return sql.NullTime{
			Time:  t,
			Valid: true,
		}
	} else {
		return sql.NullTime{
			Time:  time.Time{},
			Valid: false,
		}
	}
}
