package znet

import (
	"bytes"
	"encoding/binary"
	"errors"
	"zinx/utils"
	"zinx/ziface"
)

type DataPack struct {

}

func NewDataPack() *DataPack {
	return &DataPack{}
}

func (d *DataPack) GetDataHeader() uint32 {
	return 8
}


func (d *DataPack) Pack(message ziface.IMessage) ([]byte, error)  {

	databuff := bytes.NewBuffer([]byte{})

	// 将Msgdatalen 写入databuff
	err := binary.Write(databuff, binary.LittleEndian, message.GetMsgDataLen())
	if err != nil {
		return nil, err
	}

	// 将MsgID 写入databuff
	err = binary.Write(databuff, binary.LittleEndian, message.GetMsgID())
	if err != nil {
		return nil, err
	}

	// 将Msgdata 写入databuff
	err = binary.Write(databuff, binary.LittleEndian, message.GetMsgData())
	if err != nil {
		return nil, err
	}

	return databuff.Bytes(), nil
}


// 拆包方法（将包的header信息读出来， 之后再根据header里面的len读内容
func (d *DataPack) Unpack(data []byte) (ziface.IMessage, error) {
	// 创建io.reader, 从data里面读数据
	databuff := bytes.NewReader(data)

	// 解压header信息得到datalen， id
	msg := &Message{}

	// 读取datalen
	err := binary.Read(databuff, binary.LittleEndian, &msg.DataLen)
	if err != nil {
		return nil, err
	}

	// 读取msgID
	err = binary.Read(databuff, binary.LittleEndian, &msg.ID)
	if err != nil {
		return nil, err
	}

	// 判断datalen是否超出最大长度
	if utils.GlobalObject.MaxPackageSize > 0 && msg.DataLen > utils.GlobalObject.MaxPackageSize {
		return nil, errors.New("too large msg data recv")
	}
	return msg, nil
}


