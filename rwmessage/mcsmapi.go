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
	if u.EndTime != "" {
		t, _ := time.Parse("2006/1/2", u.EndTime)
		if t.Before(time.Now()) {
			log.Warring("服务器ID: %d ,NAME: %s 已到期,启动失败!", u.Id, u.Name)
			return fmt.Sprintf("服务器ID: %d ,NAME: %s 已到期,启动失败!", u.Id, u.Name), nil
		}
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
		log.Warring("服务器: %s 运行启动命令失败,可能是网络问题!", u.Name)
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
		log.Warring("服务器: %s 运行关闭命令失败,可能是网络问题!", u.Name)
		return "", err
	}
	return fmt.Sprintf("服务器: %s 正在关闭!", u.Name), nil
}

// 返回控制台结果，如果未查询到则返回 "运行成功"
func (u *HdGroup) RunCmd(commd string) (string, error) {
	if u.Status != 2 && u.Status != 3 {
		return fmt.Sprintf("服务器: %s 未运行!", u.Name), nil
	}
	if commd == "" {
		return "运行命令为空!", nil
	}
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
		log.Error("运行命令 %s 失败！%v", commd, err)
		return fmt.Sprintf("运行命令 %s 失败！", commd), err
	}
	time.Sleep(450 * time.Millisecond)
	return u.returnResult(commd, 2)
}

func (u *HdGroup) returnResult(command string, try uint8) (string, error) {
	if try == 0 {
		return "执行控制台命令成功! 可能由于 {网络延迟,控制台乱码} 导致运行结果返回失败", nil
	}
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
		log.Error("获取服务器 %s 命令 %s 运行结果失败！", u.Name, command)
		return u.returnResult(command, try-1)
	}
	b, _ := ioutil.ReadAll(r.Body)
	var data Data
	json.Unmarshal(b, &data)
	b2 := utils.NoColorable(&data.Data).String()
	r3, _ := regexp.Compile(`(?m)(` + command + `(\r)+?)$`)
	i := r3.FindAllStringIndex(b2, -1)
	if len(i) != 0 {
		return fmt.Sprintf("[%s] %s", u.Name, (*utils.Handle_End_Newline(&b2))[i[len(i)-1][0]:]), nil
	}
	log.Debug("终端信息: %#v\n", b2)
	return u.returnResult(command, try-1)
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
		log.Warring("服务器: %s 运行重启命令失败,可能是网络问题!", u.Name)
		return fmt.Sprintf("服务器: %s 运行重启命令失败,可能是网络问题!", u.Name), err
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
		log.Warring("服务器: %s 运行终止命令失败,可能是网络问题!", u.Name)
		return fmt.Sprintf("服务器: %s 运行终止命令失败,可能是网络问题!", u.Name), err
	}
	return fmt.Sprintf("服务器: %s 已经终止!", u.Name), nil
}

func (u *HdGroup) getStatusInfo() error {
	client := &http.Client{}
	r2, _ := http.NewRequest("GET", u.Url+"/api/instance", nil)
	r2.Close = true
	q := r2.URL.Query()
	q.Add("apikey", u.Apikey)
	q.Add("uuid", u.Uuid)
	q.Add("remote_uuid", u.Remote_uuid)
	r2.URL.RawQuery = q.Encode()
	r2.URL.RawQuery = q.Encode()
	r2.Header.Set("x-requested-with", "xmlhttprequest")
	r, err := client.Do(r2)
	if err != nil {
		log.Error("获取服务器Id: %d 信息失败! err: %v", u.Id, err)
		return err
	}
	b, _ := ioutil.ReadAll(r.Body)
	var status InstanceConfig
	json.Unmarshal(b, &status)
	u.lock.Lock()
	defer u.lock.Unlock()
	u.Status = status.Data.Status
	u.Name = status.Data.Config.Nickname
	u.EndTime = status.Data.Config.EndTime
	u.ProcessType = status.Data.Config.ProcessType
	u.Pty = status.Data.Config.TerminalOption.Pty
	u.PingIp = status.Data.Config.PingConfig.PingIp
	u.CurrentPlayers = status.Data.Info.CurrentPlayers
	u.MaxPlayers = status.Data.Info.MaxPlayers
	u.Version = status.Data.Info.Version
	return nil
}

func (u *HdGroup) GetStatus() string {
	u.lock.RLock()
	defer u.lock.RUnlock()
	if u.Status == 2 || u.Status == 3 {
		if u.CurrentPlayers == "" {
			return fmt.Sprintf("服务器: %s 状态查询 功能未开启!请前往实例中开启状态查询功能", u.Name)
		} else {
			return fmt.Sprintf("服务器: %s 正在运行!\n服务器人数: %s\n服务器最大人数: %s\n服务器版本: %s", u.Name, u.CurrentPlayers, u.MaxPlayers, u.Version)
		}
	} else {
		return fmt.Sprintf("服务器: %s 未运行!", u.Name)
	}
}
