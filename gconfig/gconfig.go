package gconfig

import (
	"io"
	"log"
	"os"

	"github.com/zijiren233/MCSM-Bot/utils"
	"gopkg.in/yaml.v3"
)

var Mconfig mConfig
var Log = log.Default()

type mConfig struct {
	McsmData []struct {
		Id                   int      `yaml:"id"`
		Url                  string   `yaml:"url"`
		Remote_uuid          string   `yaml:"remote_uuid"`
		Uuid                 string   `yaml:"uuid"`
		Apikey               string   `yaml:"apikey"`
		Group_list           []int    `yaml:"group_id"`
		User_allows_commands []string `yaml:"user_allows_commands"`
		Adminlist            []int    `yaml:"adminlist"`
	} `yaml:"mcsmdata"`
	Cqhttp struct {
		Url       string `yaml:"url"`
		AdminList []int  `yaml:"adminlist"`
	} `yaml:"cqhttp"`
}

func init() {
	if !utils.FileExists("config.yaml") {
		creatConfig()
	}
	var err error
	Mconfig, err = getMConfig()
	if err != nil {
		Log.Fatalf("读取配置文件失败!请按照config.example.yaml: %v", err)
		os.Exit(-1)
	}
}

func getMConfig() (mConfig, error) {
	var config mConfig
	f, err := os.OpenFile("config.yaml", os.O_RDONLY, 0755)
	if err != nil {
		return config, err
	}
	defer f.Close()
	b, _ := io.ReadAll(f)
	err = yaml.Unmarshal(b, &config)
	if err != nil {
		reCreatexampleConfig()
		return config, err
	}
	return config, nil
}

func creatConfig() {
	f, _ := os.OpenFile("config.yaml", os.O_CREATE|os.O_WRONLY, 0755)
	f.WriteString(config)
	f2, _ := os.OpenFile("config.example.yaml", os.O_CREATE|os.O_WRONLY, 0755)
	f2.WriteString(confit_example)
	log.Fatal("已创建配置文件config.yaml 和 config.example.yaml,请根据注释填写配置")
}

func reCreatexampleConfig() {
	f, err := os.OpenFile("config.example.yaml", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return
	}
	defer f.Close()
	f.WriteString(confit_example)
}

func GetAllDaemon() *map[string]string {
	var tmplist = make(map[string]string)
	for i := 0; i < len(Mconfig.McsmData); i++ {
		tmplist[Mconfig.McsmData[i].Url] = Mconfig.McsmData[i].Apikey
	}
	return &tmplist
}

func GetAllId() []int {
	tmp := make([]int, 0, len(Mconfig.McsmData))
	for i := 0; i < len(Mconfig.McsmData); i++ {
		tmp = append(tmp, Mconfig.McsmData[i].Id)
	}
	return tmp
}

const config = `mcsmdata:
  - id: 1 # Id 为任意小于256的数,但不可重复!
    url: https://mcsm.domain.com:443 # MCSM面板的地址,包含http(s)://,结尾不要有斜杠/
    remote_uuid: d6a27b0b13ad44ce879b5a56c88b4d34 # 守护进程的GID
    uuid: a8788991a64e4a06b76d539b35db1b16 # 实例的UID
    apikey: vmajkfnvklNSdvkjbnfkdsnv7e0f # 不可为空，用户中心->右上角点蓝色用户名->个人资料->右方生成API密钥
    group_id: # 要管理的QQ群号
      - 383033610
      - 1145141919
    user_allows_commands: # 所有群成员均可运行的命令,填正则表达式
      - ^help$
      - ^list$
      - ^status$
    adminlist: # 群管理员，第一个为主管理员，只有管理员才可以发送命令,管理员列表可以为空，则所有用户都可以发送命令
      - 1670605849
      - 1145141919
  - id: 2
    url: http://mcsm.domain.com:24444
    remote_uuid: 3ec8d0ff584c43bd95598b18949a8bac
    uuid: 76a49c5ef46a41f29b374109d58f994a
    apikey: vmajkfnvklNSdvkjbnfkdsnv7e0f
    group_id:
      - 383033610
    user_allows_commands:
      - ^help$
      - ^status$
    adminlist:
      - 1670605849
      - 1145141919

cqhttp:
url: ws://q-api.pyhdxy.com:8080 # cqhttp 请求地址,末尾不带斜杠!只能使用Ws(s)协议
adminlist: # 可以私聊机器人以访问所有实例,填服务器所有者的QQ号,用于管理所有实例
  - 1670605849
  - 1145141919
`

const confit_example = `mcsmdata:
  - id: 1 # Id 为任意小于256的数,但不可重复!
    url: https://mcsm.domain.com:443 # MCSM面板的地址,包含http(s)://,结尾不要有斜杠/
    remote_uuid: d6a27b0b13ad44ce879b5a56c88b4d34 # 守护进程的GID
    uuid: a8788991a64e4a06b76d539b35db1b16 # 实例的UID
    apikey: vmajkfnvklNSdvkjbnfkdsnv7e0f # 不可为空，用户中心->右上角点蓝色用户名->个人资料->右方生成API密钥
    group_id: # 要管理的QQ群号
      - 383033610
      - 1145141919
    user_allows_commands: # 所有群成员均可运行的命令,填正则表达式
      - ^help$
      - ^list$
      - ^status$
    adminlist: # 群管理员，第一个为主管理员，只有管理员才可以发送命令,管理员列表可以为空，则所有用户都可以发送命令
      - 1670605849
      - 1145141919
  - id: 2
    url: http://mcsm.domain.com:24444
    remote_uuid: 3ec8d0ff584c43bd95598b18949a8bac
    uuid: 76a49c5ef46a41f29b374109d58f994a
    apikey: vmajkfnvklNSdvkjbnfkdsnv7e0f
    group_id:
      - 383033610
    user_allows_commands:
      - ^help$
      - ^status$
    adminlist:
      - 1670605849
      - 1145141919

cqhttp:
url: ws://q-api.pyhdxy.com:8080 # cqhttp 请求地址,末尾不带斜杠!只能使用Ws(s)协议
adminlist: # 可以私聊机器人以访问所有实例,填服务器所有者的QQ号,用于管理所有实例
  - 1670605849
  - 1145141919
`
