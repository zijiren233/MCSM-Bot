package gconfig

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"

	"github.com/zijiren233/MCSM-Bot/utils"
)

var Mconfig mConfig
var Qconfig qConfig
var Log = log.Default()

type mConfig struct {
	McsmData []struct {
		Id                   int      `json:"id"`
		Url                  string   `json:"url"`
		Remote_uuid          string   `json:"remote_uuid"`
		Uuid                 string   `json:"uuid"`
		Apikey               string   `json:"apikey"`
		Group_id             int      `json:"group_id"`
		User_allows_commands []string `json:"user_allows_commands"`
		Adminlist            []int    `json:"adminlist"`
	} `json:"mcsmdata"`
}

type qConfig struct {
	Cqhttp struct {
		Url string `json:"url"`
		Qq  int    `json:"qq"`
		Op  int    `json:"op"`
	} `json:"cqhttp"`
}

func init() {
	if !utils.FileExists("config.json") {
		creatConfig()
	}
	var err error
	Mconfig, err = getMConfig()
	if err != nil {
		Log.Fatalf("读取配置配置文件失败! err: %v", err)
		panic(err)
	}
	Qconfig, err = getQConfig()
	if err != nil {
		Log.Fatalf("读取配置配置文件失败! err: %v", err)
		panic(err)
	}
}

func getMConfig() (mConfig, error) {
	var config mConfig
	f, err := os.OpenFile("config.json", os.O_RDONLY, 0755)
	if err != nil {
		return config, err
	}
	b, _ := ioutil.ReadAll(f)
	err = json.Unmarshal(b, &config)
	if err != nil {
		return config, err
	}
	return config, nil
}

func getQConfig() (qConfig, error) {
	var config qConfig
	f, err := os.OpenFile("config.json", os.O_RDWR, 0755)
	if err != nil {
		return config, err
	}
	b, _ := ioutil.ReadAll(f)
	err = json.Unmarshal(b, &config)
	if err != nil {
		return config, err
	}
	return config, nil
}

func creatConfig() {
	f, _ := os.OpenFile("config.json", os.O_CREATE|os.O_WRONLY, 0755)
	f.WriteString(config)
	f2, _ := os.OpenFile("config.sample.json", os.O_CREATE|os.O_WRONLY, 0755)
	f2.WriteString(confit_sample)
	log.Fatal("已创建配置文件config.json 和 config.sample.json,请根据注释填写配置")
}

const config = `{
	"mcsmdata": [
		{
			"id": 1,
			"url": "https://mcsm.domain.com:443",
			"remote_uuid": "d6a27b0b13ad44ce879b5a56c88b4d34",
			"uuid": "a8788991a64e4a06b76d539b35db1b16",
			"apikey": "vmajkfnvklNSdvkjbnfkdsnv7e0f",
			"group_id": 383033610,
			"user_allows_commands": ["help", "list", "status"],
			"adminlist": [
				1670605849,
				1145141919
			]
		},
		{
			"id": 2,
			"url": "http://mcsm.domain.com:24444",
			"remote_uuid": "3ec8d0ff584c43bd95598b18949a8bac",
			"uuid": "76a49c5ef46a41f29b374109d58f994a",
			"apikey": "vmajkfnvklNSdvkjbnfkdsnv7e0f",
			"group_id": 383033610,
			"user_allows_commands": ["help", "status"],
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
}`

const confit_sample = `{ // 真正的配置文件为标准的json格式，里面不要有注释！！！
	"mcsmdata": [
		{
			"id": 2, // Id 为任意小于256的数，但不可重复！
			"url": "https://mcsm.domain.com:443", // MCSM面板的地址，包含http(s)://，结尾不要有斜杠/
			"remote_uuid": "d6a27b0b13ad44ce879b5a56c88b4d34", // 守护进程的GID
			"uuid": "a8788991a64e4a06b76d539b35db1b16", // 实例的UID
			"apikey": "vmajkfnvklNSdvkjbnfkdsnv7e0f", // 不可为空，用户中心->右上角点蓝色用户名->个人资料->右方生成API密钥
			"group_id": 234532, // 要管理的QQ群号
			"user_allows_commands": ["help", "status"], // 所有群成员均可运行的命令
			"adminlist": [
				1145141919, // 群管理员，第一个为主管理员，只有管理员才可以发送命令
				1433223 // 管理员列表可以为空，则所有用户都可以发送命令
			]
		}, // 只有一个实例可以删掉后面的服务器，有多个则自行添加
		{
			"id": 5, // Id 不可重复！
			"url": "http://mcsm.domain.com:24444",
			"remote_uuid": "d6a27b0b13ad44ce879b5ascwfscr323",
			"uuid": "a8788991a6acasfaca76d539b35db1b16",
			"apikey": "6ewc6292daefvlksmdvjadnvjbf",
			"group_id": 234532,
			"user_allows_commands": ["help", "status"],
			"adminlist": [
				114514, // 不同实例在同一个群也可以有不同的管理员
				1919
			]
		} // <--最后一个实例配置这里没有逗号！！！
	],
	"cqhttp": {
		"url": "ws://127.0.0.1:8080", // cqhttp 请求地址，末尾不带斜杠！只能使用Ws(s)协议
		"qq": 3426898431, // 机器人QQ号
		"op": 1670605849 // op里的qq可以在私聊机器人以访问所有实例，填服务器所有者的QQ号，用于管理所有实例
	}
}`
