package base

import (
	"strconv"
	"strings"
	"time"

	"github.com/zijiren233/MCSM-Bot/utils"
	"github.com/zijiren233/go-colorlog"
)

func Update(version string) {
	for {
		gaj, err := utils.UpdateVersion(version)
		if err != nil {
			colorlog.Warringf("获取最新版失败! err: %v", err)
		} else if chackVersion(version[1:], gaj.Tag_name[1:]) {
			colorlog.Infof("当前版本: %s 获取到最新版: %s 下载地址: %s", version, gaj.Tag_name, gaj.Html_url)
		}
		time.Sleep(time.Hour * 6)
	}
}

func chackVersion(version, gitTag string) bool {
	var shuldUpdate = false
	ver := stringListToIntList(strings.Split(version, "."))
	tag := stringListToIntList(strings.Split(gitTag, "."))
	if len(ver) == len(tag) {
		for k, v := range tag {
			if v > ver[k] {
				shuldUpdate = true
			}
		}
	} else if len(ver) > len(tag) {
		for k, v := range tag {
			if v > ver[k] {
				shuldUpdate = true
			}
		}
	} else {
		for k, v := range ver {
			if v < tag[k] {
				shuldUpdate = true
			}
		}
		if !shuldUpdate && gitTag[:len(version)] == version {
			shuldUpdate = true
		}
	}
	return shuldUpdate
}

func stringListToIntList(data []string) []int {
	var ret []int
	for _, v := range data {
		i, _ := strconv.Atoi(v)
		ret = append(ret, i)
	}
	return ret
}
