package gconfig

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

type MConfig struct {
	McsmData []struct {
		Id          int    `json:"id"`
		Sendtype    string `json:"sendtype"`
		Name        string `json:"name"`
		Url         string `json:"url"`
		Remote_uuid string `json:"remote_uuid"`
		Uuid        string `json:"uuid"`
		Apikey      string `json:"apikey"`
		Group_id    int    `json:"group_id"`
		Adminlist   []int  `json:"adminlist"`
	} `json:"mcsmdata"`
}

type QConfig struct {
	Cqhttp struct {
		Url string `json:"url"`
		Qq  int    `json:"qq"`
		Op  int    `json:"op"`
	} `json:"cqhttp"`
}

func GetMConfig() MConfig {
	var config MConfig
	f, err := os.OpenFile("config.json", os.O_RDONLY, 0755)
	if err != nil {
		fmt.Printf("读取配置文件出错: %v\n", err)
		f, _ := os.OpenFile("config.json", os.O_CREATE|os.O_WRONLY, 0755)
		f.WriteString(`{
	"mcsmdata": [
		{
			"id": 1,
			"name": "LYC_01",
			"url": "https://mc.pyhdxy.com",
			"remote_uuid": "d6a27b0b13ad44ce879b5a56c88b4d34",
			"uuid": "a8788991a64e4a06b76d539b35db1b16",
			"apikey": "6e54ec06292d4b77aef4ada844ae7e0f",
			"group_id": 383033610,
			"adminlist": [
				1670605849,
				1145141919
			]
		},
		{
			"id": 2,
			"name": "java",
			"url": "http://154.209.91.180:23333",
			"remote_uuid": "3ec8d0ff584c43bd95598b18949a8bac",
			"uuid": "76a49c5ef46a41f29b374109d58f994a",
			"apikey": "11717c92b7274114ac89b9ef0abf3e7d",
			"group_id": 383033610,
			"adminlist": [
				1670605849,
				1145141919
			]
		}
	],
	"cqhttp": {
		"url": "ws://q-api.pyhdxy.com:8080",
		"qq": 3426898431,
		"op": 1670605849
	}
}`)
		f2, _ := os.OpenFile("config.sample.json", os.O_CREATE|os.O_WRONLY, 0755)
		f2.WriteString(`{ // 真正的配置文件为标准的json格式，里面不要有注释！！！
	"mcsmdata": [
		{
			"id": 2, // Id 为任意小于256的数，但不可重复！
			"name": "server1", // MCSM里面的实例名，即基本信息里的昵称，实例名不可重复！！！
			"url": "https://mcsm.domain.com:443", // MCSM面板的地址，包含http(s)://，结尾不要有斜杠/
			"remote_uuid": "d6a27b0b13ad44ce879b5a56c88b4d34", // 守护进程的GID
			"uuid": "a8788991a64e4a06b76d539b35db1b16", // 实例的UID
			"apikey": "vmajkfnvklNSdvkjbnfkdsnv7e0f", // 不可为空，用户中心->右上角个人资料->右方生成API密钥
			"group_id": 234532, // 要管理的QQ群号
			"adminlist": [
				1145141919, // 群管理员，第一个为主管理员，只有管理员才可以发送命令
				1433223 // 管理员列表可以为空，则所有用户都可以发送命令
			]
		}, // 只有一个实例可以删掉后面的服务器，有多个则自行添加
		{
			"id": 5, // Id 不可重复！
			"name": "server2",
			"url": "http://mcsm.domain.com:24444",
			"remote_uuid": "d6a27b0b13ad44ce879b5ascwfscr323",
			"uuid": "a8788991a6acasfaca76d539b35db1b16",
			"apikey": "6ewc6292daefvlksmdvjadnvjbf",
			"group_id": 234532,
			"adminlist": [
				114514, // 不同实例在同一个群也可以有不同的管理员
				1919
			]
		} // <--最后一个实例配置这里没有逗号！！！
	],
	"cqhttp": {
		"url": "ws://127.0.0.1:8080", // cqhttp 请求地址，末尾不带斜杠！只能使用Ws(s)协议
		"qq": 3426898431, // 机器人QQ号
		"op": 1670605849 // op里的qq可以在私聊机器人以访问所有实例
	}
}`)
		fmt.Println("已创建配置文件config.json 和 config.sample.json，请根据注释填写配置")
		panic(err)
	}
	b, _ := ioutil.ReadAll(f)
	err2 := json.Unmarshal(b, &config)
	if err2 != nil {
		fmt.Printf("配置文件内容出错: %v\n", err2)
		fmt.Print("可能是配置文件内容格式错误 或 配置文件格式和当前版本不匹配，删除当前配置文件重新启动以获取最新配置文件模板")
		panic(err2)
	}
	return config
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
		panic(err2)
	}
	return config
}
