package utils

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const pkgPath = "PACKAGE_BASE_PATH"
const baseUrl = "NPM_BASE_URL"
const addr = "ADDR"
const metaPath = "META_PATH"
const etagPath = "ETAG_PATH"
const localBaseUrl = "LOCAL_BASE_URL"

var DefaultConfigs = map[string]string{
	pkgPath:      "/data/_pkg",
	metaPath:     "/data/_registry",
	etagPath:     "/data/_etag",
	baseUrl:      "https://registry.npmjs.org",
	localBaseUrl: "http://npm.sunrise.lan",
	addr:         ":8888",
}

func LoadEnv() {
	for _, e := range os.Environ() {
		pair := strings.SplitN(e, "=", 2)
		//fmt.Println(pair[0])

		_, ok := DefaultConfigs[pair[0]]
		if ok {
			DefaultConfigs[pair[0]] = pair[1]
			log.Printf("load env %v", pair)
		}
	}
}

func GetPkgPath() string {
	return DefaultConfigs[pkgPath]
}

func GetBaseUrl() string {
	return DefaultConfigs[baseUrl]
}

func GetAddr() string {
	return DefaultConfigs[addr]
}

func MetaBasePath() string {
	return DefaultConfigs[metaPath]
}

func EtagBasePath() string {
	return DefaultConfigs[etagPath]
}

func BaseUrl() string {
	return DefaultConfigs[baseUrl]
}

func LocalBaseUrl() string {
	return DefaultConfigs[localBaseUrl]
}

func ReplaceBasePath(metaId string, origin string) string {
	var (
		originPackageUrl string
		localPackageUrl  string
	)
	if strings.HasPrefix(metaId, "@") {
		metaParts := strings.Split(metaId, "/")
		metaGroup := metaParts[0]
		metaIdEx := metaParts[1]
		originPackageUrl = fmt.Sprintf("%s/%s/%s/-/%s", BaseUrl(), metaGroup, metaIdEx, metaIdEx)
		localPackageUrl = fmt.Sprintf("%s/_pkg/%s/%s/-/%s", LocalBaseUrl(), metaGroup, metaIdEx, metaIdEx)
	} else {
		originPackageUrl = fmt.Sprintf("%s/%s/-/%s", BaseUrl(), metaId, metaId)
		localPackageUrl = fmt.Sprintf("%s/_pkg/%s/-/%s", LocalBaseUrl(), metaId, metaId)
	}

	return strings.Replace(origin, originPackageUrl, localPackageUrl, -1)
}

func HandleHttpError(url string, w http.ResponseWriter, err error, resp *http.Response) bool {
	if err != nil {
		w.WriteHeader(500)
		log.Printf("fetch from %s failed %v", url, err)
		return true
	}
	if resp.StatusCode > 299 {
		w.WriteHeader(resp.StatusCode)
		log.Printf("fetch from %s failed", url)
		return true
	}
	return false
}

func CloseBody(body io.ReadCloser) {

	err := body.Close()
	if err != nil {
		log.Printf("close failed %v", err)
	}

}

func CloseFile(file *os.File) {
	err := file.Close()
	if err != nil {
		log.Printf("close %s failed %v", file.Name(), err)
	}
}

func CreateFileParent(file string) error {

	parent := filepath.Dir(file)
	return CreateDirIfNotExist(parent)

}

func CreateDirIfNotExist(parent string) error {
	if _, err := os.Stat(parent); os.IsNotExist(err) {
		const PermDir = 0755
		err = os.MkdirAll(parent, PermDir)
		return err
	}
	return nil
}

func GetMetaContent(metaId string, c *http.Client) (string, error) {
	metaUrl := fmt.Sprintf("%s/%s", BaseUrl(), metaId)
	metaFilePath := buildMetaFilePath(metaId)
	var modTime time.Time
	if info, err := os.Stat(metaFilePath); err == nil {
		modTime = info.ModTime()
		fmt.Println("📁 Local cache found, mod time:", modTime)
		resp, err := c.Head(metaUrl)
		if err != nil {
			log.Println("send head request failed")
			return "", err
		}
		if lm := resp.Header.Get("Last-Modified"); lm != "" {
			if t, err := time.Parse(http.TimeFormat, lm); err == nil {
				if t.After(modTime) {
					return downloadMeta(metaId)
				}
			}

		}
		data, err := os.ReadFile(metaFilePath)
		if err != nil {
			log.Println("read meta file failed")
			return downloadMeta(metaId)
		}
		return string(data), nil

	} else {
		fmt.Println("🆕 No local cache, will fetch full content")
		return downloadMeta(metaId)

	}

}

func downloadMeta(metaId string) (string, error) {
	metaUrl := fmt.Sprintf("%s/%s", BaseUrl(), metaId)
	metaFilePath := buildMetaFilePath(metaId)
	resp, err := http.Get(metaUrl)
	if err != nil {
		return "", err
	}
	defer CloseBody(resp.Body)
	if resp.StatusCode != 200 {
		return "", errors.New(resp.Status)
	}
	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	content := string(buf)
	replacedContent := ReplaceBasePath(metaId, content)

	log.Printf("write to meta file %s", metaFilePath)
	if err := WriteStringToFile(replacedContent, metaFilePath); err != nil {
		log.Printf("write to file %s failed %v", metaFilePath, err)
	}
	if lm := resp.Header.Get("Last-Modified"); lm != "" {
		if t, err := time.Parse(http.TimeFormat, lm); err == nil {
			os.Chtimes(metaFilePath, t, t)
		}
	}
	return replacedContent, nil
}
