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
)

type HdCqOp struct {
	Op        int
	ChCqOpMsg chan *MsgData
	SendChan  chan *SendData
}

// var mcmd = []string{"help", "list", "status", "add listen"}

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
				return
			}
			params2 := flysnowRegexp.FindStringSubmatch(params)
			if params2[1] == "" {
				go p.help(params2[2])
				continue
			}
			id, _ = strconv.Atoi(params2[1])
			if InInt(id, AllId) {
				if _, ok := GOnlineMap[id]; !ok {
					Log.Warring("OP 试图访问服务器:%s ,%s 未开启监听！", Mconfig.McsmData[IdToOd[id]].Name, Mconfig.McsmData[IdToOd[id]].Name)
					p.Send_private_msg("%s 未开启监听！", Mconfig.McsmData[IdToOd[id]].Name)
				}
				go p.checkCMD(id, params2[2])
			} else {
				p.Send_private_msg("请输入正确的ID!")
				Log.Warring("OP 输入:%d 请输入正确的ID!", p.Op, params)
			}
		}
	}
}

func (p *HdCqOp) help(params string) {
	switch params {
	case "help":
		p.Send_private_msg("待完善...")
	case "list":
		var serverlist string
		serverlist += "服务器列表:\n"
		for _, v := range AllId {
			if i, ok := GOnlineMap[v]; ok {
				serverlist += fmt.Sprintf("Name: %s    Id: %d    Group:%d    监听状态: 是\n", i.Name, i.Id, i.Group_id)
			} else {
				serverlist += fmt.Sprintf("Name: %s    Id: %d    Group:%d    监听状态: 否\n", i.Name, i.Id, i.Group_id)
			}
		}
		serverlist += "查询具体服务器请输入 run id list"
		p.Send_private_msg(serverlist)
	case "status":
		var serverstatus string
		serverstatus += "已监听服务器状态:\n"
		for k, v := range GOnlineMap {
			serverstatus += fmt.Sprintf("Name: %s    Id: %d    Status: %d\n", v.Name, k, v.Status)
		}
		serverstatus += "查询具体服务器请输入 run id status"
		p.Send_private_msg(serverstatus)
	case "add listen":
		p.Send_private_msg("待完善...")
	default:
		var serverlist string
		serverlist += "服务器列表:\n"
		for _, v := range AllId {
			if i, ok := GOnlineMap[v]; ok {
				serverlist += fmt.Sprintf("Name: %s    Id: %d    监听状态: 是\n", i.Name, i.Id)
			} else {
				serverlist += fmt.Sprintf("Name: %s    Id: %d    监听状态: 否\n", i.Name, i.Id)
			}
		}
		serverlist += "请添加 Id 参数"
		p.Send_private_msg(serverlist)
	}
}

func (p *HdCqOp) checkCMD(id int, params string) {
	params = strings.ReplaceAll(params, "\n", "")
	params = strings.ReplaceAll(params, "\r", "")
	switch params {
	case "status":
		if GOnlineMap[id].Status == 1 {
			if GOnlineMap[id].CurrentPlayers == "-1" {
				p.Send_private_msg("服务器:%s 正在运行!", GOnlineMap[id].Name)
			} else {
				p.Send_private_msg("服务器:%s 正在运行!\n服务器人数:%s\n服务器最大人数:%s\n服务器版本:%s\n服务器到期日期:%s", GOnlineMap[id].Name, GOnlineMap[id].CurrentPlayers, GOnlineMap[id].MaxPlayers, GOnlineMap[id].Version, GOnlineMap[id].EndTime)
			}
		} else if GOnlineMap[id].Status == 0 {
			p.Send_private_msg("服务器:%s 未运行!", GOnlineMap[id].Name)
		}
	case "start":
		if GOnlineMap[id].Status == 0 {
			p.Start(id)
			p.Send_private_msg("服务器:%s 启动中!", GOnlineMap[id].Name)
		} else {
			p.Send_private_msg("服务器:%s 已在运行!", GOnlineMap[id].Name)
		}
	case "stop":
		if GOnlineMap[id].Status == 1 {
			p.Stop(id)
			p.Send_private_msg("服务器:%s 正在停止!", GOnlineMap[id].Name)
		} else {
			p.Send_private_msg("服务器:%s 未运行!", GOnlineMap[id].Name)
		}
	case "restart":
		p.Restart(id)
		p.Send_private_msg("服务器:%s 重启中!", GOnlineMap[id].Name)
	case "kill":
		p.Kill(id)
		p.Send_private_msg("服务器:%s 已终止!", GOnlineMap[id].Name)
	default:
		p.RunCmd(params, id)
	}
}

