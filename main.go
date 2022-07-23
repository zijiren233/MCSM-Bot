package main

import (
	"fmt"
	"time"

	"github.com/zijiren233/MCSM-Bot/cmd/bot"
	"github.com/zijiren233/MCSM-Bot/utils"
)

var version = "v1.6.0"

func main() {
	fmt.Printf("%s|\033[97;42m %s \033[0m| MCSM-BOT Version:%s\n", time.Now().Format("[2006-01-02 15:04:05] "), "INFO", version)
	bot.Main()
	utils.WaitExit()
}
