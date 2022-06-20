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
	f, err := os.OpenFile("config.json", os.O_RDWR, 0755)
	if err != nil {
		fmt.Printf("打开配置文件出错: %v\n", err)
		panic(err)
	}
	b, _ := ioutil.ReadAll(f)
	err2 := json.Unmarshal(b, &config)
	if err2 != nil {
		fmt.Printf("读取配置文件出错: %v\n", err2)
		log.Error("读取配置文件出错: %v", err2)
		panic(err2)
	}
	return config
}

func TestCqhttpStatus(order int) {
	client := &http.Client{}
	data := url.Values{}
	data.Set("group_id", mconfig.McsmData[order].Group_id)
	data.Set("access_token", qconfig.Cqhttp.Token)
	r2, _ := http.NewRequest("GET", qconfig.Cqhttp.Url+"/get_group_msg_history"+"?"+data.Encode(), nil)
	r2.Close = true
	_, err := client.Do(r2)
	if err != nil {
		fmt.Println("Cqhttp 状态检测错误，请检查配置文件或 Cqhttp 状态")
		log.Error("Cqhttp 状态检测错误，请检查配置文件或 Cqhttp 状态 err:%v", err)
		panic(err)
	}
}

func Get_group_new_msg(order int, chan_message chan Mdata) {
	client := &http.Client{}
	data := url.Values{}
	data.Set("group_id", mconfig.McsmData[order].Group_id)
	data.Set("access_token", qconfig.Cqhttp.Token)
	r2, _ := http.NewRequest("GET", qconfig.Cqhttp.Url+"/get_group_msg_history"+"?"+data.Encode(), nil)
	r2.Close = true
	r, _ := client.Do(r2)
	b, _ := ioutil.ReadAll(r.Body)
	var mesdata MesData
	json.Unmarshal(b, &mesdata)
	tmp := mesdata.Data.Messages[len(mesdata.Data.Messages)-1].Message_id
	var err error
	for {
		r, err = client.Do(r2)
		if err != nil {
			log.Error("获取群组:%s 消息失败 err:%v", mconfig.McsmData[order].Group_id, err)
			continue
		}
		b, _ = ioutil.ReadAll(r.Body)
		err = json.Unmarshal(b, &mesdata)
		if err != nil {
			log.Error("返回群组:%s 消息错误 err:%v", mconfig.McsmData[order].Group_id, err)
			continue
		}
		if mesdata.Data.Messages[len(mesdata.Data.Messages)-1].Message_id != tmp {
			go write_In_Chan_Latest_News(tmp, mesdata, chan_message, order)
			tmp = mesdata.Data.Messages[len(mesdata.Data.Messages)-1].Message_id
		}
		// 获取消息间隔
		time.Sleep(400 * time.Millisecond)
	}
}

