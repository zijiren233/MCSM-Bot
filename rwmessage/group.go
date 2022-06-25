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

type HdGroup struct {
	Id             int
	Name           string
	Url            string
	Remote_uuid    string
	Uuid           string
	Apikey         string
	Group_id       int
	Adminlist      []int
	Status         int
	EndTime        string
	CurrentPlayers string
	MaxPlayers     string
	Version        string
	ChGroupMsg     chan *MsgData
	SendChan       chan *SendData
}

type Status struct {
	Data struct {
		Status int `json:"status"`
		Config struct {
			Nickname string `json:"nickname"`
			EndTime  string `json:"endTime"`
		} `json:"config"`
		Info struct {
			CurrentPlayers string `json:"currentPlayers"`
			MaxPlayers     string `json:"maxPlayers"`
			Version        string `json:"version"`
		} `json:"info"`
	} `json:"data"`
}

func NewHdGroup(id int, send chan *SendData) *HdGroup {
	if !InInt(id, AllId) {
		fmt.Println("Id错误!")
		logger.Log.Error("监听Id: %d ,Id错误!", id)
		fmt.Println()
		return nil
	} else if _, ok := GOnlineMap[id]; ok {
		fmt.Printf("重复监听服务器: %s\n", GOnlineMap[id].Name)
		logger.Log.Warring("重复监听服务器: %s", GOnlineMap[id].Name)
		return nil
	}
	u := HdGroup{
		Id:          id,
		Url:         Mconfig.McsmData[IdToOd[id]].Url,
		Remote_uuid: Mconfig.McsmData[IdToOd[id]].Remote_uuid,
		Uuid:        Mconfig.McsmData[IdToOd[id]].Uuid,
		Apikey:      Mconfig.McsmData[IdToOd[id]].Apikey,
		Group_id:    Mconfig.McsmData[IdToOd[id]].Group_id,
		Adminlist:   Mconfig.McsmData[IdToOd[id]].Adminlist,
		SendChan:    send,
	}
	err := u.StatusTest()
	if err != nil {
		fmt.Printf("服务器Id: %d 监听失败!\n", u.Id)
		return nil
	}
	GroupToId[u.Group_id] = append(GroupToId[u.Group_id], u.Id)
	logger.Log.Debug("GroupToId: %v", GroupToId)
	u.ChGroupMsg = make(chan *MsgData, 25)
	go u.ReportStatus()
	u.Run()
	return &u
}

func (u *HdGroup) Run() {
	GOnlineMap[u.Id] = u
	fmt.Println("监听实例 ", u.Name, " 成功")
	logger.Log.Info("监听实例 %s 成功", u.Name)
	go u.HdMessage()
}

func (u *HdGroup) HdMessage() {
	var msg *MsgData
	for {
		msg = <-u.ChGroupMsg
		if InInt(msg.User_id, u.Adminlist) && msg.Group_id == u.Group_id {
			go u.HandleMessage(msg)
		}
	}
}

func (u *HdGroup) HandleMessage(mdata *MsgData) {
	flysnowRegexp, _ := regexp.Compile(`^run ([0-9]*) *(.*)`)
	params := flysnowRegexp.FindString(mdata.Message)
	if len(params) == 0 {
		return
	}
	params2 := flysnowRegexp.FindStringSubmatch(params)
	// 当一个群有两个实例监听时
	if len(GroupToId[u.Group_id]) >= 2 {
		// 则由第一个监听的实例执行
		if GroupToId[u.Group_id][0] != u.Id {
			return
		}
		if (params2)[2] == "" {
			u.Send_group_msg("请输入run help查看帮助!")
			return
		}
		if params2[1] == "" {
			u.checkCMD2(params2[2])
		} else {
			u.checkCMD1(params2[2])
		}
		return
	}
	// 如果指定了Id则只由指定Id的goroutine执行
	if len(GroupToId[u.Group_id]) >= 2 && params2[1] != strconv.Itoa(u.Id) {
		return
	}
	if (params2)[2] == "" {
		u.Send_group_msg("请输入run help查看帮助!")
		return
	}
	u.checkCMD1(params2[2])
}

// 一群一个实例
func (u *HdGroup) checkCMD1(params string) {
	params = strings.ReplaceAll(params, "\n", "")
	params = strings.ReplaceAll(params, "\r", "")
	if u.Status != 3 && u.Status != 2 && (params != "help" && params != "server" && params != "start") {
		u.Send_group_msg("服务器: %s 未启动!\n请先启动服务器:\nrun %d start", u.Name, u.Id)
		return
	}
	var msg string
	switch params {
	case "help":
		u.Send_group_msg("run status : 查看服务器状态\nrun start : 启动服务器\nrun stop : 关闭服务器\nrun restart : 重启服务器\nrun kill : 终止服务器\nrun 控制台命令 : 运行服务器命令")
	case "server":
		msg += "服务器列表:\n"
		msg = fmt.Sprintf("Name: %s    Id: %d", u.Name, u.Id)
		u.Send_group_msg(msg)
	case "status":
		u.SendStatus()
	case "start":
		if u.Status != 2 && u.Status != 3 {
			u.Start()
		} else {
			u.Send_group_msg("服务器: %s 已在运行!", u.Name)
		}
	case "stop":
		if u.Status == 2 || u.Status == 3 {
			u.Stop()
		} else {
			u.Send_group_msg("服务器: %s 未运行!", u.Name)
		}
	case "restart":
		u.Restart()
		u.Send_group_msg("服务器: %s 正在重启!", u.Name)
	case "kill":
		u.Kill()
	default:
		u.RunCmd(params)
	}
}

