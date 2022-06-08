package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

type Data struct {
	Data string `json:"data"`
}

type MConfig struct {
	McsmData []struct {
		Order       int    `json:"order"`
		Sendtype    string `json:"sendtype"`
		Name        string `json:"name"`
		Domain      string `json:"url"`
		Remote_uuid string `json:"remote_uuid"`
		Uuid        string `json:"uuid"`
		Apikey      string `json:"apikey"`
	} `json:"mcsmdata"`
}

type Status struct {
	Data struct {
		Data []struct {
			Status int `json:"status"`
		} `json:"data"`
	} `json:"data"`
}

func GetMConfig() MConfig {
	var config MConfig
	f, err := os.OpenFile("config.json", os.O_RDWR|os.O_CREATE, 0777)
	if err != nil {
		fmt.Printf("读取配置文件出错: %v\n", err)
		os.Exit(0)
	}
	b, err2 := ioutil.ReadAll(f)
	if err2 != nil {
		fmt.Printf("读取配置文件出错: %v\n", err2)
		os.Exit(0)
	}
	err3 := json.Unmarshal(b, &config)
	if err3 != nil {
		fmt.Printf("读取配置文件出错: %v\n", err3)
		os.Exit(0)
	}
	return config
}

func ReturnResult(command string, order int) {
	client := &http.Client{}
	r2, _ := http.NewRequest("GET", mconfig.McsmData[order].Domain+"/api/protected_instance/outputlog", nil)
	r2.Header.Set("x-requested-with", "xmlhttprequest")
	q := r2.URL.Query()
	q.Add("apikey", mconfig.McsmData[order].Apikey)
	q.Add("uuid", mconfig.McsmData[order].Uuid)
	q.Add("remote_uuid", mconfig.McsmData[order].Remote_uuid)
	r2.URL.RawQuery = q.Encode()
	time.Sleep(1 * time.Second)
	r, err := client.Do(r2)
	if err != nil {
		return
	}
	defer r.Body.Close()
	b, _ := ioutil.ReadAll(r.Body)
	r3 := regexp.MustCompile(`> \\r|\\r+|(\\u001b(\[|>|[a-zA-Z])*(\?)*[0-9;:]*[a-z]*=*\]*[m]*)`)
	ret := r3.ReplaceAllString(string(b), "")
	var data Data
	json.Unmarshal([]byte(ret), &data)
	str_b := string(data.Data)
	// fmt.Print(str_b)
	last := strings.LastIndex(str_b, fmt.Sprint("> ", command))
	res := str_b[last : len(str_b)-2]
	Send_group_msg(res, order)
}

func RunCmd(commd string, order int) {
	client := &http.Client{}
	r2, _ := http.NewRequest("GET", mconfig.McsmData[order].Domain+"/api/protected_instance/command", nil)
	q := r2.URL.Query()
	q.Add("apikey", mconfig.McsmData[order].Apikey)
	q.Add("uuid", mconfig.McsmData[order].Uuid)
	q.Add("remote_uuid", mconfig.McsmData[order].Remote_uuid)
	q.Add("command", commd)
	r2.URL.RawQuery = q.Encode()
	r2.Header.Set("x-requested-with", "xmlhttprequest")
	client.Do(r2)
	time.Sleep(100 * time.Millisecond)
	ReturnResult(commd, order)
}

func RunningTest(order int) bool {
	time.Sleep(1 * time.Second)
	client := &http.Client{}
	r2, err := http.NewRequest("GET", mconfig.McsmData[order].Domain+"/api/service/remote_service_instances", nil)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		os.Exit(1)
	}
	q := r2.URL.Query()
	q.Add("apikey", mconfig.McsmData[order].Apikey)
	q.Add("page", "1")
	q.Add("page_size", "1")
	q.Add("instance_name", mconfig.McsmData[order].Name)
	q.Add("remote_uuid", mconfig.McsmData[order].Remote_uuid)
	r2.URL.RawQuery = q.Encode()
	r2.Header.Set("x-requested-with", "xmlhttprequest")
	r, _ := client.Do(r2)
	b, _ := ioutil.ReadAll(r.Body)
	var status Status
	json.Unmarshal(b, &status)
	if status.Data.Data[0].Status == 3 {
		return true
	} else {
		return false
	}
}

func Start(order int) {
	client := &http.Client{}
	r2, _ := http.NewRequest("GET", mconfig.McsmData[order].Domain+"/api/protected_instance/open", nil)
	q := r2.URL.Query()
	q.Add("apikey", mconfig.McsmData[order].Apikey)
	q.Add("uuid", mconfig.McsmData[order].Uuid)
	q.Add("remote_uuid", mconfig.McsmData[order].Remote_uuid)
	r2.URL.RawQuery = q.Encode()
	r2.Header.Set("x-requested-with", "xmlhttprequest")
	client.Do(r2)
}

func Stop(order int) {
	client := &http.Client{}
	r2, _ := http.NewRequest("GET", mconfig.McsmData[order].Domain+"/api/protected_instance/stop", nil)
	q := r2.URL.Query()
	q.Add("apikey", mconfig.McsmData[order].Apikey)
	q.Add("uuid", mconfig.McsmData[order].Uuid)
	q.Add("remote_uuid", mconfig.McsmData[order].Remote_uuid)
	r2.URL.RawQuery = q.Encode()
	r2.Header.Set("x-requested-with", "xmlhttprequest")
	client.Do(r2)
}
