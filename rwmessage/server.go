package rwmessage

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/zijiren233/MCSM-Bot/gconfig"
	"github.com/zijiren233/MCSM-Bot/utils"
	"github.com/zijiren233/go-colorlog"
)

var (
	// 已监听群组 map[id](*HdGroup)
	GOnlineMap = make(map[int]*HdGroup)

	// 已监听admin *admin
	pAdmin *admin

	// map[群号](id)
	GroupToId = make(map[int]([]int))

	// map[id](config index)
	IdToOd  = make(map[int]int)
	Mconfig = gconfig.Mconfig
	Qconfig = gconfig.Qconfig
	AllId   = gconfig.GetAllId()
)

type SendData struct {
	Action string `json:"action"`
	Params struct {
		Group_id int    `json:"group_id"`
		User_id  int    `json:"user_id"`
		Message  string `json:"message"`
	} `json:"params"`
}

type Server struct {
	Url         string
	SendMessage chan *SendData

	ws   *websocket.Conn
	lock sync.RWMutex
}

type MsgData struct {
	Post_type    string `json:"post_type"`
	Message_type string `json:"message_type"`
	Message_id   int    `json:"message_id"`
	User_id      int    `json:"user_id"`
	Group_id     int    `json:"group_id"`
	Message      string `json:"message"`
	Sender       struct {
		Nickname string `json:"nickname"`
	} `json:"sender"`

	Params []string
}

func NewServer(url string) *Server {
	w := Server{
		Url: url,
	}
	w.init()
	w.SendMessage = make(chan *SendData, 50)
	var err error
	w.ws, _, err = websocket.DefaultDialer.Dial(w.Url, nil)
	if err != nil {
		colorlog.Warningf("Cqhttp 连接失败,正在重连...")
		w.retrydial()
	}
	colorlog.Infof("Cqhttp 连接成功!")
	return &w
}

func (s *Server) init() {
	for k, v := range Mconfig.McsmData {
		IdToOd[v.Id] = k
	}
}

func (s *Server) Run() {
	go s.sendMsg()
	var data []byte
	var err error
	re, _ := regexp.Compile(`^run ?([0-9\*]*) ?(.*)`)
	for {
		s.lock.RLock()
		_, data, err = s.ws.ReadMessage()
		s.lock.RUnlock()
		if err != nil {
			s.retrydial()
			colorlog.Infof("Cqhttp 连接成功!")
			continue
		}
		var msgdata MsgData
		json.Unmarshal(data, &msgdata)
		if msgdata.Post_type == "message" {
			params := re.FindStringSubmatch(msgdata.Message)
			if len(params) == 0 {
				continue
			}
			if msgdata.Message_type == "group" {
				colorlog.Infof("获取到群组信息:Group_id:%d,User_id:%d,Nickname:%s,Message:%s", msgdata.Group_id, msgdata.User_id, msgdata.Sender.Nickname, msgdata.Message)
			} else if msgdata.Message_type == "private" {
				colorlog.Infof("获取到私聊信息:User_id:%d,Nickname:%s,Message:%s", msgdata.User_id, msgdata.Sender.Nickname, msgdata.Message)
			}
			params[2] = strings.ReplaceAll(params[2], "\n", "")
			params[2] = strings.ReplaceAll(params[2], "\r", "")
			msgdata.Params = params
			go s.broadCast(&msgdata)
		}
	}
}

func (s *Server) send_group_msg(group_id int, msg string, a ...interface{}) {
	var tmp SendData
	tmp.Action = "send_group_msg"
	tmp.Params.Group_id = group_id
	tmp.Params.Message = fmt.Sprintf(msg, a...)
	s.SendMessage <- &tmp
}

func (s *Server) send_private_msg(user_id int, msg string, a ...interface{}) {
	var tmp SendData
	tmp.Action = "send_private_msg"
	tmp.Params.User_id = user_id
	tmp.Params.Message = fmt.Sprintf(msg, a...)
	s.SendMessage <- &tmp
}

