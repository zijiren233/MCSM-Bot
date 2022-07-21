package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/zijiren233/MCSM-Bot/gconfig"
	"github.com/zijiren233/MCSM-Bot/logger"
	"github.com/zijiren233/MCSM-Bot/rwmessage"
	"github.com/zijiren233/MCSM-Bot/utils"
)

var version = "v1.5.3-rc1"

var logLevle uint
var disableLogPrint bool

func init() {
	flag.UintVar(&logLevle, "log", 1, "记录命令日志的级别 0:Debug 1:Info 2:Warning 3:Error 4:Fatal 5:None")
	flag.BoolVar(&disableLogPrint, "dlp", false, "Disable Log Print")
}

func main() {
	fmt.Printf("%s[%s] MCSM-BOT Version:%s\n", time.Now().Format("[2006-01-02 15:04:05] "), "INFO", version)
	flag.Parse()
	log := logger.GetLog()
	if disableLogPrint {
		logger.DisableLogPrint()
	}
	log.SetLogLevle(logLevle)
	fmt.Printf("%s[%s] 当前日志级别:%s\n", time.Now().Format("[2006-01-02 15:04:05] "), "DEBUG", logger.IntToLevle(logLevle))
	// 检查配置文件内是否存在重复ID
	if utils.IsListDuplicated(rwmessage.AllId) {
		log.Error("配置文件中存在重复 id")
		return
	}
	serve := rwmessage.NewServer(gconfig.Qconfig.Cqhttp.Url)
	go serve.Run()

	for i := 0; i < len(rwmessage.Mconfig.McsmData); i++ {
		hg := rwmessage.NewHdGroup(rwmessage.Mconfig.McsmData[i].Id, serve.SendMessage)
		if hg == nil {
			continue
		}
		hg.Run()
	}

	p := rwmessage.NewHdOp(serve.SendMessage)
	go p.Run()

	log.Info("MCSM-Bot 启动成功")

	utils.WaitExit()
}