func write_In_Chan_Latest_News(tmp int, mesdata MesData, chan_message chan Mdata, order int) {
	i := len(mesdata.Data.Messages) - 1
	for i >= 0 {
		if mesdata.Data.Messages[i].Message_id == tmp {
			i++
			break
		}
		i--
	}
	if i == -1 {
		log.Error("群组:%s 消息刷新过快! 请调低获取消息间隔!", mconfig.McsmData[order].Group_id)
		return
	}
	for i <= len(mesdata.Data.Messages)-1 {
		if strconv.Itoa(mesdata.Data.Messages[i].User_id) != qconfig.Cqhttp.Qq {
			chan_message <- mesdata.Data.Messages[i]
			log.Debug("群组:%s 获取到最新消息 QQ:%d ,iD:%d", mconfig.McsmData[order].Group_id, mesdata.Data.Messages[i].User_id, mesdata.Data.Messages[i].Message_id)
		}
		i++
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
	r2.Close = true
	_, err := client.Do(r2)
	if err != nil {
		log.Warring("发送消息到群组:%s 失败,可能是cqhttp或网络问题!", mconfig.McsmData[order].Group_id)
		return
	}
}

func in(target string, str_array []string) bool {
	for _, element := range str_array {
		if target == element {
			return true
		}
	}
	return false
}

// func getKeys(m map[int]int) []int {
// 	keys := make([]int, 0, len(m))
// 	for k := range m {
// 		keys = append(keys, k)
// 	}
// 	return keys
// }

func AddQListen(order int) {
	TestCqhttpStatus(order)
	fmt.Println("监听实例 ", mconfig.McsmData[order].Name, " 成功！")
	log.Info("监听实例 %s 成功", mconfig.McsmData[order].Name)
	// 获取已监听的 []ID
	// i := getKeys(listenmap)
	var listen = make([]int, 0, len(GetMConfig().McsmData))
	// fmt.Printf("listen: %v\n", listen)
	listenmap.Range(func(key, value interface{}) bool {
		listen = append(listen, key.(int))
		return true
	})
	// fmt.Printf("test: %v\n", test)
	for j := range listen {
		if mconfig.McsmData[j].Group_id == mconfig.McsmData[order].Group_id && j != order {
			// fmt.Println("监听相同的群")
			log.Info("服务器:%s 监听相同的群:%s", mconfig.McsmData[order].Name, mconfig.McsmData[order].Group_id)
			go ReportStatus(order)
			listenmap.Store(order, 1)
			return
		} else if j == order {
			log.Info("服务器:%s 重复监听", mconfig.McsmData[order].Name)
			return
		}
	}
	// 设置缓存大小为25
	chan_message := make(chan Mdata, 25)
	var od int
	var params string
	var params2 []string
	var mdata Mdata
	go Get_group_new_msg(order, chan_message)
	go ReportStatus(order)
	listenmap.Store(order, 1)
	flysnowRegexp, _ := regexp.Compile(`^run ([0-9]*) *(.*)`)
	for mdata = range chan_message {
		params = flysnowRegexp.FindString(mdata.Message)
		if len(params) == 0 {
			continue
		}
		params2 = flysnowRegexp.FindStringSubmatch(params)
		if params2[1] != "" {
			od, _ = strconv.Atoi(params2[1])
		} else {
			od = order
		}
		log.Info("群组:%s QQ:%d 输入命令: %s", mconfig.McsmData[od].Group_id, mdata.User_id, mdata.Message)
		if od >= len(mconfig.McsmData) {
			go Send_group_msg(fmt.Sprint(`[CQ:at,qq=`, mdata.User_id, `]`, "ID错误！"), order)
			log.Info("群组:%s QQ:%d 试图访问服务器ID:%d ,ID超出范围", mconfig.McsmData[od].Group_id, mdata.User_id, od)
			continue
		}
		if mconfig.McsmData[order].Group_id != mconfig.McsmData[od].Group_id {
			go Send_group_msg(fmt.Sprint(`[CQ:at,qq=`, mdata.User_id, `]`, "ID错误！"), order)
			log.Info("群组:%s QQ:%d 试图访问服务器:%s ,服务器属于群组:%s", mconfig.McsmData[od].Group_id, mdata.User_id, mconfig.McsmData[od].Name, mconfig.McsmData[od].Group_id)
			continue
		}
		if !in(strconv.Itoa(mdata.User_id), mconfig.McsmData[od].Adminlist) {
			go Send_group_msg(fmt.Sprint(`[CQ:at,qq=`, mdata.User_id, `]`, "权限不足！"), order)
			log.Info("群组:%s QQ:%d 试图访问服务器:%s ,权限不足", mconfig.McsmData[od].Group_id, mdata.User_id, mconfig.McsmData[od].Name)
			continue
		}
		tmpl, _ := listenmap.Load(od)
		if tmpl != 1 {
			go Send_group_msg(fmt.Sprint(`[CQ:at,qq=`, mdata.User_id, `]`, mconfig.McsmData[od].Name, "未开启监听！"), order)
			log.Info("群组:%s QQ:%d 试图访问服务器:%s ,服务器未开启监听", mconfig.McsmData[od].Group_id, mdata.User_id, mconfig.McsmData[od].Name)
			continue
		}
		if params2[2] == "" {
			log.Info("群组:%s QQ:%d 输入命令为空！", mconfig.McsmData[od].Group_id, mdata.User_id)
			continue
		}
		tmp, _ := statusmap.Load(mconfig.McsmData[order].Name)
		if tmp == 0 && params2[2] != "start" {
			go Send_group_msg(fmt.Sprint(`[CQ:at,qq=`, mdata.User_id, `]`, mconfig.McsmData[od].Name, "未运行！"), order)
			log.Info("群组:%s QQ:%d 试图访问服务器:%s ,服务器未运行", mconfig.McsmData[od].Group_id, mdata.User_id, mconfig.McsmData[od].Name)
			continue
		}
		go checkCMD(params2[2], od)
	}
}

func checkCMD(params string, order int) {
	params = strings.ReplaceAll(params, "\n", "")
	params = strings.ReplaceAll(params, "\r", "")
	log.Debug("服务器:%s 运行命令:%s", mconfig.McsmData[order].Name, params)
	tmp, _ := statusmap.Load(mconfig.McsmData[order].Name)
	switch params {
	case "status":
		SendStatus(order)
	case "start":
		if tmp == 0 {
			Send_group_msg(fmt.Sprint("服务器 ", mconfig.McsmData[order].Name, " 正在启动"), order)
			Start(order)
		} else {
			Send_group_msg(fmt.Sprint("服务器 ", mconfig.McsmData[order].Name, " 已在运行"), order)
		}
	case "stop":
		if tmp == 1 {
			Send_group_msg(fmt.Sprint("服务器 ", mconfig.McsmData[order].Name, " 正在关闭"), order)
			Stop(order)
		} else {
			Send_group_msg(fmt.Sprint("服务器 ", mconfig.McsmData[order].Name, " 未在运行"), order)
		}
	case "restart":
		Send_group_msg(fmt.Sprint("服务器 ", mconfig.McsmData[order].Name, " 正在重启"), order)
		Restart(order)
	case "kill":
		Kill(order)
	default:
		RunCmd(params, order)
	}
}

func ReportStatus(order int) {
	var tmp any
	for {
		tmp, _ = statusmap.Load(mconfig.McsmData[order].Name)
		if !RunningTest(order) && tmp == 1 {
			statusmap.Store(mconfig.McsmData[order].Name, 0)
			log.Info("实例 %s 已停止", mconfig.McsmData[order].Name)
			Send_group_msg(fmt.Sprint(`[CQ:at,qq=`, mconfig.McsmData[order].Adminlist[0], `]`, "服务器 ", mconfig.McsmData[order].Name, " 已停止！"), order)
		} else if RunningTest(order) && tmp == 0 {
			statusmap.Store(mconfig.McsmData[order].Name, 1)
			log.Info("实例 %s 已运行", mconfig.McsmData[order].Name)
			Send_group_msg(fmt.Sprint(`[CQ:at,qq=`, mconfig.McsmData[order].Adminlist[0], `]`, "服务器 ", mconfig.McsmData[order].Name, " 已启动！"), order)
		}
		time.Sleep(2 * time.Second)
	}
}
