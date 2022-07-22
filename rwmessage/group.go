package rwmessage

import (
	"fmt"
	"strconv"
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
		if msg.Group_id == u.Group_id {
			// 当一个群有两个实例监听,且没有指定ID,则由第一个监听的实例执行
			if len(GroupToId[u.Group_id]) >= 2 && msg.Params[1] == "" {
				// 只保留一个goroutine执行
				if GroupToId[u.Group_id][0] != u.Id {
					continue
				}
				if utils.InInt(msg.User_id, u.Adminlist) || utils.InString(msg.Params[2], u.UserCmd) {
					u.checkCMD2(msg)
				}
				log.Warring("权限不足:群组:%d,用户:%d,命令:%#v", msg.Group_id, msg.User_id, msg.Params[0])
				continue
				// 当一个群有两个实例监听,且指定ID,则由指定ID的实例执行
			} else if len(GroupToId[u.Group_id]) >= 2 && msg.Params[1] != "" {
				// 检测id是否监听此群
				id, _ := strconv.Atoi(msg.Params[1])
				if utils.InInt(id, GroupToId[u.Group_id]) {
					// 只保留指定 id goroutine 执行
					if msg.Params[1] != strconv.Itoa(u.Id) {
						continue
					}
				} else {
					// 只保留一个goroutine执行
					if GroupToId[u.Group_id][0] != u.Id {
						continue
					}
					u.Send_group_msg("[CQ:reply,id=%d] ID 不存在", msg.Message_id)
					continue
				}
			}
			if msg.Params[2] == "" {
				u.Send_group_msg("命令为空!\n请输入run help查看帮助!")
				continue
			}
			if utils.InInt(msg.User_id, u.Adminlist) || utils.InString(msg.Params[2], u.UserCmd) {
				go u.checkCMD1(msg)
			} else {
				log.Warring("权限不足:群组:%d,用户:%d,命令:%#v", msg.Group_id, msg.User_id, msg.Params[0])
				u.Send_group_msg("[CQ:reply,id=%d] 权限不足!", msg.Message_id)
			}
		}
	}
}

func (u *HdGroup) checkCMD1(msg *MsgData) {
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
		sendmsg = "run status : 查看服务器状态\nrun start : 启动服务器\nrun stop : 关闭服务器\nrun restart : 重启服务器\nrun kill : 终止服务器\nrun 服务器命令 : 运行服务器命令"
		sendmsg += "\n\n普通用户可用命令:\n"
		for _, v := range u.UserCmd {
			sendmsg += "run " + v + "\n"
		}
		sendmsg += "要在控制台内运行 help 命令请输入 run terminal help"
		sendmsg = *utils.Handle_End_Newline(&sendmsg)
	case "terminal help":
		sendmsg, err = u.RunCmd("help")
	case "server":
		sendmsg += "服务器列表:\n"
		sendmsg = fmt.Sprintf("Name: %s    Id: %d", u.Name, u.Id)
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

// 不指定ID
func (u *HdGroup) checkCMD2(msgdata *MsgData) {
	u.lock.RLock()
	defer u.lock.RUnlock()
	var msg string
	switch msgdata.Params[2] {
	case "help":
		msg = "run server : 查看服务器列表\nrun status : 查看服务器状态\nrun id start : 启动服务器\nrun id stop : 关闭服务器\nrun id restart : 重启服务器\nrun id kill : 终止服务器\nrun id 控制台命令 : 运行服务器命令"
		msg += "\n\n普通用户可用命令:\n请输入 run id help 查询"
		msg = *utils.Handle_End_Newline(&msg)
	case "server":
		msg += "服务器列表:\n"
		for _, v := range GroupToId[u.Group_id] {
			if GOnlineMap[v].Status == 2 || GOnlineMap[v].Status == 3 {
				msg += fmt.Sprintf("Id: %-5dName: %s    Status: 运行中\n", GOnlineMap[v].Id, GOnlineMap[v].Name)
			} else {
				msg += fmt.Sprintf("Id: %-5dName: %s    Status: 已停止\n", GOnlineMap[v].Id, GOnlineMap[v].Name)
			}
		}
		msg = *utils.Handle_End_Newline(&msg)
	default:
		msg += "服务器列表:\n"
		for _, v := range GroupToId[u.Group_id] {
			if GOnlineMap[v].Status == 2 || GOnlineMap[v].Status == 3 {
				msg += fmt.Sprintf("Id: %-5dName: %s    Status: 运行中\n", GOnlineMap[v].Id, GOnlineMap[v].Name)
			} else {
				msg += fmt.Sprintf("Id: %-5dName: %s    Status: 已停止\n", GOnlineMap[v].Id, GOnlineMap[v].Name)
			}
		}
		msg += fmt.Sprintf("查询具体服务器请输入 run id %s", msgdata.Params[2])
	}
	u.Send_group_msg("[CQ:reply,id=%d]%s", msgdata.Message_id, msg)
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
