package bot

import (
	"fmt"
	"os"
	"time"

	"github.com/zijiren233/MCSM-Bot/base"
	"github.com/zijiren233/MCSM-Bot/gconfig"
	"github.com/zijiren233/MCSM-Bot/logger"
	"github.com/zijiren233/MCSM-Bot/rwmessage"
	"github.com/zijiren233/MCSM-Bot/utils"
)

var version = "v1.6.3"

func Main() {
	fmt.Printf("%s|\033[97;42m %s \033[0m| MCSM-BOT Version:%s\n", time.Now().Format("[2006-01-02 15:04:05] "), "INFO", version)
	base.Parse()
	log := logger.GetLog()
	if base.DisableLogPrint {
		logger.DisableLogPrint()
	}
	log.SetLogLevle(base.LogLevle)
	go base.Update(version)
	fmt.Printf("%s|\033[97;44m %s \033[0m| 当前日志级别:%s\n", time.Now().Format("[2006-01-02 15:04:05] "), "DEBUG", logger.IntToLevle(base.LogLevle))
	// 检查配置文件内是否存在重复ID
	if dou, existdou := utils.IsListDuplicated(rwmessage.AllId); existdou {
		log.Fatal("配置文件中存在重复id: %s", dou)
		fmt.Printf("%s|\033[97;45m %s \033[0m| 配置文件中存在重复id: %s\n", time.Now().Format("[2006-01-02 15:04:05] "), "FATAL", dou)
		os.Exit(-1)
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
}
