package main

import (
	"fmt"
	"strconv"
	"time"
)

var mconfig MConfig
var qconfig QConfig
var statusmap map[string]int
var listenmap map[string]int

func AddListen() {
	fmt.Print("请输入服务器Order(-1表示启动所有):")
	var order int
	i := 0
	fmt.Scan(&order)
	if i >= len(mconfig.McsmData) {
		fmt.Println("未找到此Order序号,Order序号应从0开始依次增加！")
		time.Sleep(3 * time.Second)
	} else if order == -1 {
		for i < len(mconfig.McsmData) {
			startlisten(i)
			i++
		}
	} else {
		startlisten(order)
	}
}

func startlisten(order int) {
	if RunningTest(order) {
		go AddQListen(order)
		fmt.Println("监听实例 ", mconfig.McsmData[order].Name, " 成功！")
		statusmap[mconfig.McsmData[order].Name] = 1
		listenmap[mconfig.McsmData[order].Name] = 1
		time.Sleep(1 * time.Second)
		fmt.Println()
	} else {
		fmt.Println("实例 ", mconfig.McsmData[order].Name, " 未启动，是否启动(yes/no):")
		var chose string
		fmt.Scan(&chose)
		if chose == "yes" {
			Start(order)
			time.Sleep(2 * time.Second)
			startlisten(order)
		}
	}
}

func Chose() {
	var what string
	fmt.Println("MCSM-BOT V0.1")
	fmt.Println("1.添加监听服务器")
	fmt.Println("2.查看已监听列表")
	fmt.Println("3.查看服务器状态")
	fmt.Print("请输入序号:")
	fmt.Scan(&what)
	switch what {
	case "1":
		AddListen()
	case "2":
		fmt.Println(listenmap)
		time.Sleep(3 * time.Second)
		fmt.Printf("\n")
	case "3":
		fmt.Println(statusmap)
		time.Sleep(3 * time.Second)
		fmt.Printf("\n")
	default:
		fmt.Println("输入错误,请重新输入...")
		time.Sleep(1 * time.Second)
	}
}

func main() {
	mconfig = GetMConfig()
	qconfig = GetQConfig()
	statusmap = make(map[string]int)
	listenmap = make(map[string]int)
	fmt.Print("启动所有服务器?/启动Order (yes/order):")
	var chose string
	fmt.Scan(&chose)
	if chose == "yes" {
		i := 0
		for i < len(mconfig.McsmData) {
			startlisten(i)
			time.Sleep(3 * time.Second)
			i++
		}
	} else {
		i, _ := strconv.Atoi(chose)
		if i < len(mconfig.McsmData) && i >= 0 {
			startlisten(i)
			time.Sleep(3 * time.Second)
		} else {
			fmt.Println("请输入正确的Order!")
			fmt.Println()
			time.Sleep(2 * time.Second)
		}
	}
	for {
		Chose()
	}
}
