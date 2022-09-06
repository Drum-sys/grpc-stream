package ziface

/*

	路由抽象接口
	路由里面的数据都是irequest
*/

type IRouter interface {

	// 处理conn业务之前的钩子方法Hook
	PreHandle(request IRequest)

	// 处理conn业务方法hook
	Handle(request IRequest)

	// 处理conn业务之前的钩子方法Hook
	PostHandle(request IRequest)
}
