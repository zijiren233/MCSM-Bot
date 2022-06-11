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
		Order       int      `json:"order"`
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
	f, err := os.OpenFile("config.json", os.O_RDWR, 0777)
	if err != nil {
		fmt.Printf("读取配置文件出错: %v\n", err)
		f, _ := os.OpenFile("config.json", os.O_CREATE, 0777)
		f.WriteString(`{
	"mcsmdata": [
		{
			"order": 0,
			"sendtype": "QQ",
			"name": "server1",
			"url": "https://mcsm.domain.com:443",
			"remote_uuid": "d6a27b0b13ad44ce879b5a56c88b4d34",
			"uuid": "a8788991a64e4a06b76d539b35db1b16",
			"apikey": "vmajkfnvklNSdvkjbnfkdsnv7e0f",
			"group_id": "234532",
			"adminlist": [
				"1145141919",
				"1145141919"
			]
		},
		{
			"order": 1,
			"sendtype": "TG",
			"name": "server2",
			"url": "http://mcsm.domain.com:24444",
			"remote_uuid": "d6a27b0b13ad44ce879b5ascwfscr323",
			"uuid": "a8788991a6acasfaca76d539b35db1b16",
			"apikey": "6ewc6292daefvlksmdvjadnvjbf",
			"group_id": "234532",
			"adminlist": [
				"1145141919",
				"1145141919"
			]
		}
	],
	"cqhttp": {
		"token": "test",
		"url": "http://10.10.10.4:5700",
		"qq": "3333446431"
	}
}`)
		f2, _ := os.OpenFile("config.sample.json", os.O_CREATE, 0777)
		f2.WriteString(`{
	"mcsmdata": [
		{
			"order": 0, // 按顺序填,此项为监听服务器的序号，从0开始依次增加，用于启动监听时填的要监听哪一个服务器
			"sendtype": "QQ", // 暂时只有QQ
			"name": "server1", // MCSM里面的实例名，即基本信息里的昵称，实例名不可重复！！！
			"url": "https://mcsm.domain.com:443", // MCSM面板的地址，包含http(s)//，结尾不要有斜杠/
			"remote_uuid": "d6a27b0b13ad44ce879b5a56c88b4d34", // 守护进程的GID （守护进程标识符）
			"uuid": "a8788991a64e4a06b76d539b35db1b16", // 实例的UID （远程/本地实例标识符）
			"apikey": "vmajkfnvklNSdvkjbnfkdsnv7e0f", // 不可为空，用户中心->右上角个人资料->右方生成API密钥
			"group_id": "234532", // 要管理的QQ群号，如果多个实例要监听同一个群，那么下面的群管理员列表应该设置相同
			"adminlist": [
				"1145141919", // 群管理员，第一个为主管理员，只有管理员才可以发送命令
				"1145141919" // 管理员列表可以为空，则所有用户都可以发送命令
			]
		}, // 只有一个实例则可以删掉后面的这个order，有多个则自行添加
		{
			"order": 1,
			"sendtype": "TG",
			"name": "server2",
			"url": "http://mcsm.domain.com:24444",
			"remote_uuid": "d6a27b0b13ad44ce879b5ascwfscr323",
			"uuid": "a8788991a6acasfaca76d539b35db1b16",
			"apikey": "6ewc6292daefvlksmdvjadnvjbf",
			"group_id": "234532", // 多个实例监听同一个群，下面的管理员列表应设置一样
			"adminlist": [
				"1145141919", // 多个实例监听同一个群，管理员列表应设置一样
				"1145141919"
			]
		} // <--最后一个实例配置这里没有逗号！！！
	],
	"cqhttp": {
		"token": "test", // cqhttp配置文件里的一个配置项，即 默认中间件锚点 中的 access-token ，不可为空
		"url": "http://10.10.10.4:5700", // cqhttp 请求地址，末尾不带斜杠！
		"qq": "3333446431" // 机器人QQ号
	}
}`)
		fmt.Println("已创建配置文件config.json 和 config.sample.json，请根据注释填写配置")
		os.Exit(0)
	}
	b, err2 := ioutil.ReadAll(f)
	if err2 != nil {
		fmt.Printf("读取配置文件出错: %v\n", err2)
		os.Exit(0)
	}
	err3 := json.Unmarshal(b, &config)
	if err3 != nil {
		fmt.Printf("配置文件内容出错: %v\n", err3)
		fmt.Print("可能是配置文件内容格式错误 或 配置文件格式和当前版本不匹配，删除当前配置文件重新启动以获取最新配置文件模板")
		os.Exit(0)
	}
	defer f.Close()
	return config
}

