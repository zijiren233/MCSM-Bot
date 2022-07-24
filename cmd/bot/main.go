package bot

import (
	"fmt"
	"time"

	"github.com/zijiren233/MCSM-Bot/base"
	"github.com/zijiren233/MCSM-Bot/gconfig"
	"github.com/zijiren233/MCSM-Bot/logger"
	"github.com/zijiren233/MCSM-Bot/rwmessage"
	"github.com/zijiren233/MCSM-Bot/utils"
)

var version = "v1.6.0"

func Main() {
	fmt.Printf("%s|\033[97;42m %s \033[0m| MCSM-BOT Version:%s\n", time.Now().Format("[2006-01-02 15:04:05] "), "INFO", version)
	base.Parse()
	log := logger.GetLog()
	if base.DisableLogPrint {
		logger.DisableLogPrint()
	}
	log.SetLogLevle(base.LogLevle)
	fmt.Printf("%s|\033[97;44m %s \033[0m| 当前日志级别:%s\n", time.Now().Format("[2006-01-02 15:04:05] "), "DEBUG", logger.IntToLevle(base.LogLevle))
	base.Update(version)
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
}
