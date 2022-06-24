package server

import (
	"bufio"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/smallnest/rpcx/log"
	"github.com/smallnest/rpcx/protocol"
	"github.com/smallnest/rpcx/share"
	"github.com/soheilhy/cmux"
	"net/http"

	"runtime"
	"strconv"

	"reflect"

	"net"
	"sync"
	"sync/atomic"
	"time"

	"io"
)

// ErrServerClosed is returned by the Server's Serve, ListenAndServe after a call to Shutdown or Close.
var (
	ErrServerClosed  = errors.New("http: Server closed")
	ErrReqReachLimit = errors.New("request reached rate limit")
)

const (
	// ReaderBuffsize is used for bufio reader.
	ReaderBuffsize = 1024
	// WriterBuffsize is used for bufio writer.
	WriterBuffsize = 1024

	// WriteChanSize is used for response.
	WriteChanSize = 1024 * 1024
)

type contextKey struct {
	name string
}

var (
	// RemoteConnContextKey is a context key. It can be used in
	// services with context.WithValue to access the connection arrived on.
	// The associated value will be of type net.Conn.
	RemoteConnContextKey = &contextKey{"remote-conn"}
	// StartRequestContextKey records the start time
	StartRequestContextKey = &contextKey{"start-parse-request"}
	// StartSendRequestContextKey records the start time
	StartSendRequestContextKey = &contextKey{"start-send-request"}
	// TagContextKey is used to record extra info in handling services. Its value is a map[string]interface{}
	TagContextKey = &contextKey{"service-tag"}
	// HttpConnContextKey is used to store http connection.
	HttpConnContextKey = &contextKey{"http-conn"}
)

type Handler func(ctx *Context) error

type Server struct {
	Plugins PluginContainer
	// AuthFunc can be used to auth.
	AuthFunc func(ctx context.Context, req *protocol.Message, token string) error // 认证服务
	options  map[string]interface{}                                               //配置文件，字典形式

	mu         sync.RWMutex //读写锁，保护activeConn的并发读写
	In         net.Listener //监听器，用来监听服务端的端口
	activeConn map[net.Conn]struct{}
	doneChan   chan struct{} //关闭连接的管道

	// 更加满意的关闭方式
	onShutdown []func(s *Server)
	inShutdown int32
	onRestart  []func(s *Server)

	// TLSConfig for creating tls tcp connection.
	// TLS 传输安全性，类似于SSL
	tlsConfig *tls.Config

	readTimeout  time.Duration // 读取client请求数据包的超时时间
	writeTimeout time.Duration // 给client写响应数据包的超时时间

	AsyncWrite         bool
	DisableHTTPGateway bool // should disable http invoke or not.
	DisableJSONRPC     bool // should disable json rpc or not.
	gatewayHTTPServer  *http.Server

	handlerMsgNum int32

	router       map[string]Handler
	serviceMapMu sync.RWMutex
	serviceMap   map[string]*service

	HandleServiceError func(error)
}

func NewServer(options ...OptionFn) *Server {
	s := &Server{
		options:    make(map[string]interface{}),
		Plugins:    &pluginContainer{},
		activeConn: make(map[net.Conn]struct{}),
		doneChan:   make(chan struct{}),
		serviceMap: make(map[string]*service),
		router:     make(map[string]Handler),
		AsyncWrite: false,
	}

	for _, op := range options {
		op(s)
	}

	if s.options["TCPKeepAlivePeriod"] == nil {
		s.options["TCPKeepAlivePeriod"] = 3 * time.Minute
	}
	return s
}

func (s *Server) Address() net.Addr {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.In == nil {
		return nil
	}
	return s.In.Addr()
}

// Close immediately
func (s *Server) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var err error
	if s.In != nil {
		err = s.In.Close() //如果没有断开连接返回错误
	}

	//获取连接map，手动断开连接，从map中删除
	for c := range s.activeConn {
		c.Close()
		delete(s.activeConn, c)
		s.Plugins.DoPostConnClose(c)
	}
	s.closeDoneChanLocked()
	return err
}

