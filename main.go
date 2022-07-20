package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/zijiren233/MCSM-Bot/gconfig"
	"github.com/zijiren233/MCSM-Bot/logger"
	"github.com/zijiren233/MCSM-Bot/rwmessage"
	"github.com/zijiren233/MCSM-Bot/utils"
)

var version = "v1.6.0"

func Chose() {
	var chose string
	fmt.Println("MCSM-BOT", version)
	fmt.Println("1.查看已监听列表")
	fmt.Println("2.查看服务器状态")
	fmt.Print("请输入序号:")
	fmt.Scan(&chose)
	switch chose {
	case "1":
		for _, v := range rwmessage.GOnlineMap {
			fmt.Printf("%s 监听中\n", v.Name)
		}
		fmt.Println()
	case "2":
		for _, v := range rwmessage.GOnlineMap {
			fmt.Printf("%s : %d\n", v.Name, v.Status)
		}
		fmt.Println()
	case "exit":
		os.Exit(0)
	case "stop":
		os.Exit(0)
	default:
		fmt.Println("输入错误,请重新输入...")
		time.Sleep(2 * time.Second)
		fmt.Println()
	}
}

var LogLevle uint

func init() {
	flag.UintVar(&LogLevle, "log", 1, "记录命令日志的级别 0:Debug 1:Info 2:Warning 3:Error 4:None")
}

func main() {
	flag.Parse()
	log := logger.Getlog()
	log.Levle = LogLevle
	// 检查配置文件内是否存在重复ID
	if utils.IsListDuplicated(rwmessage.AllId) {
		log.Error("配置文件中存在重复 id")
		fmt.Printf("配置文件中存在重复 id")
		return
	}
	serve := rwmessage.NewServer(gconfig.Qconfig.Cqhttp.Url)
	go serve.Run()

	for i := 0; i < len(rwmessage.Mconfig.McsmData); i++ {
		hg := rwmessage.NewHdGroup(rwmessage.Mconfig.McsmData[i].Id, serve.SendMessage)
		if hg == nil {
			continue
		}
		go hg.Run()
		time.Sleep(time.Second)
	}
	fmt.Println()

	p := rwmessage.NewHdCqOp(serve.SendMessage)
	go p.Run()
	for {
		Chose()
	}
}
