package Service

import (
	"bufio"
	"fmt"
	"forward-core/Constant"
	"forward-core/Models"
	"forward-core/NetUtils"
	"github.com/astaxie/beego/logs"
	"io"
	"net"
	"sync"
	"time"
)

type MagicServiceV1 struct {
	MagicClientMap     map[string]net.Conn
	MagicClientMapLock sync.Mutex
	MagicListener      net.Listener
	agentRunType       int
	MagicTargetAddr string
	sessionId 		int
	idLock             sync.Mutex
	sessionConnMap     map[int]net.Conn
	ForwardInfo *Models.PortForward
}

func NewMagicServiceV1() *MagicServiceV1 {
	return &MagicServiceV1{
		MagicClientMap:make(map[string]net.Conn,200),
		sessionConnMap:make(map[int]net.Conn,200),
	}
}

func (_self *MagicServiceV1) GetNewSessionId() int {
	_self.idLock.Lock()
	defer _self.idLock.Unlock()
	_self.sessionId++

	return _self.sessionId
}

func (_self *MagicServiceV1) GetKeyByEntity(entity *Models.PortForward) string {

	fromAddr := fmt.Sprint(entity.Addr, ":", entity.Port)
	toAddr := fmt.Sprint(entity.TargetAddr, ":", entity.TargetPort)
	key := _self.GetKey(fromAddr, toAddr, entity.FType)

	return key
}

func (_self *MagicServiceV1) GetKey(sourcePort, targetPort string, fType int) string {

	return fmt.Sprint(sourcePort, "_", fType, "_TCP_", targetPort)

}

func (_self *MagicServiceV1) RegistryMagicClient(key string, conn net.Conn) {
	_self.MagicClientMapLock.Lock()
	defer _self.MagicClientMapLock.Unlock()

	_self.MagicClientMap[key] = conn

}

func (_self *MagicServiceV1) GetTopMagicClient() net.Conn {
	_self.MagicClientMapLock.Lock()
	defer _self.MagicClientMapLock.Unlock()

	for _, v := range _self.MagicClientMap {
		return v
	}

	return nil

}

func (_self *MagicServiceV1) GetMagicListener() net.Listener {

	return _self.MagicListener

}

func (_self *MagicServiceV1) UnRegistryMagicClient(key string) {
	_self.MagicClientMapLock.Lock()
	defer _self.MagicClientMapLock.Unlock()

	delete(_self.MagicClientMap, key)
	logs.Debug("UnRegistryMagicClient key: ", key)

}

func (_self *MagicServiceV1) CountMagicClient() int {
	_self.MagicClientMapLock.Lock()
	defer _self.MagicClientMapLock.Unlock()

	return len(_self.MagicClientMap)

}

func (_self *MagicServiceV1) GetMagicClient() map[string]net.Conn {

	return _self.MagicClientMap

}


func (_self *MagicServiceV1) StartMagicService(addr string, result chan Models.FuncResult) {
	//?????????????????????
	resultData := &Models.FuncResult{Code: 0, Msg: ""}
	var err error
	_self.MagicListener, err = net.Listen("tcp", addr)
	if err != nil {
		logs.Error("Magic Listen err:", err)
		resultData.Code = 1
		resultData.Msg = err.Error()
		result <- *resultData
		return
	}

	result <- *resultData

	for {
		logs.Debug("Magic Ready to Accept ...")
		magic_client_Conn, err := _self.MagicListener.Accept()
		if err != nil {
			logs.Error("Accept err:", err)
			break
		}

		if _self.CountMagicClient() > 0 && _self.CurrentAgentRunType() != 1 {
			logs.Debug("???????????????????????????Agent???????????????????????????????????????")
			NetUtils.WriteConn(magic_client_Conn, -1, Constant.MagicCmd_Refused, []byte(""))
			magic_client_Conn.Close()
			continue
		}

		if _self.CountMagicClient() == 0 {
			magicId := magic_client_Conn.RemoteAddr().String()
			_self.RegistryMagicClient(magicId, magic_client_Conn)
		} else {
			if _self.CurrentAgentRunType() == 1 {
				_self.MagicJustCopy(magic_client_Conn, _self.MagicTargetAddr)
			}

		}

	}

}

func (_self *MagicServiceV1) StopMagicService(result chan Models.FuncResult) {
	resultData := &Models.FuncResult{Code: 0, Msg: ""}

	for k, conn := range _self.MagicClientMap {
		conn.Close()
		_self.UnRegistryMagicClient(k)

	}
	_self.MagicListener.Close()
	_self.MagicListener = nil

	result <- *resultData

}

