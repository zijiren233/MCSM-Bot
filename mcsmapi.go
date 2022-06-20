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
		Order       int      `json:"id"`
		Sendtype    string   `json:"sendtype"`
		Name        string   `json:"name"`
		Domain      string   `json:"url"`
		Remote_uuid string   `json:"remote_uuid"`
		Uuid        string   `json:"uuid"`
		Apikey      string   `json:"apikey"`
		Group_id    string   `json:"group_id"`
		Adminlist   []string `json:"adminlist"`
	} `json:"mcsmdata"`
}

type Status struct {
	Data struct {
		Data []struct {
			Status int `json:"status"`
		} `json:"data"`
	} `json:"data"`
}

type CmdData struct {
	Time_unix int64 `json:"time"`
}

func GetMConfig() MConfig {
	var config MConfig
	f, err := os.OpenFile("config.json", os.O_RDONLY, 0755)
	if err != nil {
		fmt.Printf("读取配置文件出错: %v\n", err)
		go log.Error("读取配置文件出错: %v", err)
		f, _ := os.OpenFile("config.json", os.O_CREATE|os.O_WRONLY, 0755)
		f.WriteString(`{
	"mcsmdata": [
		{
			"id": 0,
			"name": "server1",
			"url": "https://mcsm.domain.com:443",
			"remote_uuid": "d6a27b0b13ad44ce879b5a56c88b4d34",
			"uuid": "a8788991a64e4a06b76d539b35db1b16",
			"apikey": "vmajkfnvklNSdvkjbnfkdsnv7e0f",
			"group_id": "234532",
			"adminlist": [
				"1145141919",
				"1433223"
			]
		},
		{
			"id": 1,
			"name": "server2",
			"url": "http://mcsm.domain.com:24444",
			"remote_uuid": "d6a27b0b13ad44ce879b5ascwfscr323",
			"uuid": "a8788991a6acasfaca76d539b35db1b16",
			"apikey": "6ewc6292daefvlksmdvjadnvjbf",
			"group_id": "234532",
			"adminlist": [
				"114514",
				"1919"
			]
		}
	],
	"cqhttp": {
		"token": "test",
		"url": "https://q-api.pyhdxy.com:443",
		"qq": "3426898431"
	}
}`)
		f2, _ := os.OpenFile("config.sample.json", os.O_CREATE|os.O_WRONLY, 0755)
		f2.WriteString(`{ // 真正的配置文件为标准的json格式，里面不要有注释！！！
	"mcsmdata": [
		{
			"id": 0, // 按顺序填,此项为监听服务器的序号，从0开始依次增加，用于启动监听时填的要监听哪一个服务器
			"name": "server1", // MCSM里面的实例名，即基本信息里的昵称，实例名不可重复！！！
			"url": "https://mcsm.domain.com:443", // MCSM面板的地址，包含http(s)://，结尾不要有斜杠/
			"remote_uuid": "d6a27b0b13ad44ce879b5a56c88b4d34", // 守护进程的GID （守护进程标识符）
			"uuid": "a8788991a64e4a06b76d539b35db1b16", // 实例的UID （远程/本地实例标识符）
			"apikey": "vmajkfnvklNSdvkjbnfkdsnv7e0f", // 不可为空，用户中心->右上角个人资料->右方生成API密钥
			"group_id": "234532", // 要管理的QQ群号
			"adminlist": [
				"1145141919", // 群管理员，第一个为主管理员，只有管理员才可以发送命令
				"1433223" // 管理员列表可以为空，则所有用户都可以发送命令
			]
		}, // 只有一个实例可以删掉后面的服务器，有多个则自行添加
		{
			"id": 1, // 按顺序填，0，1，2，3 ......
			"name": "server2",
			"url": "http://mcsm.domain.com:24444",
			"remote_uuid": "d6a27b0b13ad44ce879b5ascwfscr323",
			"uuid": "a8788991a6acasfaca76d539b35db1b16",
			"apikey": "6ewc6292daefvlksmdvjadnvjbf",
			"group_id": "234532",
			"adminlist": [
				"114514", // 不同实例在同一个群也可以有不同的管理员
				"1919"
			]
		} // <--最后一个实例配置这里没有逗号！！！
	],
	"cqhttp": {
		"token": "test", // cqhttp配置文件里的一个配置项，即 默认中间件锚点 中的 access-token ，不可为空
		"url": "https://q-api.pyhdxy.com:443", // cqhttp 请求地址，末尾不带斜杠！
		"qq": "3426898431" // 机器人QQ号
	}
}`)
		fmt.Println("已创建配置文件config.json 和 config.sample.json，请根据注释填写配置")
		go log.Error("已创建配置文件config.json 和 config.sample.json，请根据注释填写配置")
		panic(err)
	}
	b, _ := ioutil.ReadAll(f)
	err2 := json.Unmarshal(b, &config)
	if err2 != nil {
		fmt.Printf("配置文件内容出错: %v\n", err2)
		go log.Error("配置文件内容出错: %v", err2)
		fmt.Print("可能是配置文件内容格式错误 或 配置文件格式和当前版本不匹配，删除当前配置文件重新启动以获取最新配置文件模板")
		panic(err2)
	}
	return config
}

