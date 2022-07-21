package utils

import (
	"bytes"
	"fmt"
	"os"
	"strings"
)

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

func WaitExit() {
	var chose string
	for {
		fmt.Scan(&chose)
		switch chose {
		case "exit":
			os.Exit(0)
		case "stop":
			os.Exit(0)
		default:
		}
	}
}

func FileExists(path string) bool {
	_, err := os.Stat(path) //os.Stat获取文件信息
	if err != nil {
		return os.IsExist(err)
	}
	return true
}
