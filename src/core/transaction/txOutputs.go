package transaction

import (
	"bytes"
	"encoding/gob"
	"log"
)

// 输出集合
type TXOutputs struct {
	Outputs []TXOutput
}

// 序列化输出数组
func SerializeOutputs(outs TXOutputs) []byte {
	var buff bytes.Buffer
	enc := gob.NewEncoder(&buff)
	err := enc.Encode(outs)
	if err != nil {
		log.Panic(err)
	}

	return buff.Bytes()
}

// 反序列化输出
func DeserializeOutputs(data []byte) TXOutputs{
	var outputs TXOutputs
	dec := gob.NewDecoder(bytes.NewReader(data))
	err := dec.Decode(&outputs)
	if err != nil {
		log.Panic(err)
	}

	return outputs
}
