package rwmessage

import (
	"fmt"
	"sync"
	"time"

	"github.com/zijiren233/MCSM-Bot/utils"
)

type HdGroup struct {
	config
	instanceConfig
	ChGroupMsg chan *MsgData
	SendChan   chan *SendData

	lock sync.RWMutex
}

type config struct {
	Id          int
	Url         string
	Remote_uuid string
	Uuid        string
	Apikey      string
	Group_id    int
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
			Group_id:    Mconfig.McsmData[IdToOd[id]].Group_id,
			UserCmd:     Mconfig.McsmData[IdToOd[id]].User_allows_commands,
			Adminlist:   Mconfig.McsmData[IdToOd[id]].Adminlist},
		SendChan: serveSend,
	}
	err := u.getStatusInfo()
	if err != nil {
		log.Fatal("服务器Id: %d 监听失败!可能是 mcsm-web 端地址错误\n", u.Id)
		return nil
	}
	log.Debug("ID: %d ,NAME: %s ,TYPE:%s ,PTY: %v", u.Id, u.Name, u.ProcessType, u.Pty)
	if u.ProcessType != "docker" && !u.Pty {
		log.Error("实例:%s 未开启 仿真终端 或 未使用 docker 启动！", u.Name)
		log.Fatal("实例:%s 监听失败", u.Name)
		return nil
	}
	GroupToId[u.Group_id] = append(GroupToId[u.Group_id], u.Id)
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
		if u.Group_id == msg.Group_id {
			if utils.InInt(msg.User_id, u.Adminlist) || utils.InString(msg.Params[2], u.UserCmd) {
				go u.runCMD(msg)
			} else {
				log.Warring("权限不足:群组: %d,用户: %d,命令: %#v", msg.Group_id, msg.User_id, msg.Params[0])
				u.Send_group_msg("[CQ:reply,id=%d]权限不足!", msg.Message_id)
			}
		}

	}
}

func (u *HdGroup) runCMD(msg *MsgData) {
	var sendmsg string
	var err error
	u.lock.RLock()
	defer u.lock.RUnlock()
	if u.Status != 3 && u.Status != 2 && (msg.Params[2] != "help" && msg.Params[2] != "server" && msg.Params[2] != "start") {
		u.Send_group_msg("服务器: %s 未启动!\n请先启动服务器:\nrun %d start", u.Name, u.Id)
		return
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
		sendmsg += "要在控制台内运行 help 命令请输入 run id terminal help"
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
		return
	}
	u.Send_group_msg("[CQ:reply,id=%d]%s", msg.Message_id, sendmsg)
}

func (u *HdGroup) reportStatus() {
	go func() {
		for {
			u.getStatusInfo()
			time.Sleep(3000 * time.Millisecond)
		}
	}()
	var status = u.Status
	for {
		u.lock.RLock()
		if status != u.Status {
			if (u.Status == 2 && status != 3) || (u.Status == 3 && status != 2) {
				u.Send_group_msg("服务器ID: %-4d NAME: %s 已运行!", u.Id, u.Name)
			} else if u.Status == 0 {
				u.Send_group_msg("服务器ID: %-4d NAME: %s 已停止!", u.Id, u.Name)
			}
			status = u.Status
		}
		u.lock.RUnlock()
		time.Sleep(2000 * time.Millisecond)
	}
}

func (u *HdGroup) Send_group_msg(msg string, a ...interface{}) {
	var tmp SendData
	tmp.Action = "send_group_msg"
	tmp.Params.Group_id = u.Group_id
	tmp.Params.Message = fmt.Sprintf(msg, a...)
	u.SendChan <- &tmp
}
