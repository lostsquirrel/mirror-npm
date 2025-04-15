package utils

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
)

const HeaderEtag = "Etag"
const WithEtag = true

func UpdateMetaOnDisk() {
	files, err := os.ReadDir(EtagBasePath())
	if err != nil {
		log.Fatal(err)
	}
	client := &http.Client{}
	for _, file := range files {
		fileName := file.Name()
		var metaId = getMetaIdFromFileName(fileName)
		_, err := GetMetaContent(metaId, client)
		if err != nil {
			log.Printf("update meta %s failed %v", file, err)
		}
	}

}

func ReadEtagFromFile(fileName string) (string, error) {
	etagFilePath := fmt.Sprintf("%s/%s", EtagBasePath(), fileName)
	log.Printf("read %s", etagFilePath)
	buf, err := os.ReadFile(etagFilePath)
	if err != nil {
		return "", err
	}
	return string(buf), nil
}

func buildMetaFilePath(metaId string) string {
	return fmt.Sprintf("%s/%s", MetaBasePath(), metaId)
}

func BuildMetaUrl(metaId string) string {
	return fmt.Sprintf("%s/%s", BaseUrl(), metaId)
}

func BuildEtagFilePath(metaId string) string {
	return fmt.Sprintf("%s/%s", EtagBasePath(), GetEtagFileName(metaId))
}

func CheckUpdate(metaId string) (bool, error) {
	metaUrl := BuildMetaUrl(metaId)
	resp, err := http.Head(metaUrl)
	if err != nil {
		return true, err
	}
	defer func() {
		if resp.Body != nil {
			if err := resp.Body.Close(); err != nil {
				log.Printf("close body failed %v", err)
			}
		}
	}()
	newTag, ok := resp.Header[HeaderEtag]
	if ok {
		old, err := ReadEtagFromFile(GetEtagFileName(metaId))
		if err != nil {
			return true, err
		}
		log.Printf("compare --%s-- %s", newTag[0], old)
		if newTag[0] == old {
			return false, nil
		}
	}
	return true, nil
}

func WriteStringToFile(content, filePath string) error {
	err := CreateFileParent(filePath)
	if err != nil {
		return err
	}
	bytes := []byte(content)
	return os.WriteFile(filePath, bytes, 0644) // updated for Go 1.16+
}

func getMetaIdFromFileName(fileName string) string {
	if strings.HasPrefix(fileName, "@") {
		namePart := strings.Split(fileName, "_")
		return strings.Join(namePart, "/")
	}
	return fileName
}

func GetEtagFileName(metaId string) string {
	if strings.HasPrefix(metaId, "@") {
		namePart := strings.Split(metaId, "/")
		return strings.Join(namePart, "_")
	}
	return metaId
}
