package handler

import (
	"blog-service-go/database"
	"blog-service-go/entity"
	"blog-service-go/upload"
	"blog-service-go/util"
	"bytes"
	"context"
	"embed"
	"errors"
	"fmt"
	"github.com/goccy/go-json"
	"golang.org/x/oauth2"
	"html/template"
	"log"
	"net/http"
	"strings"
)

var (
	pageSize   = int32(5)
	pageNumber = int32(1)
)

const (
	authServerURL = "http://localhost:9096"
	checkTokenURL = "http://localhost:9096/oauth/checkToken"
)

var (
	config = oauth2.Config{
		ClientID:     "222222",
		ClientSecret: "22222222",
		Scopes:       []string{"all"},
		RedirectURL:  "http://localhost:9096/oauth2",
		Endpoint: oauth2.Endpoint{
			AuthURL:  authServerURL + "/oauth/authorize",
			TokenURL: authServerURL + "/oauth/token",
		},
	}
	//globalToken *oauth2.Token // Non-concurrent security
)

func DeletePostById() http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		postID := strings.TrimPrefix(request.URL.Path, "/post/delete/")
		id, flag := util.ParseStr(postID)
		if !flag {
			log.Printf("Invalid post id : %s \n", postID)
			http.Error(writer, errors.New(fmt.Sprintf("Invalid post id : %s", postID)).Error(), http.StatusBadRequest)
			return
		}
		_, err := database.DeletePostById(id)
		if nil != err {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
	})
}
func UploadMarkdownHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fileByteInfo := upload.WriteFile2Tmp(w, r)
		if len(fileByteInfo.Bytes) > 0 {
			markdownFileContentInfo := upload.ProcessMarkdownFile(fileByteInfo)
			database.SaveMarkdown(markdownFileContentInfo)
		}
	})
}
func GetPagingPosts(writer http.ResponseWriter, request *http.Request) {
	err := request.ParseForm()
	if nil != err {
		http.Error(writer,
			fmt.Sprintf("Request parseForm err : %s \n", err.Error()),
			http.StatusInternalServerError)
		return
	}
	pageSizeStr := request.PostForm.Get("pageSize")
	pageNumberStr := request.PostForm.Get("pageNumber")
	i1, flag := util.ParseStr(pageSizeStr)
	if flag {
		pageSize = i1
	}
	i2, flag := util.ParseStr(pageNumberStr)
	if flag {
		pageNumber = i2
	}
	posts := database.QueryPagingPost(pageNumber, pageSize)
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", " ")
	err = encoder.Encode(posts)
	if nil != err {
		http.Error(writer,
			fmt.Sprintf("Encode JSON err : %s \n", err.Error()),
			http.StatusInternalServerError)
		return
	}
}
func GetPostsByCategoryName(writer http.ResponseWriter, request *http.Request) {
	categoryName := strings.TrimPrefix(request.URL.Path, "/category/")
	postDTOS := database.QueryPostByCategoryName(&categoryName)
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", " ")
	err := encoder.Encode(postDTOS)
	if nil != err {
		http.Error(writer,
			fmt.Sprintf("Encode JSON err : %s \n", err.Error()),
			http.StatusInternalServerError)
		return
	}
}
func GetPostsByTagName(writer http.ResponseWriter, request *http.Request) {
	tagName := strings.TrimPrefix(request.URL.Path, "/tag/")
	postDTOS := database.QueryPostByTagName(tagName)
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", " ")
	err := encoder.Encode(postDTOS)
	if nil != err {
		http.Error(writer,
			fmt.Sprintf("Encode JSON err : %s \n", err.Error()),
			http.StatusInternalServerError)
		return
	}
}
func GetPostById(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/post/id/")
	id, flag := util.ParseStr(idStr)
	if !flag {
		return
	}
	id64 := int64(id)
	postDTO := database.QueryPostById(&id64)
	_, err := w.Write([]byte(postDTO.Content))
	if nil != err {
		http.Error(w,
			fmt.Sprintf("Write Content to response body err : %s \n", err.Error()),
			http.StatusInternalServerError)
		return
	}
}
func GetPostByTitle(w http.ResponseWriter, r *http.Request) {
	title := strings.TrimPrefix(r.URL.Path, "/post/")
	post := database.QueryPostByTitle(&title)
	_, err := w.Write([]byte(post.Content))
	if nil != err {
		http.Error(w,
			fmt.Sprintf("Write Content to response body err : %s \n", err.Error()),
			http.StatusInternalServerError)
		return
	}
}
func HandleHome(fs embed.FS) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		posts := database.QueryPagingPost(1, 520)
		data := entity.HtmlDTO{
			PageTitle: "HandleHome",
			Data:      posts,
		}
		parse, err := template.ParseFS(fs, "template/home.html")
		if nil != err {
			log.Println(err)
			return
		} else {
			err := parse.Execute(w, data)
			if nil != err {
				http.Error(w,
					fmt.Sprintf("Execute template err : %s \n", err.Error()),
					http.StatusInternalServerError)
				return
			}
		}
	})
}
func CheckToken(handler http.Handler, method string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		//validation request method
		if req.Method != method {
			http.Error(w, errors.New(fmt.Sprintf("Only support request method : %s", method)).Error(), http.StatusBadRequest)
			return
		}
		//before request data validation
		request, err := http.NewRequest(http.MethodGet, checkTokenURL, nil)
		if nil != err {
			log.Printf("Create new request to check token error : %s \n", err.Error())
		}
		request.Header.Set("Authorization", req.Header.Get("Authorization"))
		client := &http.Client{}
		response, _ := client.Do(request)
		var b bytes.Buffer
		_, err = b.ReadFrom(response.Body)
		if nil != err {
			log.Println(fmt.Sprintf("Read byte from response body err : %s \n", err.Error()))
		}
		if http.StatusOK != response.StatusCode {
			http.Error(w, b.String(), http.StatusBadRequest)
			return
		}
		handler.ServeHTTP(w, req)
		//after request do something
	})
}
func Login(res http.ResponseWriter, req *http.Request) {
	//get username and password from request body
	err := req.ParseForm()
	if nil != err {
		http.Error(res,
			fmt.Sprintf("Request parseForm err : %s \n", err.Error()),
			http.StatusInternalServerError)
		return
	}
	username := req.PostForm.Get("username")
	password := req.PostForm.Get("password")
	if "" == username || "" == password {
		http.Error(res, errors.New("invalid username or password").Error(), http.StatusBadRequest)
		return
	}
	token, err := config.PasswordCredentialsToken(context.Background(), username, password)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	e := json.NewEncoder(res)
	e.SetIndent("", "  ")
	err = e.Encode(token)
	if nil != err {
		http.Error(res,
			fmt.Sprintf("Encode JSON err : %s \n", err.Error()),
			http.StatusInternalServerError)
		return
	}
}