func (s *Server) closeDoneChanLocked() {
	select {
	case <-s.doneChan:
		// already closed, don't clsoe again
	default:
		close(s.doneChan)
	}
}

// RegisterOnShutdown registers a function to call on Shutdown.
// This can be used to gracefully shutdown connections.
func (s *Server) RegisterOnShutdown(f func(s *Server)) {
	s.mu.Lock()
	s.onShutdown = append(s.onShutdown, f)
	s.mu.Unlock()
}

// RegisterOnRestart registers a function to call on Restart.
func (s *Server) RegisterOnRestart(f func(s *Server)) {
	s.mu.Lock()
	s.onRestart = append(s.onRestart, f)
	s.mu.Unlock()
}

// Serve starts and listens RPC requests.
// It is blocked until receiving connections from clients.
func (s *Server) Serve(network, address string) (err error) {
	var In net.Listener
	In, err = s.makeListener(network, address)
	if err != nil {
		return err
	}

	//先只实现tcp
	/*if network == "http" {
		s.serveByHTTP(ln, "")
		return nil
	}

	if network == "ws" || network == "wss" {
		s.serveByWS(ln, "")
		return nil
	}*/
	In = s.startGateway(network, In)

	return s.serveListener(In)
}

// serveListener accepts incoming connections on the Listener ln,
// creating a new service goroutine for each.
// The service goroutines read requests and then call services to reply to them.
func (s *Server) serveListener(ln net.Listener) error {
	var tempDelay time.Duration

	//加锁防止破坏
	s.mu.Lock()
	s.In = ln
	s.mu.Unlock()

	for {
		conn, e := ln.Accept()
		if e != nil {
			if s.isShutdown() {
				<-s.doneChan
				return ErrServerClosed
			}
			if ne, ok := e.(net.Error); ok && ne.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}

				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}

				log.Errorf("rpcx: Accept error: %v; retrying in %v", e, tempDelay)
				time.Sleep(tempDelay)
				continue
			}

			if errors.Is(e, cmux.ErrListenerClosed) {
				return ErrServerClosed
			}
			return e
		}
		tempDelay = 0

		// 判断是不是tcp连接
		if tc, ok := conn.(*net.TCPConn); ok {
			period := s.options["TCPKeepAlivePeriod"]
			if period != nil {
				tc.SetKeepAlive(true)
				tc.SetKeepAlivePeriod(period.(time.Duration))
				tc.SetLinger(10)
			}
		}

		conn, ok := s.Plugins.DoPostConnAccept(conn)
		if !ok {
			conn.Close()
			continue
		}

		s.mu.Lock()
		s.activeConn[conn] = struct{}{}
		s.mu.Unlock()

		if share.Trace {
			log.Debugf("server accepted an conn: %v", conn.RemoteAddr().String())
		}

		// 循环，启动一个goroutine处理连接, 启动goroutine异步处理请求，继续开始下一次循环
		go s.serveConn(conn)
	}
}

