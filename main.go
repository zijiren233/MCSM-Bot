package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/zijiren233/MCSM-Bot/logger"
	"github.com/zijiren233/MCSM-Bot/rwmessage"
)

var version = "v1.5.2"
var alone bool

func Chose() {
	var chose string
	fmt.Println("MCSM-BOT", version)
	fmt.Println("1.添加监听服务器")
	fmt.Println("2.查看已监听列表")
	fmt.Println("3.查看服务器状态")
	// fmt.Println("4.重连cqhttp服务")
	fmt.Print("请输入序号:")
	fmt.Scan(&chose)
	switch chose {
	case "1":
		addListen()
	case "2":
		for _, v := range rwmessage.GOnlineMap {
			fmt.Printf("%s 监听中\n", v.Name)
		}
		fmt.Println()
	case "3":
		for _, v := range rwmessage.GOnlineMap {
			fmt.Printf("%s : %d\n", v.Name, v.Status)
		}
		fmt.Println()
	// case "4":
	// 	var err error
	// 	s.Ws, _, err = websocket.DefaultDialer.Dial(s.Url, nil)
	// 	if err != nil {
	// 		fmt.Println("cqhttp重连失败!")
	// 	} else {
	// 		fmt.Println("cqhttp重连成功!")
	// 		fmt.Println()
	// 	}
	default:
		fmt.Println("输入错误,请重新输入...")
		time.Sleep(2 * time.Second)
		fmt.Println()
	}
}

func addListen() {
	fmt.Print("请输入要监听的服务器Id(输入任意字母则监听所有):")
	var id int
	_, err := fmt.Scan(&id)
	if err != nil {
		for i := 0; i < len(rwmessage.Mconfig.McsmData); i++ {
			rwmessage.NewHdGroup(rwmessage.Mconfig.McsmData[i].Id, rwmessage.S.SendMessage)
			fmt.Println()
		}
		return
	}
	rwmessage.NewHdGroup(id, rwmessage.S.SendMessage)
	time.Sleep(1 * time.Second)
	fmt.Println()
}

var LogLevle uint

func init() {
	flag.BoolVar(&alone, "a", false, "运行时自动监听所有服务器 (default false)")
	flag.UintVar(&LogLevle, "log", 1, "记录命令日志的级别 0:Debug 1:Info 2:Warning 3:Error 4:None")
}

func main() {
	flag.Parse()
	logger.Log = logger.Newlog(LogLevle)
	if rwmessage.IsListDuplicated(rwmessage.GetAllId()) {
		panic("有重复ID!")
	}
	if alone {
		addListen()
	} else {
		for i := 0; i < len(rwmessage.Mconfig.McsmData); i++ {
			rwmessage.NewHdGroup(rwmessage.Mconfig.McsmData[i].Id, rwmessage.S.SendMessage)
		}
		fmt.Println()
	}
	p := rwmessage.NewHdCqOp(rwmessage.S.SendMessage)
	go p.HdCqOp()
	for {
		Chose()
	}
}
