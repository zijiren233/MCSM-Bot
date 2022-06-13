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
		Messages []struct {
			Message    string `json:"message"`
			User_id    int    `json:"user_id"`
			Message_id int    `json:"message_id"`
		} `json:"messages"`
	} `json:"data"`
}

type Mdata struct {
	Message    string `json:"message"`
	User_id    int    `json:"user_id"`
	Message_id int    `json:"message_id"`
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

func Get_group_new_msg(order int, chan_message chan Mdata) {
	client := &http.Client{}
	data := url.Values{}
	data.Set("group_id", mconfig.McsmData[order].Group_id)
	data.Set("access_token", qconfig.Cqhttp.Token)
	r2, _ := http.NewRequest("GET", qconfig.Cqhttp.Url+"/get_group_msg_history"+"?"+data.Encode(), nil)
	r, _ := client.Do(r2)
	b, _ := ioutil.ReadAll(r.Body)
	var mesdata MesData
	json.Unmarshal(b, &mesdata)
	tmp := mesdata.Data.Messages[19].Message_id
	for {
		r, _ = client.Do(r2)
		b, _ = ioutil.ReadAll(r.Body)
		json.Unmarshal(b, &mesdata)
		if mesdata.Data.Messages[19].Message_id != tmp {
			tmp = mesdata.Data.Messages[19].Message_id
			chan_message <- mesdata.Data.Messages[19]
		}
		time.Sleep(40 * time.Millisecond)
		/*The detection frequency is 40ms once, and it should not be set too large,
		otherwise messages will be missed*/
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
	chan_message := make(chan Mdata, 25)
	var od int
	go Get_group_new_msg(order, chan_message)
	go ReportStatus(order)
	flysnowRegexp, _ := regexp.Compile(`^run ([0-9]*) *(.*)`)
	var params string
	var params2 []string
	var mdata Mdata
	for {
		mdata = <-chan_message
		params = flysnowRegexp.FindString(mdata.Message)
		if len(params) == 0 {
			continue
		}
		params2 = flysnowRegexp.FindStringSubmatch(params)
		if params2[1] != "" {
			od, _ = strconv.Atoi(params2[1])
			if od >= len(mconfig.McsmData) {
				go Send_group_msg("Order错误！", order)
				continue
			}
			if mconfig.McsmData[order].Group_id != mconfig.McsmData[od].Group_id {
				go Send_group_msg("Order错误！", order)
				continue
			}
			if !in(strconv.Itoa(mdata.User_id), mconfig.McsmData[od].Adminlist) {
				go Send_group_msg("权限不足！", order)
				continue
			}
			if listenmap[od] != 1 {
				go Send_group_msg("未开启监听！", order)
				continue
			}
		} else {
			od = order
		}
		if params2[2] == "" {
			continue
		}
		go func(params string, order int) {
			params = strings.ReplaceAll(params, "\n", "")
			params = strings.ReplaceAll(params, "\r", "")
			switch params {
			case "status":
				SendStatus(order)
			case "start":
				if statusmap[mconfig.McsmData[order].Name] == 0 {
					Start(order)
				} else {
					Send_group_msg(fmt.Sprint("服务器 ", mconfig.McsmData[order].Name, " 已在运行"), order)
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
				RunCmd(params, order)

			}
		}(params2[2], od)
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
