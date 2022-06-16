package main

import (
	"fmt"
	"strconv"
	"time"
)

var version = "v1.0.0"
var mconfig MConfig
var qconfig QConfig
var statusmap map[string]int
var listenmap map[int]int

func AddListen() {
	fmt.Print("启动所有服务器?/某个服务器ID (y/id):")
	var chose string
	fmt.Scan(&chose)
	if chose == "y" {
		i := 0
		for i < len(mconfig.McsmData) {
			if listenmap[i] != 1 {
				StartListen(i)
			}
			i++
		}
	} else {
		i, _ := strconv.Atoi(chose)
		if i < len(mconfig.McsmData) && i >= 0 {
			if listenmap[i] != 1 {
				StartListen(i)
			} else {
				fmt.Println("监听实例 ", mconfig.McsmData[i].Name, " 成功！")
				fmt.Println()
			}
		} else {
			fmt.Println("请输入正确的Id!")
			time.Sleep(2 * time.Second)
			fmt.Println()
		}
	}
}

func StartListen(order int) {
	TestMcsmStatus(order)
	if RunningTest(order) {
		go AddQListen(order)
		statusmap[mconfig.McsmData[order].Name] = 1
		time.Sleep(2 * time.Second)
		fmt.Println()
	} else {
		fmt.Print("实例 ", mconfig.McsmData[order].Name, " 未启动，是否启动(y/n):")
		var chose string
		fmt.Scan(&chose)
		if chose == "y" || chose == "yes" {
			Start(order)
			time.Sleep(3 * time.Second)
			StartListen(order)
		} else {
			go AddQListen(order)
			statusmap[mconfig.McsmData[order].Name] = 0
			time.Sleep(2 * time.Second)
			fmt.Println()
		}
	}
}

func Chose() {
	var chose string
	fmt.Println("MCSM-BOT", version)
	fmt.Println("1.添加监听服务器")
	fmt.Println("2.查看已监听列表")
	fmt.Println("3.查看服务器状态")
	fmt.Println("4.动态读取配置信息(用于增加了配置信息后直接添加监听，不用重启软件。减少了配置信息不要动态读取！请重启软件！)")
	fmt.Print("请输入序号:")
	fmt.Scan(&chose)
	switch chose {
	case "1":
		AddListen()
	case "2":
		fmt.Println("[服务器Id: 监听状态(1:监听中)]  没显示的Id均未监听")
		fmt.Println(listenmap)
		time.Sleep(2 * time.Second)
		fmt.Println()
	case "3":
		fmt.Println("[服务器Name: 监听状态(1:运行 2:停止)]")
		fmt.Println(statusmap)
		time.Sleep(2 * time.Second)
		fmt.Println()
	case "4":
		mconfig = GetMConfig()
		fmt.Println("读取配置信息成功！")
		time.Sleep(2 * time.Second)
		fmt.Println()
	default:
		fmt.Println("输入错误,请重新输入...")
		time.Sleep(2 * time.Second)
		fmt.Println()
	}
}

func main() {
	mconfig = GetMConfig()
	qconfig = GetQConfig()
	statusmap = make(map[string]int)
	listenmap = make(map[int]int)
	AddListen()
	for {
		Chose()
	}
}
