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
	CurrentPlayers int
	MaxPlayers     string
	Version        string
	ChGroupMsg     chan *MsgData
	SendChan       chan *SendData
}

type Status struct {
	Data struct {
		Status int `json:"status"`
		Config struct {
			EndTime string `json:"endTime"`
		} `json:"config"`
		Info struct {
			CurrentPlayers int    `json:"currentPlayers"`
			MaxPlayers     string `json:"maxPlayers"`
			Version        string `json:"version"`
		} `json:"info"`
	} `json:"data"`
}

func NewHdGroup(id int, send chan *SendData) *HdGroup {
	if !In(id, AllId) {
		fmt.Println("Id错误!")
		Log.Error("监听Id:%d ,Id错误!", id)
		fmt.Println()
		return nil
	} else if _, ok := GOnlineMap[id]; ok {
		fmt.Printf("重复监听服务器:%s\n", Mconfig.McsmData[IdToOd[id]].Name)
		Log.Warring("重复监听服务器:%s", Mconfig.McsmData[IdToOd[id]].Name)
		return nil
	}
	u := HdGroup{
		Id:          id,
		Name:        Mconfig.McsmData[IdToOd[id]].Name,
		Url:         Mconfig.McsmData[IdToOd[id]].Url,
		Remote_uuid: Mconfig.McsmData[IdToOd[id]].Remote_uuid,
		Uuid:        Mconfig.McsmData[IdToOd[id]].Uuid,
		Apikey:      Mconfig.McsmData[IdToOd[id]].Apikey,
		Group_id:    Mconfig.McsmData[IdToOd[id]].Group_id,
		Adminlist:   Mconfig.McsmData[IdToOd[id]].Adminlist,
		SendChan:    send,
	}
	GroupToId[u.Group_id] = append(GroupToId[u.Group_id], u.Id)
	Log.Debug("GroupToId: %v", GroupToId)
	u.ChGroupMsg = make(chan *MsgData, 25)
	go u.ReportStatus()
	go u.Run()
	return &u
}

func (u *HdGroup) Run() {
	GOnlineMap[u.Id] = u
	fmt.Println("监听实例 ", u.Name, " 成功！")
	u.HdMessage()
}

func (u *HdGroup) HdMessage() {
	var msg *MsgData
	for {
		msg = <-u.ChGroupMsg
		if In(msg.User_id, u.Adminlist) && msg.Group_id == u.Group_id {
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
	if len(GroupToId[u.Group_id]) >= 2 && params2[1] == "" {
		if GroupToId[u.Group_id][0] != u.Id {
			return
		}
	} else if len(GroupToId[u.Group_id]) >= 2 && params2[1] != strconv.Itoa(u.Id) {
		return
	}
	if (params2)[2] == "" {
		return
	}
	u.checkCMD(params2[2])
}

func (u *HdGroup) checkCMD(params string) {
	params = strings.ReplaceAll(params, "\n", "")
	params = strings.ReplaceAll(params, "\r", "")
	switch params {
	case "status":
		u.SendStatus()
	case "start":
		if u.Status == 0 {
			u.Start()
		} else {
			u.Send_group_msg("服务器:%s 正在运行!", u.Name)
		}
	case "stop":
		if u.Status == 1 {
			u.Stop()
		} else {
			u.Send_group_msg("服务器:%s 未运行!", u.Name)
		}
	case "restart":
		u.Restart()
	case "kill":
		u.Kill()
	default:
		u.RunCmd(params)
	}
}

func (u *HdGroup) SendStatus() {
	if u.Status == 1 {
		if u.CurrentPlayers == -1 {
			u.Send_group_msg("服务器:%s 正在运行!", u.Name)
		} else {
			u.Send_group_msg("服务器:%s 正在运行!\n服务器人数:%d\n服务器最大人数:%s\n服务器版本:%s", u.Name, u.CurrentPlayers, u.MaxPlayers, u.Version)
		}
	} else if u.Status == 0 {
		u.Send_group_msg("服务器:%s 未运行!", u.Name)
	}
}

func (u *HdGroup) ReportStatus() {
	if u.RunningTest() {
		u.Status = 1
	} else {
		u.Status = 0
	}
	var status bool
	for {
		status = u.RunningTest()
		if !status && u.Status == 1 {
			u.Status = 0
			u.Send_group_msg("服务器:%s 已停止!", u.Name)
		} else if status && u.Status == 0 {
			u.Status = 1
			u.Send_group_msg("服务器:%s 已运行!", u.Name)
		}
		time.Sleep(3 * time.Second)
	}
}

func (u *HdGroup) RunningTest() bool {
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
		return u.Status-1 == 0
	}
	b, _ := ioutil.ReadAll(r.Body)
	var status Status
	json.Unmarshal(b, &status)
	u.EndTime = status.Data.Config.EndTime
	u.CurrentPlayers = status.Data.Info.CurrentPlayers
	u.MaxPlayers = status.Data.Info.MaxPlayers
	u.Version = status.Data.Info.Version
	if status.Data.Status != 3 && status.Data.Status != 2 {
		return false
	} else {
		return true
	}
}

func In(target int, str_array []int) bool {
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
