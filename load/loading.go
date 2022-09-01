package load

import (
	"blog-service-go/database"
	"blog-service-go/repository"
	"blog-service-go/util"
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
)

func SaveAsciiDoc2DB(dir *string) {
	log.Printf("Start to loading file from dir:%s \n", *dir)
	files, err := ioutil.ReadDir(*dir)
	if nil != err {
		log.Println(err)
		return
	}
	//clear html document exist on dir
	for i := 0; i < len(files); i++ {
		if filename := files[i].Name(); isFileSuffixMatch(&filename, "html") {
			err := os.Remove(fmt.Sprintf("%s/%s", *dir, filename))
			if nil != err {
				log.Println(err)
			} else {
				log.Println(fmt.Sprintf("Removed file : %s/%s", *dir, filename))
			}
		}
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	for i := 0; i < len(files); i++ {
		file := files[i]
		if !file.IsDir() {
			filename := file.Name()
			if isFileSuffixMatch(&filename, "md") {
				//convert markdown doc to asciidoc
				markdownDocPath := fmt.Sprintf("%s/%s", *dir, filename)
				asciiDocPath := fmt.Sprintf("%s/%s.adoc", *dir, FilenamePrefix(&filename))
				//	asciiCssPath := fmt.Sprintf("%s/my.css", *dir)
				cmd := exec.Command("pandoc",
					"--template", fmt.Sprintf("%s/%s", *dir, "default.asciidoctor"),
					markdownDocPath,
					"-f",
					"markdown",
					"-t",
					"asciidoctor", "-s", "-o",
					asciiDocPath)
				cmd.Stdout = &stdout
				cmd.Stderr = &stderr
				log.Println(cmd)
				err := cmd.Run()
				if nil != err {
					log.Println(stdout.String())
					log.Println(stderr.String())
				} else {
					//convert asciidoc to html
					// asciidoctor -a stylesheet=my.css  test.adoc
					cmd := exec.Command("asciidoctor",
						"-a", "stylesheet=my.css",
						fmt.Sprintf("%s/%s.adoc", *dir, FilenamePrefix(&filename)))
					if err := cmd.Run(); nil != err {
						log.Println(fmt.Sprintf("Cmd run error:%s\n", err.Error()))
					}
				}
			} else if isFileSuffixMatch(&filename, "adoc") {
				//convert asciidoc to html
				cmd := exec.Command("asciidoctor",
					"-a", "stylesheet=my.css",
					fmt.Sprintf("%s/%s.adoc", *dir, FilenamePrefix(&filename)))
				if err := cmd.Run(); nil != err {
					log.Println(fmt.Sprintf("Cmd run error:%s\n", err.Error()))
				}
			}
		}
	}
	//save html file content file to db
	saveHtml2DB(dir)
}

func saveHtml2DB(dir *string) {
	var Posts []repository.Post
	files, err := ioutil.ReadDir(*dir)
	if nil != err {
		log.Println(err.Error())
	} else {
		for i := 0; i < len(files); i++ {
			if filename := files[i].Name(); isFileSuffixMatch(&filename, "html") {
				filePath := fmt.Sprintf("%s/%s", *dir, filename)
				file, err := ioutil.ReadFile(filePath)
				if nil != err {
					log.Printf("Read file [%s] error:%s \n", filePath, err.Error())
				} else {
					Post := repository.Post{
						ID:        0,
						Title:     FilenamePrefix(&filename),
						Content:   util.ToSqlNullString(string(file)),
						CreatedAt: sql.NullTime{},
						UpdatedAt: sql.NullTime{},
						Deleted:   0,
						PubDate:   sql.NullTime{},
						CoverImg:  sql.NullString{},
					}
					Posts = append(Posts, Post)
				}
			}
		}
		if nil != Posts && len(Posts) > 0 {
			ctx := context.Background()
			db := database.OpenDB()
			defer database.Close(db)
			queries := database.OpenQueries(db)
			for i := 0; i < len(Posts); i++ {
				_, err := queries.SavePost(ctx, repository.SavePostParams{
					Title:    Posts[i].Title,
					Content:  util.ToSqlNullString(Posts[i].Content.String),
					CoverImg: sql.NullString{},
					PubDate:  sql.NullTime{},
				})
				if nil != err {
					log.Printf("Save post to db err : %s \n", err.Error())
				}
			}
		}
	}
}
func isFileSuffixMatch(filename *string, targetSuffix string) bool {
	if "" == targetSuffix || "" == *filename {
		return false
	} else {
		return "" != *filename && targetSuffix == strings.Split(strings.ToLower(*filename), ".")[1]
	}
}
func FilenamePrefix(filename *string) string {
	if "" != *filename {
		return strings.Split(*filename, ".")[0]
	} else {
		return *filename
	}
}
