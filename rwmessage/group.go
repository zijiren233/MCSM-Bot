package rwmessage

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/zijiren233/MCSM-Bot/utils"
)

type HdGroup struct {
	config
	instanceConfig
	ChGroupMsg  chan *MsgData
	SendChan    chan *SendData
	performance int64

	lock sync.RWMutex
}

type config struct {
	Id          int
	Url         string
	Remote_uuid string
	Uuid        string
	Apikey      string
	Group_list  []int
	UserCmd     []string
	Adminlist   []int
}

type instanceConfig struct {
	Name           string
	Status         int
	EndTime        string
	ProcessType    string
	Pty            bool
	PingIp         string
	CurrentPlayers string
	MaxPlayers     string
	Version        string
}

type InstanceConfig struct {
	Data struct {
		Status int `json:"status"`
		Config struct {
			Nickname       string `json:"nickname"`
			EndTime        string `json:"endTime"`
			ProcessType    string `json:"processType"`
			TerminalOption struct {
				Pty bool `json:"pty"`
			} `json:"terminalOption"`
			PingConfig struct {
				PingIp string `json:"ip"`
			} `json:"pingConfig"`
		} `json:"config"`
		Info struct {
			CurrentPlayers string `json:"currentPlayers"`
			MaxPlayers     string `json:"maxPlayers"`
			Version        string `json:"version"`
		} `json:"info"`
	} `json:"data"`
}

func NewHdGroup(id int, serveSend chan *SendData) *HdGroup {
	u := HdGroup{
		config: config{
			Id:          id,
			Url:         Mconfig.McsmData[IdToOd[id]].Url,
			Remote_uuid: Mconfig.McsmData[IdToOd[id]].Remote_uuid,
			Uuid:        Mconfig.McsmData[IdToOd[id]].Uuid,
			Apikey:      Mconfig.McsmData[IdToOd[id]].Apikey,
			Group_list:  Mconfig.McsmData[IdToOd[id]].Group_list,
			UserCmd:     Mconfig.McsmData[IdToOd[id]].User_allows_commands,
			Adminlist:   Mconfig.McsmData[IdToOd[id]].Adminlist},
		SendChan: serveSend,
	}
	err := u.getStatusInfo()
	if err != nil {
		log.Fatal("服务器Id: %d 监听失败!可能是 mcsm-web 端地址错误\n", id)
		return nil
	}
	log.Debug("ID: %d ,NAME: %s ,TYPE:%s ,PTY: %v", id, u.Name, u.ProcessType, u.Pty)
	if u.ProcessType != "docker" && !u.Pty {
		log.Error("实例:%s 未开启 仿真终端 或 未使用 docker 启动！", u.Name)
		log.Fatal("Id: %d, 实例:%s 监听失败", id, u.Name)
		return nil
	}
	for _, v := range u.Group_list {
		if !utils.InInt(id, GroupToId[v]) {
			GroupToId[v] = append(GroupToId[v], id)
		}
	}
	log.Debug("GroupToId: %v", GroupToId)
	u.ChGroupMsg = make(chan *MsgData, 25)
	return &u
}

func (u *HdGroup) Run() {
	GOnlineMap[u.Id] = u
	log.Info("监听实例 %s 成功", u.Name)
	if u.PingIp == "" {
		log.Warring("ID: %d ,NAME: %s 未开启 状态查询,请开启 状态查询 以获得完整体验", u.Id, u.Name)
	}
	go u.reportStatus()
	go u.hdChMessage()
}

func (u *HdGroup) hdChMessage() {
	var msg *MsgData
	for {
		msg = <-u.ChGroupMsg
		if utils.InInt(msg.Group_id, u.Group_list) {
			u.lock.RLock()
			if u.isAdmin(msg.User_id) || utils.InString(strings.Split(msg.Params[2], " ")[0], u.UserCmd) {
				go func(msg *MsgData) {
					u.Send_group_msg(msg.Group_id, u.runCMD(msg))
				}(msg)
			} else {
				log.Warring("权限不足:群组: %d,用户: %d,命令: %#v, 实例: %s", msg.Group_id, msg.User_id, msg.Params[0], u.Name)
				u.Send_group_msg(msg.Group_id, "[CQ:reply,id=%d]权限不足!", msg.Message_id)
			}
			u.lock.RUnlock()
		}
	}
}

