package main

import (
	"flag"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/zijiren233/MCSM-Bot/logger"
)

var version = "v1.2.0"
var mconfig MConfig
var qconfig QConfig
var statusmap = sync.Map{}
var listenmap = sync.Map{}
var log *logger.Logger

// var errfile *os.File

func AddListen() {
	if len(mconfig.McsmData) == 1 {
		StartListen(0)
		return
	}
	fmt.Print("启动所有服务器?/某个服务器 ID ? (y/id):")
	var chose string
	fmt.Scan(&chose)
	if chose == "y" {
		for i := 0; i < len(mconfig.McsmData); i++ {
			tmp, _ := listenmap.Load(i)
			if tmp != 1 {
				StartListen(i)
			}
		}
	} else {
		i, err := strconv.Atoi(chose)
		if err != nil {
			fmt.Println("输入不正确!")
			return
		}
		if i < len(mconfig.McsmData) && i >= 0 {
			tmp, _ := listenmap.Load(i)
			if tmp != 1 {
				StartListen(i)
			} else {
				fmt.Println("监听实例 ", mconfig.McsmData[i].Name, " 成功！")
				log.Info("监听实例 ", mconfig.McsmData[i].Name, " 成功！")
				fmt.Println()
			}
		} else {
			fmt.Println("请输入正确的Id! (或配置文件中id错误,id必须从0依次增加!)")
			time.Sleep(2 * time.Second)
			fmt.Println()
		}
	}
}

func StartListen(order int) {
	if !TestMcsmStatus(order) {
		return
	}
	if RunningTest(order) {
		go AddQListen(order)
		statusmap.Store(mconfig.McsmData[order].Name, 1)
		time.Sleep(1 * time.Second)
		fmt.Println()
	} else {
		go AddQListen(order)
		statusmap.Store(mconfig.McsmData[order].Name, 0)
		time.Sleep(1 * time.Second)
		fmt.Println()
	}
}

func Chose() {
	var chose string
	fmt.Println("MCSM-BOT", version)
	fmt.Println("1.添加监听服务器")
	fmt.Println("2.查看已监听列表")
	fmt.Println("3.查看服务器状态")
	fmt.Print("请输入序号:")
	fmt.Scan(&chose)
	switch chose {
	case "1":
		AddListen()
	case "2":
		fmt.Println()
		fmt.Println("[没显示的均未监听]")
		listenmap.Range(func(key, value interface{}) bool {
			fmt.Println(mconfig.McsmData[key.(int)].Name)
			return true
		})
		time.Sleep(2 * time.Second)
		fmt.Println()
	case "3":
		fmt.Println()
		fmt.Println("[服务器Name: 监听状态(0:停止 1:运行)]")
		statusmap.Range(func(key, value interface{}) bool {
			fmt.Printf("%s: %d\n", key, value)
			return true
		})
		time.Sleep(2 * time.Second)
		fmt.Println()
	default:
		fmt.Println("输入错误,请重新输入...")
		time.Sleep(2 * time.Second)
		fmt.Println()
	}
}

var all bool
var loglevle uint

func init() {
	flag.BoolVar(&all, "a", false, "运行时自动监听所有服务器 (default false)")
	flag.UintVar(&loglevle, "log", 1, "记录命令日志的级别 0:Debug 1:Info 2:Warning 3:Error 4:None")
}

func main() {
	flag.Parse()
	log = logger.Newlog(loglevle)
	mconfig = GetMConfig()
	qconfig = GetQConfig()
	if !all {
		AddListen()
	} else {
		for i := 0; i < len(mconfig.McsmData); i++ {
			StartListen(i)
		}
	}
	for {
		Chose()
	}
}
