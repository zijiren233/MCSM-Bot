package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type QConfig struct {
	Cqhttp struct {
		Token string `json:"token"`
		Url   string `json:"url"`
		Qq    string `json:"qq"`
	} `json:"cqhttp"`
}

type MesData struct {
	Data struct {
		Message string `json:"message"`
		Sender  struct {
			Nickname string `json:"nickname"`
			User_id  int    `json:"user_id"`
		} `json:"sender"`
	} `json:"data"`
}

func GetQConfig() QConfig {
	var config QConfig
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

func Get_group_last_msg_id(order int) string {
	client := &http.Client{}
	data := url.Values{}
	data.Set("group_id", mconfig.McsmData[order].Group_id)
	data.Set("access_token", qconfig.Cqhttp.Token)
	r2, _ := http.NewRequest("GET", qconfig.Cqhttp.Url+"/get_group_msg_history"+"?"+data.Encode(), nil)
	r, _ := client.Do(r2)
	b, _ := ioutil.ReadAll(r.Body)
	strb := string(b)
	index := strings.LastIndex(strb, `"message_id":`)
	getdot := string(b)[index:]
	regexp, _ := regexp.Compile(":.*?[0-9]+")
	ret := regexp.FindString(getdot)
	return ret[1:]
}

func Get_Group_New_Mesage_Id(order int, chan_message_id chan string) {
	tmp := Get_group_last_msg_id(order)
	for {
		tmp2 := Get_group_last_msg_id(order)
		if tmp2 != "1" && tmp != tmp2 {
			tmp = tmp2
			chan_message_id <- tmp2
		}
		time.Sleep(40 * time.Millisecond)
		/*The detection frequency is 40ms once, and it should not be set too large,
		otherwise messages will be missed*/
	}
}

func TestCqhttpStatus(order int) {
	client := &http.Client{}
	data := url.Values{}
	data.Set("group_id", mconfig.McsmData[order].Group_id)
	data.Set("access_token", qconfig.Cqhttp.Token)
	r2, err := http.NewRequest("GET", qconfig.Cqhttp.Url+"/get_group_msg_history"+"?"+data.Encode(), nil)
	if err != nil {
		fmt.Println("Cqhttp状态检测错误，请检查配置文件或Cqhttp状态")
		os.Exit(1)
	}
	r, err2 := client.Do(r2)
	if err2 != nil {
		fmt.Println("Cqhttp状态检测错误，请检查配置文件或Cqhttp状态")
		os.Exit(1)
	}
	defer r.Body.Close()
}

func Get_msg(order int, chan_message_id chan string, chan_message chan string) {
	client := &http.Client{}
	var mesdata MesData
	for {
		data := url.Values{}
		data.Set("access_token", qconfig.Cqhttp.Token)
		data.Set("message_id", <-chan_message_id)
		r2, _ := http.NewRequest("GET", qconfig.Cqhttp.Url+"/get_msg"+"?"+data.Encode(), nil)
		r, _ := client.Do(r2)
		b, _ := ioutil.ReadAll(r.Body)
		json.Unmarshal(b, &mesdata)
		user_id := strconv.Itoa(mesdata.Data.Sender.User_id)
		// Check admin list
		if len(mconfig.McsmData[order].Adminlist) == 0 {
			chan_message <- mesdata.Data.Message
		} else if in(user_id, mconfig.McsmData[order].Adminlist) && user_id != qconfig.Cqhttp.Qq {
			chan_message <- mesdata.Data.Message
		}
	}
}

func Send_group_msg(message string, order int) {
	client := &http.Client{}
	data := url.Values{}
	data.Set("group_id", mconfig.McsmData[order].Group_id)
	data.Set("message", message)
	data.Set("auto_escape", "false")
	data.Set("access_token", qconfig.Cqhttp.Token)
	r2, _ := http.NewRequest("GET", qconfig.Cqhttp.Url+"/send_group_msg"+"?"+data.Encode(), nil)
	client.Do(r2)
}

func in(target string, str_array []string) bool {
	for _, element := range str_array {
		if target == element {
			return true
		}
	}
	return false
}

func AddQListen(order int) {
	// 设置缓存大小为 1000 / 40 = 25
	TestCqhttpStatus(order)
	TestMcsmStatus(order)
	chan_message := make(chan string, 25)
	chan_message_id := make(chan string, 25)
	go Get_Group_New_Mesage_Id(order, chan_message_id)
	go Get_msg(order, chan_message_id, chan_message)
	go ReportStatus(order)
	flysnowRegexp := regexp.MustCompile(`^run ([0-9]*) *([a-zA-Z]+)`)
	for {
		params := flysnowRegexp.FindStringSubmatch(<-chan_message)
		if len(params) == 0 {
			continue
		} else if params[2] != "" {
			// cmd := strings.TrimLeft(params[2], " ")
			go RunCmd(params[2], order)
		}
	}
}

func ReportStatus(order int) {
	for {
		if !RunningTest(order) && statusmap[mconfig.McsmData[order].Name] == 1 {
			statusmap[mconfig.McsmData[order].Name] = 0
			Send_group_msg(fmt.Sprint(`[CQ:at,qq=`, mconfig.McsmData[order].Adminlist[0], `]`, "服务器", mconfig.McsmData[order].Name, "已停止！"), order)
		} else if RunningTest(order) && statusmap[mconfig.McsmData[order].Name] == 0 {
			statusmap[mconfig.McsmData[order].Name] = 1
			Send_group_msg(fmt.Sprint(`[CQ:at,qq=`, mconfig.McsmData[order].Adminlist[0], `]`, "服务器", mconfig.McsmData[order].Name, "以启动！"), order)
		}
		fmt.Printf("statusmap[mconfig.McsmData[order].Name]: %v\n", statusmap[mconfig.McsmData[order].Name])
		fmt.Printf("RunningTest(order): %v\n", RunningTest(order))
		time.Sleep(3 * time.Second)
	}
}