func ReturnResult(command string, order int, time_now int64) {
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
	r3, _ := regexp.Compile(`\\r+|\\n|\\u001b\[?=?[a-zA-Z]?\?*[0-9]*[hl]*>? ?[0-9;]*m*`)
	ret := r3.ReplaceAllString(string(b), "")
	last := strings.LastIndex(ret, `","time":`)
	r4, _ := regexp.Compile(`([0-1][0-9]|2[0-3]):([0-5][0-9]):([0-5][0-9])`)
	index := strings.Index(ret, r4.FindString(time.Unix(time_now/1000, 0).Format("2006-01-02 15:04:05")))
	if strings.IndexAny(ret[index:], `]`) == 8 {
		ret = fmt.Sprint(ret[:index+8] + " " + ret[index+9:])
	}
	if index == -1 {
		index = strings.Index(ret, r4.FindString(time.Unix((time_now/1000)+1, 0).Format("2006-01-02 15:04:05")))
		if strings.IndexAny(ret[index:], `]`) == 8 {
			ret = fmt.Sprint(ret[:index+8] + " " + ret[index+9:])
		}
		if index == -1 {
			Send_group_msg("运行命令成功！", order)
			return
		}
		Send_group_msg(ret[index:last], order)
		return
	}
	Send_group_msg(ret[index:last], order)
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
	r, _ := client.Do(r2)
	b, _ := ioutil.ReadAll(r.Body)
	var time_unix CmdData
	json.Unmarshal(b, &time_unix)
	ReturnResult(commd, order, time_unix.Time_unix)
}

func RunningTest(order int) bool {
	client := &http.Client{}
	r2, _ := http.NewRequest("GET", mconfig.McsmData[order].Domain+"/api/service/remote_service_instances", nil)
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
	err := json.Unmarshal(b, &status)
	if err != nil {
		fmt.Println("未检测到实例！请检查实例状态或配置文件是否填写正确！")
		os.Exit(1)
	}
	if len(status.Data.Data) == 0 {
		fmt.Println("未检测到实例！请检查实例状态或配置文件是否填写正确！")
		os.Exit(1)
	}
	if status.Data.Data[0].Status == 3 || status.Data.Data[0].Status == 2 {
		return true
	} else {
		return false
	}
}

func SendStatus(order int) {
	if statusmap[mconfig.McsmData[order].Name] == 1 {
		Send_group_msg(fmt.Sprint("服务器", mconfig.McsmData[order].Name, "正在运行"), order)
	} else if statusmap[mconfig.McsmData[order].Name] == 0 {
		Send_group_msg(fmt.Sprint("服务器", mconfig.McsmData[order].Name, "未运行"), order)
	} else {
		Send_group_msg("未监听Order", order)
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

func TestMcsmStatus(order int) {
	client := &http.Client{}
	r2, err := http.NewRequest("GET", mconfig.McsmData[order].Domain+"/api/service/remote_service_instances", nil)
	if err != nil {
		fmt.Println("检测MCSM后端连接失败，请检查配置文件是否填写正确或MCSM是否启动")
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
	r, err2 := client.Do(r2)
	if err2 != nil {
		fmt.Println("检测MCSM后端连接失败，请检查配置文件是否填写正确或MCSM是否启动")
		os.Exit(1)
	}
	defer r.Body.Close()
}
