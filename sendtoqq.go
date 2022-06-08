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

var chan_message_id chan string
var chan_message chan string

type QConfig struct {
	Cqhttp struct {
		Token     string   `json:"token"`
		Url       string   `json:"url"`
		Qq        string   `json:"qq"`
		Group_id  string   `json:"group_id"`
		Adminlist []string `json:"adminlist"`
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

func Get_group_last_msg_id() string {
	client := &http.Client{}
	data := url.Values{}
	data.Set("group_id", qconfig.Cqhttp.Group_id)
	data.Set("access_token", qconfig.Cqhttp.Token)
	r2, err := http.NewRequest("GET", qconfig.Cqhttp.Url+"/get_group_msg_history"+"?"+data.Encode(), nil)
	if err != nil {
		return "1"
	}
	r, err2 := client.Do(r2)
	if err2 != nil {
		return "1"
	}
	defer r.Body.Close()
	b, _ := ioutil.ReadAll(r.Body)
	strb := string(b)
	index := strings.LastIndex(strb, `"message_id":`)
	getdot := string(b)[index:]
	regexp, _ := regexp.Compile(":.*?[0-9]+")
	ret := regexp.FindString(getdot)
	return ret[1:]
}

func Get_Group_New_Mesage() {
	var tmp string
	tmp = Get_group_last_msg_id()
	go Get_msg()
	for {
		tmp2 := Get_group_last_msg_id()
		if tmp2 != "1" && tmp != tmp2 {
			tmp = tmp2
			chan_message_id <- tmp
		}
		time.Sleep(40 * time.Millisecond)
		/*The detection frequency is 40ms once, and it should not be set too large,
		otherwise messages will be missed*/
	}
}

func Get_msg() {
	client := &http.Client{}
	for {
		data := url.Values{}
		data.Set("message_id", <-chan_message_id)
		data.Set("access_token", qconfig.Cqhttp.Token)
		r2, err := http.NewRequest("GET", qconfig.Cqhttp.Url+"/get_msg"+"?"+data.Encode(), nil)
		if err != nil {
			fmt.Printf("err: %v\n", err)
		}
		r, err2 := client.Do(r2)
		if err != nil {
			fmt.Printf("err: %v\n", err2)
			// fmt.Println("可能是网络问题")
		}
		b, _ := ioutil.ReadAll(r.Body)
		// fmt.Printf("b: %v\n", string(b))
		var mesdata MesData
		json.Unmarshal(b, &mesdata)
		user_id := strconv.Itoa(mesdata.Data.Sender.User_id)
		// Check admin list
		if len(qconfig.Cqhttp.Adminlist) == 0 {
			chan_message <- mesdata.Data.Message
		} else if in(user_id, qconfig.Cqhttp.Adminlist) && user_id != qconfig.Cqhttp.Qq {
			chan_message <- mesdata.Data.Message
		}
	}
}

func Send_group_msg(message string, order int) {
	client := &http.Client{}
	data := url.Values{}
	data.Set("group_id", qconfig.Cqhttp.Group_id)
	data.Set("message", message)
	data.Set("auto_escape", "false")
	data.Set("access_token", qconfig.Cqhttp.Token)
	r2, err := http.NewRequest("GET", qconfig.Cqhttp.Url+"/send_group_msg"+"?"+data.Encode(), nil)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return
	}
	r3, err2 := client.Do(r2)
	if err != nil {
		fmt.Printf("err: %v\n", err2)
		// fmt.Println("发送失败")
		return
	}
	defer r3.Body.Close()
}

func in(target string, str_array []string) bool {
	for _, element := range str_array {
		if target == element {
			return true
		}
	}
	return false
}

func ProcessMessag(order int) {
	// 设置缓存大小为 1000 / 40 = 25
	chan_message_id = make(chan string, 25)
	chan_message = make(chan string, 25)
	go Get_Group_New_Mesage()
	go func() {
		for {
			if !RunningTest(order) && mstatus[mconfig.McsmData[order].Name] == 1 {
				mstatus[mconfig.McsmData[order].Name] = 0
				Send_group_msg(fmt.Sprint(`[CQ:at,qq=all]`, "服务器", mconfig.McsmData[order].Name, "已停止！"), order)
			} else if RunningTest(order) && mstatus[mconfig.McsmData[order].Name] == 0 {
				Send_group_msg(fmt.Sprint(`[CQ:at,qq=`, qconfig.Cqhttp.Qq, `]`, "服务器", mconfig.McsmData[order].Name, "以启动！"), order)
			}
			time.Sleep(5 * time.Second)
		}
	}()
	tmp := ""
	for {
		tmp = <-chan_message
		flysnowRegexp := regexp.MustCompile(`^run ([0-9]*)(.*)`)
		params := flysnowRegexp.FindStringSubmatch(tmp)
		if params[2] != "" {
			cmd := strings.TrimLeft(params[2], " ")
			go RunCmd(cmd, order)
		}
	}
}
