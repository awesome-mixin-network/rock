package memo

import (
	"encoding/base64"

	"github.com/vmihailenco/msgpack"
)

func Marshal(object interface{}) (string, error) {
	data, err := msgpack.Marshal(object)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(data), nil
}

func Unmarshal(memo string, object interface{}) error {
	data, err := base64.StdEncoding.DecodeString(memo)
	if err != nil {
		return err
	}

	return msgpack.Unmarshal(data, object)
}
