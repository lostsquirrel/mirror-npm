package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

func main() {
	buf, err := ioutil.ReadFile("/tmp/etag")
	if err != nil {
		return
	}
	var data map[string]string
	err = json.Unmarshal(buf, &data)
	if err != nil {
		return
	}
	//runtime.NumGoroutine()
	//wg := sync.WaitGroup{}
	//wg.Add(len(data))
	for key, _ := range data {
		//go func() {
		//	defer wg.Done()
		url := fmt.Sprintf("http://npm.sunrise.lan/%s", key)
		resp, err := http.Get(url)
		if err != nil {
			log.Println(err)
			return
		}
		fmt.Printf("get %s %d\n", key, resp.StatusCode)
		//}()
	}
	//wg.Wait()
}
