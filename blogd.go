package main

import (
	"blog-service-go/handler"
	"blog-service-go/load"
	"embed"
	"fmt"
	"log"
	"net/http"
	"os/user"
)

//go:embed template/*
var content embed.FS

func init() {
	current, err := user.Current()
	if nil != err {
		log.Fatal(err)
		return
	}
	dir := fmt.Sprintf("%s/%s", current.HomeDir, "markdown")
	load.SaveAsciiDoc2DB(&dir)
}
func main() {
	//http.HandleFunc("/login", handler.Login)
	//http.HandleFunc("/hello", handler.CategoryOfPost)
	http.Handle("/", handler.HandleHome(content))
	http.HandleFunc("/post/", handler.GetPostByTitle)
	//http.Handle("/upload", handler.CheckToken(handler.UploadMarkdownHandler(), http.MethodPost))
	//http.Handle("/post/delete/", handler.CheckToken(handler.DeletePostById(), http.MethodPost))
	//http.HandleFunc("/post/id/", handler.GetPostById)
	//http.HandleFunc("/posts", handler.GetPagingPosts)
	//http.HandleFunc("/category/", handler.GetPostsByCategoryName)
	//http.HandleFunc("/tag/", handler.GetPostsByTagName)
	err := http.ListenAndServe(":8089", nil)
	if nil != err {
		log.Fatal(err)
	}
	//database := database.ConnDB()
	//defer database.Close()
	//insertStmt := `insert into markdown(title,content) values ($1,$2)`
	//_, e := database.Exec(insertStmt, "This is a title2", "This is a content2")
	//database.CheckError(e)
}
