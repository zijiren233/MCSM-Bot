package base

import (
	"time"

	"github.com/zijiren233/MCSM-Bot/logger"
	"github.com/zijiren233/MCSM-Bot/utils"
)

func Update(version string) {
	log := logger.GetLog()
	for {
		gaj, err := utils.UpdateVersion(version)
		if err != nil {
			log.Warring("获取最新版失败! err: %v", err)
			time.Sleep(time.Hour * 12)
			continue
		}
		if gaj.Tag_name != version {
			log.Info("当前版本: %s 获取到最新版: %s 下载地址: %s", version, gaj.Tag_name, gaj.Html_url)
			time.Sleep(time.Hour * 12)
			continue
		}
		time.Sleep(time.Hour * 12)
	}
}