func (u *HdGroup) isAdmin(user_id int) bool {
	return utils.InInt(user_id, u.Adminlist)
}

func (u *HdGroup) runCMD(msg *MsgData) string {
	var sendmsg string
	var err error
	u.lock.RLock()
	defer u.lock.RUnlock()
	if u.Status != 3 && u.Status != 2 && (msg.Params[2] != "help" && msg.Params[2] != "server" && msg.Params[2] != "start") {
		return fmt.Sprintf("服务器: %s 未启动!\n请先启动服务器:\nrun %d start", u.Name, u.Id)
	}
	switch msg.Params[2] {
	case "help":
		sendmsg = fmt.Sprintf("run %d status : 查看服务器状态\nrun %d start : 启动服务器\nrun %d stop : 关闭服务器\nrun %d restart : 重启服务器\nrun %d kill : 终止服务器\nrun %d 服务器命令 : 运行服务器命令\n\n普通用户可用命令:\n",
			u.Id,
			u.Id,
			u.Id,
			u.Id,
			u.Id,
			u.Id,
		)
		for _, v := range u.UserCmd {
			sendmsg += fmt.Sprintf("run %d %s\n", u.Id, v)
		}
		sendmsg += fmt.Sprintf("要在控制台内运行 help 命令请输入 run %d terminal help", u.Id)
	case "terminal help":
		sendmsg, err = u.RunCmd("help")
	case "status":
		sendmsg = u.GetStatus()
	case "start":
		sendmsg, err = u.Start()
	case "stop":
		sendmsg, err = u.Stop()
	case "restart":
		sendmsg, err = u.Restart()
	case "kill":
		sendmsg, err = u.Kill()
	default:
		sendmsg, err = u.RunCmd(msg.Params[2])
	}
	if err != nil {
		return ""
	}
	return fmt.Sprintf("[CQ:reply,id=%d]%s", msg.Message_id, sendmsg)
}

func (u *HdGroup) reportStatus() {
	go func() {
		var performance int64
		for {
			u.getStatusInfo()
			u.lock.RLock()
			performance = u.performance
			u.lock.RUnlock()
			if performance <= 200 {
				time.Sleep(time.Second)
			} else {
				time.Sleep(3 * time.Second)
			}
		}
	}()
	var status = u.Status
	var performance int64
	for {
		u.lock.RLock()
		if status != u.Status {
			if (u.Status == 2 && status != 3) || (u.Status == 3 && status != 2) {
				u.Send_all_group_msg("服务器 %s 已运行!\nID: %d", u.Name, u.Id)
			} else if (u.Status == 0 && status != 1) || (u.Status == 1 && status != 0) {
				u.Send_all_group_msg("服务器 %s 已停止!\nID: %d", u.Name, u.Id)
			}
			status = u.Status
		}
		performance = u.performance
		u.lock.RUnlock()
		if performance <= 200 {
			time.Sleep(time.Second)
		} else {
			time.Sleep(3 * time.Second)
		}
	}
}

func (u *HdGroup) Send_group_msg(group int, msg string, a ...interface{}) {
	var tmp SendData
	tmp.Action = "send_group_msg"
	tmp.Params.Group_id = group
	tmp.Params.Message = fmt.Sprintf(msg, a...)
	u.SendChan <- &tmp
}

func (u *HdGroup) Send_all_group_msg(msg string, a ...interface{}) {
	for _, v := range u.Group_list {
		var tmp SendData
		tmp.Action = "send_group_msg"
		tmp.Params.Group_id = v
		tmp.Params.Message = fmt.Sprintf(msg, a...)
		u.SendChan <- &tmp
	}
}
