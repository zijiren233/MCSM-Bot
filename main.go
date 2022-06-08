package main

import (
	"fmt"
	"time"
)

var mconfig MConfig
var qconfig QConfig
var mstatus map[string]int

func AddDaemon() {
	fmt.Print("请输入服务器Order(-1表示启动所有):")
	var order int
	var chose string
	i := 0
	fmt.Scan(&order)
	if order == -1 {
		for i < len(mconfig.McsmData) {
			if RunningTest(i) {
				go ProcessMessag(i)
				time.Sleep(1 * time.Second)
				mstatus[mconfig.McsmData[i].Name] = 1
				fmt.Print("监听实例 ", mconfig.McsmData[i].Name, " 成功！")
				time.Sleep(1 * time.Second)
				fmt.Println()
			} else {
				fmt.Println("实例 ", mconfig.McsmData[i].Name, " 未启动，是否启动(yes/no):")
				fmt.Scan(&chose)
				if chose == "yes" {
					Start(i)
					time.Sleep(1 * time.Second)
					if RunningTest(i) {
						mstatus[mconfig.McsmData[i].Name] = 1
						fmt.Println("实例 ", mconfig.McsmData[i].Name, " 启动成功！")
						go ProcessMessag(i)
						time.Sleep(1 * time.Second)
						fmt.Println()
					}
				}
			}
			i++
		}
	} else {
		if RunningTest(order) {
			go ProcessMessag(order)
			mstatus[mconfig.McsmData[order].Name] = 1
			fmt.Println("监听实例 ", mconfig.McsmData[order].Name, " 成功！")
			time.Sleep(1 * time.Second)
			fmt.Println()
		} else {
			fmt.Println("实例 ", mconfig.McsmData[order].Name, " 未启动，是否启动(yes/no):")
			fmt.Scan(&chose)
			if chose == "yes" {
				Start(order)
				time.Sleep(1 * time.Second)
				if RunningTest(order) {
					mstatus[mconfig.McsmData[order].Name] = 1
					fmt.Println("实例 ", mconfig.McsmData[order].Name, " 启动成功！")
					go ProcessMessag(i)
					time.Sleep(1 * time.Second)
					fmt.Println()
				}
			} else {
				return
			}
		}
	}
	if i >= len(mconfig.McsmData) {
		fmt.Println("未找到此Order序号,Order序号应从0开始依次增加！")
		time.Sleep(3 * time.Second)
	}
}

func main() {
	mconfig = GetMConfig()
	qconfig = GetQConfig()
	mstatus = make(map[string]int)
	for {
		var what string
		fmt.Println("MCSM-BOT V0.1")
		fmt.Println("1.添加监听服务器")
		fmt.Println("2.查看已启动服务器")
		fmt.Print("请输入序号:")
		fmt.Scan(&what)
		switch what {
		case "1":
			AddDaemon()
		case "2":
			fmt.Printf("%v\n", mstatus)
			time.Sleep(3 * time.Second)
		default:
			fmt.Println("输入错误,请重新输入...")
			time.Sleep(1 * time.Second)
		}
	}
}
