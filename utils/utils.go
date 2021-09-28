package utils

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
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
	originPackageUrl := fmt.Sprintf("%s/%s/-/%s", BaseUrl(), metaId, metaId)
	localPackageUrl := fmt.Sprintf("%s/%s/-/%s", LocalBaseUrl(), metaId, metaId)
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

func CreateFileParent(filepath string) error {
	parent := filepath[:strings.LastIndex(filepath, "/")]
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

func GetMetaContentWithEtag(metaId string, withEtag bool) (string, string, error) {
	metaUrl := fmt.Sprintf("%s/%s", BaseUrl(), metaId)
	client := &http.Client{}
	req, err := http.NewRequest("GET", metaUrl, nil)
	if withEtag {
		etag, err := ReadEtagFromFile(GetEtagFileName(metaId))
		if err != nil {
			log.Printf("cannot read etag %s", err)
		} else {
			req.Header.Add("If-None-Match", etag)
		}
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer CloseBody(resp.Body)
	if resp.StatusCode != 200 {
		return "", "", errors.New(resp.Status)
	}
	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}
	content := string(buf)
	var etag string
	newTag, ok := resp.Header[HeaderEtag]
	if ok {
		etag = newTag[0]
		log.Printf("read etag --%s--", etag)
	}
	return content, etag, nil
}

