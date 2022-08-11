package ziface

type IDataPack interface {

	Unpack([]byte) (IMessage, error)
	Pack(IMessage) ([]byte, error)
	GetDataHeader() uint32

}
