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
	Id          int
	Name        string
	Url         string
	Remote_uuid string
	Uuid        string
	Apikey      string
	Group_id    int
	Adminlist   []int
	Status      int
	ChGroupMsg  chan MsgData
	SendChan    chan SendData
}

type Status struct {
	Data struct {
		Data []struct {
			Status int `json:"status"`
		} `json:"data"`
	} `json:"data"`
}

// var Statusmap = make(map[string]int)

func NewHdGroup(id int, send chan SendData) *HdGroup {
	if IdToOd[id] == 0 && Mconfig.McsmData[0].Id != id {
		Log.Error("输入的ID:%d 不存在!", id)
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
	u.ChGroupMsg = make(chan MsgData, 25)
	go u.ReportStatus()
	go u.Run()
	return &u
}

func (u *HdGroup) Run() {
	GOnlineMap[u.Id] = u
	fmt.Println("监听实例 ", u.Name, " 成功！")
	u.HdMessage()
}

func (u *HdGroup) ReportStatus() {
	if u.RunningTest() {
		u.Status = 1
	} else {
		u.Status = 0
	}
	for {
		if !u.RunningTest() && u.Status == 1 {
			u.Status = 0
			u.Send_group_msg("服务器:%s 已停止!", u.Name)
		} else if u.RunningTest() && u.Status == 0 {
			u.Status = 1
			u.Send_group_msg("服务器:%s 已运行!", u.Name)
		}
		time.Sleep(2 * time.Second)
	}
}

func (u *HdGroup) RunningTest() bool {
	client := &http.Client{}
	r2, _ := http.NewRequest("GET", u.Url+"/api/service/remote_service_instances", nil)
	r2.Close = true
	q := r2.URL.Query()
	q.Add("apikey", u.Apikey)
	q.Add("page", "1")
	q.Add("page_size", "1")
	q.Add("instance_name", u.Name)
	q.Add("remote_uuid", u.Remote_uuid)
	r2.URL.RawQuery = q.Encode()
	r2.Header.Set("x-requested-with", "xmlhttprequest")
	r, err := client.Do(r2)
	if err != nil {
		return u.Status-1 == 0
	}
	b, _ := ioutil.ReadAll(r.Body)
	var status Status
	json.Unmarshal(b, &status)
	if status.Data.Data[0].Status != 3 && status.Data.Data[0].Status != 2 {
		return false
	} else {
		return true
	}
}

func (u *HdGroup) HdMessage() {
	var tmp MsgData
	for {
		tmp = <-u.ChGroupMsg
		if In(tmp.User_id, u.Adminlist) && tmp.Group_id == u.Group_id {
			go u.HandleMessage(tmp)
		}
	}
}

func (u *HdGroup) HandleMessage(mdata MsgData) {
	flysnowRegexp, _ := regexp.Compile(`^run ([0-9]*) *(.*)`)
	params := flysnowRegexp.FindString(mdata.Message)
	// fmt.Printf("params: %v\n", params)
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
	go u.checkCMD(params2[2])
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
		u.Send_group_msg("服务器:%s 正在运行!", u.Name)
	} else if u.Status == 0 {
		u.Send_group_msg("服务器:%s 已停止!", u.Name)
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
	u.SendChan <- tmp
}
