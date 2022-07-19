package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mirror-npm/utils"
	"testing"
)

func TestInit(t *testing.T) {
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
	for key := range data {
		//go func() {
		//	defer wg.Done()
		//url := fmt.Sprintf("http://npm.sunrise.lan/%s", key)
		//resp, err := http.Get(url)
		//if err != nil {
		//	log.Println(err)
		//	return
		//}
		fmt.Printf("touch %s\n", utils.GetEtagFileName(key))
		//}()
	}
	//wg.Wait()
}