// 处理 net.conn
func (s *Server) serveConn(conn net.Conn) {
	if s.isShutdown() {
		s.closeConn(conn)
		return
	}

	// 处理连接错误的情况
	defer func() {
		if err := recover(); err != nil {
			const size = 64 << 10
			buf := make([]byte, size)
			ss := runtime.Stack(buf, false)
			if ss > size {
				ss = size
			}
			buf = buf[:ss]
			log.Errorf("serving %s panic error: %s, stack:\n %s", conn.RemoteAddr(), err, buf)
		}

		if share.Trace {
			log.Debugf("server closed conn: %v", conn.RemoteAddr().String())
		}

		// make sure all inflight requests are handled and all drained
		if s.isShutdown() {
			<-s.doneChan
		}

		s.closeConn(conn)
	}()

	// 判断是不是TLs连接，出错如何处理
	/*if tlsConn, ok := conn.(*tls.Conn); ok {
		if d := s.readTimeout; d != 0 {
			conn.SetReadDeadline(time.Now().Add(d))
		}
		if d := s.writeTimeout; d != 0 {
			conn.SetWriteDeadline(time.Now().Add(d))
		}
		if err := tlsConn.Handshake(); err != nil {
			log.Errorf("rpcx: TLS handshake error from %s: %v", conn.RemoteAddr(), err)
			return
		}
	}*/

	// 服务器从客户端的连接中读取客户端发送来的数据
	// 申请空间，存放请求数据
	r := bufio.NewReaderSize(conn, ReaderBuffsize)

	// 是否同步写入
	var writeCh chan *[]byte
	/*if s.AsyncWrite {
		writeCh = make(chan *[]byte, 1)
		defer close(writeCh)
		go s.serveAsyncWrite(conn, writeCh)
	}*/
	for {
		if s.isShutdown() {
			return
		}

		t0 := time.Now()
		if s.readTimeout != 0 {
			conn.SetDeadline(t0.Add(s.readTimeout))
		}

		ctx := share.WithValue(context.Background(), RemoteConnContextKey, conn)

		// 处理客户端发送来的数据
		req, err := s.readRequest(ctx, r)

		// 判断错误类型， 处理错误，最后protocol.FreeMsg(req)
		if err != nil {
			if err == io.EOF {
				log.Infof("client has closed this connection: %s", conn.RemoteAddr().String())
			} else if errors.Is(err, net.ErrClosed) {
				log.Infof("rpcx: connection %s is closed", conn.RemoteAddr().String())
			} else if errors.Is(err, ErrReqReachLimit) {
				if !req.IsOneway() {
					res := req.Clone()
					res.SetMessageType(protocol.Response)

					handleError(res, err)

					// 返回数据到客户端
					s.sendResponse(ctx, conn, writeCh, err, req, res)
					protocol.FreeMsg(res)
				} else {
					s.Plugins.DoPreWriteResponse(ctx, req, nil, err)
				}
				protocol.FreeMsg(req)
				continue
			} else {
				log.Warnf("rpcx: failed to read request: %v", err)
			}

			protocol.FreeMsg(req)

			return
		}
		if share.Trace {
			log.Debugf("server received an request %+v from conn: %v", req, conn.RemoteAddr().String())
		}

		// 更新上下文
		ctx = share.WithLocalValue(ctx, StartRequestContextKey, time.Now().UnixNano())
		closeConn := false

		// message.go 消息编码过程中判断是否是核心消息
		if !req.IsHeartbeat() {

			// 认证客户端发送的消息
			err = s.auth(ctx, req)
			closeConn = err != nil
		}

		// 客户端认证失败，处理消息返回客户端
		if err != nil {
			if !req.IsOneway() {
				res := req.Clone()
				res.SetMessageType(protocol.Response)
				handleError(res, err)
				s.sendResponse(ctx, conn, writeCh, err, req, res)
				protocol.FreeMsg(res)
			} else {
				s.Plugins.DoPreWriteResponse(ctx, req, nil, err)
			}
			protocol.FreeMsg(req)
			// auth failed, closed the connection
			if closeConn {
				log.Infof("auth failed for conn %s: %v", conn.RemoteAddr().String(), err)
				return
			}
			continue
		}

		go func() {
			defer func() {
				if r := recover(); r != nil {
					log.Errorf("failed to handle request: %v", r)
				}
			}()
			atomic.AddInt32(&s.handlerMsgNum, 1)
			defer atomic.AddInt32(&s.handlerMsgNum, -1)

			if req.IsHeartbeat() {
				s.Plugins.DoHeartbeatRequest(ctx, req)
				req.SetMessageType(protocol.Response)
				data := req.EncodeSlicePointer()
				if s.AsyncWrite {
					writeCh <- data
				} else {
					if s.writeTimeout != 0 {
						conn.SetWriteDeadline(time.Now().Add(s.writeTimeout))
					}
					conn.Write(*data)
					protocol.PutData(data)
				}
				protocol.FreeMsg(req)
				return
			}

			// 回复客户端的信息
			resMetadata := make(map[string]string)
			if req.Metadata == nil {
				// req 接受客户端的消息
				req.Metadata = make(map[string]string)
			}
			ctx = share.WithLocalValue(share.WithLocalValue(ctx, share.ReqMetaDataKey, req.Metadata), share.ResMetaDataKey, resMetadata)

			cancelFunc := parseServerTimeout(ctx, req)
			if cancelFunc != nil {
				defer cancelFunc()
			}

			s.Plugins.DoPreHandleRequest(ctx, req)
			if share.Trace {
				log.Debugf("server handle request %+v from conn: %v", req, conn.RemoteAddr().String())
			}

			// 第一次使用handle
			if handler, ok := s.router[req.ServicePath+"."+req.ServiceMethod]; ok {
				sctx := NewContext(ctx, conn, req, writeCh)
				err := handler(sctx)
				if err != nil {
					log.Errorf("[handler internal error]: servicePath:%s, serviceMethod, err:%v", req.ServicePath, req.ServiceMethod, err)
				}
				protocol.FreeMsg(req)
				return
			}

			res, err := s.handleRequest(ctx, req)
			if err != nil {
				if s.HandleServiceError != nil {
					s.HandleServiceError(err)
				} else {
					log.Warnf("rpcx: failed to handle request: %v", err)
				}
			}

			if !req.IsOneway() {
				if len(resMetadata) > 0 { // copy meta in context to request
					meta := res.Metadata
					if meta == nil {
						res.Metadata = resMetadata
					} else {
						for k, v := range resMetadata {
							if meta[k] == "" {
								meta[k] = v
							}
						}
					}
				}

				s.sendResponse(ctx, conn, writeCh, err, req, res)
			}

			if share.Trace {
				log.Debugf("server write response %+v for an request %+v from conn: %v", res, req, conn.RemoteAddr().String())
			}

			protocol.FreeMsg(req)
			protocol.FreeMsg(res)
		}()
	}
}

