package utils

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"testing"
)

func TestReadDir(t *testing.T) {
	files, err := ioutil.ReadDir("/data")
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		fmt.Println(file.Name(), file.IsDir())
	}
}

func TestIfNoneMatch(t *testing.T) {

	client := &http.Client{}

	req, _ := http.NewRequest("GET", "https://registry.npmjs.org/@ant-design/pro-form", nil)
	// ...
	etag := `W/"e91d01bc21eec26d15454b7c13eabef3"`
	req.Header.Add("If-None-Match", etag)
	resp, _ := client.Do(req)
	defer resp.Body.Close()

	buf, _ := ioutil.ReadAll(resp.Body)

	content := string(buf)
	fmt.Println(resp.Header[HeaderEtag])
	fmt.Println(content)
	fmt.Println(resp.StatusCode)

}

func TestReloadEtag(t *testing.T) {

}