func (_self *MagicServiceV1) StartMagicForward(portForward *Models.PortForward, result chan Models.FuncResult) {
	resultData := &Models.FuncResult{Code: 0, Msg: ""}

	agentConn := _self.GetTopMagicClient()

	if agentConn == nil {
		resultData.Code = 1
		resultData.Msg = "????????????Agent??????"
		result <- *resultData
		return
	}

	if _self.CurrentAgentRunType() != 0 {
		resultData.Code = 1
		resultData.Msg = "??????????????????Agent???????????????????????????"
		result <- *resultData
		return
	}

	if portForward.FType == 2 {
		//??????????????????
		go _self.ReverseListenForClient(portForward, agentConn, result)
		callback := func(conn net.Conn, sessionId int, cmd byte, payload []byte) {
			//payload ?????????????????????
			_self.OnTunnelRecv(_self.sessionConnMap[sessionId], sessionId, cmd, payload)
		}
		logs.Debug("??? magic_client_Conn ??????????????? client_Conn")
		go NetUtils.ReadConn(agentConn, callback)
	} else {
		//????????????
		localListenAddr := fmt.Sprint(portForward.Addr, ":", portForward.Port)
		NetUtils.WriteConn(agentConn, -1, Constant.MagicCmd_AgentListenerOpen, []byte(localListenAddr))
		result <- *resultData
		_self.agentRunType = 1
		_self.MagicTargetAddr = fmt.Sprint(portForward.TargetAddr, ":", portForward.TargetPort)

		//key := _self.GetKeyByEntity(portForward)
		//_self.RegistryPort(key, nil)
	}

	_self.ForwardInfo = portForward

}

func (_self *MagicServiceV1) StopMagicForward() error {

	return nil
}

func (_self *MagicServiceV1) MagicJustCopy(toConn net.Conn, targetAddr string) {

	localConn, err := net.DialTimeout("tcp", targetAddr, 30*time.Second)
	if err != nil {
		logs.Error("try dial err", err)
		return
	}

	go func() {
		_, err = io.Copy(localConn, toConn)
		if err != nil {
			logs.Error("JustCopy to local ?????????????????????", err)
			localConn.Close()
		}
	}()

	go func() {
		_, err = io.Copy(toConn, localConn)
		if err != nil {
			logs.Error("JustCopy to local ??????????????????2???", err)
			toConn.Close()
		}
	}()

}

func (_self *MagicServiceV1) ReverseListenForClient(portForward *Models.PortForward, magic_client_Conn net.Conn, result chan Models.FuncResult) {
	resultData := &Models.FuncResult{Code: 0, Msg: ""}

	localListenAddr := fmt.Sprint(portForward.Addr, ":", portForward.Port)
	//?????????????????????????????????????????????????????????
	remote := fmt.Sprint(portForward.TargetAddr, ":", portForward.TargetPort)
	//fType := portForward.FType

	client_listener, err := net.Listen("tcp", localListenAddr)
	if err != nil {
		logs.Error("ListenForClient err:", err)
		resultData.Code = 1
		resultData.Msg = err.Error()
		result <- *resultData
		return
	}

	result <- *resultData
	_self.agentRunType = 2
	//key := _self.GetKeyByEntity(portForward)
	//_self.RegistryPort(key, client_listener)

	//??? client_Conn ??????????????? magic_client_Conn
	//??? magic_client_Conn ??????????????? client_Conn
	for {
		logs.Debug("ListenForClient Ready to Accept ...")
		client_Conn, err := client_listener.Accept()
		if err != nil {
			logs.Error("Accept err:", err)
			break
		}

		//id := client_Conn.RemoteAddr().String()
		//_self.RegistryClient(fmt.Sprint(localListenAddr, "_", fType, "_", id), client_Conn)

		//????????????????????????????????????sessionId
		sessionId := _self.GetNewSessionId()
		_self.sessionConnMap[sessionId] = client_Conn
		logs.Debug("????????????????????????sessionId???", sessionId)

		NetUtils.WriteConn(magic_client_Conn, sessionId, Constant.MagicCmd_AgentConnOpen, []byte(remote))

		logs.Debug("??? sessionId???", sessionId, " ?????? AgentConnOpen ??????")

		logs.Debug("??? client_Conn ??????????????? magic_client_Conn sessionId???", sessionId)
		go _self.ReadRawConn(client_Conn, magic_client_Conn, sessionId, Constant.MagicCmd_DataToAgent)

	}

}

func (_self *MagicServiceV1) OnTunnelRecv(client_Conn net.Conn, sessionId int, cmd byte, payload []byte) {
	logs.Debug("??????????????? sessionId???", sessionId, " ?????????????????????????????????", cmd)
	switch cmd {
	case Constant.MagicCmd_DataToMagic:
		client_Conn.Write(payload)
	}

}

func (_self *MagicServiceV1) ReadRawConn(from net.Conn, magic_client_Conn net.Conn, sessionId int, cmd byte) {

	arr := make([]byte, 5000)
	reader := bufio.NewReader(from)

	for {
		size, err := reader.Read(arr)
		if err != nil {
			break
		}

		err = NetUtils.WriteConn(magic_client_Conn, sessionId, cmd, arr[0:size])

		if err != nil {
			//?????????
			logs.Error(err)
			break
		}
	}
}

func (_self *MagicServiceV1) CurrentAgentRunType() int {
	// 0:?????????1?????????????????????????????????2??????????????????????????????
	return _self.agentRunType
}


