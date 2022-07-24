package base

import (
	"github.com/zijiren233/MCSM-Bot/logger"
	"github.com/zijiren233/MCSM-Bot/utils"
)

func Update(version string) {
	gaj, err := utils.UpdateVersion(version)
	log := logger.GetLog()
	if err != nil {
		log.Warring("获取最新版失败! err: %v", err)
		return
	}
	if gaj.Tag_name != version {
		log.Info("获取到最新版: %s 下载地址: %s", gaj.Tag_name, gaj.Html_url)
		return
	}
	log.Info("当前已是最新版")
}