func (s *Server) readRequest(ctx context.Context, r io.Reader) (req *protocol.Message, err error) {
	err = s.Plugins.DoPreReadRequest(ctx)
	if err != nil {
		return nil, err
	}

	// 建立空的 protol message结构体
	req = protocol.GetPooledMsg()

	// decode客户端发送来的数据到req， r为发送来的数据
	err = req.Decode(r)
	if err != io.EOF {
		return req, err
	}
	perr := s.Plugins.DoPostReadRequest(ctx, req, err)
	if err == nil {
		err = perr
	}
	return req, err
}

// 认证客户端发---对应于AuthFun
func (s *Server) auth(ctx context.Context, req *protocol.Message) error {
	if s.AuthFunc != nil {
		token := req.Metadata[share.AuthKey]
		return s.AuthFunc(ctx, req, token)
	}

	return nil
}

func (s *Server) isShutdown() bool {
	return atomic.LoadInt32(&s.inShutdown) == 1
}

// 关闭连接
func (s *Server) closeConn(conn net.Conn) {

	// 加锁防止多个客户端冲突
	s.mu.Lock()

	// 从字典中删除连接
	delete(s.activeConn, conn)
	s.mu.Unlock()

	conn.Close()

	s.Plugins.DoPostConnClose(conn)
}

func handleError(res *protocol.Message, err error) (*protocol.Message, error) {
	res.SetMessageStatusType(protocol.Error)
	if res.Metadata == nil {
		res.Metadata = make(map[string]string)
	}
	res.Metadata[protocol.ServiceError] = err.Error()
	return res, err
}

func (s *Server) sendResponse(ctx *share.Context, conn net.Conn, writeCh chan *[]byte, err error, req, res *protocol.Message) {
	// 设置压缩类型
	if len(res.Payload) > 1024 && req.CompressType() != protocol.None {
		res.SetCompressType(req.CompressType())
	}

	// encode message to a slice pointer
	data := res.EncodeSlicePointer()
	s.Plugins.DoPreWriteResponse(ctx, req, res, err)

	if s.AsyncWrite {
		writeCh <- data
	} else {
		if s.writeTimeout != 0 {
			conn.SetWriteDeadline(time.Now().Add(s.writeTimeout))
		}
		conn.Write(*data)
		protocol.PutData(data)
	}
	s.Plugins.DoPostWriteResponse(ctx, req, res, err)
}

