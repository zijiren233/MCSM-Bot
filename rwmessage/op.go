package rwmessage

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/zijiren233/MCSM-Bot/logger"
)

type HdCqOp struct {
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

func NewHdCqOp(send chan *SendData) *HdCqOp {
	p := HdCqOp{
		Op:        Qconfig.Cqhttp.Op,
		ChCqOpMsg: make(chan *MsgData, 25),
		SendChan:  send,
	}
	return &p
}

func (p *HdCqOp) HdCqOp() {
	POnlineMap[0] = p
	var msg *MsgData
	var id int
	for {
		msg = <-p.ChCqOpMsg
		if msg.User_id == Qconfig.Cqhttp.Op {
			flysnowRegexp, _ := regexp.Compile(`^run ([0-9]*) *(.*)`)
			params := flysnowRegexp.FindString(msg.Message)
			if len(params) == 0 {
				continue
			}
			params2 := flysnowRegexp.FindStringSubmatch(params)
			if params2[1] == "" {
				if params2[2] == "addlisten" {
					var serverlist string
					serverlist += "未监听服务器列表:\n"
					for _, v := range AllId {
						if _, ok := GOnlineMap[v]; !ok {
							serverlist += fmt.Sprintf("Id: %-5dGroup: %d\n", id, Mconfig.McsmData[IdToOd[id]].Group_id)
						}
					}
					serverlist += "添加具体服务器请输入 run id addlisten"
					p.Send_private_msg(serverlist)
					continue
				}
				go p.help(params2[2])
				continue
			}
			id, _ = strconv.Atoi(params2[1])
			if InInt(id, AllId) {
				if params2[2] == "addlisten" {
					NewHdGroup(id, S.SendMessage)
					if v, ok := GOnlineMap[id]; ok {
						p.Send_private_msg("添加服务器: %s 监听成功!", v.Name)
					}
					continue
				}
				if _, ok := GOnlineMap[id]; !ok {
					logger.Log.Warring("OP 试图访问服务器Id: %d ,但未开启监听！", id)
					p.Send_private_msg("服务器Id: %d 未开启监听！", id)
					continue
				}
				go p.checkCMD(id, params2[2])
			} else {
				p.Send_private_msg("请输入正确的ID!")
				logger.Log.Warring("OP 输入: %d 请输入正确的ID!", p.Op, params)
			}
		}
	}
}

func (p *HdCqOp) help(params string) {
	switch params {
	case "help":
		p.Send_private_msg("run list : 查看服务器列表\nrun status : 查看已监听服务器状态\nrun id addlisten : 添加监听服务器\nrun server : 查看MCSM后端状态\nrun add listen : 添加监听服务器\nrun id start : 启动服务器\nrun id stop : 关闭服务器\nrun id restart : 重启服务器\nrun id kill : 终止服务器\nrun id 控制台命令 : 运行服务器命令")
	case "list":
		var serverlist string
		serverlist += "服务器列表:\n"
		for _, v := range AllId {
			if i, ok := GOnlineMap[v]; ok {
				serverlist += fmt.Sprintf("Id: %-5dGroup: %-13dName: %s    监听状态: 是\n", i.Id, i.Group_id, i.Name)
			} else {
				serverlist += fmt.Sprintf("Id: %-5dGroup: %-13d    监听状态: 否\n", v, Mconfig.McsmData[IdToOd[v]].Group_id)
			}
		}
		serverlist += "运行具体服务器命令请输入 run id list"
		p.Send_private_msg(serverlist)
	case "status":
		var serverstatus string
		serverstatus += "已监听服务器状态:\n"
		for k, v := range GOnlineMap {
			serverstatus += fmt.Sprintf("Id: %-5dStatus: %-5dName: %s\n", k, v.Status, v.Name)
		}
		serverstatus += "查询具体服务器请输入 run id status"
		p.Send_private_msg(serverstatus)
	case "server":
		p.GetDaemonStatus()
	default:
		var serverlist string
		serverlist += "服务器列表:\n"
		for _, v := range AllId {
			if i, ok := GOnlineMap[v]; ok {
				serverlist += fmt.Sprintf("Id: %-5d监听状态: 是    Name: %s\n", i.Id, i.Name)
			} else {
				serverlist += fmt.Sprintf("Id: %-5d监听状态: 否\n", v)
			}
		}
		serverlist += "请添加 Id 参数"
		p.Send_private_msg(serverlist)
	}
}

func (p *HdCqOp) checkCMD(id int, params string) {
	params = strings.ReplaceAll(params, "\n", "")
	params = strings.ReplaceAll(params, "\r", "")
	// if GOnlineMap[id].Status != 3 && GOnlineMap[id].Status != 2 && (params != "help" && params != "server" && params != "start") {
	// 	p.Send_private_msg("服务器: %s 未启动!\n请先启动服务器:\nrun %d start", GOnlineMap[id].Name, id)
	// 	return
	// }
	switch params {
	case "status":
		if GOnlineMap[id].Status == 2 || GOnlineMap[id].Status == 3 {
			if GOnlineMap[id].CurrentPlayers == "-1" {
				p.Send_private_msg("服务器: %s 正在运行!", GOnlineMap[id].Name)
			} else {
				p.Send_private_msg("服务器: %s 正在运行!\n服务器人数: %s\n服务器最大人数: %s\n服务器版本: %s\n服务器到期日期: %s", GOnlineMap[id].Name, GOnlineMap[id].CurrentPlayers, GOnlineMap[id].MaxPlayers, GOnlineMap[id].Version, GOnlineMap[id].EndTime)
			}
		} else {
			p.Send_private_msg("服务器: %s 未运行!", GOnlineMap[id].Name)
		}
	case "start":
		p.Start(id)
	case "stop":
		p.Stop(id)
	case "restart":
		p.Restart(id)
	case "kill":
		p.Kill(id)
	default:
		p.RunCmd(params, id)
	}
}

func (p *HdCqOp) RunCmd(commd string, id int) {
	if GOnlineMap[id].Status == 2 || GOnlineMap[id].Status == 3 {
		client := &http.Client{}
		r2, _ := http.NewRequest("GET", GOnlineMap[id].Url+"/api/protected_instance/command", nil)
		r2.Close = true
		q := r2.URL.Query()
		q.Add("apikey", GOnlineMap[id].Apikey)
		q.Add("uuid", GOnlineMap[id].Uuid)
		q.Add("remote_uuid", GOnlineMap[id].Remote_uuid)
		q.Add("command", commd)
		r2.URL.RawQuery = q.Encode()
		r2.Header.Set("x-requested-with", "xmlhttprequest")
		r, err := client.Do(r2)
		if err != nil {
			p.Send_private_msg("运行命令 %s 失败！", commd)
			logger.Log.Error("运行命令 %s 失败！%v", commd, err)
			return
		}
		b, _ := ioutil.ReadAll(r.Body)
		var time_unix CmdData
		json.Unmarshal(b, &time_unix)
		time.Sleep(70 * time.Millisecond)
		p.ReturnResult(commd, time_unix.Time_unix, id)
	} else {
		p.Send_private_msg("服务器: %s 未运行!", GOnlineMap[id].Name)
	}

}

func (p *HdCqOp) ReturnResult(command string, time_now int64, id int) {
	client := &http.Client{}
	r2, _ := http.NewRequest("GET", GOnlineMap[id].Url+"/api/protected_instance/outputlog", nil)
	r2.Close = true
	r2.Header.Set("x-requested-with", "xmlhttprequest")
	q := r2.URL.Query()
	q.Add("apikey", GOnlineMap[id].Apikey)
	q.Add("uuid", GOnlineMap[id].Uuid)
	q.Add("remote_uuid", GOnlineMap[id].Remote_uuid)
	r2.URL.RawQuery = q.Encode()
	r, err := client.Do(r2)
	if err != nil {
		logger.Log.Error("获取服务器 %s 命令 %s 运行结果失败！", GOnlineMap[id].Name, command)
		return
	}
	b, _ := ioutil.ReadAll(r.Body)
	r3, _ := regexp.Compile(`\\r+|\\u001b\[?=?[a-zA-Z]?\?*[0-9]*[hl]*>? ?[0-9;]*m*`)
	ret := r3.ReplaceAllString(string(b), "")
	last := strings.LastIndex(ret, `","time":`)
	var index int
	var i int64
	logger.Log.Debug("服务器 %s 运行命令 %s 返回时间: %s", GOnlineMap[id].Name, command, time.Unix((time_now/1000)+i, 0).Format("15:04:05"))
	for i = 0; i <= 2; i++ {
		index = strings.Index(ret, time.Unix((time_now/1000)+i, 0).Format("15:04:05"))
		if index == -1 {
			continue
		}
		p.Send_private_msg("> [%s] %s\n%s", GOnlineMap[id].Name, command, *(p.handle_End_Newline(ret[index-1 : last])))
		return
	}
	index = strings.Index(ret, time.Unix((time_now/1000)-1, 0).Format("15:04:05"))
	if index == -1 {
		p.Send_private_msg("运行命令成功！")
		logger.Log.Warring("服务器 %s 命令 %s 成功,但未查找到返回时间: %s", GOnlineMap[id].Name, command, time.Unix((time_now/1000)+i, 0).Format("15:04:05"))
		return
	}
	p.Send_private_msg("> [%s] %s\n%s", GOnlineMap[id].Name, command, *(p.handle_End_Newline(ret[index-1 : last])))
}

func (p *HdCqOp) handle_End_Newline(msg string) *string {
	var data Data
	last := strings.LastIndex(msg, `\n`)
	if last == len(msg)-2 {
		msg = (msg)[:last]
	}
	msg = fmt.Sprint(`{"data":"`, msg, `"}`)
	json.Unmarshal([]byte(msg), &data)
	return &data.Data
}

func (p *HdCqOp) Start(id int) {
	if GOnlineMap[id].Status != 2 && GOnlineMap[id].Status != 3 {
		client := &http.Client{}
		r2, _ := http.NewRequest("GET", GOnlineMap[id].Url+"/api/protected_instance/open", nil)
		r2.Close = true
		q := r2.URL.Query()
		q.Add("apikey", GOnlineMap[id].Apikey)
		q.Add("uuid", GOnlineMap[id].Uuid)
		q.Add("remote_uuid", GOnlineMap[id].Remote_uuid)
		r2.URL.RawQuery = q.Encode()
		r2.Header.Set("x-requested-with", "xmlhttprequest")
		_, err := client.Do(r2)
		if err != nil {
			logger.Log.Warring("服务器: %s 运行启动命令失败,可能是网络问题!", GOnlineMap[id].Name)
			return
		}
		p.Send_private_msg("服务器: %s 正在启动!", GOnlineMap[id].Name)
	} else {
		p.Send_private_msg("服务器: %s 已运行!", GOnlineMap[id].Name)
	}
}

func (p *HdCqOp) Stop(id int) {
	if GOnlineMap[id].Status == 2 || GOnlineMap[id].Status == 3 {
		client := &http.Client{}
		r2, _ := http.NewRequest("GET", GOnlineMap[id].Url+"/api/protected_instance/stop", nil)
		r2.Close = true
		q := r2.URL.Query()
		q.Add("apikey", GOnlineMap[id].Apikey)
		q.Add("uuid", GOnlineMap[id].Uuid)
		q.Add("remote_uuid", GOnlineMap[id].Remote_uuid)
		r2.URL.RawQuery = q.Encode()
		r2.Header.Set("x-requested-with", "xmlhttprequest")
		_, err := client.Do(r2)
		if err != nil {
			logger.Log.Warring("服务器: %s 运行关闭命令失败,可能是网络问题!", GOnlineMap[id].Name)
			return
		}
		p.Send_private_msg("服务器: %s 正在停止!", GOnlineMap[id].Name)
	} else {
		p.Send_private_msg("服务器: %s 已停止!", GOnlineMap[id].Name)
	}
}

func (p *HdCqOp) Restart(id int) {
	client := &http.Client{}
	r2, _ := http.NewRequest("GET", GOnlineMap[id].Url+"/api/protected_instance/restart", nil)
	r2.Close = true
	q := r2.URL.Query()
	q.Add("apikey", GOnlineMap[id].Apikey)
	q.Add("uuid", GOnlineMap[id].Uuid)
	q.Add("remote_uuid", GOnlineMap[id].Remote_uuid)
	r2.URL.RawQuery = q.Encode()
	r2.Header.Set("x-requested-with", "xmlhttprequest")
	_, err := client.Do(r2)
	if err != nil {
		p.Send_private_msg("服务器: %s 运行重启命令失败!", GOnlineMap[id].Name)
		logger.Log.Warring("服务器: %s 运行重启命令失败,可能是网络问题!", GOnlineMap[id].Name)
		return
	}
	p.Send_private_msg("服务器: %s 重启中!", GOnlineMap[id].Name)
}

func (p *HdCqOp) Kill(id int) {
	if GOnlineMap[id].Status == 2 || GOnlineMap[id].Status == 3 {
		client := &http.Client{}
		r2, _ := http.NewRequest("GET", GOnlineMap[id].Url+"/api/protected_instance/kill", nil)
		r2.Close = true
		q := r2.URL.Query()
		q.Add("apikey", GOnlineMap[id].Apikey)
		q.Add("uuid", GOnlineMap[id].Uuid)
		q.Add("remote_uuid", GOnlineMap[id].Remote_uuid)
		r2.URL.RawQuery = q.Encode()
		r2.Header.Set("x-requested-with", "xmlhttprequest")
		_, err := client.Do(r2)
		if err != nil {
			p.Send_private_msg("服务器: %s 运行终止命令失败!", GOnlineMap[id].Name)
			logger.Log.Warring("服务器: %s 运行终止命令失败,可能是网络问题!", GOnlineMap[id].Name)
			return
		}
		p.Send_private_msg("服务器: %s 已终止!", GOnlineMap[id].Name)
	} else {
		p.Send_private_msg("服务器: %s 已停止!", GOnlineMap[id].Name)
	}
}

func (p *HdCqOp) GetDaemonStatus() {
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

func (p *HdCqOp) Send_private_msg(msg string, a ...interface{}) {
	var tmp SendData
	tmp.Action = "send_private_msg"
	tmp.Params.User_id = p.Op
	tmp.Params.Message = fmt.Sprintf(msg, a...)
	p.SendChan <- &tmp
}
