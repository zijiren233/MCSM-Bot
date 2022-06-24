package rwmessage

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gorilla/websocket"
	"github.com/zijiren233/MCSM-Bot/gconfig"
	"github.com/zijiren233/MCSM-Bot/logger"
)

var GOnlineMap = make(map[int]*HdGroup)
var POnlineMap = make(map[int]*HdCqOp)
var LogLevle uint
var Log = logger.Newlog(LogLevle)
var GroupToId = make(map[int]([]int))
var IdToOd = make(map[int]int)
var Mconfig = gconfig.GetMConfig()
var Qconfig = gconfig.GetQConfig()
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
	Ws          *websocket.Conn
	SendMessage chan *SendData
}

type MsgData struct {
	Post_type    string `json:"post_type"`
	Message_type string `json:"message_type"`
	Message_id   int    `json:"message_id"`
	User_id      int    `json:"user_id"`
	Group_id     int    `json:"group_id"`
	Message      string `json:"message"`
}

func NewServer(url string) *Server {
	w := Server{
		Url: url,
	}
	w.init()
	w.SendMessage = make(chan *SendData, 50)
	var err error
	w.Ws, _, err = websocket.DefaultDialer.Dial(w.Url, nil)
	if err != nil {
		fmt.Println("cqhttp 连接失败，等待重连...")
	}
	go w.Run()
	return &w
}

func (s *Server) init() {
	for k, v := range Mconfig.McsmData {
		IdToOd[v.Id] = k
	}
}

func (s *Server) Run() {
	go s.SendMsg()
	var data []byte
	var err error
	for {
		_, data, err = s.Ws.ReadMessage()
		if err != nil {
			logger.Log.Error("cqhttp 连接失败!")
			for i := 0; ; i++ {
				logger.Log.Error("cqhttp 第 %d 次重连", i)
				s.Ws, _, err = websocket.DefaultDialer.Dial(s.Url, nil)
				if err != nil {
					time.Sleep(5 * time.Second)
					continue
				}
				logger.Log.Info("cqhttp 重连成功!")
				break
			}
		}
		var msgdata MsgData
		json.Unmarshal(data, &msgdata)
		if msgdata.Post_type == "message" {
			s.BroadCast(&msgdata)
		}
	}
}

func (s *Server) BroadCast(msg *MsgData) {
	if msg.Message_type == "group" {
		for _, v := range GOnlineMap {
			select {
			case v.ChGroupMsg <- msg:
			default:
				logger.Log.Warring("ChGroupMsg 堵塞!会造成消息丢失!")
			}
		}
	} else if msg.Message_type == "private" {
		for _, v := range POnlineMap {
			select {
			case v.ChCqOpMsg <- msg:
			default:
				logger.Log.Warring("ChPrivatemsg 堵塞!会造成消息丢失!")
			}
		}
	}
}

func (s *Server) SendMsg() {
	var tmp []byte
	for {
		tmp, _ = json.Marshal(*<-s.SendMessage)
		s.Ws.WriteMessage(websocket.TextMessage, tmp)
	}
}

func GetAllId() []int {
	tmp := make([]int, 0, len(Mconfig.McsmData))
	for i := 0; i < len(Mconfig.McsmData); i++ {
		tmp = append(tmp, Mconfig.McsmData[i].Id)
	}
	return tmp
}

func IsListDuplicated(list []int) bool {
	tmpMap := make(map[int]int)
	for _, value := range list {
		tmpMap[value] += 1
	}
	for _, v := range tmpMap {
		if v > 1 {
			return true
		}
	}
	return false
}

func GetAllDaemon() *map[string]string {
	var tmplist = make(map[string]string)
	for i := 0; i < len(Mconfig.McsmData); i++ {
		tmplist[Mconfig.McsmData[i].Url] = Mconfig.McsmData[i].Apikey
	}
	return &tmplist
}

// func RemoveRep(list []string) []string {

// }