func ReturnResult(command string, order int, time_now int64) {
	client := &http.Client{}
	r2, _ := http.NewRequest("GET", mconfig.McsmData[order].Domain+"/api/protected_instance/outputlog", nil)
	r2.Close = true
	r2.Header.Set("x-requested-with", "xmlhttprequest")
	q := r2.URL.Query()
	q.Add("apikey", mconfig.McsmData[order].Apikey)
	q.Add("uuid", mconfig.McsmData[order].Uuid)
	q.Add("remote_uuid", mconfig.McsmData[order].Remote_uuid)
	r2.URL.RawQuery = q.Encode()
	r, err := client.Do(r2)
	if err != nil {
		Send_group_msg("获取运行结果失败！", order)
		go log.Error("获取服务器 %s 命令 %s 运行结果失败！", mconfig.McsmData[order].Name, command)
		return
	}
	b, _ := ioutil.ReadAll(r.Body)
	r3, _ := regexp.Compile(`\\r+|\\u001b\[?=?[a-zA-Z]?\?*[0-9]*[hl]*>? ?[0-9;]*m*`)
	ret := r3.ReplaceAllString(string(b), "")
	last := strings.LastIndex(ret, `","time":`)
	var index int
	var i int64
	go log.Debug("服务器 %s 运行命令 %s 返回时间: %s", mconfig.McsmData[order].Name, command, time.Unix((time_now/1000)+i, 0).Format("15:04:05"))
	for i = 0; i <= 2; i++ {
		index = strings.Index(ret, time.Unix((time_now/1000)+i, 0).Format("15:04:05"))
		if index == -1 {
			continue
		}
		Send_group_msg(fmt.Sprintf("> [%s] %s\n%s", mconfig.McsmData[order].Name, command, handle_End_Newline(ret[index-1:last])), order)
		return
	}
	index = strings.Index(ret, time.Unix((time_now/1000)-1, 0).Format("15:04:05"))
	if index == -1 {
		Send_group_msg("运行命令成功！", order)
		go log.Warring("服务器 %s 命令 %s 成功,但未查找到返回时间: %s", mconfig.McsmData[order].Name, command, time.Unix((time_now/1000)+i, 0).Format("15:04:05"))
		return
	}
	Send_group_msg(fmt.Sprintf("> [%s] %s\n%s", mconfig.McsmData[order].Name, command, handle_End_Newline(ret[index-1:last])), order)
}

func handle_End_Newline(msg string) string {
	var data Data
	last := strings.LastIndex(msg, `\n`)
	if last != len(msg)-2 {
		last = len(msg) - 1
	}
	msg = fmt.Sprint(`{"data":"`, msg[:last], `"}`)
	json.Unmarshal([]byte(msg), &data)
	return data.Data
}

