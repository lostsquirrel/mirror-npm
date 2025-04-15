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

const pkgPath = "_pkg"

func Handler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	log.Printf("%s url %s", r.Method, path)
	if r.Method == http.MethodPost {
		realPath := fmt.Sprintf("%s/%s", utils.BaseUrl(), path)

		proxyReq, err := http.NewRequest(r.Method, realPath, r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		proxyReq.Header = make(http.Header)
		for h, val := range r.Header {
			proxyReq.Header[h] = val
		}
		httpClient := &http.Client{}
		resp, err := httpClient.Do(proxyReq)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		for h, val := range resp.Header {
			for _, v := range val {
				w.Header().Add(h, v)
			}
		}
		w.WriteHeader(resp.StatusCode)
		io.Copy(w, resp.Body)
		return
	}
	if r.Method == http.MethodDelete {
		realPath := fmt.Sprintf("%s%s", utils.MetaBasePath(), strings.TrimPrefix(path, "/_operation"))
		err := os.Remove(realPath)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		return
	}
	path = strings.TrimPrefix(path, "/")

	if len(path) == 0 || strings.HasPrefix(path, pkgPath) {
		w.WriteHeader(http.StatusNotFound)
		log.Println("return for null", path)
		return
	}

	var realPath string = path

	log.Printf("downloading %s", realPath)

	metaId := realPath
	if len(metaId) == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	content, err := utils.GetMetaContent(metaId, &http.Client{})
	if err != nil {
		log.Printf("get meta content failed %s %v", metaId, err)
		return
	}

	_, err = w.Write([]byte(content))
	if err != nil {
		log.Printf("write http resp %s failed %v", realPath, err)
	}

}
