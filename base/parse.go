package base

import "flag"

var LogLevle uint

func init() {
	flag.UintVar(&LogLevle, "log", 1, "记录命令日志的级别 0:Debug 1:Info 2:Warning 3:Error 4:Fatal 5:None")
}

func Parse() {
	flag.Parse()
}
