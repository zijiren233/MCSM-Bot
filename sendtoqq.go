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
	f, err := os.OpenFile("config.json", os.O_RDWR, 0777)
	if err != nil {
		fmt.Printf("读取配置文件出错1: %v\n", err)
		os.Exit(0)
	}
	b, err2 := ioutil.ReadAll(f)
	if err2 != nil {
		fmt.Printf("读取配置文件出错2: %v\n", err2)
		os.Exit(0)
	}
	err3 := json.Unmarshal(b, &config)
	if err3 != nil {
		fmt.Printf("读取配置文件出错3: %v\n", err3)
		os.Exit(0)
	}
	defer f.Close()
	return config
}

func Get_group_last_msg_id(order int, chan_last_message_id chan string) {
	client := &http.Client{}
	data := url.Values{}
	data.Set("group_id", mconfig.McsmData[order].Group_id)
	data.Set("access_token", qconfig.Cqhttp.Token)
	r2, _ := http.NewRequest("GET", qconfig.Cqhttp.Url+"/get_group_msg_history"+"?"+data.Encode(), nil)
	for {
		r, _ := client.Do(r2)
		b, _ := ioutil.ReadAll(r.Body)
		strb := string(b)
		index := strings.LastIndex(strb, `"message_id":`)
		if index == -1 {
			continue
		}
		// fmt.Printf("index: %v\n", index)
		getdot := string(b)[index:]
		regexp, _ := regexp.Compile(":.*?[0-9]+")
		ret := regexp.FindString(getdot)
		chan_last_message_id <- ret[1:]
		time.Sleep(40 * time.Millisecond)
		/*The detection frequency is 40ms once, and it should not be set too large,
		otherwise messages will be missed*/
	}
}

func Get_Group_New_Mesage_Id(order int, chan_message_id chan string) {
	chan_last_message_id := make(chan string, 25)
	go Get_group_last_msg_id(order, chan_last_message_id)
	tmp := <-chan_last_message_id
	var tmp2 string
	for {
		tmp2 = <-chan_last_message_id
		if tmp2 != "1" && tmp != tmp2 {
			tmp = tmp2
			chan_message_id <- tmp2
		}
	}
}

func TestCqhttpStatus(order int) {
	client := &http.Client{}
	data := url.Values{}
	data.Set("group_id", mconfig.McsmData[order].Group_id)
	data.Set("access_token", qconfig.Cqhttp.Token)
	r2, err := http.NewRequest("GET", qconfig.Cqhttp.Url+"/get_group_msg_history"+"?"+data.Encode(), nil)
	if err != nil {
		fmt.Println("Cqhttp 状态检测错误，请检查配置文件或 Cqhttp 状态")
		os.Exit(1)
	}
	r, err2 := client.Do(r2)
	if err2 != nil {
		fmt.Println("Cqhttp 状态检测错误，请检查配置文件或 Cqhttp 状态")
		os.Exit(1)
	}
	defer r.Body.Close()
}

func Get_msg(order int, chan_message_id chan string, chan_message chan string) {
	for {
		go func(message_id string, chan_message chan string) {
			client := &http.Client{}
			var mesdata MesData
			data := url.Values{}
			data.Set("access_token", qconfig.Cqhttp.Token)
			data.Set("message_id", message_id)
			r2, _ := http.NewRequest("GET", qconfig.Cqhttp.Url+"/get_msg"+"?"+data.Encode(), nil)
			r, _ := client.Do(r2)
			b, _ := ioutil.ReadAll(r.Body)
			// fmt.Printf("b: %v\n", string(b))
			json.Unmarshal(b, &mesdata)
			user_id := strconv.Itoa(mesdata.Data.Sender.User_id)
			// Check admin list
			if len(mconfig.McsmData[order].Adminlist) == 0 {
				chan_message <- mesdata.Data.Message
			} else if in(user_id, mconfig.McsmData[order].Adminlist) && user_id != qconfig.Cqhttp.Qq {
				chan_message <- mesdata.Data.Message
			}
		}(<-chan_message_id, chan_message)
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

func getKeys(m map[int]int) []int {
	keys := make([]int, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func AddQListen(order int) {
	TestCqhttpStatus(order)
	fmt.Println("监听实例 ", mconfig.McsmData[order].Name, " 成功！")
	// 获取已监听的 []Order
	i := getKeys(listenmap)
	for j := range i {
		if mconfig.McsmData[j].Group_id == mconfig.McsmData[order].Group_id || j == order {
			// fmt.Println("监听相同的群/已监听")
			listenmap[order] = 1
			return
		}
	}
	listenmap[order] = 1
	// 设置缓存大小为 1000 / 40 = 25 每秒最多处理25条消息
	chan_message := make(chan string, 25)
	chan_message_id := make(chan string, 25)
	go Get_Group_New_Mesage_Id(order, chan_message_id)
	go Get_msg(order, chan_message_id, chan_message)
	go ReportStatus(order)
	flysnowRegexp, _ := regexp.Compile(`^run ([0-9]*) *(.*)`)
	for {
		params := flysnowRegexp.FindString(<-chan_message)
		if len(params) == 0 {
			continue
		}
		// fmt.Printf("params: %v\n", params)
		go func(params string, order int) {
			params = strings.ReplaceAll(params, "\n", "")
			params = strings.ReplaceAll(params, "\r", "")
			params2 := flysnowRegexp.FindStringSubmatch(params)
			if params2[1] != "" {
				order, _ = strconv.Atoi(params2[1])
			}
			if params2[2] == "" || order >= len(mconfig.McsmData) || listenmap[order] != 1 {
				return
			}
			if params2[2] != "" {
				switch params2[2] {
				case "status":
					SendStatus(order)
				case "start":
					if statusmap[mconfig.McsmData[order].Name] == 0 {
						Start(order)
						Send_group_msg(fmt.Sprint("服务器 ", mconfig.McsmData[order].Name, " 正在启动"), order)
					} else {
					}
				case "stop":
					if statusmap[mconfig.McsmData[order].Name] == 1 {
						Stop(order)
					} else {
						Send_group_msg(fmt.Sprint("服务器 ", mconfig.McsmData[order].Name, " 未在运行"), order)
					}
				case "restart":
					Restart(order)
					Send_group_msg(fmt.Sprint("服务器 ", mconfig.McsmData[order].Name, " 正在重启"), order)
				case "kill":
					Kill(order)
				default:
					go RunCmd(params2[2], order)
				}
			}
		}(params, order)
	}
}

func ReportStatus(order int) {
	for {
		if !RunningTest(order) && statusmap[mconfig.McsmData[order].Name] == 1 {
			statusmap[mconfig.McsmData[order].Name] = 0
			Send_group_msg(fmt.Sprint(`[CQ:at,qq=`, mconfig.McsmData[order].Adminlist[0], `]`, "服务器 ", mconfig.McsmData[order].Name, " 已停止！"), order)
		} else if RunningTest(order) && statusmap[mconfig.McsmData[order].Name] == 0 {
			statusmap[mconfig.McsmData[order].Name] = 1
			Send_group_msg(fmt.Sprint(`[CQ:at,qq=`, mconfig.McsmData[order].Adminlist[0], `]`, "服务器 ", mconfig.McsmData[order].Name, " 已启动！"), order)
		}
		time.Sleep(1500 * time.Millisecond)
	}
}