func (p *HdCqOp) RunCmd(commd string, id int) {
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
		Log.Error("运行命令 %s 失败！%v", commd, err)
		return
	}
	b, _ := ioutil.ReadAll(r.Body)
	var time_unix CmdData
	json.Unmarshal(b, &time_unix)
	time.Sleep(100 * time.Millisecond)
	p.ReturnResult(commd, time_unix.Time_unix, id)
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
		Log.Error("获取服务器 %s 命令 %s 运行结果失败！", GOnlineMap[id].Name, command)
		return
	}
	b, _ := ioutil.ReadAll(r.Body)
	r3, _ := regexp.Compile(`\\r+|\\u001b\[?=?[a-zA-Z]?\?*[0-9]*[hl]*>? ?[0-9;]*m*`)
	ret := r3.ReplaceAllString(string(b), "")
	last := strings.LastIndex(ret, `","time":`)
	var index int
	var i int64
	Log.Debug("服务器 %s 运行命令 %s 返回时间: %s", GOnlineMap[id].Name, command, time.Unix((time_now/1000)+i, 0).Format("15:04:05"))
	for i = 0; i <= 2; i++ {
		index = strings.Index(ret, time.Unix((time_now/1000)+i, 0).Format("15:04:05"))
		if index == -1 {
			continue
		}
		tmp := ret[index-1 : last]
		p.Send_private_msg("> [%s] %s\n%s", GOnlineMap[id].Name, command, *(GOnlineMap[id].handle_End_Newline(&tmp)))
		return
	}
	index = strings.Index(ret, time.Unix((time_now/1000)-1, 0).Format("15:04:05"))
	if index == -1 {
		p.Send_private_msg("运行命令成功！")
		Log.Warring("服务器 %s 命令 %s 成功,但未查找到返回时间: %s", GOnlineMap[id].Name, command, time.Unix((time_now/1000)+i, 0).Format("15:04:05"))
		return
	}
	tmp := ret[index-1 : last]
	p.Send_private_msg("> [%s] %s\n%s", GOnlineMap[id].Name, command, *(GOnlineMap[id].handle_End_Newline(&tmp)))
}

func (p *HdCqOp) Start(id int) {
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
		Log.Warring("服务器:%s 运行启动命令失败,可能是网络问题!", GOnlineMap[id].Name)
		return
	}
	p.Send_private_msg("服务器:%s 正在启动!", GOnlineMap[id].Name)
}

func (p *HdCqOp) Stop(id int) {
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
		Log.Warring("服务器:%s 运行关闭命令失败,可能是网络问题!", GOnlineMap[id].Name)
		return
	}
	p.Send_private_msg("服务器:%s 正在停止!", GOnlineMap[id].Name)
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
		p.Send_private_msg("服务器:%s 运行重启命令失败!", GOnlineMap[id].Name)
		Log.Warring("服务器:%s 运行重启命令失败,可能是网络问题!", GOnlineMap[id].Name)
		return
	}
	p.Send_private_msg("服务器:%s 重启中!", GOnlineMap[id].Name)
}

func (p *HdCqOp) Kill(id int) {
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
		p.Send_private_msg("服务器:%s 运行终止命令失败!", GOnlineMap[id].Name)
		Log.Warring("服务器:%s 运行终止命令失败,可能是网络问题!", GOnlineMap[id].Name)
		return
	}
	p.Send_private_msg("服务器:%s 已终止!", GOnlineMap[id].Name)
}

func (p *HdCqOp) Send_private_msg(msg string, a ...interface{}) {
	var tmp SendData
	tmp.Action = "send_private_msg"
	tmp.Params.User_id = p.Op
	tmp.Params.Message = fmt.Sprintf(msg, a...)
	p.SendChan <- &tmp
}
