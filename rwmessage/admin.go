package rwmessage

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/zijiren233/MCSM-Bot/gconfig"
	"github.com/zijiren233/MCSM-Bot/utils"
)

type admin struct {
	adminList []int
	ChCqOpMsg chan *MsgData
	SendChan  chan *SendData
}

type RemoteStatus struct {
	Data struct {
		RemoteCount struct {
			Total     int `json:"total"`
			Available int `json:"available"`
		} `json:"remoteCount"`
		Remote []struct {
			Instance struct {
				Total   int `json:"total"`
				Running int `json:"running"`
			} `json:"instance"`
			System struct {
				Platform string  `json:"platform"`
				CpuUsage float64 `json:"cpuUsage"`
				MemUsage float64 `json:"memUsage"`
			} `json:"system"`
			Ip        string `json:"ip"`
			Port      string `json:"port"`
			Available bool   `json:"available"`
			Remarks   string `json:"remarks"`
		} `json:"remote"`
	} `json:"data"`
}

func NewHdOp(send chan *SendData) *admin {
	PAdmin := admin{
		adminList: Qconfig.Cqhttp.AdminList,
		ChCqOpMsg: make(chan *MsgData, 25),
		SendChan:  send,
	}
	return &PAdmin
}

func (p *admin) Run() {
	var msg *MsgData
	for {
		msg = <-p.ChCqOpMsg
		if msg.Params[2] == "" {
			msg.Params[2] = "help"
			p.Send_private_msg(msg.User_id, p.help(msg))
			continue
		} else if msg.Params[1] == "" {
			p.Send_private_msg(msg.User_id, p.help(msg))
			continue
		}
		go p.handleMessage(msg)

	}
}

func (p *admin) handleMessage(msg *MsgData) {
	id, err := strconv.Atoi(msg.Params[1])
	if err != nil {
		log.Error("strconv.Atoi error:%v", err)
		p.Send_private_msg(msg.User_id, "命令格式错误!\n请输入run help查看帮助!")
		return
	}
	if utils.InInt(id, AllId) {
		p.Send_private_msg(msg.User_id, GOnlineMap[id].runCMD(msg))
	} else {
		p.Send_private_msg(msg.User_id, "请输入正确的ID!")
		log.Warring("OP 输入: %d 请输入正确的ID!", msg.User_id, msg.Params[1])
	}
}

func (p *admin) help(msgdata *MsgData) string {
	var msg string
	switch msgdata.Params[2] {
	case "help":
		return "run list : 查看服务器列表\nrun stopped list : 查看未运行的服务器列表\nrun status : 查看服务器状态\nrun daemon status : 查看MCSM后端状态\nrun id start : 启动服务器\nrun id stop : 关闭服务器\nrun id restart : 重启服务器\nrun id kill : 终止服务器\nrun id 服务器命令 : 运行服务器命令"
	case "list":
		msg += "服务器列表:\n"
		for _, v := range AllId {
			if i, ok := GOnlineMap[v]; ok {
				msg += fmt.Sprintf("Name: %s Id: %-5d Group_list: %v\n", i.Name, i.Id, i.Group_list)
			}
		}
		msg += "运行具体服务器命令请输入 run id list"
	case "stopped list":
		msg = p.getStoppedList()
	case "status":
		msg = p.getStatus()
	case "daemon status":
		p.getDaemonStatus(msgdata.User_id)
	default:
		msg += "服务器列表:\n"
		for _, v := range AllId {
			if i, ok := GOnlineMap[v]; ok {
				msg += fmt.Sprintf("Id: %-5d Name: %s\n", i.Id, i.Name)
			}
		}
		msg += "请添加 Id 参数"
	}
	return msg
}

func (p *admin) getStatus() string {
	var msg string
	msg += "服务器状态:\n"
	for k, v := range GOnlineMap {
		if v.Status == 2 || v.Status == 3 {
			msg += fmt.Sprintf("Id: %-5dStatus: RUNNING  Name: %s\n", k, v.Name)
		} else {
			msg += fmt.Sprintf("Id: %-5dStatus: STOPPED  Name: %s\n", k, v.Name)
		}
	}
	msg += "查询具体服务器请输入 run id status"
	return msg
}

func (p *admin) getStoppedList() string {
	var msg string
	msg += "未运行服务器列表:\n"
	for k, v := range GOnlineMap {
		if v.Status != 2 && v.Status != 3 {
			msg += fmt.Sprintf("Id: %-5dStatus: STOPPED  Name: %s\n", k, v.Name)
		}
	}
	return msg
}

func (p *admin) getDaemonStatus(user_id int) {
	UrlAndKey := gconfig.GetAllDaemon()
	client := &http.Client{}
	var data RemoteStatus
	for url, key := range *UrlAndKey {
		r2, _ := http.NewRequest("GET", url+"/api/overview", nil)
		r2.Close = true
		q := r2.URL.Query()
		q.Add("apikey", key)
		r2.URL.RawQuery = q.Encode()
		r2.Header.Set("x-requested-with", "xmlhttprequest")
		r, err := client.Do(r2)
		if err != nil {
			return
		}
		b, _ := io.ReadAll(r.Body)
		json.Unmarshal(b, &data)
		var sendmsg string
		sendmsg += fmt.Sprintf("前端面板地址: %s\n后端总数量: %d\n后端在线数量: %d", url, data.Data.RemoteCount.Total, data.Data.RemoteCount.Available)
		p.Send_private_msg(user_id, sendmsg)
		time.Sleep(time.Second)
		for _, tmpdata := range data.Data.Remote {
			sendmsg = ""
			sendmsg += fmt.Sprintf("后端地址: %s:%s\n连接状态: %v\n备注: %s\n平台: %s\nCpu: %.2f", tmpdata.Ip, tmpdata.Port, tmpdata.Available, tmpdata.Remarks, tmpdata.System.Platform, tmpdata.System.CpuUsage*100)
			sendmsg += "%%\n"
			sendmsg += fmt.Sprintf("Mem: %.2f", tmpdata.System.MemUsage*100)
			sendmsg += "%%\n"
			sendmsg += fmt.Sprintf("实例总个数: %d\n运行中实例个数: %d", tmpdata.Instance.Total, tmpdata.Instance.Running)
			p.Send_private_msg(user_id, sendmsg)
			time.Sleep(time.Second)
		}
	}
}

func (p *admin) Send_private_msg(user_id int, msg string, a ...interface{}) {
	var tmp SendData
	tmp.Action = "send_private_msg"
	tmp.Params.User_id = user_id
	tmp.Params.Message = fmt.Sprintf(msg, a...)
	p.SendChan <- &tmp
}
