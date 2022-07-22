package main

import (
	"fmt"
	"time"

	"github.com/zijiren233/MCSM-Bot/bot"
)

var version = "v1.5.4"

func main() {
	fmt.Printf("%s[%s] MCSM-BOT Version:%s\n", time.Now().Format("[2006-01-02 15:04:05] "), "INFO", version)
	bot.Main()
}