func RunCmd(commd string, order int) {
	client := &http.Client{}
	r2, _ := http.NewRequest("GET", mconfig.McsmData[order].Domain+"/api/protected_instance/command", nil)
	r2.Close = true
	q := r2.URL.Query()
	q.Add("apikey", mconfig.McsmData[order].Apikey)
	q.Add("uuid", mconfig.McsmData[order].Uuid)
	q.Add("remote_uuid", mconfig.McsmData[order].Remote_uuid)
	q.Add("command", commd)
	r2.URL.RawQuery = q.Encode()
	r2.Header.Set("x-requested-with", "xmlhttprequest")
	r, err := client.Do(r2)
	if err != nil {
		Send_group_msg(fmt.Sprintf("运行命令 %s 失败！", commd), order)
		go log.Error("运行命令 %s 失败！%v", commd, err)
		return
	}
	b, _ := ioutil.ReadAll(r.Body)
	var time_unix CmdData
	json.Unmarshal(b, &time_unix)
	time.Sleep(75 * time.Millisecond)
	ReturnResult(commd, order, time_unix.Time_unix)
}

func RunningTest(order int) bool {
	client := &http.Client{}
	r2, _ := http.NewRequest("GET", mconfig.McsmData[order].Domain+"/api/service/remote_service_instances", nil)
	r2.Close = true
	q := r2.URL.Query()
	q.Add("apikey", mconfig.McsmData[order].Apikey)
	q.Add("page", "1")
	q.Add("page_size", "1")
	q.Add("instance_name", mconfig.McsmData[order].Name)
	q.Add("remote_uuid", mconfig.McsmData[order].Remote_uuid)
	r2.URL.RawQuery = q.Encode()
	r2.Header.Set("x-requested-with", "xmlhttprequest")
	r, err := client.Do(r2)
	if err != nil {
		go log.Warring("检测服务器 %s 运行状况失败,可能是网络原因导致!", mconfig.McsmData[order].Name)
		return statusmap[mconfig.McsmData[order].Name]-1 == 0
	}
	b, err2 := ioutil.ReadAll(r.Body)
	if err2 != nil {
		go log.Warring("检测服务器 %s 状态读取结果错误!", mconfig.McsmData[order].Name)
		return statusmap[mconfig.McsmData[order].Name]-1 == 0
	}
	var status Status
	json.Unmarshal(b, &status)
	if len(status.Data.Data) == 0 {
		go log.Warring("检测服务器 %s 运行状况失败,可能是网络原因导致!", mconfig.McsmData[order].Name)
		return statusmap[mconfig.McsmData[order].Name]-1 == 0
	}
	if status.Data.Data[0].Status != 3 && status.Data.Data[0].Status != 2 {
		return false
	} else {
		return true
	}
}

func SendStatus(order int) {
	if statusmap[mconfig.McsmData[order].Name] == 1 {
		Send_group_msg(fmt.Sprint(`[CQ:at,qq=`, mconfig.McsmData[order].Adminlist[0], `]`, "服务器", mconfig.McsmData[order].Name, "正在运行"), order)
	} else if statusmap[mconfig.McsmData[order].Name] == 0 {
		Send_group_msg(fmt.Sprint(`[CQ:at,qq=`, mconfig.McsmData[order].Adminlist[0], `]`, "服务器", mconfig.McsmData[order].Name, "未运行"), order)
	} else {
		Send_group_msg(fmt.Sprint(`[CQ:at,qq=`, mconfig.McsmData[order].Adminlist[0], `]`, "未监听", mconfig.McsmData[order].Name), order)
	}
}

func Start(order int) {
	client := &http.Client{}
	r2, _ := http.NewRequest("GET", mconfig.McsmData[order].Domain+"/api/protected_instance/open", nil)
	r2.Close = true
	q := r2.URL.Query()
	q.Add("apikey", mconfig.McsmData[order].Apikey)
	q.Add("uuid", mconfig.McsmData[order].Uuid)
	q.Add("remote_uuid", mconfig.McsmData[order].Remote_uuid)
	r2.URL.RawQuery = q.Encode()
	r2.Header.Set("x-requested-with", "xmlhttprequest")
	_, err := client.Do(r2)
	if err != nil {
		Send_group_msg(fmt.Sprintf("服务器:%s 运行启动命令失败!", mconfig.McsmData[order].Name), order)
		go log.Warring("服务器:%s 运行启动命令失败,可能是网络问题!", mconfig.McsmData[order].Name)
		return
	}
}

