package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)


func ClearOldIndicesInElastic(host string) {
	deleteAll, _ := json.Marshal(map[string]interface{}{
		"query": map[string]interface{}{
			"match_all": map[string]interface{}{},
		},
	})

	resp1, err := http.Post("http://" + host + ":9200/simulation-map/_delete_by_query?conflicts=proceed", "application/json", bytes.NewBuffer(deleteAll))
	if err != nil {
		panic(err)
	}
	defer resp1.Body.Close()

	resp2, err := http.Post("http://" +  host + ":9200/simulation-vehicle/_delete_by_query","application/json", bytes.NewBuffer(deleteAll))
	if err != nil {
		panic(err)
	}
	defer resp2.Body.Close()

	if resp1.StatusCode != 200 {
		body, _ := ioutil.ReadAll(resp1.Body)
		fmt.Println("Request to ES failed:")
		fmt.Println(string(body))
		panic("ES error")
	}

	if resp2.StatusCode != 200 {
		body, _ := ioutil.ReadAll(resp2.Body)
		fmt.Println("Request to ES failed:")
		fmt.Println(string(body))
		panic("ES error")
	}
}

