package rwmessage

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/zijiren233/MCSM-Bot/utils"
)

type HdGroup struct {
	Id             int
	Name           string
	Url            string
	Remote_uuid    string
	Uuid           string
	Apikey         string
	Group_id       int
	UserCmd        []string
	Adminlist      []int
	Status         int
	EndTime        string
	CurrentPlayers string
	MaxPlayers     string
	Version        string
	ChGroupMsg     chan *MsgData
	SendChan       chan *SendData

	lock sync.RWMutex
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

func NewHdGroup(id int, serveSend chan *SendData) *HdGroup {
	u := HdGroup{
		Id:          id,
		Url:         Mconfig.McsmData[IdToOd[id]].Url,
		Remote_uuid: Mconfig.McsmData[IdToOd[id]].Remote_uuid,
		Uuid:        Mconfig.McsmData[IdToOd[id]].Uuid,
		Apikey:      Mconfig.McsmData[IdToOd[id]].Apikey,
		Group_id:    Mconfig.McsmData[IdToOd[id]].Group_id,
		UserCmd:     Mconfig.McsmData[IdToOd[id]].User_allows_commands,
		Adminlist:   Mconfig.McsmData[IdToOd[id]].Adminlist,
		SendChan:    serveSend,
	}
	err := u.statusTest()
	if err != nil {
		fmt.Printf("服务器Id: %d 监听失败!可能是 mcsm-web 端地址错误\n", u.Id)
	}
	GroupToId[u.Group_id] = append(GroupToId[u.Group_id], u.Id)
	Log.Debug("GroupToId: %v", GroupToId)
	u.ChGroupMsg = make(chan *MsgData, 25)
	return &u
}

func (u *HdGroup) Run() {
	GOnlineMap[u.Id] = u
	fmt.Println("监听实例 ", u.Name, " 成功")
	Log.Info("监听实例 %s 成功", u.Name)
	go u.reportStatus()
	u.hdChMessage()
}

func (u *HdGroup) hdChMessage() {
	var msg *MsgData
	for {
		msg = <-u.ChGroupMsg
		if msg.Group_id == u.Group_id {
			// 当一个群有两个实例监听,且没有指定ID,则由第一个监听的实例执行
			if len(GroupToId[u.Group_id]) >= 2 && msg.Params[1] == "" {
				if GroupToId[u.Group_id][0] != u.Id {
					continue
				}
				// 当一个群有两个实例监听,且指定ID,则由指定ID的实例执行
			} else if len(GroupToId[u.Group_id]) >= 2 && msg.Params[1] != "" {
				// 检测id是否监听此群
				id, _ := strconv.Atoi(msg.Params[1])
				if utils.InInt(id, GroupToId[u.Group_id]) {
					if msg.Params[1] != strconv.Itoa(u.Id) {
						continue
					}
				} else {
					if GroupToId[u.Group_id][0] != u.Id {
						continue
					}
					GOnlineMap[GroupToId[u.Group_id][0]].Send_group_msg("[CQ:at,qq=%d] ID 不存在", msg.User_id)
					GOnlineMap[GroupToId[u.Group_id][0]].checkCMD2("server")
					continue
				}
			}
			if msg.Params[2] == "" {
				u.Send_group_msg("命令为空!\n请输入run help查看帮助!")
				continue
			}
			if utils.InInt(msg.User_id, u.Adminlist) || utils.InString(msg.Params[2], u.UserCmd) {
				go u.handleMessage(msg)
			}
		}
	}
}

func (u *HdGroup) handleMessage(msg *MsgData) {
	if len(GroupToId[u.Group_id]) >= 2 && msg.Params[1] == "" {
		u.checkCMD2(msg.Params[2])
	} else {
		u.checkCMD1(msg.Params[2])
	}
}

func (u *HdGroup) checkCMD1(params string) {
	var msg string
	var err error
	u.lock.RLock()
	defer u.lock.RUnlock()
	if u.Status != 3 && u.Status != 2 && (params != "help" && params != "server" && params != "start") {
		u.Send_group_msg("服务器: %s 未启动!\n请先启动服务器:\nrun %d start", u.Name, u.Id)
		return
	}
	switch params {
	case "help":
		msg = "run status : 查看服务器状态\nrun start : 启动服务器\nrun stop : 关闭服务器\nrun restart : 重启服务器\nrun kill : 终止服务器\nrun 服务器命令 : 运行服务器命令"
		msg += "\n\n普通用户可用命令:\n"
		for _, v := range u.UserCmd {
			msg += "run " + v + "\n"
		}
		msg = *utils.Handle_End_Newline(&msg)
	case "server":
		msg += "服务器列表:\n"
		msg = fmt.Sprintf("Name: %s    Id: %d", u.Name, u.Id)
	case "status":
		msg = u.SendStatus()
	case "start":
		msg, err = u.Start()
	case "stop":
		msg, err = u.Stop()
	case "restart":
		msg, err = u.Restart()
	case "kill":
		msg, err = u.Kill()
	default:
		msg, err = u.RunCmd(params)
	}
	if err != nil {
		return
	}
	u.Send_group_msg(msg)
}

// 不指定ID
func (u *HdGroup) checkCMD2(params string) {
	u.lock.RLock()
	defer u.lock.RUnlock()
	var msg string
	switch params {
	case "help":
		u.Send_group_msg("run server : 查看服务器列表\nrun status : 查看服务器状态\nrun id start : 启动服务器\nrun id stop : 关闭服务器\nrun id restart : 重启服务器\nrun id kill : 终止服务器\nrun id 控制台命令 : 运行服务器命令")
	case "server":
		msg += "服务器列表:\n"
		for _, v := range GroupToId[u.Group_id] {
			if GOnlineMap[v].Status == 2 || GOnlineMap[v].Status == 3 {
				msg += fmt.Sprintf("Id: %-5dName: %s    Status: 运行中\n", GOnlineMap[v].Id, GOnlineMap[v].Name)
			} else {
				msg += fmt.Sprintf("Id: %-5dName: %s    Status: 已停止\n", GOnlineMap[v].Id, GOnlineMap[v].Name)
			}
		}
		u.Send_group_msg(msg[:len(msg)-1])
	default:
		msg += "服务器列表:\n"
		for _, v := range GroupToId[u.Group_id] {
			if GOnlineMap[v].Status == 2 || GOnlineMap[v].Status == 3 {
				msg += fmt.Sprintf("Id: %-5dName: %s    Status: 运行中\n", GOnlineMap[v].Id, GOnlineMap[v].Name)
			} else {
				msg += fmt.Sprintf("Id: %-5dName: %s    Status: 已停止\n", GOnlineMap[v].Id, GOnlineMap[v].Name)
			}
		}
		msg += fmt.Sprintf("查询具体服务器请输入 run id %s", params)
		u.Send_group_msg(msg)
	}
}

func (u *HdGroup) SendStatus() string {
	u.lock.RLock()
	defer u.lock.RUnlock()
	if u.Status == 2 || u.Status == 3 {
		if u.CurrentPlayers == "-1" {
			return fmt.Sprintf("服务器: %s 正在运行!", u.Name)
		} else {
			return fmt.Sprintf("服务器: %s 正在运行!\n服务器人数: %s\n服务器最大人数: %s\n服务器版本: %s", u.Name, u.CurrentPlayers, u.MaxPlayers, u.Version)
		}
	} else {
		return fmt.Sprintf("服务器: %s 未运行!", u.Name)
	}
}

func (u *HdGroup) reportStatus() {
	go func() {
		for {
			u.statusTest()
			time.Sleep(3000 * time.Millisecond)
		}
	}()
	var status = u.Status
	for {
		u.lock.RLock()
		if status != u.Status {
			if (u.Status == 2 && status != 3) || (u.Status == 3 && status != 2) {
				u.Send_group_msg("服务器: %s 已运行!", u.Name)
			} else if u.Status == 0 {
				u.Send_group_msg("服务器: %s 已停止!", u.Name)
			}
			status = u.Status
		}
		u.lock.RUnlock()
		time.Sleep(1500 * time.Millisecond)
	}
}

func (u *HdGroup) statusTest() error {
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
		Log.Error("获取服务器Id: %d 信息失败! err: %v", u.Id, err)
		return err
	}
	b, _ := ioutil.ReadAll(r.Body)
	var status Status
	json.Unmarshal(b, &status)
	// Log.Debug("status: %v", status)
	if u.Status != status.Data.Status {
		u.lock.Lock()
		u.Status = status.Data.Status
		u.lock.Unlock()
	}
	if u.Name != status.Data.Config.Nickname {
		u.lock.Lock()
		u.Name = status.Data.Config.Nickname
		u.lock.Unlock()
	}
	if u.EndTime != status.Data.Config.EndTime {
		u.lock.Lock()
		u.EndTime = status.Data.Config.EndTime
		u.lock.Unlock()
	}
	if u.CurrentPlayers != status.Data.Info.CurrentPlayers {
		u.lock.Lock()
		u.CurrentPlayers = status.Data.Info.CurrentPlayers
		u.lock.Unlock()
	}
	if u.MaxPlayers != status.Data.Info.MaxPlayers {
		u.lock.Lock()
		u.MaxPlayers = status.Data.Info.MaxPlayers
		u.lock.Unlock()
	}
	if u.Version != status.Data.Info.Version {
		u.lock.Lock()
		u.Version = status.Data.Info.Version
		u.lock.Unlock()
	}
	return nil
}

func (u *HdGroup) Send_group_msg(msg string, a ...interface{}) {
	var tmp SendData
	tmp.Action = "send_group_msg"
	tmp.Params.Group_id = u.Group_id
	tmp.Params.Message = fmt.Sprintf(msg, a...)
	u.SendChan <- &tmp
}
