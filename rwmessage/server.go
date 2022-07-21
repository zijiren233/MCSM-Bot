package rwmessage

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/zijiren233/MCSM-Bot/gconfig"
	"github.com/zijiren233/MCSM-Bot/logger"
	"github.com/zijiren233/MCSM-Bot/utils"
)

// 已监听群组 map[id](*HdGroup)
var GOnlineMap = make(map[int]*HdGroup)

// 已监听OP qq map[0](*HdGroup)
var POnlineMap = make(map[int]*Op)

// map[群号](id)
var GroupToId = make(map[int]([]int))

// map[id](config index)
var IdToOd = make(map[int]int)
var LogLevle uint
var Log = logger.Getlog()
var Mconfig = gconfig.Mconfig
var Qconfig = gconfig.Qconfig
var AllId = utils.GetAllId()

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
		fmt.Println("Cqhttp 连接失败,正在重连...")
		w.retrydial()
	}
	fmt.Printf("Cqhttp 连接成功!\n")
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
	Log.Error("cqhttp 连接失败!")
	for i := 0; ; i++ {
		Log.Error("cqhttp 第 %d 次重连", i)
		s.lock.Lock()
		s.ws, _, err = websocket.DefaultDialer.Dial(s.Url, nil)
		s.lock.Unlock()
		if err != nil {
			time.Sleep(5 * time.Second)
			continue
		}
		Log.Info("cqhttp 重连成功!")
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
		for _, v := range GOnlineMap {
			select {
			case v.ChGroupMsg <- msg:
			default:
				Log.Warring("ChGroupMsg 堵塞!会造成消息丢失!")
			}
		}
	} else if msg.Message_type == "private" {
		for _, v := range POnlineMap {
			select {
			case v.ChCqOpMsg <- msg:
			default:
				Log.Warring("ChPrivatemsg 堵塞!会造成消息丢失!")
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
			continue
		}
		s.lock.RLock()
		err = s.ws.WriteMessage(websocket.TextMessage, tmp)
		s.lock.RUnlock()
		if err != nil {
			continue
		}
	}
}
