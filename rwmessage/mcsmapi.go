package rwmessage

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"time"

	"github.com/zijiren233/MCSM-Bot/utils"
)

type Data struct {
	Data string `json:"data"`
}

func (u *HdGroup) Start() (string, error) {
	if u.Status == 2 || u.Status == 3 {
		return fmt.Sprintf("服务器: %s 已经运行!", u.Name), nil
	}
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
		Log.Warring("服务器: %s 运行启动命令失败,可能是网络问题!", u.Name)
		return "", err
	}
	return fmt.Sprintf("服务器: %s 正在启动!", u.Name), nil
}

func (u *HdGroup) Stop() (string, error) {
	if u.Status != 2 && u.Status != 3 {
		return fmt.Sprintf("服务器: %s 未运行!", u.Name), nil
	}
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
		Log.Warring("服务器: %s 运行关闭命令失败,可能是网络问题!", u.Name)
		return "", err
	}
	return fmt.Sprintf("服务器: %s 正在关闭!", u.Name), nil
}

// 返回控制台结果，如果未查询到则返回 "运行成功"
func (u *HdGroup) RunCmd(commd string) (string, error) {
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
	_, err := client.Do(r2)
	if err != nil {
		Log.Error("运行命令 %s 失败！%v", commd, err)
		return fmt.Sprintf("运行命令 %s 失败！", commd), err
	}
	time.Sleep(250 * time.Millisecond)
	return u.returnResult(commd)
}

func (u *HdGroup) returnResult(command string) (string, error) {
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
		return "", err
	}
	b, _ := ioutil.ReadAll(r.Body)
	var data Data
	json.Unmarshal(b, &data)
	b2, _ := utils.NoColorable(&data.Data)
	r3, _ := regexp.Compile(`(?m)(` + command + `(\r)+?)$`)
	i := r3.FindAllStringIndex(b2.String(), -1)
	if len(i) != 0 {
		msg := b2.String()
		return fmt.Sprintf("[%s] %s", u.Name, (*utils.Handle_End_Newline(&msg))[i[len(i)-1][0]:]), nil
	}
	Log.Error("查找命令失败:%#v\ncmd:%s", b2.String(), command)
	return "运行命令成功!", nil
}

func (u *HdGroup) Restart() (string, error) {
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
		// u.Send_group_msg("服务器: %s 运行重启命令失败!", u.Name)
		Log.Warring("服务器: %s 运行重启命令失败,可能是网络问题!", u.Name)
		return "", err
	}
	return fmt.Sprintf("服务器: %s 重启中!", u.Name), nil
}

func (u *HdGroup) Kill() (string, error) {
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
		// u.Send_group_msg("服务器: %s 运行终止命令失败!", u.Name)
		Log.Warring("服务器: %s 运行终止命令失败,可能是网络问题!", u.Name)
		return "", err
	}
	return fmt.Sprintf("服务器: %s 已经终止!", u.Name), nil
}
