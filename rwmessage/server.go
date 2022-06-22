package rwmessage

import (
	"encoding/json"

	"github.com/gorilla/websocket"
	"github.com/zijiren233/MCSM-Bot/gconfig"
	"github.com/zijiren233/MCSM-Bot/logger"
)

var GOnlineMap = make(map[int]*HdGroup)
var POnlineMap = make(map[int]*HdPrivate)
var LogLevle uint
var Log = logger.Newlog(LogLevle)
var GroupToId = make(map[int]([]int))
var IdToOd = make(map[int]int)
var Mconfig = gconfig.GetMConfig()
var Qconfig = gconfig.GetQConfig()
var AllId = GetAllId()

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
	SendMessage chan SendData
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
	w.SendMessage = make(chan SendData, 25)
	w.Ws, _, _ = websocket.DefaultDialer.Dial(w.Url, nil)
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
	for {
		_, data, err := s.Ws.ReadMessage()
		if err != nil {
			continue
		}
		var msgdata MsgData
		json.Unmarshal(data, &msgdata)
		if msgdata.Post_type == "message" {
			s.BroadCast(msgdata)
		}
	}
}

func (s *Server) BroadCast(msg MsgData) {
	if msg.Message_type == "group" {
		for _, v := range GOnlineMap {
			select {
			case v.ChGroupMsg <- msg:
			default:
				Log.Warring("ChGroupMsg 堵塞！")
			}
		}
	} else if msg.Message_type == "private" {
		for _, v := range POnlineMap {
			select {
			case v.ChPrivateMsg <- msg:
			default:
				Log.Warring("ChPrivatemsg 堵塞！")
			}
		}
	}
}

func (s *Server) SendMsg() {
	for {
		data := <-s.SendMessage
		b, _ := json.Marshal(data)
		// fmt.Printf("b: %v\n", string(b))
		s.Ws.WriteMessage(websocket.TextMessage, b)
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

func RemoveRepByLoop(slc []int) []int {
	result := []int{} // 存放结果
	for i := range slc {
		flag := true
		for j := range result {
			if slc[i] == result[j] {
				flag = false // 存在重复元素，标识为false
				break
			}
		}
		if flag { // 标识为false，不添加进结果
			result = append(result, slc[i])
		}
	}
	return result
}