func help(msgdata *MsgData) string {
	var msg string
	switch msgdata.Params[2] {
	case "help":
		msg = "run list : 查看实例列表\nrun status : 查看实例状态\nrun id start : 启动实例\nrun id stop : 关闭实例\nrun id restart : 重启实例\nrun id kill : 终止实例\nrun id 控制台命令 : 运行实例命令"
		msg += "\n\n普通用户可用命令:\n请输入 run id help 查询"
		utils.Handle_End_Newline(&msg)
	default:
		msg += "实例列表:\n"
		for _, v := range GroupToId[msgdata.Group_id] {
			if GOnlineMap[v].Status == 2 || GOnlineMap[v].Status == 3 {
				msg += fmt.Sprintf("%s (id:%d) | 运行中\n", GOnlineMap[v].Name, GOnlineMap[v].Id)
			} else {
				msg += fmt.Sprintf("%s (id:%d) | 已停止\n", GOnlineMap[v].Name, GOnlineMap[v].Id)
			}
		}
		msg += fmt.Sprintf("查询具体实例请输入 run id %s", msgdata.Params[2])
	}
	return fmt.Sprintf("[CQ:reply,id=%d]%s", msgdata.Message_id, msg)
}

func (s *Server) retrydial() {
	var err error
	colorlog.Errorf("cqhttp 连接失败!")
	var ws *websocket.Conn
	for i := 1; ; i++ {
		colorlog.Errorf("cqhttp 第 %d 次重连", i)
		s.lock.Lock()
		ws, _, err = websocket.DefaultDialer.Dial(s.Url, nil)
		s.lock.Unlock()
		if err != nil {
			time.Sleep(5 * time.Second)
			continue
		}
		s.ws = ws
		return
	}
}

func (s *Server) broadCast(msg *MsgData) {
	if msg.Message_type == "group" {
		if idList, ok := GroupToId[msg.Group_id]; ok {
			if msg.Params[1] == "*" && msg.Params[2] == "help" {
				return
			} else if msg.Params[1] == "*" && msg.Params[2] != "help" {
				for _, id := range idList {
					if GOnlineMap[id].isAdmin(msg.User_id) {
						s.send_group_msg(msg.Group_id, GOnlineMap[id].runCMD(msg))
					} else {
						continue
					}
				}
				return
			} else if msg.Params[1] == "" && len(idList) >= 2 {
				s.send_group_msg(msg.Group_id, help(msg))
				return
			} else if msg.Params[1] == "" && len(idList) == 1 {
				GOnlineMap[idList[0]].ChGroupMsg <- msg
				return
			} else {
				id, err := strconv.Atoi(msg.Params[1])
				if err != nil {
					colorlog.Errorf("接收 id 失败: %v", err)
					return
				}
				if !utils.InInt(id, AllId) {
					colorlog.Warningf("接收的 id: %d 不存在!", id)
					return
				}
				GOnlineMap[id].ChGroupMsg <- msg
			}
		}
		return
	} else if msg.Message_type == "private" && pAdmin != nil {
		pAdmin.ChCqOpMsg <- msg
	}
}

func (s *Server) sendMsg() {
	var tmp []byte
	var err error
	var data *SendData
	var index int
	for {
		data = <-s.SendMessage
		if len(data.Params.Message) >= 5000 {
			colorlog.Warningf("消息过长,将采用分段发送...")
			s.fragmentSend(data)
			continue
		}
		tmp, err = json.Marshal(*data)
		if err != nil {
			colorlog.Errorf("解析待发送的消息失败:%v", err)
			continue
		}
		s.lock.RLock()
		err = s.ws.WriteMessage(websocket.TextMessage, tmp)
		s.lock.RUnlock()
		if err != nil {
			colorlog.Errorf("发送消息: %s 失败:%v", string(tmp), err)
			continue
		}
		if len(string(tmp)) <= 200 {
			colorlog.Infof("发送消息:%s ...", string(tmp))
		} else {
			index = strings.LastIndex(string(tmp)[:200], "\n")
			if index > 0 {
				colorlog.Infof("发送消息:%s ...", string(tmp)[:strings.LastIndex(string(tmp)[:200], "\n")])
			} else {
				colorlog.Infof("发送消息:%s ...", string(tmp)[:200])
			}
		}
	}
}

func (s *Server) fragmentSend(data *SendData) {
	new := strings.LastIndex(data.Params.Message[:4000], "\n")
	if new != -1 {
		newdata := *data
		newdata.Params.Message = data.Params.Message[:new]
		time.Sleep(time.Second)
		s.SendMessage <- &newdata
		data.Params.Message = data.Params.Message[new:]
		time.Sleep(time.Second)
		s.SendMessage <- data
	} else {
		newdata := *data
		newdata.Params.Message = data.Params.Message[:4000]
		time.Sleep(time.Second)
		s.SendMessage <- &newdata
		data.Params.Message = data.Params.Message[4000:]
		time.Sleep(time.Second)
		s.SendMessage <- data
	}
}