func parseServerTimeout(ctx *share.Context, req *protocol.Message) context.CancelFunc {
	if req == nil || req.Metadata == nil {
		return nil
	}

	st := req.Metadata[share.ServerTimeout]
	if st == "" {
		return nil
	}

	timeout, err := strconv.ParseInt(st, 10, 64)
	if err != nil {
		return nil
	}

	newCtx, cancel := context.WithTimeout(ctx.Context, time.Duration(timeout)*time.Millisecond)
	ctx.Context = newCtx
	return cancel
}

func (s *Server) handleRequest(ctx context.Context, req *protocol.Message) (res *protocol.Message, err error) {
	serviceName := req.ServicePath
	methodName := req.ServiceMethod

	res = req.Clone()

	res.SetMessageType(protocol.Response)

	// 处理服务service
	s.serviceMapMu.RLock()
	service := s.serviceMap[serviceName]

	if share.Trace {
		log.Debugf("server get service %+v for an request %+v", service, req)
	}

	s.serviceMapMu.RUnlock()
	if service == nil {
		err = errors.New("rpcx: can't find service " + serviceName)
		return handleError(res, err)
	}

	// 处理方法function
	mtype := service.method[methodName]
	if mtype == nil {
		if service.function[methodName] != nil { // check raw functions
			return s.handleRequestForFunction(ctx, req)
		}
		err = errors.New("rpcx: can't find method " + methodName)
		return handleError(res, err)
	}

	argv := reflectTypePools.Get(mtype.ArgType)

	codec := share.Codecs[req.SerializeType()]
	if codec == nil {
		err = fmt.Errorf("cana not find codec for %d", req.SerializeType())
		return handleError(res, err)
	}

	err = codec.Decode(req.Payload, argv)
	if err != nil {
		return handleError(res, err)
	}

	// and get a reply object from object pool
	replyv := reflectTypePools.Get(mtype.ReplyType)

	argv, err = s.Plugins.DoPreCall(ctx, serviceName, methodName, argv)
	if err != nil {
		reflectTypePools.Put(mtype.ReplyType, replyv)
		return handleError(res, err)
	}

	if mtype.ArgType.Kind() != reflect.Ptr {
		err = service.call(ctx, mtype, reflect.ValueOf(argv).Elem(), reflect.ValueOf(replyv))
	} else {
		err = service.call(ctx, mtype, reflect.ValueOf(argv), reflect.ValueOf(replyv))
	}

	if err == nil {
		replyv, err = s.Plugins.DoPostCall(ctx, serviceName, methodName, argv, replyv)
	}

	//  return argc to object pool
	reflectTypePools.Put(mtype.ArgType, argv)

	if err != nil {
		if replyv != nil {
			data, err := codec.Encode(replyv)
			// return reply to object pool
			reflectTypePools.Put(mtype.ReplyType, replyv)
			if err != nil {
				return handleError(res, err)
			}
			res.Payload = data
		}
		return handleError(res, err)
	}

	if !req.IsOneway() {
		data, err := codec.Encode(replyv)
		// return reply to object pool
		reflectTypePools.Put(mtype.ReplyType, replyv)
		if err != nil {
			return handleError(res, err)
		}
		res.Payload = data
	} else if replyv != nil {
		reflectTypePools.Put(mtype.ReplyType, replyv)
	}

	if share.Trace {
		log.Debugf("server called service %+v for an request %+v", service, req)
	}

	return res, nil
}

