package upload

import (
	"blog-service-go/entity"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

// WriteFile2Tmp  upload file to /tmp dir and return byte info of this upload
func WriteFile2Tmp(res http.ResponseWriter, req *http.Request) entity.FileByteInfo {
	err := req.ParseMultipartForm(5e+6)
	if nil != err {
		fmt.Printf("Parse request body to multipart form error : %s \n", err)
	}
	if files := req.MultipartForm.File["file"]; len(files) > 0 {
		//Only upload first index upload
		open, err := files[0].Open()
		defer open.Close()
		size := files[0].Size
		filename := files[0].Filename
		if nil == err {
			bytesOfFile := make([]byte, size)
			_, err := open.Read(bytesOfFile)
			if nil == err {
				filePath := fmt.Sprintf("/tmp/%s", filename)
				err := ioutil.WriteFile(filePath, bytesOfFile, 0644)
				if nil == err {
					fmt.Printf("Export upload to path:%s success \n", filePath)
				}
			} else {
				fmt.Printf("Read upload error : %s \n", err)
			}
			categories, tags, coverImg, pubDate := parseCategoriesAndTags(req)
			return entity.FileByteInfo{
				Filename:   filename,
				Bytes:      bytesOfFile,
				Categories: categories,
				Tags:       tags,
				CoverImg:   coverImg,
				PubDate:    pubDate,
			}
		} else {
			fmt.Printf("Open upload error : %s \n", err)
		}
	}
	return entity.FileByteInfo{}
}
func parseCategoriesAndTags(req *http.Request) ([]string, []string, string, string) {
	categories := req.MultipartForm.Value["categories"]
	tags := req.MultipartForm.Value["tags"]
	coverImgs := req.MultipartForm.Value["coverImg"]
	pubDates := req.MultipartForm.Value["pubDate"]
	var categoriesOfPost []string
	var tagsOfPost []string
	var coverImg string
	var pubDate string
	if nil != categories && len(categories) > 0 {
		categoriesOfPost = strings.Split(categories[0], ",")
	}
	if nil != tags && len(tags) > 0 {
		tagsOfPost = strings.Split(tags[0], ",")
	}
	if nil != coverImgs && len(coverImgs) > 0 {
		coverImg = coverImgs[0]
	}
	if nil != pubDates && len(pubDates) > 0 {
		pubDate = pubDates[0]
	}
	return categoriesOfPost, tagsOfPost, coverImg, pubDate
}
func ProcessMarkdownFile(fileByteInfo entity.FileByteInfo) entity.FileContentInfo {
	md := string(fileByteInfo.Bytes)
	lines := strings.Split(md, "\n")
	newLines := make([]string, 0, len(lines))
	open := false
	close := false
	for _, line := range lines {
		line = strings.Trim(line, " ")
		if "---" == line {
			if open == true {
				close = true
			} else {
				open = true
			}
		} else {
			if open == false {
				newLines = append(newLines, line)
			} else if close == true {
				newLines = append(newLines, line)
			}
		}
	}
	newMd := strings.Join(newLines, "\n")
	return entity.FileContentInfo{
		Filename:   fileByteInfo.Filename,
		Content:    newMd,
		Categories: fileByteInfo.Categories,
		Tags:       fileByteInfo.Tags,
		CoverImg:   fileByteInfo.CoverImg,
		PubDate:    fileByteInfo.PubDate,
	}
}
