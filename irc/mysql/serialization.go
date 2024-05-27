package mysql

import (
	"encoding/json"

	"github.com/ergochat/ergo/irc/history"
)

// 123 / '{' is the magic number that means JSON;
// if we want to do a binary encoding later, we just have to add different magic version numbers

func marshalItem(item *history.Item) (result []byte, err error) {
	return json.Marshal(item)
}

func unmarshalItem(data []byte, result *history.Item) (err error) {
	return json.Unmarshal(data, result)
}

// TODO: probably should convert the internal mysql column to uint
func decodeMsgid(msgid string) ([]byte, error) {
	return []byte(msgid), nil
}
