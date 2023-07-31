package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
)

type gitApiJson struct {
	Html_url string `json:"html_url"`
	Tag_name string `json:"tag_name"`
}

func InInt(target int, str_array []int) bool {
	for _, element := range str_array {
		if target == element {
			return true
		}
	}
	return false
}

func IsListDuplicated(list []int) (string, bool) {
	tmpMap := make(map[int]int)
	for _, value := range list {
		tmpMap[value] += 1
	}
	for k, v := range tmpMap {
		if v > 1 {
			return strconv.Itoa(k), true
		}
	}
	return "", false
}

func WaitExit() {
	var chose string
	for {
		fmt.Scan(&chose)
		switch chose {
		case "exit":
			os.Exit(0)
		case "stop":
			os.Exit(0)
		default:
		}
	}
}

func FileExists(path string) bool {
	_, err := os.Stat(path) //os.Stat获取文件信息
	if err != nil {
		return os.IsExist(err)
	}
	return true
}

func UpdateVersion(version string) (gitApiJson, error) {
	var gitapi gitApiJson
	client := &http.Client{}
	r, err := client.Get("https://api.github.com/repos/zijiren233/MCSM-Bot/releases/latest")
	if err != nil {
		return gitapi, err
	}
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return gitapi, err
	}
	json.Unmarshal(body, &gitapi)
	return gitapi, nil
}
