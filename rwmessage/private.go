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

type HdPrivate struct {
	Op           int
	ChPrivateMsg chan MsgData
	SendChan     chan SendData
}

func NewHdPrivate(send chan SendData) *HdPrivate {
	p := HdPrivate{
		Op:           Qconfig.Cqhttp.Op,
		ChPrivateMsg: make(chan MsgData, 25),
		SendChan:     send,
	}
	return &p
}

func (p *HdPrivate) HdOpPrivate() {
	POnlineMap[0] = p
	var tmp MsgData
	for {
		tmp = <-p.ChPrivateMsg
		if tmp.User_id == Qconfig.Cqhttp.Op {
			flysnowRegexp, _ := regexp.Compile(`^run ([0-9]*) *(.*)`)
			params := flysnowRegexp.FindString(tmp.Message)
			if len(params) == 0 {
				return
			}
			params2 := flysnowRegexp.FindStringSubmatch(params)
			id, _ := strconv.Atoi(params2[1])
			if In(id, AllId) {
				if _, ok := GOnlineMap[id]; !ok {
					Log.Warring("%s 未开启监听！", Mconfig.McsmData[IdToOd[id]].Name)
					p.Send_private_msg("%s 未开启监听！", Mconfig.McsmData[IdToOd[id]].Name)
				}
				go p.checkCMD(id, params2[2])
			} else {
				p.Send_private_msg("请输入正确的ID!")
				Log.Warring("%d 输入:%d 请输入正确的ID!", p.Op, params)
			}
		}
	}
}

func (p *HdPrivate) checkCMD(id int, params string) {
	params = strings.ReplaceAll(params, "\n", "")
	params = strings.ReplaceAll(params, "\r", "")
	switch params {
	case "status":
		if GOnlineMap[id].Status == 1 {
			p.Send_private_msg("服务器:%s 正在运行!", GOnlineMap[id].Name)
		} else if GOnlineMap[id].Status == 0 {
			p.Send_private_msg("服务器:%s 未运行!", GOnlineMap[id].Name)
		}
	case "start":
		if GOnlineMap[id].Status == 0 {
			p.Start(id)
		} else {
			p.Send_private_msg("服务器:%s 已在运行!", GOnlineMap[id].Name)
		}
	case "stop":
		if GOnlineMap[id].Status == 1 {
			p.Stop(id)
		} else {
			p.Send_private_msg("服务器:%s 未运行!", GOnlineMap[id].Name)
		}
	case "restart":
		p.Restart(id)
	case "kill":
		p.Kill(id)
	default:
		p.RunCmd(params, id)
	}
}

func (p *HdPrivate) RunCmd(commd string, id int) {
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
	time.Sleep(40 * time.Millisecond)
	p.ReturnResult(commd, time_unix.Time_unix, id)
}

func (p *HdPrivate) ReturnResult(command string, time_now int64, id int) {
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

func (p *HdPrivate) Start(id int) {
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

func (p *HdPrivate) Stop(id int) {
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

func (p *HdPrivate) Restart(id int) {
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

func (p *HdPrivate) Kill(id int) {
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

func (p *HdPrivate) Send_private_msg(msg string, a ...interface{}) {
	var tmp SendData
	tmp.Action = "send_private_msg"
	tmp.Params.User_id = p.Op
	tmp.Params.Message = fmt.Sprintf(msg, a...)
	p.SendChan <- tmp
}
