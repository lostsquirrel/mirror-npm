package handlers

import (
	"fmt"
	"io"
	"log"
	"mirror-npm/utils"
	"net/http"
	"os"
	"strings"
)

const pkgPrefix = "_pkg"
const prefixLength = len(pkgPrefix)
const withEtag = false

func Handler(w http.ResponseWriter, r *http.Request) {
	// omit the prefix splash
	path := r.URL.Path
	log.Printf("request url %s", path)
	if strings.HasPrefix(path, "/") {
		path = path[1:]
	}
	if strings.HasSuffix(path, "/") {
		path = path[:len(path) - 1]
	}
	var realPath string
	var isMeta bool

	if strings.HasPrefix(path, pkgPrefix) {
		realPath = path[prefixLength:]
	} else {
		realPath = path
		isMeta = true
	}

	log.Printf("downloading %s", realPath)

	if isMeta {
		metaId := realPath
		if len(metaId) == 0 {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		content, etag, err := utils.GetMetaContentWithEtag(metaId, withEtag)
		if err != nil {
			log.Printf("get meta content failed %s %v", metaId, err)
			return
		}
		replacedContent := utils.ReplaceBasePath(metaId, content)
		metaFilePath := fmt.Sprintf("%s/%s", utils.MetaBasePath(), metaId)
		err = utils.WriteStringToFile(replacedContent, metaFilePath)
		if err != nil {
			log.Printf("write to file %s failed %v", metaFilePath, err)
		}
		tagPath := utils.BuildEtagFilePath(metaId)
		log.Printf("tag path %s", tagPath)
		err = utils.WriteStringToFile(etag, tagPath)
		if err != nil {
			log.Printf("write to file %s failed %v", tagPath, err)
		}
		_, err = w.Write([]byte(replacedContent))
		if err != nil {
			log.Printf("write http resp %s failed %v", realPath, err)
		}
	} else {
		resourceUrl := fmt.Sprintf("%s%s", utils.GetBaseUrl(), realPath)
		resp, err := http.Get(resourceUrl)
		defer utils.CloseBody(resp.Body)
		if utils.HandleHttpError(resourceUrl, w, err, resp) {
			return
		}
		filepath := fmt.Sprintf("%s%s", utils.GetPkgPath(), realPath)
		err = utils.CreateFileParent(filepath)
		if err != nil {
			w.WriteHeader(500)
			log.Printf("create parent dir for  %s failed %v", filepath, err)
			return
		}
		file, err := os.Create(filepath)
		if err != nil {
			w.WriteHeader(500)
			log.Printf("write to %s failed %v", filepath, err)
			return
		}
		defer utils.CloseFile(file)
		out := io.MultiWriter(file, w)
		_, err = io.Copy(out, resp.Body)
		if err != nil {
			w.WriteHeader(500)
			log.Printf("write respone stream failed %v", err)
		}
		log.Printf("saved %s", filepath)
	}

}
