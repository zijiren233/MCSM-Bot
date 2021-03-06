package rwmessage

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/zijiren233/MCSM-Bot/utils"
)

type Op struct {
	Op        int
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

func NewHdOp(send chan *SendData) *Op {
	p := Op{
		Op:        Qconfig.Cqhttp.Op,
		ChCqOpMsg: make(chan *MsgData, 25),
		SendChan:  send,
	}
	return &p
}

func (p *Op) Run() {
	POnlineMap[0] = p
	var msg *MsgData
	for {
		msg = <-p.ChCqOpMsg
		if msg.User_id == p.Op {
			if msg.Params[2] == "" {
				p.Send_private_msg("命令为空!")
				p.help("help")
				continue
			} else if msg.Params[1] == "" {
				p.help(msg.Params[2])
				continue
			}
			go p.handleMessage(msg)
		}
	}
}

func (p *Op) handleMessage(msg *MsgData) {
	id, err := strconv.Atoi(msg.Params[1])
	if err != nil {
		log.Error("strconv.Atoi error:%v", err)
		p.Send_private_msg("命令格式错误!\n请输入run help查看帮助!")
		return
	}
	if utils.InInt(id, AllId) {
		p.checkCMD(id, msg.Params[2])
	} else {
		p.Send_private_msg("请输入正确的ID!")
		log.Warring("OP 输入: %d 请输入正确的ID!", p.Op, msg.Params[1])
	}
}

func (p *Op) help(params string) {
	switch params {
	case "help":
		p.Send_private_msg("run list : 查看服务器列表\nrun status : 查看服务器状态\nrun daemon status : 查看MCSM后端状态\nrun id start : 启动服务器\nrun id stop : 关闭服务器\nrun id restart : 重启服务器\nrun id kill : 终止服务器\nrun id 服务器命令 : 运行服务器命令")
	case "list":
		var serverlist string
		serverlist += "服务器列表:\n"
		for _, v := range AllId {
			if i, ok := GOnlineMap[v]; ok {
				serverlist += fmt.Sprintf("Id: %-5dGroup: %-13dName: %s\n", i.Id, i.Group_id, i.Name)
			}
		}
		serverlist += "运行具体服务器命令请输入 run id list"
		p.Send_private_msg(serverlist)
	case "status":
		var serverstatus string
		serverstatus += "服务器状态:\n"
		for k, v := range GOnlineMap {
			if v.Status == 2 || v.Status == 3 {
				serverstatus += fmt.Sprintf("Id: %-5dStatus: RUNNING  Name: %s\n", k, v.Name)
			} else {
				serverstatus += fmt.Sprintf("Id: %-5dStatus: STOPPED  Name: %s\n", k, v.Name)
			}
		}
		serverstatus += "查询具体服务器请输入 run id status"
		p.Send_private_msg(serverstatus)
	case "daemon status":
		p.getDaemonStatus()
	default:
		var serverlist string
		serverlist += "服务器列表:\n"
		for _, v := range AllId {
			if i, ok := GOnlineMap[v]; ok {
				serverlist += fmt.Sprintf("Id: %-5d Name: %s\n", i.Id, i.Name)
			}
		}
		serverlist += "请添加 Id 参数"
		p.Send_private_msg(serverlist)
	}
}

func (p *Op) checkCMD(id int, params string) {
	var msg string
	var err error
	switch params {
	case "status":
		msg = GOnlineMap[id].GetStatus()
	case "start":
		msg, err = GOnlineMap[id].Start()
	case "stop":
		msg, err = GOnlineMap[id].Stop()
	case "restart":
		msg, err = GOnlineMap[id].Restart()
	case "kill":
		msg, err = GOnlineMap[id].Kill()
	default:
		msg, err = GOnlineMap[id].RunCmd(params)
	}
	if err != nil {
		return
	}
	p.Send_private_msg(msg)
}

func (p *Op) getDaemonStatus() {
	UrlAndKey := GetAllDaemon()
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
		b, _ := ioutil.ReadAll(r.Body)
		json.Unmarshal(b, &data)
		var sendmsg string
		sendmsg += fmt.Sprintf("前端面板地址: %s\n后端总数量: %d\n后端在线数量: %d", url, data.Data.RemoteCount.Total, data.Data.RemoteCount.Available)
		p.Send_private_msg(sendmsg)
		time.Sleep(time.Second)
		for _, tmpdata := range data.Data.Remote {
			sendmsg = ""
			sendmsg += fmt.Sprintf("后端地址: %s:%s\n连接状态: %v\n备注: %s\n平台: %s\nCpu: %.2f", tmpdata.Ip, tmpdata.Port, tmpdata.Available, tmpdata.Remarks, tmpdata.System.Platform, tmpdata.System.CpuUsage*100)
			sendmsg += "%%\n"
			sendmsg += fmt.Sprintf("Mem: %.2f", tmpdata.System.MemUsage*100)
			sendmsg += "%%\n"
			sendmsg += fmt.Sprintf("实例总个数: %d\n运行中实例个数: %d", tmpdata.Instance.Total, tmpdata.Instance.Running)
			p.Send_private_msg(sendmsg)
			time.Sleep(time.Second)
		}
	}
}

func (p *Op) Send_private_msg(msg string, a ...interface{}) {
	var tmp SendData
	tmp.Action = "send_private_msg"
	tmp.Params.User_id = p.Op
	tmp.Params.Message = fmt.Sprintf(msg, a...)
	p.SendChan <- &tmp
}
