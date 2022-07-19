package rwmessage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

type Data struct {
	Data string `json:"data"`
}

type CmdData struct {
	Time_unix int64 `json:"time"`
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
	r, err := client.Do(r2)
	if err != nil {
		// u.Send_group_msg("运行命令 %s 失败！", commd)
		Log.Error("运行命令 %s 失败！%v", commd, err)
		return fmt.Sprintf("运行命令 %s 失败！", commd), err
	}
	b, _ := ioutil.ReadAll(r.Body)
	var time_unix CmdData
	json.Unmarshal(b, &time_unix)
	time.Sleep(150 * time.Millisecond)
	return u.ReturnResult(commd, time_unix.Time_unix)
}

func (u *HdGroup) ReturnResult(command string, time_now int64) (string, error) {
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
	b2, _ := nocolorable(&data.Data)
	var index int
	// var i int64
	// Log.Debug("服务器 %s 运行命令 %s 返回时间: %s", u.Name, command, time.Unix((time_now/1000)+i, 0).Format("15:04:05"))
	// for i = 0; i < 2; i++ {
	// 	index = strings.Index(b2.String(), time.Unix((time_now/1000)+i, 0).Format("15:04:05"))
	// 	if index == -1 {
	// 		continue
	// 	}
	// 	return fmt.Sprintf("> [%s] %s\n%s", u.Name, command, *(handle_End_Newline(b2.String()[index-1:]))), nil
	// }
	// Log.Warring("服务器 %s 命令 %s 成功,但未查找到返回时间: %s", u.Name, command, time.Unix((time_now/1000), 0).Format("15:04:05"))
	index = strings.LastIndex(b2.String(), command+"\r\n")
	if index == -1 {
		return "运行命令成功!", nil
	}
	return fmt.Sprintf("[%s] %s", u.Name, *(handle_End_Newline(b2.String()[index:]))), nil
}

func handle_End_Newline(msg string) *string {
	last := strings.LastIndex(msg, "\r\n")
	if last == len(msg)-3 {
		msg = (msg)[:last]
	}
	return &msg
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

func nocolorable(data *string) (*bytes.Buffer, error) {
	er := bytes.NewReader([]byte(*data))
	var plaintext bytes.Buffer
loop:
	for {
		c1, err := er.ReadByte()
		if err != nil {
			// plaintext.WriteTo(w.out)
			break loop
		}
		if c1 != 0x1b {
			plaintext.WriteByte(c1)
			continue
		}
		// _, err = plaintext.WriteTo(w.out)
		// if err != nil {
		// 	break loop
		// }
		c2, err := er.ReadByte()
		if err != nil {
			break loop
		}
		if c2 != 0x5b {
			continue
		}

		for {
			c, err := er.ReadByte()
			if err != nil {
				break loop
			}
			if ('a' <= c && c <= 'z') || ('A' <= c && c <= 'Z') || c == '@' {
				break
			}
		}
	}
	return &plaintext, nil
}
