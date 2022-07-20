package utils

import (
	"bytes"
	"strings"

	"github.com/zijiren233/MCSM-Bot/gconfig"
)

var Mconfig = gconfig.Mconfig

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

func NoColorable(data *string) (*bytes.Buffer, error) {
	er := bytes.NewReader([]byte(*data))
	var plaintext bytes.Buffer
loop:
	for {
		c1, err := er.ReadByte()
		if err != nil {
			break loop
		}
		if c1 != 0x1b {
			plaintext.WriteByte(c1)
			continue
		}
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

func Handle_End_Newline(msg *string) *string {
	last := strings.LastIndex(*msg, "\n")
	if last == len(*msg)-2 {
		*msg = (*msg)[:last]
	}
	return msg
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