func Stop(order int) {
	client := &http.Client{}
	r2, _ := http.NewRequest("GET", mconfig.McsmData[order].Domain+"/api/protected_instance/stop", nil)
	r2.Close = true
	q := r2.URL.Query()
	q.Add("apikey", mconfig.McsmData[order].Apikey)
	q.Add("uuid", mconfig.McsmData[order].Uuid)
	q.Add("remote_uuid", mconfig.McsmData[order].Remote_uuid)
	r2.URL.RawQuery = q.Encode()
	r2.Header.Set("x-requested-with", "xmlhttprequest")
	_, err := client.Do(r2)
	if err != nil {
		Send_group_msg(fmt.Sprintf("服务器:%s 运行关闭命令失败!", mconfig.McsmData[order].Name), order)
		go log.Warring("服务器:%s 运行关闭命令失败,可能是网络问题!", mconfig.McsmData[order].Name)
		return
	}
}

func TestMcsmStatus(order int) bool {
	client := &http.Client{}
	r2, _ := http.NewRequest("GET", mconfig.McsmData[order].Domain+"/api/service/remote_service_instances", nil)
	r2.Close = true
	q := r2.URL.Query()
	q.Add("apikey", mconfig.McsmData[order].Apikey)
	q.Add("page", "1")
	q.Add("page_size", "1")
	q.Add("instance_name", mconfig.McsmData[order].Name)
	q.Add("remote_uuid", mconfig.McsmData[order].Remote_uuid)
	r2.URL.RawQuery = q.Encode()
	r2.Header.Set("x-requested-with", "xmlhttprequest")
	_, err2 := client.Do(r2)
	if err2 != nil {
		fmt.Printf("服务器:%s MCSM前端连接失败，请检查配置文件是否填写正确或MCSM是否启动\n", mconfig.McsmData[order].Name)
		go log.Error("服务器:%s MCSM前端连接失败，请检查配置文件是否填写正确或MCSM是否启动", mconfig.McsmData[order].Name)
		return false
	}
	return true
}

func Restart(order int) {
	client := &http.Client{}
	r2, _ := http.NewRequest("GET", mconfig.McsmData[order].Domain+"/api/protected_instance/restart", nil)
	r2.Close = true
	q := r2.URL.Query()
	q.Add("apikey", mconfig.McsmData[order].Apikey)
	q.Add("uuid", mconfig.McsmData[order].Uuid)
	q.Add("remote_uuid", mconfig.McsmData[order].Remote_uuid)
	r2.URL.RawQuery = q.Encode()
	r2.Header.Set("x-requested-with", "xmlhttprequest")
	_, err := client.Do(r2)
	if err != nil {
		Send_group_msg(fmt.Sprintf("服务器:%s 运行重启命令失败!", mconfig.McsmData[order].Name), order)
		go log.Warring("服务器:%s 运行重启命令失败,可能是网络问题!", mconfig.McsmData[order].Name)
		return
	}
}

func Kill(order int) {
	client := &http.Client{}
	r2, _ := http.NewRequest("GET", mconfig.McsmData[order].Domain+"/api/protected_instance/kill", nil)
	r2.Close = true
	q := r2.URL.Query()
	q.Add("apikey", mconfig.McsmData[order].Apikey)
	q.Add("uuid", mconfig.McsmData[order].Uuid)
	q.Add("remote_uuid", mconfig.McsmData[order].Remote_uuid)
	r2.URL.RawQuery = q.Encode()
	r2.Header.Set("x-requested-with", "xmlhttprequest")
	_, err := client.Do(r2)
	if err != nil {
		Send_group_msg(fmt.Sprintf("服务器:%s 运行终止命令失败!", mconfig.McsmData[order].Name), order)
		go log.Warring("服务器:%s 运行终止命令失败,可能是网络问题!", mconfig.McsmData[order].Name)
		return
	}
}
