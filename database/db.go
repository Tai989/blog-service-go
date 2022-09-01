package database

import (
	"blog-service-go/entity"
	"blog-service-go/repository"
	"blog-service-go/util"
	"context"
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"log"
	"os"
	"strings"
)

const (
	host   = "localhost"
	port   = 5432
	user   = "postgres"
	dbname = "postgres"
)

func OpenDB() *sql.DB {
	pwd := os.Getenv("POSTGRES_PWD")
	psqlConn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, pwd, dbname)
	db, err := sql.Open("postgres", psqlConn)
	CheckError(err)
	err = db.Ping()
	CheckError(err)
	fmt.Println("DB connected!")
	return db
}
func OpenQueries(db *sql.DB) *repository.Queries {
	return repository.New(db)
}
func CheckError(err error) {
	if nil != err {
		panic(err)
	}
}
func QueryPagingPost(pageNumber int32, pageSize int32) []entity.PostDTO {
	ctx := context.Background()
	offset := (pageNumber - 1) * pageSize
	db := OpenDB()
	defer Close(db)
	queries := OpenQueries(db)
	posts, err := queries.FindAllPost(ctx, repository.FindAllPostParams{
		Limit:  pageSize,
		Offset: offset,
	})
	if nil == err {
		postDTOS := util.CopyPostArray(&posts)
		//get post category and tag
		for i := range postDTOS {
			postDTOS[i].Categories = queryPostCategory(ctx, queries, postDTOS[i].ID)
			postDTOS[i].Tags = queryPostTag(ctx, queries, postDTOS[i].ID)
		}
		return postDTOS
	} else {
		log.Println(err)
		return make([]entity.PostDTO, 0)
	}
}
func queryPostTag(ctx context.Context, queries *repository.Queries, postID int64) []string {
	ts, err := queries.FindTagByPostId(ctx, postID)
	var tags []string
	if nil != err {
		log.Println(err)
		return make([]string, 0)
	} else {
		for i := range ts {
			if ts[i].TagName.Valid {
				tags = append(tags, ts[i].TagName.String)
			}
		}
		if nil != tags && len(tags) > 0 {
			return tags
		} else {
			return make([]string, 0)
		}
	}
}
func queryPostCategory(ctx context.Context, queries *repository.Queries, postID int64) []string {
	cs, err := queries.FindCategoryByPostId(ctx, postID)
	var categories []string
	if nil != err {
		log.Println(err)
		return make([]string, 0)
	} else {
		for i := range cs {
			if cs[i].CategoryName.Valid {
				categories = append(categories, cs[i].CategoryName.String)
			}
		}
		if nil != categories && len(categories) > 0 {
			return categories
		} else {
			return make([]string, 0)
		}
	}
}
func DeletePostById(postID int32) (int64, error) {
	ctx := context.Background()
	db := OpenDB()
	defer Close(db)
	queries := OpenQueries(db)
	id, err := queries.DeletePostById(ctx, int64(postID))
	if nil != err {
		log.Println(err)
		return 0, err
	} else {
		log.Printf("deleted post, post id: %d \n", id)
	}
	err = queries.DeletePostCategoryRelByPostId(ctx, int64(postID))
	if nil != err {
		log.Printf("Delete post category relation err : %s \n", err.Error())
	}
	err = queries.DeletePostTagRelByPostId(ctx, int64(postID))
	if nil != err {
		log.Printf("Delete post tag relation err : %s \n", err.Error())
	}
	return id, nil
}
func QueryPostByTagName(tagName string) []entity.PostDTO {
	ctx := context.Background()
	db := OpenDB()
	defer Close(db)
	queries := OpenQueries(db)
	posts, err := queries.FindPostByTagName(ctx, util.ToSqlNullString(tagName))
	if nil != err {
		log.Println(err)
		return make([]entity.PostDTO, 0)
	} else {
		postDTOS := util.CopyPostArray(&posts)
		for index := range postDTOS {
			postDTOS[index].Categories = queryPostCategory(ctx, queries, postDTOS[index].ID)
			postDTOS[index].Tags = queryPostTag(ctx, queries, postDTOS[index].ID)
		}
		return postDTOS
	}
}
func QueryPostById(id *int64) entity.PostDTO {
	ctx := context.Background()
	db := OpenDB()
	defer Close(db)
	queries := OpenQueries(db)
	post, err := queries.FindPostById(ctx, *id)
	if nil != err {
		log.Println(err)
	}
	return util.CopyPost(&post)
}
func QueryPostByTitle(title *string) entity.PostDTO {
	ctx := context.Background()
	db := OpenDB()
	defer Close(db)
	queries := OpenQueries(db)
	post, err := queries.FindPostByTitle(ctx, *title)
	if nil != err {
		log.Println(err)
		return entity.PostDTO{}
	}
	return util.CopyPost(&post)
}
func QueryPostByCategoryName(categoryName *string) []entity.PostDTO {
	ctx := context.Background()
	db := OpenDB()
	defer Close(db)
	queries := OpenQueries(db)
	posts, err := queries.FindPostByCategoryName(ctx, util.ToSqlNullString(*categoryName))
	if nil != err {
		log.Println(err)
		return make([]entity.PostDTO, 0)
	}
	postDTOs := util.CopyPostArray(&posts)
	for index := range postDTOs {
		postDTOs[index].Categories = queryPostCategory(ctx, queries, postDTOs[index].ID)
		postDTOs[index].Tags = queryPostTag(ctx, queries, postDTOs[index].ID)
	}
	return postDTOs
}
func SaveMarkdown(info entity.FileContentInfo) {
	ctx := context.Background()
	db := OpenDB()
	defer Close(db)
	queries := OpenQueries(db)
	//save post category
	var categoryIDs []int64
	var tagIDs []int64
	if nil != info.Categories && len(info.Categories) > 0 {
		for _, c := range info.Categories {
			category, err := queries.SaveCategory(ctx, util.ToSqlNullString(c))
			if nil == err {
				categoryIDs = append(categoryIDs, category.ID)
			}
		}
	}
	//save post tag
	if nil != info.Tags && len(info.Tags) > 0 {
		for _, t := range info.Tags {
			tag, err := queries.SaveTag(ctx, util.ToSqlNullString(t))
			if nil == err {
				tagIDs = append(tagIDs, tag.ID)
			}
		}
	}
	//save post
	title := strings.Split(info.Filename, ".")[0]
	post, _ := queries.SavePost(ctx, repository.SavePostParams{
		Title:    title,
		Content:  util.ToSqlNullString(info.Content),
		CoverImg: util.ToSqlNullString(info.CoverImg),
		PubDate:  util.ToSqlNullTime(info.PubDate),
	})
	//save post and category id relation
	for _, categoryId := range categoryIDs {
		err := queries.SavePostCategory(ctx, repository.SavePostCategoryParams{
			PostID:     post.ID,
			CategoryID: categoryId,
		})
		if nil != err {
			log.Printf("Save post category relation err : %s \n", err.Error())
		}
	}
	//save post and tag id relation
	for _, tagId := range tagIDs {
		err := queries.SavePostTag(ctx, repository.SavePostTagParams{
			PostID: post.ID,
			TagID:  tagId,
		})
		if nil != err {
			log.Printf("Save post tag relation err : %s \n", err.Error())
		}
	}
}
func Close(db *sql.DB) {
	err := db.Close()
	if nil != err {
		log.Println(fmt.Sprintf("DB close err : %s \n", err))
	} else {
		log.Println(fmt.Sprintln("DB close successful."))
	}
}
