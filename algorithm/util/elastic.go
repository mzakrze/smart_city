package util

import (
	"net/http"
	"net/url"
)

func ClearOldIndicesInElastic(host string) {
	client := &http.Client{}

	urlLog := &url.URL{
		Scheme:  "http",
		Host: host + ":9200",
		Path: "simulation-map",
	}

	urlMap := &url.URL{
		Scheme:  "http",
		Host: host + ":9200",
		Path: "simulation-vehicle",
	}

	urlTrip := &url.URL{
		Scheme:  "http",
		Host: host + ":9200",
		Path: "simulation-intersection",
	}

	_, err := client.Do(&http.Request{
		Method: http.MethodDelete,
		URL: urlLog,
	})
	if err != nil {
		panic(err)
	}

	_, err = client.Do(&http.Request{
		Method: http.MethodDelete,
		URL: urlMap,
	})
	if err != nil {
		panic(err)
	}

	_, err = client.Do(&http.Request{
		Method: http.MethodDelete,
		URL: urlTrip,
	})
	if err != nil {
		panic(err)
	}
}

