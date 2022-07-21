package rwmessage

import (
	"encoding/json"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/zijiren233/MCSM-Bot/gconfig"
	"github.com/zijiren233/MCSM-Bot/logger"
)

// 已监听群组 map[id](*HdGroup)
var GOnlineMap = make(map[int]*HdGroup)

// 已监听OP qq map[0](*HdGroup)
var POnlineMap = make(map[int]*Op)

// map[群号](id)
var GroupToId = make(map[int]([]int))

// map[id](config index)
var IdToOd = make(map[int]int)
var log = logger.GetLog()
var Mconfig = gconfig.Mconfig
var Qconfig = gconfig.Qconfig
var AllId = GetAllId()

// var AllDaemon = make(map[string]([]string))

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
		log.Warring("Cqhttp 连接失败,正在重连...")
		w.retrydial()
	}
	log.Info("Cqhttp 连接成功!")
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
	for {
		s.lock.RLock()
		_, data, err = s.ws.ReadMessage()
		s.lock.RUnlock()
		if err != nil {
			s.retrydial()
			continue
		}
		var msgdata MsgData
		json.Unmarshal(data, &msgdata)
		if msgdata.Post_type == "message" {
			s.broadCast(&msgdata)
		}
	}
}

func (s *Server) retrydial() {
	var err error
	log.Error("cqhttp 连接失败!")
	for i := 0; ; i++ {
		log.Error("cqhttp 第 %d 次重连", i)
		s.lock.Lock()
		s.ws, _, err = websocket.DefaultDialer.Dial(s.Url, nil)
		s.lock.Unlock()
		if err != nil {
			time.Sleep(5 * time.Second)
			continue
		}
		log.Info("cqhttp 重连成功!")
		return
	}
}

func (s *Server) broadCast(msg *MsgData) {
	re, _ := regexp.Compile(`^run ([0-9]*) *(.*)`)
	params := re.FindStringSubmatch(msg.Message)
	if len(params) == 0 {
		return
	}
	params[2] = strings.ReplaceAll(params[2], "\n", "")
	params[2] = strings.ReplaceAll(params[2], "\r", "")
	msg.Params = params
	if msg.Message_type == "group" {
		log.Info("获取到群组信息:Group_id:%d,User_id:%d,Nickname:%s,Message:%s", msg.Group_id, msg.User_id, msg.Sender.Nickname, msg.Message)
		for _, v := range GOnlineMap {
			select {
			case v.ChGroupMsg <- msg:
			default:
				log.Warring("ChGroupMsg 堵塞!会造成消息丢失!")
			}
		}
	} else if msg.Message_type == "private" {
		log.Info("获取到私聊信息:User_id:%d,Nickname:%s,Message:%s", msg.User_id, msg.Sender.Nickname, msg.Message)
		for _, v := range POnlineMap {
			select {
			case v.ChCqOpMsg <- msg:
			default:
				log.Warring("ChPrivatemsg 堵塞!会造成消息丢失!")
			}
		}
	}
}

func (s *Server) sendMsg() {
	var tmp []byte
	var err error
	for {
		tmp, err = json.Marshal(*<-s.SendMessage)
		if err != nil {
			log.Error("解析待发送的消息失败:%v", err)
			continue
		}
		s.lock.RLock()
		err = s.ws.WriteMessage(websocket.TextMessage, tmp)
		s.lock.RUnlock()
		if err != nil {
			log.Error("发送消息: %s 失败:%v", string(tmp), err)
			continue
		}
		log.Info("发送消息成功:%s", string(tmp))
	}
}

func GetAllDaemon() *map[string]string {
	var tmplist = make(map[string]string)
	for i := 0; i < len(Mconfig.McsmData); i++ {
		tmplist[Mconfig.McsmData[i].Url] = Mconfig.McsmData[i].Apikey
	}
	return &tmplist
}

func GetAllId() []int {
	tmp := make([]int, 0, len(Mconfig.McsmData))
	for i := 0; i < len(Mconfig.McsmData); i++ {
		tmp = append(tmp, Mconfig.McsmData[i].Id)
	}
	return tmp
}
