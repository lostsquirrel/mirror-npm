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
	log.Printf("%s url %s", r.Method, path)
	if r.Method == http.MethodPost {
		realPath := fmt.Sprintf("%s/%s", utils.BaseUrl(), path)
		// body, err := ioutil.ReadAll(r.Body)

		// if err != nil {
		// 	http.Error(w, err.Error(), http.StatusInternalServerError)
		// 	return
		// }
		// log.Println("req", string(body))

		// you can reassign the body if you need to parse it as multipart
		// r.Body = ioutil.NopCloser(bytes.NewReader(body))

		// create a new url from the raw RequestURI sent by the client
		// url := fmt.Sprintf("%s://%s%s", proxyScheme, proxyHost, req.RequestURI)

		proxyReq, err := http.NewRequest(r.Method, realPath, r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		// We may want to filter some headers, otherwise we could just use a shallow copy
		// proxyReq.Header = req.Header
		// proxyReq.Header = make(http.Header)
		// for h, val := range r.Header {
		// 	proxyReq.Header[h] = val
		// }
		httpClient := &http.Client{}
		resp, err := httpClient.Do(proxyReq)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()
		// buf, err := ioutil.ReadAll(resp.Body)
		// log.Println("resp", string(buf))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(resp.StatusCode)
		io.Copy(w, resp.Body)
		return
	}
	if r.Method == http.MethodDelete {
		realPath := fmt.Sprintf("%s%s", utils.MetaBasePath(), path)
		err := os.Remove(realPath)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		return
	}
	path = strings.TrimPrefix(path, "/")

	if len(path) == 0 {
		w.WriteHeader(http.StatusNotFound)
		log.Println("return for null")
		return
	}

	path = strings.TrimSuffix(path, "/")

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