// 一群多个实例不指定ID
func (u *HdGroup) checkCMD2(params string) {
	params = strings.ReplaceAll(params, "\n", "")
	params = strings.ReplaceAll(params, "\r", "")
	if u.Status != 3 && u.Status != 2 && (params != "help" && params != "server" && params != "start") {
		u.Send_group_msg("服务器: %s 未启动!\n请先启动服务器:\nrun %d start", u.Name, u.Id)
		return
	}
	var msg string
	switch params {
	case "help":
		u.Send_group_msg("run server : 查看服务器列表\nrun status : 查看服务器状态\nrun id start : 启动服务器\nrun id stop : 关闭服务器\nrun id restart : 重启服务器\nrun id kill : 终止服务器\nrun id 控制台命令 : 运行服务器命令")
	case "server":
		msg += "服务器列表:\n"
		for _, v := range GroupToId[u.Group_id] {
			if GOnlineMap[v].Status == 2 || GOnlineMap[v].Status == 3 {
				msg += fmt.Sprintf("Id: %-5dName: %s    Status: RUN\n", GOnlineMap[v].Id, GOnlineMap[v].Name)
			} else {
				msg += fmt.Sprintf("Id: %-5dName: %s    Status: STOP\n", GOnlineMap[v].Id, GOnlineMap[v].Name)
			}
		}
		u.Send_group_msg(msg[:len(msg)-1])
	default:
		msg += "服务器列表:\n"
		for _, v := range GroupToId[u.Group_id] {
			if GOnlineMap[v].Status == 2 || GOnlineMap[v].Status == 3 {
				msg += fmt.Sprintf("Id: %-5dName: %s    Status: RUN\n", GOnlineMap[v].Id, GOnlineMap[v].Name)
			} else {
				msg += fmt.Sprintf("Id: %-5dName: %s    Status: STOP\n", GOnlineMap[v].Id, GOnlineMap[v].Name)
			}
		}
		msg += fmt.Sprintf("查询具体服务器请输入 run id %s", params)
		u.Send_group_msg(msg)
	}
}

func (u *HdGroup) SendStatus() {

	if u.Status == 2 || u.Status == 3 {
		if u.CurrentPlayers == "-1" {
			u.Send_group_msg("服务器: %s 正在运行!", u.Name)
		} else {
			u.Send_group_msg("服务器: %s 正在运行!\n服务器人数: %s\n服务器最大人数: %s\n服务器版本: %s", u.Name, u.CurrentPlayers, u.MaxPlayers, u.Version)
		}
	} else {
		u.Send_group_msg("服务器: %s 未运行!", u.Name)
	}
}

func (u *HdGroup) ReportStatus() {
	go func() {
		for {
			u.StatusTest()
			time.Sleep(3000 * time.Millisecond)
		}
	}()
	var status = u.Status
	for {
		if status != u.Status {
			if u.Status == 2 || u.Status == 3 {
				u.Send_group_msg("服务器: %s 已运行!", u.Name)
			} else if u.Status == 0 {
				u.Send_group_msg("服务器: %s 已停止!", u.Name)
			}
			status = u.Status
		}
		time.Sleep(1500 * time.Millisecond)
	}
}

func (u *HdGroup) StatusTest() error {
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
		logger.Log.Error("获取服务器Id: %d 信息失败! err: %v", u.Id, err)
		return err
	}
	b, _ := ioutil.ReadAll(r.Body)
	var status Status
	json.Unmarshal(b, &status)
	u.Status = status.Data.Status
	u.Name = status.Data.Config.Nickname
	u.EndTime = status.Data.Config.EndTime
	u.CurrentPlayers = status.Data.Info.CurrentPlayers
	u.MaxPlayers = status.Data.Info.MaxPlayers
	u.Version = status.Data.Info.Version
	return nil
}

func InInt(target int, str_array []int) bool {
	for _, element := range str_array {
		if target == element {
			return true
		}
	}
	return false
}

func InString(target string, str_array []string) bool {
	for _, element := range str_array {
		if target == element {
			return true
		}
	}
	return false
}

func (u *HdGroup) Send_group_msg(msg string, a ...interface{}) {
	var tmp SendData
	tmp.Action = "send_group_msg"
	tmp.Params.Group_id = u.Group_id
	tmp.Params.Message = fmt.Sprintf(msg, a...)
	u.SendChan <- &tmp
}
