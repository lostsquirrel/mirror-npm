package utils

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"sync"
)

const HeaderEtag = "Etag"
const WithEtag = true

func UpdateMetaOnDisk() {
	files, err := ioutil.ReadDir(EtagBasePath())
	if err != nil {
		log.Fatal(err)
	}
	wg := sync.WaitGroup{}
	wg.Add(len(files))
	for _, file := range files {
		fileName := file.Name()
		go func() {
			defer wg.Done()
			var metaId = getMetaIdFromFileName(fileName)
			log.Printf("check meta update %s", metaId)

			err := updateMeta(metaId)
			if err != nil {
				log.Printf("update meta %s failed %v", metaId, err)
			}

		}()

	}
	wg.Wait()
}

func ReadEtagFromFile(fileName string) (string, error) {
	etagFilePath := fmt.Sprintf("%s/%s", EtagBasePath(), fileName)
	log.Printf("read %s", etagFilePath)
	buf, err := ioutil.ReadFile(etagFilePath)
	if err != nil {
		return "", nil
	}
	etag := string(buf)
	return etag, nil
}

func updateMeta(metaId string) error {
	log.Printf("start to update %s", metaId)
	content, etag, err := GetMetaContentWithEtag(metaId, WithEtag)
	if err != nil {
		return err
	}
	replacedContent := ReplaceBasePath(metaId, content)
	metaFilePath := buildMetaFilePath(metaId)
	log.Printf("write to meta file %s", metaFilePath)
	err = WriteStringToFile(replacedContent, metaFilePath)
	if err != nil {
		log.Printf("write to file %s failed %v", metaFilePath, err)
	}
	tagPath := BuildEtagFilePath(metaId)
	log.Printf("write to etag %s", tagPath)
	err = WriteStringToFile(etag, tagPath)
	if err != nil {
		log.Printf("write to file %s failed %v", tagPath, err)
	}
	return nil
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

func checkUpdate(metaId string) (bool, error) {
	metaUrl := BuildMetaUrl(metaId)
	resp, err := http.Head(metaUrl)
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			log.Printf("close body failed %v", err)
			return
		}
	}()
	if err != nil {
		return true, err
	}
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
	return ioutil.WriteFile(filePath, bytes, 0644)
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
