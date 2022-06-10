package main

import (
	"fmt"
	"strconv"
	"time"
)

var mconfig MConfig
var qconfig QConfig
var statusmap map[string]int
var listenmap map[int]int

func AddListen() {
	fmt.Print("请输入服务器Order(-1表示启动所有):")
	var order int
	i := 0
	fmt.Scan(&order)
	if i >= len(mconfig.McsmData) {
		fmt.Println("未找到此Order序号,Order序号应从0开始依次增加！")
		time.Sleep(2 * time.Second)
	} else if order == -1 {
		for i < len(mconfig.McsmData) {
			StartListen(i)
			i++
		}
	} else {
		StartListen(order)
	}
}

func StartListen(order int) {
	TestMcsmStatus(order)
	if RunningTest(order) {
		go AddQListen(order)
		statusmap[mconfig.McsmData[order].Name] = 1
		time.Sleep(3 * time.Second)
		fmt.Println()
	} else {
		fmt.Print("实例 ", mconfig.McsmData[order].Name, " 未启动，是否启动(y/n):")
		var chose string
		fmt.Scan(&chose)
		if chose == "y" || chose == "yes" {
			Start(order)
			time.Sleep(3 * time.Second)
			StartListen(order)
		}
	}
}

func Chose() {
	var chose string
	fmt.Println("MCSM-BOT V0.1")
	fmt.Println("1.添加监听服务器")
	fmt.Println("2.查看已监听列表")
	fmt.Println("3.查看服务器状态")
	fmt.Print("请输入序号:")
	fmt.Scan(&chose)
	switch chose {
	case "1":
		AddListen()
	case "2":
		fmt.Println("Order: status")
		fmt.Println(listenmap)
		time.Sleep(2 * time.Second)
		fmt.Printf("\n")
	case "3":
		fmt.Println(statusmap)
		time.Sleep(2 * time.Second)
		fmt.Printf("\n")
	default:
		fmt.Println("输入错误,请重新输入...")
		time.Sleep(2 * time.Second)
	}
}

func main() {
	mconfig = GetMConfig()
	qconfig = GetQConfig()
	statusmap = make(map[string]int)
	listenmap = make(map[int]int)
	fmt.Print("启动所有服务器?/启动Order (y/order):")
	var chose string
	fmt.Scan(&chose)
	if chose == "y" || chose == "yes" {
		i := 0
		for i < len(mconfig.McsmData) {
			StartListen(i)
			i++
		}
	} else {
		i, _ := strconv.Atoi(chose)
		if i < len(mconfig.McsmData) && i >= 0 {
			StartListen(i)
		} else {
			fmt.Println("请输入正确的Order!")
			time.Sleep(2 * time.Second)
			fmt.Println()
		}
	}
	for {
		Chose()
	}
}
