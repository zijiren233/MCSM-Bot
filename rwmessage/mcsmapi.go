package rwmessage

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"time"
)

type Data struct {
	Data string `json:"data"`
}

type CmdData struct {
	Time_unix int64 `json:"time"`
}

func (u *HdGroup) Start() {
	client := &http.Client{}
	r2, _ := http.NewRequest("GET", u.Url+"/api/protected_instance/open", nil)
	r2.Close = true
	q := r2.URL.Query()
	q.Add("apikey", u.Apikey)
	q.Add("uuid", u.Uuid)
	q.Add("remote_uuid", u.Remote_uuid)
	r2.URL.RawQuery = q.Encode()
	r2.Header.Set("x-requested-with", "xmlhttprequest")
	_, err := client.Do(r2)
	if err != nil {
		Log.Warring("服务器:%s 运行启动命令失败,可能是网络问题!", u.Name)
		return
	}
}

func (u *HdGroup) Stop() {
	client := &http.Client{}
	r2, _ := http.NewRequest("GET", u.Url+"/api/protected_instance/stop", nil)
	r2.Close = true
	q := r2.URL.Query()
	q.Add("apikey", u.Apikey)
	q.Add("uuid", u.Uuid)
	q.Add("remote_uuid", u.Remote_uuid)
	r2.URL.RawQuery = q.Encode()
	r2.Header.Set("x-requested-with", "xmlhttprequest")
	_, err := client.Do(r2)
	if err != nil {
		Log.Warring("服务器:%s 运行关闭命令失败,可能是网络问题!", u.Name)
		return
	}
}

func (u *HdGroup) RunCmd(commd string) {
	client := &http.Client{}
	r2, _ := http.NewRequest("GET", u.Url+"/api/protected_instance/command", nil)
	r2.Close = true
	q := r2.URL.Query()
	q.Add("apikey", u.Apikey)
	q.Add("uuid", u.Uuid)
	q.Add("remote_uuid", u.Remote_uuid)
	q.Add("command", commd)
	r2.URL.RawQuery = q.Encode()
	r2.Header.Set("x-requested-with", "xmlhttprequest")
	r, err := client.Do(r2)
	if err != nil {
		u.Send_group_msg("运行命令 %s 失败！", commd)
		Log.Error("运行命令 %s 失败！%v", commd, err)
		return
	}
	b, _ := ioutil.ReadAll(r.Body)
	var time_unix CmdData
	json.Unmarshal(b, &time_unix)
	time.Sleep(100 * time.Millisecond)
	u.ReturnResult(commd, time_unix.Time_unix)
}

func (u *HdGroup) ReturnResult(command string, time_now int64) {
	client := &http.Client{}
	r2, _ := http.NewRequest("GET", u.Url+"/api/protected_instance/outputlog", nil)
	r2.Close = true
	r2.Header.Set("x-requested-with", "xmlhttprequest")
	q := r2.URL.Query()
	q.Add("apikey", u.Apikey)
	q.Add("uuid", u.Uuid)
	q.Add("remote_uuid", u.Remote_uuid)
	r2.URL.RawQuery = q.Encode()
	r, err := client.Do(r2)
	if err != nil {
		Log.Error("获取服务器 %s 命令 %s 运行结果失败！", u.Name, command)
		return
	}
	b, _ := ioutil.ReadAll(r.Body)
	r3, _ := regexp.Compile(`\\r+|\\u001b\[?=?[a-zA-Z]?\?*[0-9]*[hl]*>? ?[0-9;]*m*`)
	ret := r3.ReplaceAllString(string(b), "")
	last := strings.LastIndex(ret, `","time":`)
	var index int
	var i int64
	Log.Debug("服务器 %s 运行命令 %s 返回时间: %s", u.Name, command, time.Unix((time_now/1000)+i, 0).Format("15:04:05"))
	for i = 0; i <= 2; i++ {
		index = strings.Index(ret, time.Unix((time_now/1000)+i, 0).Format("15:04:05"))
		if index == -1 {
			continue
		}
		tmp := ret[index-1 : last]
		u.Send_group_msg("> [%s] %s\n%s", u.Name, command, *(u.handle_End_Newline(&tmp)))
		return
	}
	index = strings.Index(ret, time.Unix((time_now/1000)-1, 0).Format("15:04:05"))
	if index == -1 {
		u.Send_group_msg("运行命令成功！")
		Log.Warring("服务器 %s 命令 %s 成功,但未查找到返回时间: %s", u.Name, command, time.Unix((time_now/1000)+i, 0).Format("15:04:05"))
		return
	}
	tmp := ret[index-1 : last]
	u.Send_group_msg("> [%s] %s\n%s", u.Name, command, *(u.handle_End_Newline(&tmp)))
}

func (u *HdGroup) handle_End_Newline(msg *string) *string {
	var data Data
	last := strings.LastIndex(*msg, `\n`)
	if last == len(*msg)-2 {
		*msg = (*msg)[:last]
	}
	*msg = fmt.Sprint(`{"data":"`, *msg, `"}`)
	json.Unmarshal([]byte(*msg), &data)
	return &data.Data
}

func (u *HdGroup) TestMcsmStatus() bool {
	client := &http.Client{}
	r2, _ := http.NewRequest("GET", u.Url+"/api/instance", nil)
	r2.Close = true
	q := r2.URL.Query()
	q.Add("apikey", u.Apikey)
	q.Add("uuid", u.Uuid)
	q.Add("remote_uuid", u.Remote_uuid)
	r2.URL.RawQuery = q.Encode()
	r2.Header.Set("x-requested-with", "xmlhttprequest")
	_, err2 := client.Do(r2)
	if err2 != nil {
		fmt.Printf("服务器:%s MCSM前端连接失败，请检查配置文件是否填写正确或MCSM是否启动\n", u.Name)
		Log.Error("服务器:%s MCSM前端连接失败，请检查配置文件是否填写正确或MCSM是否启动", u.Name)
		return false
	}
	return true
}

func (u *HdGroup) Restart() {
	client := &http.Client{}
	r2, _ := http.NewRequest("GET", u.Url+"/api/protected_instance/restart", nil)
	r2.Close = true
	q := r2.URL.Query()
	q.Add("apikey", u.Apikey)
	q.Add("uuid", u.Uuid)
	q.Add("remote_uuid", u.Remote_uuid)
	r2.URL.RawQuery = q.Encode()
	r2.Header.Set("x-requested-with", "xmlhttprequest")
	_, err := client.Do(r2)
	if err != nil {
		u.Send_group_msg("服务器:%s 运行重启命令失败!", u.Name)
		Log.Warring("服务器:%s 运行重启命令失败,可能是网络问题!", u.Name)
		return
	}
}

func (u *HdGroup) Kill() {
	client := &http.Client{}
	r2, _ := http.NewRequest("GET", u.Url+"/api/protected_instance/kill", nil)
	r2.Close = true
	q := r2.URL.Query()
	q.Add("apikey", u.Apikey)
	q.Add("uuid", u.Uuid)
	q.Add("remote_uuid", u.Remote_uuid)
	r2.URL.RawQuery = q.Encode()
	r2.Header.Set("x-requested-with", "xmlhttprequest")
	_, err := client.Do(r2)
	if err != nil {
		u.Send_group_msg("服务器:%s 运行终止命令失败!", u.Name)
		Log.Warring("服务器:%s 运行终止命令失败,可能是网络问题!", u.Name)
		return
	}
}