func (s *Server) handleRequestForFunction(ctx context.Context, req *protocol.Message) (res *protocol.Message, err error) {
	res = req.Clone()

	res.SetMessageType(protocol.Response)

	serviceName := req.ServicePath
	methodName := req.ServiceMethod
	s.serviceMapMu.RLock()
	service := s.serviceMap[serviceName]
	s.serviceMapMu.RUnlock()
	if service == nil {
		err = errors.New("rpcx: can't find service  for func raw function")
		return handleError(res, err)
	}
	mtype := service.function[methodName]
	if mtype == nil {
		err = errors.New("rpcx: can't find method " + methodName)
		return handleError(res, err)
	}

	// get a argv object from object pool
	argv := reflectTypePools.Get(mtype.ArgType)

	// req.SerializeType() 返回序列化类型， codeC 对应encoder or decoder
	codec := share.Codecs[req.SerializeType()]
	if codec == nil {
		err = fmt.Errorf("can not find codec for %d", req.SerializeType())
		return handleError(res, err)
	}

	err = codec.Decode(req.Payload, argv)
	if err != nil {
		return handleError(res, err)
	}
	replyv := reflectTypePools.Get(mtype.ReplyType)

	if mtype.ArgType.Kind() != reflect.Ptr {
		err = service.callForFunction(ctx, mtype, reflect.ValueOf(argv).Elem(), reflect.ValueOf(replyv))
	} else {
		err = service.callForFunction(ctx, mtype, reflect.ValueOf(argv), reflect.ValueOf(replyv))
	}

	reflectTypePools.Put(mtype.ArgType, argv)

	if err != nil {
		reflectTypePools.Put(mtype.ReplyType, replyv)
		return handleError(res, err)
	}

	if !req.IsOneway() {
		data, err := codec.Encode(replyv)
		reflectTypePools.Put(mtype.ReplyType, replyv)
		if err != nil {
			return handleError(res, err)
		}
		res.Payload = data
	} else if replyv != nil {
		reflectTypePools.Put(mtype.ReplyType, replyv)
	}
	return res, nil
}

var shutdownPollInterval = 1000 * time.Millisecond

// Shutdown gracefully shuts down the server without interrupting any
// active connections. Shutdown works by first closing the
// listener, then closing all idle connections, and then waiting
// indefinitely for connections to return to idle and then shut down.
// If the provided context expires before the shutdown is complete,
// Shutdown returns the context's error, otherwise it returns any
// error returned from closing the Server's underlying Listener.
func (s *Server) Shutdown(ctx context.Context) error {
	var err error
	if atomic.CompareAndSwapInt32(&s.inShutdown, 0, 1) {
		log.Info("shutdown begin")

		s.mu.Lock()

		// 主动注销注册的服务
		if s.Plugins != nil {
			for name := range s.serviceMap {
				s.Plugins.DoUnregister(name)
			}
		}

		s.In.Close()
		for conn := range s.activeConn {
			if tcpConn, ok := conn.(*net.TCPConn); ok {
				tcpConn.CloseRead()
			}
		}
		s.mu.Unlock()

		// wait all in-processing requests finish.
		ticker := time.NewTicker(shutdownPollInterval)
		defer ticker.Stop()
	outer:
		for {
			if s.checkProcessMsg() {
				break
			}
			select {
			case <-ctx.Done():
				err = ctx.Err()
				break outer
			case <-ticker.C:
			}
		}

		/*if s.gatewayHTTPServer != nil {
			if err := s.closeHTTP1APIGateway(ctx); err != nil {
				log.Warnf("failed to close gateway: %v", err)
			} else {
				log.Info("closed gateway")
			}
		}*/

		s.mu.Lock()
		for conn := range s.activeConn {
			conn.Close()
			delete(s.activeConn, conn)
			s.Plugins.DoPostConnClose(conn)
		}
		s.closeDoneChanLocked()

		s.mu.Unlock()

		log.Info("shutdown end")

	}
	return err
}

func (s *Server) checkProcessMsg() bool {
	size := atomic.LoadInt32(&s.handlerMsgNum)
	log.Info("need handle in-processing msg size:", size)
	return size == 0
}
