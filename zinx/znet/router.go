package znet

import "zinx/ziface"

// 实现router时， 先嵌入Router基类， 然后根据需要对这个基类的进行重写
type BaseRouter struct {
	
}


// 具体的业务需求只需要继承BaseRouter， 实现自己的需求
func (b *BaseRouter) PreHandle(request ziface.IRequest)  {
	
}

func (b *BaseRouter) Handle(request ziface.IRequest)  {

}

func (b *BaseRouter) PostHandle(request ziface.IRequest)  {

}
