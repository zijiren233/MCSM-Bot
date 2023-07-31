package rwmessage

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/zijiren233/go-colorable"
	"github.com/zijiren233/go-colorlog"
)

type Data struct {
	Data string `json:"data"`
}

func (u *HdGroup) Start() (string, error) {
	if u.Status == 2 || u.Status == 3 {
		return fmt.Sprintf("实例: %s 已经运行!", u.Name), nil
	}
	if u.EndTime != "" {
		t, _ := time.Parse("2006/1/2", u.EndTime)
		if t.Before(time.Now()) {
			colorlog.Warningf("实例ID: %d ,NAME: %s 已到期,启动失败!", u.Id, u.Name)
			return fmt.Sprintf("实例ID: %d ,NAME: %s 已到期,启动失败!", u.Id, u.Name), nil
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
	r, err := client.Do(r2)
	if err != nil {
		colorlog.Warningf("实例: %s 运行启动命令失败,可能是网络问题!", u.Name)
		return "", err
	}
	r.Body.Close()
	return fmt.Sprintf("实例: %s 正在启动!", u.Name), nil
}

func (u *HdGroup) Stop() (string, error) {
	if u.Status != 2 && u.Status != 3 {
		return fmt.Sprintf("实例: %s 未运行!", u.Name), nil
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
	r, err := client.Do(r2)
	if err != nil {
		colorlog.Warningf("实例: %s 运行关闭命令失败,可能是网络问题!", u.Name)
		return "", err
	}
	r.Body.Close()
	return fmt.Sprintf("实例: %s 正在关闭!", u.Name), nil
}

// 返回控制台结果，如果未查询到则返回 "运行成功"
func (u *HdGroup) RunCmd(cmd string) (string, error) {
	if u.Status != 2 && u.Status != 3 {
		return fmt.Sprintf("实例: %s 未运行!", u.Name), nil
	}
	if cmd == "" {
		return "运行命令为空!", nil
	}
	client := &http.Client{}
	r2, _ := http.NewRequest("GET", u.Url+"/api/protected_instance/command", nil)
	r2.Close = true
	q := r2.URL.Query()
	q.Add("apikey", u.Apikey)
	q.Add("uuid", u.Uuid)
	q.Add("remote_uuid", u.Remote_uuid)
	q.Add("command", cmd)
	r2.URL.RawQuery = q.Encode()
	r2.Header.Set("x-requested-with", "xmlhttprequest")
	r, err := client.Do(r2)
	if err != nil {
		colorlog.Errorf("运行命令 %s 失败！%v", cmd, err)
		return fmt.Sprintf("运行命令 %s 失败！", cmd), err
	}
	r.Body.Close()
	time.Sleep(500 * time.Millisecond)
	return u.returnResult(cmd, 2)
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
		colorlog.Errorf("获取实例 %s 命令 %s 运行结果失败！", u.Name, command)
		return u.returnResult(command, try-1)
	}
	defer r.Body.Close()
	var data Data
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		return u.returnResult(command, try-1)
	}
	d, err := io.ReadAll(colorable.NewNonColorableReader(strings.NewReader(data.Data)))
	if err != nil {
		return u.returnResult(command, try-1)
	}
	dataString := string(d)
	colorlog.Debugf("终端信息: %s", dataString)
	if index := strings.LastIndex(dataString, command+"\r\n"); index != -1 {
		return fmt.Sprintf("[%s]\n%s", u.Name, strings.TrimRight(dataString, "\r\n")[index:]), nil
	} else if index = strings.LastIndex(dataString, command+"\r\r\n"); index != -1 {
		return fmt.Sprintf("[%s]\n%s", u.Name, strings.TrimRight(dataString, "\r\n")[index:]), nil
	}
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
		colorlog.Warningf("实例: %s 运行重启命令失败,可能是网络问题!", u.Name)
		return fmt.Sprintf("实例: %s 运行重启命令失败,可能是网络问题!", u.Name), err
	}
	return fmt.Sprintf("实例: %s 重启中!", u.Name), nil
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
		colorlog.Warningf("实例: %s 运行终止命令失败,可能是网络问题!", u.Name)
		return fmt.Sprintf("实例: %s 运行终止命令失败,可能是网络问题!", u.Name), err
	}
	return fmt.Sprintf("实例: %s 已经终止!", u.Name), nil
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
	t := time.Now()
	r, err := client.Do(r2)
	sub := time.Since(t).Milliseconds()
	if err != nil {
		colorlog.Errorf("获取实例Id: %d 信息失败! err: %v", u.Id, err)
		return err
	}
	b, _ := io.ReadAll(r.Body)
	if b == nil {
		return errors.New("get instance info error: body is nil")
	}
	var status InstanceConfig
	json.Unmarshal(b, &status)
	if status.Data.Config.Nickname != "" {
		u.lock.Lock()
		u.performance = (u.performance + sub) / 2
		u.Status = status.Data.Status
		u.Name = status.Data.Config.Nickname
		u.EndTime = status.Data.Config.EndTime
		u.ProcessType = status.Data.Config.ProcessType
		u.Pty = status.Data.Config.TerminalOption.Pty
		u.PingIp = status.Data.Config.PingConfig.PingIp
		u.CurrentPlayers = status.Data.Info.CurrentPlayers
		u.MaxPlayers = status.Data.Info.MaxPlayers
		u.Version = status.Data.Info.Version
		u.lock.Unlock()
	} else {
		return errors.New("get instance info error: instance name is nil")
	}
	return nil
}

func (u *HdGroup) GetStatus() string {
	u.lock.RLock()
	defer u.lock.RUnlock()
	if u.Status == 2 || u.Status == 3 {
		if u.CurrentPlayers == "" {
			return fmt.Sprintf("实例: %s 状态查询 功能未开启!请前往面板实例中开启状态查询功能", u.Name)
		} else {
			return fmt.Sprintf("实例: %s 正在运行!\n人数: %s\n最大人数: %s\n版本: %s", u.Name, u.CurrentPlayers, u.MaxPlayers, u.Version)
		}
	} else {
		return fmt.Sprintf("实例: %s 未运行!", u.Name)
	}
}
