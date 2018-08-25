package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	//"bytes"
	"crypto/md5"
	//"encoding/binary"
	"encoding/hex"
	"fmt"
	//	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	//"github.com/imroc/biu" //二进制库  废弃
)

type AP struct {

	// AP 的设备ID
	DEV_ID string
	// AP IP地址
	AP_IP string
	// AP MAC地址
	AP_MAC string
	// AP的索引
	AP_Index int
	// AP 硬件型号
	DEV_BoardName string
	// AP 产品名称
	DEV_ProductName string
	// 设备类型
	DEV_Class string
	// 厂商ID
	DEV_Vendor string
	// 固件版本
	DEV_FirmwareVersion string
	// 国家代码
	DEV_CountryCode string
	// BRC ID
	DEV_BRCID string
	//////////////////

	//// AP 内部控制参数变量初始化
	// AP 心跳间隔
	HeartTime int
	// AP 状态码
	APStatusCode int
	// udp socket连接对象
	udpCliSock *net.UDPConn
	// 初始化解码器 TLV  二码合一
	tlv *TLV
	// 出事化解码器Action TLV
	//tlvAction *TLV

	// send header 消息部分
	sendHeaderDict map[string]string
	// data 默认不变化的属性
	sendDataDict map[string]string
	// 返回结果
	rsDataDict   map[string]string
	rsHeaderCode int
	rsErrorCode  string

	// AP离线次数计数器
	offline_counter int
	// 心跳请求间隔时间  默认1秒  如4接收消息则使用接收到的值
	isHeartDefault bool // true时 不使用设置的值
	OfflineTimes   int  // 心跳包最大重试次数 -- 默认重配置获取，会覆盖当前值
	// 消息计数器
	counter int
	// 第一条消息开始时间
	startTime time.Time
	// 是否记录日志，如果不记录则都不记录（包括错误日志） --需要优化
	logStatus bool

	//////AP 配置更新 终端上报 终端授权 终端在线 终端离线 终端记账 等的AP属性
	Action string
	// AP的配置MD5
	Config_MD5 string
	// AP 配置文件长度
	Config_Length int
	// AP配置文件的块
	Config_Blocks int
	// AP配置的 配置文件分块序号
	Config_Block_Index int
	// Config 配置列表，根据Index的顺序存放AP的配置文件成为一个临时表，0为拼接后的整体内容
	Config_Block_Data     []string
	Device_State_Interval int
	// 默认AP的配置状态 未配置
	Config_Status string

	// 离线是的时间点
	AP_User_OffLine_Time int64
	// 杂线时间点
	AP_User_OnLine_Time int64

	// AP 用户对象
	APUsersObject *APUsers
	// AP 用户列表
	APUsersList []map[string]string
	// 无线频道序号索引
	AP_PHY int64
	// 对应频道关联配置索引
	AP_VAP int64
	// 对应频道关联配置的ssid名称
	AP_SSID string

	// portal登录失败或成功心跳心跳状态  默认是8的 其他的是7  默认不开启
	isPortalHeartStatus  bool // portal 接收AC请求走8
	isRequestHeartStatus bool // 正常rquest的走7
	BILL_WAIT_TIME       int  // 计费时间-- 默认重配置获取，会覆盖当前值 ，不会再变化 除非配置更新
	BillWaitTime         int  // 动态的计费时间 可能会动态 变化

	// 记录当前的是否为chong发状态  默认不是 若发送失败则将状态改为true 用于心跳判断
	isReSendMessageStatus bool
	// 心跳自己错误 记录状态  不等待
	isHeartErr bool
}

func NewAP(ip string, mac string, devid string, index int) *AP {
	ap := new(AP)

	// 初始化 AP的变化参数：
	ap.AP_IP = ip_to_hex(ip)      // AP IP地址
	ap.AP_MAC = str_to_hex(mac)   // AP MAC地址
	ap.AP_Index = index           // AP的索引
	ap.DEV_ID = str_to_hex(devid) // 设备ID

	// AP 硬件型号
	ap.DEV_BoardName = str_to_hex(DEV_BoardName) //("A180p-%d" % i)
	// AP 产品名称
	ap.DEV_ProductName = str_to_hex(DEV_ProductName)
	// 设备类型 设备类型整数值，取值如下：1 – AP	2 – AC 3 – ROUTER 4 – SWITCH 5 – NAC 6 – PLC MASTER 7 – PLC SLAVE 8 – CPE
	ap.DEV_Class = int_to_hex(DEV_Class, 8)
	// 厂商ID
	ap.DEV_Vendor = int_to_hex(DEV_Vendor, 8)
	// 固件版本
	ap.DEV_FirmwareVersion = str_to_hex(DEV_FirmwareVersion)
	// 国家代码
	ap.DEV_CountryCode = fmt.Sprintf("%0.4x", DEV_CountryCode)
	// BRC ID
	ap.DEV_BRCID = str_to_hex(DEV_BRCID)
	ap.AP_PHY = 0 // 无线频道序号索引  初始化自动获取
	ap.AP_VAP = 0 // 对应频道关联配置索引 初始化自动获取

	//// AP 内部控制参数变量初始化
	////send header 消息部分
	ap.sendHeaderDict = make(map[string]string)
	//返回结果
	ap.rsDataDict = make(map[string]string)
	// version + type
	ap.sendHeaderDict["header_tag"] = "0200" // 2 2
	ap.sendHeaderDict["payLoadType"] = "02"  // 2t

	// data 默认不变化的属性
	ap.sendDataDict = make(map[string]string)

	//默认初始状态为1 ，0为停止退出
	ap.APStatusCode = 1

	// AP的配置MD5
	ap.Config_MD5 = fmt.Sprintf("%0.32d", 0)
	// AP 配置文件长度
	ap.Config_Length = 0
	// AP配置文件的块
	ap.Config_Blocks = 0
	// AP配置的 配置文件分块序号
	ap.Config_Block_Index = 0
	// Config 配置列表，根据Index的顺序存放AP的配置文件成为一个临时表，0为拼接后的整体内容
	//Config_Block_Data = []string
	ap.Device_State_Interval = 0
	// 默认AP的配置状态 未配置
	ap.Config_Status = "00000000"

	// 离线是的时间点
	ap.AP_User_OffLine_Time = 0
	// 杂线时间点
	ap.AP_User_OnLine_Time = 0

	// 初始化解码器 TLV
	ap.tlv = NewTLV()
	// 出事化解码器Action TLV
	//ap.tlvAction = NewTLV(true)
	// AP离线次数计数器
	ap.offline_counter = 1
	// 心跳请求间隔时间  默认1秒  如4接收消息则使用接收到的值
	ap.isHeartDefault = true // true时 不使用设置的值
	ap.HeartTime = 1
	ap.OfflineTimes = 3 // 心跳包最大重试次数 -- 默认重配置获取，会覆盖当前值
	// 消息计数器
	ap.counter = 0
	// 是否记录日志，如果不记录则都不记录（包括错误日志） --需要优化
	ap.logStatus = true

	//////AP 配置更新 终端上报 终端授权 终端在线 终端离线 终端记账 等的AP属性
	ap.Action = "00000001" // 只有7 才有这个属性 默认00000001 为需要获取配置状态 00000000 表示默认状态

	// portal登录失败或成功心跳心跳状态  默认是8的 其他的是7  默认不开启
	ap.isPortalHeartStatus = false  // portal 接收AC请求走8
	ap.isRequestHeartStatus = false // 正常rquest的走7
	ap.BILL_WAIT_TIME = 0           // 计费时间-- 默认重配置获取，会覆盖当前值 ，不会再变化 除非配置更新
	ap.BillWaitTime = 0             // 动态的计费时间 可能会动态 变化

	// 记录当前的是否为chong发状态  默认不是 若发送失败则将状态改为true 用于心跳判断
	ap.isReSendMessageStatus = false
	// 心跳自己错误 记录状态  不等待
	ap.isHeartErr = false

	//(ip,port) = AC.split(":")
	// 检查AC 和ORG ID是否合法
	if (len(AC_IP) < 7) || (len(AC_Port) < 4) || (len(ORGID) != 36) {
		logPrint(fmt.Sprintf("请检查错误的AC参数:AC_IP=%s,AC_Port=%s,ORGID=%s", AC_IP, AC_Port, ORGID))
		os.Exit(3)
	}

	logPrint(fmt.Sprintf("######[AP INIT]######[初始化index=[%d]虚拟AP成功！", index))
	return ap
}

// 启动AP模拟流程
func (this *AP) start() bool {

	// 记录启动时间
	this.startTime = time.Now()
	// 初始化 sock client
	udpAddr, err := net.ResolveUDPAddr("udp4", fmt.Sprintf("%v:%v", AC_IP, AC_Port))
	if !this.check("检查udpAddr初始化错误：", err) {
		return false
	}
	this.udpCliSock, err = net.DialUDP("udp", nil, udpAddr)
	if !this.check("检查udpCliSock连接错误：", err) {
		return false
	}

	//this.logApPrint(fmt.Sprintf("1111")
	for this.APStatusCode != 0 {
		switch this.APStatusCode {
		// 该状态下AP向AC发送发现请求（Discovery Request），并设置等待回复
		case 1:
			this.discoveryRequest()
		case 2:
			this.discoveryResponse()
		case 3:
			this.authorizeRequest()
		case 4:
			this.authorizeResponse()
		case 5: // 发送心跳
			this.echoRequest()
		case 6: // 接收心跳
			this.echoResponse()
		case 7: // AP msg 通讯
			this.actionRequest()
		case 8: // AP msg 消息返回
			this.actionResponse()
		default:
			time.Sleep(1 * time.Second)
			this.logApErrPrint(fmt.Sprintf("未知的APStatusCode: %v", this.APStatusCode))
		}
	}
	return false
}

// status:1 发送DISCOVER 报文
func (this *AP) discoveryRequest() {
	//time.Sleep(1 * 1e9)
	//this.logApPrint("discoveryRequest1111")
	sMsg := this.tlv.packDisc(this.updateHeaderMsg(), this.updateDataMsg())
	this.sendMsg(sMsg)
	this.APStatusCode = 2
	// 清空离线计数器
	this.offline_counter = 1
}

// status:2 发现响应（Discover Response)
func (this *AP) discoveryResponse() {
	//time.Sleep(1 * 1e9)
	//this.logApPrint("discoveryResponse2222")
	if this.recvMsg(false) {
		this.APStatusCode = 3
	} else {
		this.APStatusCode = 1
	}
}

// code:3 发送Authorize Request 报文
func (this *AP) authorizeRequest() {
	// print 33333
	sMsg := this.tlv.packDisc(this.updateHeaderMsg(), this.updateDataMsg())
	this.sendMsg(sMsg)
	this.APStatusCode = 4
}

// code:4 发送Authorize Request 报文
func (this *AP) authorizeResponse() {
	// print 44444
	// 如果接收到结构则按返回的echo time 时间进行心跳请求  目前默认写死
	if this.recvMsg(false) {
		if this.isHeartDefault {
			this.HeartTime = int(hex_to_int(this.rsDataDict["0028"]))    // 获取心跳间隔时间
			this.OfflineTimes = int(hex_to_int(this.rsDataDict["0029"])) // 获取心跳最大重试次数
			this.logApPrint(fmt.Sprintf("############### get config HeartTime=%d ,OfflineTimes=%d", this.HeartTime, this.OfflineTimes))
		}
		this.APStatusCode = 5
		// time.sleep(WAITTIME)
	} else {
		this.APStatusCode = 3
	}
}

// code{5 发送心跳请求（Echo Request）报文
func (this *AP) echoRequest() {

	// 要排除第一次获取配置时不等待
	if this.Action == "00000001" {
		//pass
	} else if this.isHeartErr { // 心跳自己失败
		//pass
		// 终端上线/终端计费/终端left 失败时  走心跳时不等待
	} else if this.isReSendMessageStatus {
		// this.isReSendMessageStatus = false
		//pass
		// 账单请求过来走只有时间大于心跳时间才等待
	} else if this.Action == "00000028" {
		tempTime := this.BillWaitTime - this.HeartTime
		// 如果账单时间 大于心跳 则先一次心跳时间
		if tempTime >= 0 { //继续走心跳
			this.BillWaitTime = tempTime
			time.Sleep(time.Duration(this.HeartTime) * time.Second)
		}
	} else if this.isPortalHeartStatus { // portal 心跳不等待
		//pass
	} else { // 正常等待心跳
		time.Sleep(time.Duration(this.HeartTime) * time.Second)
	}
	// print 55555
	sMsg := this.tlv.packDisc(this.updateHeaderMsg(), this.updateDataMsg())
	this.sendMsg(sMsg)
	this.APStatusCode = 6
}

// code:6 接收心跳响应（Echo Response） 报文
func (this *AP) echoResponse() {
	// print 66666
	// 如果接收到结构则按返回的echo time 时间进行心跳请求  目前默认写死
	if this.recvMsg(false) {
		// 重置未收到心跳计数
		this.offline_counter = 0
		//心跳OK 清理状态
		this.isHeartErr = false

		if this.Action == "00000001" { // 首次获取配置，其他情况下走心跳
			this.APStatusCode = 7
			// 7 和8 类型请求走心跳的
		} else if this.isPortalHeartStatus { // 正常8类型走心跳的收到消息 继续
			this.APStatusCode = 8
			return
		} else if this.isRequestHeartStatus { // 正常7 类型走心跳的
			this.APStatusCode = 7
			return
			// 终端上线/终端计费/终端left 失败时  关闭心跳状态
		} else if this.isReSendMessageStatus {
			this.isReSendMessageStatus = false
			// 如果是新终端响应失败了，需要重新响应 并设置对应状态
			this.APStatusCode = 7
		} else if this.Action == "00000028" { // 账单请求 大于心跳
			// 如果大于等于0 表示还需要走心跳等待
			if this.BillWaitTime >= this.HeartTime {
				this.APStatusCode = 5
				return
			} else { // 否则直接走账单
				this.APStatusCode = 7
				return
			}
			// 完成心跳后进行终端状态的判断
		} else if this.Action == "00000000" { // 判断使用是否需要重新上线
			// 检查是否需要开始上用户
			//if AP_User_Reline_Bool && (this.AP_User_OffLine_Time || this.AP_User_OffLine_Time == 0) {
			if AP_User_Reline_Bool {
				// 需要上线 离线时间等于0 或 和 在线时间大于 设定的离线时间间隔
				if (time.Now().Unix() - this.AP_User_OffLine_Time) >= AP_User_Reline_WaitTime {
					this.logApPrint("########满足上线时间，用户开始重新现上线")
					// 开始走 新终端上线流程
					this.APStatusCode = 7
					this.Action = "0000001e"
				} else { // 否则不用处理继续心跳
					this.logApPrint("########不满足上线时间继续等待")
					this.APStatusCode = 5
					this.Action = "00000000"
				}
			} else if this.AP_User_OffLine_Time == 0 && len(this.APUsersList) == 0 { // 第一次上线
				// 开始走 新终端上线流程
				this.APStatusCode = 7
				this.Action = "0000001e"
			} else {
				this.logApErrPrint("########异常，理论上不应该走这里,重新走心跳")
				// 开始走 新终端上线流程
				this.APStatusCode = 5
				this.Action = "00000000"
			}
			return
		} else {
			this.APStatusCode = 5
			this.Action = "00000000"
		}
	} else {
		// 如果没有心跳 开始计数
		this.offline_counter += 1
		if this.offline_counter >= this.OfflineTimes {
			// 如果次数达到后仍没有收到心跳则退出
			this.APStatusCode = 0
			this.Action = "00000000"
			this.logApErrPrint(fmt.Sprintf("err:没有收AC的心跳响应，第%v次！退出！exit ！", this.offline_counter))
			return
		} else {
			this.isHeartErr = true
			this.logApErrPrint(fmt.Sprintf("err:没有收AC的心跳响应，第%v次！", this.offline_counter))
		}
		time.Sleep(time.Duration(this.HeartTime-WAITTIME) * time.Second)
		this.APStatusCode = 5
	}
	return
}

// code:7 发送AP的Action （Action Request） 报文
func (this *AP) actionRequest() {
	// print 7777
	//  Action 30 新终端请求
	if this.Action == "00000001" { // Action: 1 配置请求
		// this.Action = "00000001"
		this.__getConfigRequest()
	} else if this.Action == "00000003" { // Action: 3 配置块请求
		// this.Action = "00000003"
		this.__getConfigBlockRequest()
	} else if this.Action == "0000001e" { // Action 30 新终端请求
		this.__newStationRequest()
	} else if this.Action == "00000021" { // Action：33 终端授权响应 AP -> AC
		this.__stationAccessResponse()
	} else if this.Action == "00000028" { // Action 40 终端账单请求
		this.__newStationBillRequest()
	} else if this.Action == "00000025" { // Action：37 踢终端响应 AP -> AC
		this.__stationDropResponse()
	} else if this.Action == "00000026" { // Action：38 终端离开请求（Station Left Request）
		this.__stationLeftRequest()
	} else {
		time.Sleep(time.Duration(10) * time.Millisecond)
		if this.Action == "00000020" { //
			this.APStatusCode = 8
			this.logApErrPrint("err:登录请求消息接收失败!!! ")
		} else {
			this.logApErrPrint(fmt.Sprintf("err:未知的终端请求Action: %v", this.Action))
		}
	}
}

// code:8 接AP的Action （Action Response） 报文
func (this *AP) actionResponse() {
	// print 88888
	// if this.Action == "00000020"{ print this.APUsersList[this.APUsersList[0]]["userStatus"]
	// 如果是要走登录认证流程 ,且用户是1：未登录状态   则需先发送登录请求
	if this.Action == "00000020" {
		index, _ := strconv.Atoi(this.APUsersList[0]["userIndex"])
		user := this.APUsersList[index]
		if user["userStatus"] == "1" {
			if !this.__portalHttpLogin() {
				// 心跳登录 如果登录失败 先走心跳再走登录
				this.logApErrPrint("登录失败走心跳")
				this.APStatusCode = 5
				return
			}
		} else { // 不需要认证 然后开始记录当前时间为用户在线时间 清空离线时间
			if AP_User_Reline_Bool {
				this.AP_User_OnLine_Time = time.Now().Unix()
				this.AP_User_OffLine_Time = 0
			}
		}
		// 如果是走离线流程
	} else if this.Action == "00000024" {
		// 且用户是2：登录状态 则发送注销请求
		index, _ := strconv.Atoi(this.APUsersList[0]["userIndex"])
		user := this.APUsersList[index]
		if user["userStatus"] == "2" {
			if !this.__portalHttpLogout() {
				// 心跳登录 如果登录失败 先走心跳再走注销
				this.APStatusCode = 5
				return
			}
		} else { // 不需要认证 然后开始记录当前时间为用户离线时间 清空在线时间
			if this.AP_User_OnLine_Time > 0 {
				this.AP_User_OffLine_Time = time.Now().Unix()
				this.AP_User_OnLine_Time = 0
			}
		}
	}
	// 如果接收成功
	if this.recvMsg(false) {
		// print "88881", this.rsDataDict
		// 获取Action
		// this.Action = this.rsDataDict["0001"]
		Action, ok := this.rsDataDict["0001"]
		if !ok {
			if this.Action == "00000020" { // 如果登录异常则忽略
				this.logApErrPrint(fmt.Sprintf("err:终端认证请求回复异常的，跳过当前用户进行进行下个，rsDataDict:%v", this.rsDataDict))
				// 设置 下次发送走 777
				this.APStatusCode = 7
				// 准备开始 下个 用户认证
				this.Action = "00000021"
			} else {
				this.logApErrPrint(fmt.Sprintf("err:终端响应异常的Action-%s rsDataDict:%v，重新走心跳！", this.Action, this.rsDataDict))
				this.APStatusCode = 5
				//	丢掉异常的包
				this.recvMsg(true)
			}
			return
		}

		// 特殊情况下，在对接小区宽带时，用户登录4个后，第五个会登录的用户返回的Action时 00000024，导致后面流程走错了
		if this.Action == "00000020" && Action == "00000024" { // 返回值不可用，需要忽略掉
			this.logApErrPrint("注意：请检查对应测试账号允许使用的最大终端数量！！！！！")
		} else { // 正常情况下使用返回值
			this.Action = Action
		}

		//  2 配置响应
		if this.Action == "00000002" {
			this.__getConfigResponse()
		} else if this.Action == "00000004" { // Action: 4 配置块响应
			this.__UpdateConfigBlockResponse()
		} else if this.Action == "00000005" { // Action: 5 AC-AP Update Configuration Request
			this.__updateConfigurationRequest()
		} else if this.Action == "0000001f" { // Action：31 新终端响应
			this.__newStationResponse()
		} else if this.Action == "00000020" { // Action：32 终端授权请求 AC -> AP（Station Access Request）
			this.__stationAccessRequest()
		} else if this.Action == "00000029" { // Action 41 终端账单响应
			this.__newStationBillResponse()
			// 检查是否需要开始离线用户 同时满足 需要下线 在线时间存在 和 在线时间大于 设定的离线时间间隔
			if AP_User_Reline_Bool && this.AP_User_OnLine_Time > 0 && this.APUsersList[0]["userIndex"] == "1" && (time.Now().Unix()-this.AP_User_OnLine_Time) >= AP_User_Online_WaitTime {
				//this.logApPrint("##########//记账完成后满足下线条件，开始进行下线操作！", isdebug=true)
				if AP_User_Is_Left {
					//  // 直接踢用户
					this.APStatusCode = 7
					this.Action = "00000026"
				} else {
					// 开始走 新终端下线流程
					this.APStatusCode = 8
					this.Action = "00000024" // 先走portal踢用户 再离线
				}
			}
		} else if this.Action == "00000024" { // Action：36 踢终端请求 AP -> AC  (drop http)
			this.__stationDropRequest()
		} else if this.Action == "00000027" { // Action：39 终端离开响应（Station Left Response）
			this.__stationLeftResponse()
		} else { // 找不到Action
			// time.sleep(this.HeartTime)
			this.logApErrPrint(fmt.Sprintf("err:终端响应异常的 Action: %v，重新走心跳！", this.Action))
			this.APStatusCode = 5
			this.Action = "00000000"
		}
		return
	} else {
		// 表示消息接收失败或没有timeout了 重新走心跳消息 ,并记住当前的状态
		// this.APStatusCode_BAK = this.APStatusCode
		this.logApErrPrint(fmt.Sprintf("err:终端响应失败Action: %v，重新走心跳！", this.Action))
		this.APStatusCode = 5
	}

	// 如果是新终端响应失败了，需要重新响应 并设置对应状态
	if this.Action == "0000001f" {
		this.Action = "0000001e"
		this.isReSendMessageStatus = true
		// 如果是终端计费失败了，需要重新响应
	} else if this.Action == "00000029" {
		// // 如果连续出现 多次失败 可能是终端没有上线成功  需要记录登录的错误次数
		// userIndex = this.APUsersList[0]["userIndex"]
		// //  如果登注销功 比较 返回的用户IP是否与发送的一致
		// if this.APUsersList[userIndex]["userIP"]

		this.Action = "00000028"
		this.isReSendMessageStatus = true
		// 如果是终端left失败了，需要重新响应
	} else if this.Action == "00000027" {
		this.Action = "00000026"
		this.isReSendMessageStatus = true
		// // 如果是终端login失败了，需要重新响应
		// } else if this.Action == "00000027"{
		//     this.Action = "00000026"
		//     this.isReSendMessageStatus = true
	} else { // 否则重新开始
		this.Action = "00000000"
	}
	time.Sleep(time.Duration(10) * time.Millisecond)
}

////  ACtion 的部分
// Action: 1 配置请求（Get Configuration Request）
func (this *AP) __getConfigRequest() {
	sMsg := this.tlv.packDisc(this.updateHeaderMsg(), this.updateDataMsg())
	this.sendMsg(sMsg)
	// 设置下个code 和Action
	this.APStatusCode = 8
	// AP 动作
	this.Action = "00000002"
}

// Action: 2 配置响应（Get Configuration Response）
func (this *AP) __getConfigResponse() {
	// Configuration Status: 配置文件状态
	this.Config_Status = this.rsDataDict["0009"]

	v, ok := this.rsDataDict["000e"]
	if ok { // Device State Data Interval
		this.Device_State_Interval = int(hex_to_int(v))
	}

	v, ok = this.rsDataDict["0007"]
	if ok { // Configuration MD5
		this.Config_MD5 = hex_to_str(v)
	}

	v, ok = this.rsDataDict["000a"]
	if ok { // Configuration Size
		this.Config_Length = int(hex_to_int(v))
	}

	v, ok = this.rsDataDict["000c"]
	if ok { // Configuration Blocks
		this.Config_Blocks = int(hex_to_int(v))
		// 创建与快大小的数组
		//this.Config_Block_Data = [""] * (this.Config_Blocks + 1)
		this.Config_Block_Data = make([]string, this.Config_Blocks+1)
	}

	if this.Config_Status == "00000000" || this.Config_Status == "00000001" { // 如果未配置 或 无变化 等待下个心跳再次请求
		//如果是无变化 ，则判断是是否满足上线条件
		if this.Config_Status == "00000001" {
			this.logApPrint("############Config_Status=00000001 请求配置后无变化")
			// 开始走 新终端上线流程
			this.APUsersList = []map[string]string{}
			this.APUsersObject = nil
			this.Action = "00000000"
			this.APStatusCode = 5
			//// 设置用户离线时间 开始记时离线时间 清理在西安时间
			if AP_User_Reline_Bool {
				this.AP_User_OnLine_Time = 0
				this.AP_User_OffLine_Time = time.Now().Unix()
			}
		} else {
			// 未配置继续 心跳 和配置检查
			this.logApPrint("#############Config_Status=00000000 AP is ! get configure,pleas configuore AP !!!")
			this.APStatusCode = 5
			this.Action = "00000001"
			time.Sleep(time.Duration(this.HeartTime) * time.Second)
		}
	} else if this.Config_Status == "00000002" || this.Config_Status == "00000003" { // 如果配置更新 开始请求
		if this.Config_Status == "00000002" {
			this.logApPrint("###########Config_Status=00000002 有更新，使用配置块方式（一次获取-待定） 开始请求")
		} else {
			this.logApPrint("###########Config_Status=00000003 有更新，使用配置块方式(多次拼接) 开始请求")
		}
		// 2 和 3 都统一用配置块方式请求   （正常情况下都是3 2 需要单独处理 ---暂未处理）
		this.Action = "00000003" // 设置走Action 3 的动作消息
		this.APStatusCode = 7
		this.Config_Block_Index = 0 // 索引开始计数
	} else {
		this.Action = "00000000"
		this.APStatusCode = 5
		// this.isRequestHeartStatus = true
		this.logApErrPrint("err:异常的Config_Status=" + this.Config_Status)
	}
}

// Action: 3 配置块请求（Get Configuration Block Request）
func (this *AP) __getConfigBlockRequest() {
	sMsg := this.tlv.packDisc(this.updateHeaderMsg(), this.updateDataMsg())
	this.sendMsg(sMsg)
	// 设置下个code 和Action
	this.APStatusCode = 8
	// AP 动作
	this.Action = "00000004"
}

// Action: 4 配置块响应（Get Configuration Block Response）
func (this *AP) __UpdateConfigBlockResponse() {
	// Configuration Status: 配置文件状态
	this.Config_Status = this.rsDataDict["0009"]
	// 如果判断 配置文件状态情况
	if this.Config_Status == "00000000" || this.Config_Status == "00000001" { // 如果未配置 或 无变化 等待下个心跳再次请求
		this.Action = "00000001"
		time.Sleep(time.Duration(this.Device_State_Interval) * time.Second)
	} else if this.Config_Status == "00000003" || this.Config_Status == "00000005" { // 如果配置块存在更新 开始请求
		// print "//////Config_Status:%s" % this.Config_Status
		this.Config_Block_Index = int(hex_to_int(this.rsDataDict["000d"]))
		//fmt.Println("this.Config_Block_Index=", this.Config_Block_Index)
		this.Config_MD5 = hex_to_str(this.rsDataDict["0007"])
		// Config_Length = int(this.rsDataDict["000a"],16) // 长度无用 使用md5校验即可
		Config_Content := hex_to_str(this.rsDataDict["000b"])

		tmpMd5 := this.getStrMD5(Config_Content)
		// 检查接收的的值 如果校验不通过重新获取
		if this.Config_MD5 != tmpMd5 {
			this.logApPrint(fmt.Sprintf("err:获取配置失败，第%d块配置AC的MD5值（%v）与实际文件的MD5(%v)值不一样", this.Config_Block_Index, this.Config_MD5, tmpMd5))
			this.Action = "00000003" // 从新获取
			// this.Action = "00000001"  // 从新开始
			// this.Config_MD5 = "0" * 32
			// 设置 下次发送走 777
			this.APStatusCode = 7
			// time.sleep(1)
			return
		}
		//fmt.Printf("configBlockIndex=%v ,data=%v", this.Config_Block_Index, Config_Content)
		// 如果通过加入到Block 列表中
		this.Config_Block_Data[this.Config_Block_Index+1] = Config_Content

		this.Config_Block_Index += 1
		// 如果获取配置未完，则继续，
		if this.Config_Block_Index < this.Config_Blocks {
			// 继续请求数据
			this.Action = "00000003"
		} else { // 如果配置完成 则进行最后拼接和校验
			// 检查接收的的MD5值 忽略校验（正常应该校验），节省CPU 将拼接好的放在0位
			this.Config_Block_Data[0] = strings.Join(this.Config_Block_Data[1:], "")

			// print  this.Config_Block_Data[0]
			//ap_config_json = json.loads(this.Config_Block_Data[0])
			//ap_radio_list = ap_config_json["wireless"]["radio"]
			var ap_config_json map[string]interface{}
			json.Unmarshal([]byte(this.Config_Block_Data[0]), &ap_config_json)
			//fmt.Println("this.Config_Block_Data[0]:", this.Config_Block_Data[0])
			//fmt.Println("ap_config_json:", ap_config_json)
			ap_wireless := ap_config_json["wireless"].(map[string]interface{})
			//fmt.Println("ap_config_json[wireless]", ap_config_json["wireless"])
			ap_radio_list := ap_wireless["radio"].([]interface{})

			this.AP_PHY = 0   // 无线频道序号索引
			this.AP_VAP = 0   // 对应频道关联配置索引
			this.AP_SSID = "" // 对应频道关联配置的ssid名称

			// 获取频道 数组 可能配置多个 2.4G 或 5.8G
			//            for radio in ap_radio_list{
			//                // print radio
			//				ifindex,ok := radio["ifindex"]
			//                if ok {
			//                    this.AP_PHY = ifindex
			//                }
			//				vap,ok1 := radio["vap"]
			//				enabled,ok2 := radio["enabled"]
			//				// 如果radio 中存在关联配置 且 开启该频道，则使用默认的第一个配置
			//                if ok1 && ok2 {
			//                    // 默认使用该频段的 第一个 ssid 配置
			//                    this.AP_SSID = vap[this.AP_VAP]["essid"]
			//                    this.AP_VAP = vap[this.AP_VAP]["index"] // 通过索引获取vap
			//                    this.BILL_WAIT_TIME = radio["vap"][this.AP_VAP]["agentargs"]["accttimeout"]
			//                    this.BillWaitTime = this.BILL_WAIT_TIME
			//                    // print "获取计费时间：",this.BillWaitTime
			//                    break
			//				}
			//			}
			// 获取频道 数组 可能配置多个 2.4G 或 5.8G
			for _, rad := range ap_radio_list {
				radio := rad.(map[string]interface{})
				// print radio
				ifindex, ok := radio["ifindex"]
				if ok {
					this.AP_PHY = int64(ifindex.(float64))
				}
				vap, ok1 := radio["vap"]
				_, ok2 := radio["enabled"]
				// 如果radio 中存在关联配置 且 开启该频道，则使用默认的第一个配置
				if ok1 && ok2 {
					v := (vap.([]interface{}))[int(this.AP_VAP)].(map[string]interface{})

					// 默认使用该频段的 第一个 ssid 配置
					this.AP_SSID = string(v["essid"].(string))
					this.AP_VAP = int64(v["index"].(float64)) // 通过索引获取vap
					this.BILL_WAIT_TIME = int((v["agentargs"].(map[string]interface{}))["accttimeout"].(float64))
					this.BillWaitTime = this.BILL_WAIT_TIME
					// print "获取计费时间：",this.BillWaitTime
					break
				}
			}

			// print this.AP_SSID
			// 如果ssid 没有获取到，表示配置文件无效 则更新失败
			if len(this.AP_SSID) > 0 {
				//this.logApPrint(fmt.Sprintf("更新 AP配置文件成功!!!phy=%v,vap=%v,ssid=%v,hearttime=%v,accttimeout=%v",this.AP_PHY, this.AP_VAP, this.AP_SSID.encode("utf8"),this.HeartTime,this.BILL_WAIT_TIME))
				this.logApPrint(fmt.Sprintf("更新 AP配置文件成功!!!phy=%v,vap=%v,ssid=%v,hearttime=%v,accttimeout=%v", this.AP_PHY, this.AP_VAP, this.AP_SSID, this.HeartTime, this.BILL_WAIT_TIME))
			} else {
				this.logApErrPrint("AP配置文件被解除关联 !!! ssid is None ")
			}
			// 配置更新完成开始走 心跳 和配置检查
			this.APStatusCode = 5
			this.Action = "00000000"
			return
		}
	}
	// 设置 下次发送走 777
	this.APStatusCode = 7
	// time.sleep(1)
}

// Action: 5 AC-AP Update Configuration Request
func (this *AP) __updateConfigurationRequest() {
	// 正常已经被接收 直接发送响应保温桶
	this.Action = "00000006"
	this.APStatusCode = 7
	this.__updateConfigurationResponse()
}

// Action: 6  Update AP-AC Configuration Response
func (this *AP) __updateConfigurationResponse() {
	// AP 给AC发送响应请求
	sMsg := this.tlv.packDisc(this.updateHeaderMsg(), this.updateDataMsg())
	this.sendMsg(sMsg)
	// 设置下个code 和Action
	this.APStatusCode = 7
	// AP 动作 Action 1 请求配置
	this.Action = "00000001"
}

// Action 30 新终端请求（New    Station  Request
func (this *AP) __newStationRequest() {
	// print "Action 30 新终端请求"
	//   首次 创建 10个用户
	if len(this.APUsersList) < 1 {
		//fmt.Println("创建10AP用户")
		this.APUsersObject = NewAPUsers(this.AP_Index, hex_to_ip(this.AP_IP), hex_to_str(this.AP_MAC))
		this.APUsersList = this.APUsersObject.getAPUsersList()
	}
	// AP 上线用户的列表
	userIndex, _ := strconv.Atoi(this.APUsersList[0]["userIndex"])

	// 第一次上线需要移动索引
	if userIndex == 0 {
		this.APUsersList[0]["userIndex"] = "1"
	}

	// 重新获取 AP 上线用户的列表索引
	userIndex, _ = strconv.Atoi(this.APUsersList[0]["userIndex"])
	// 获取当前需要处理的用户
	user := this.APUsersList[userIndex]

	// print userIndex,len(this.APUsersList)
	// print this.APUsersList
	// 如果用户未上线 则上线 如果上线了则 继续下个
	if user["userStatus"] == "0" { // 如果用户未上线
		//pass
	} else if user["userStatus"] == "2" { // 如果用户已经上线
		userIndex += 1
		this.APUsersList[0]["userIndex"] = strconv.Itoa(userIndex)
	} else if user["userStatus"] == "1" { // 继续 AC action 授权 32
		this.Action = "00000020"
		return
	}
	sMsg := this.tlv.packDisc(this.updateHeaderMsg(), this.updateDataMsg())
	this.sendMsg(sMsg)
	// 设置下个code 和Action
	this.APStatusCode = 8
	// AP 动作 Action 31 上线响应
	this.Action = "0000001f"
}

// Action：31 新终端响应（New Station Response）
func (this *AP) __newStationResponse() {
	// print "Action：31 新终端响应"
	// print this.rsDataDict
	this.isPortalHeartStatus = false
	// AP 上线用户的列表
	userIndex, _ := strconv.Atoi(this.APUsersList[0]["userIndex"])

	//  如果上线成功 比较 返回的用户IP是否与发送的一致
	if this.APUsersList[userIndex]["userIP"] != hex_to_ip(this.rsDataDict["0032"]) {
		this.logApErrPrint(fmt.Sprintf("err:[%s]终端上线消息收到到错误返回的IP地址[%s],重新走心跳！！！！", this.APUsersList[userIndex]["userIP"], hex_to_ip(this.rsDataDict["0032"])))
		// 走心跳后进行重新上线  不等待
		this.Action = "0000001e"
		this.APStatusCode = 5
		this.isPortalHeartStatus = true
		return
	}

	// 根据返回结果更新用户状态 1：未认证 2：不用认证
	if hex_to_int(this.rsDataDict["003b"]) == 1 { // 如果登录返回结果是1:free 则直接设置状态为2 （跳过认证）
		this.APUsersList[userIndex]["userStatus"] = "2"
	} else if hex_to_int(this.rsDataDict["003b"]) == 2 { // 如果登录返回结果是2:free Limit 表示漫游 则直接设置状态为2 （跳过认证）
		this.APUsersList[userIndex]["userStatus"] = "2"
	} else if hex_to_int(this.rsDataDict["003b"]) == 3 { // 如果登录返回结果是3:portal 则直接设置状态为1 （需要走认证流程）
		this.APUsersList[userIndex]["userStatus"] = "1"
	} else {
		this.logApErrPrint(fmt.Sprintf("#########异常的认证类型rsDataDict[\"003b\"]=%d ", hex_to_int(this.rsDataDict["003b"])))
	}

	// print "len(this.APUsersList)",len(this.APUsersList) -1
	// 如果用户已经没有了，则上线完成
	if userIndex >= len(this.APUsersList)-1 {
		// 重置索引状态 为1 方便下次取用户
		this.APUsersList[0]["userIndex"] = "1"
		this.logApPrint(fmt.Sprintf("###########一共 %d 个用户上线完成(没有走认证流)", userIndex))
		// print "1111一共 %d 个用户上线完成(没有认证)" % userIndex
		time.Sleep(time.Duration(Portal_Online_To_Login_WaitTime) * time.Second)
		// 如果需要重新上线 并且不管portal 直接等待一段时间（不能大于心跳）后下线    -- 特殊测试流程
		if AP_User_Reline_Bool && AP_User_Ignore_Portal_Bool {
			// 然后开始记录当前时间为用户在线时间 清空离线时间
			// this.AP_User_OnLine_Time = time.Now().Uinx()
			// this.AP_User_OffLine_Time = 0
			// 直接下线等待时间不能超过10秒
			if int(AP_User_Ignore_Portal_Offline_WaitTime) <= this.HeartTime {
				this.logApPrint(fmt.Sprintf("##########用户直接等待%v秒后下线（请不要超过心跳时间%d秒)", AP_User_Ignore_Portal_Offline_WaitTime, this.HeartTime))
			} else {
				this.logApPrint(fmt.Sprintf("##########AP_User_Ignore_Portal_Offline_WaitTime=%v秒,请不要超过心跳时间%d秒,矫正等待时间为%d秒)", AP_User_Ignore_Portal_Offline_WaitTime, this.HeartTime, this.HeartTime))
				AP_User_Ignore_Portal_Offline_WaitTime = int64(this.HeartTime)
			}
			time.Sleep(time.Duration(AP_User_Ignore_Portal_Offline_WaitTime) * time.Second)
			if AP_User_Is_Left {
				//  // 直接踢用户
				this.APStatusCode = 7
				this.Action = "00000026"
			} else {
				// 开始走 新终端下线流程
				this.APStatusCode = 8
				this.Action = "00000024" // 先走portal踢用户 再离线
			}
			return
		}
		// 上线完成后开始 走 认证流程 （由认证流程判断用户是否需要登录） --- 正常流程
		// AP 上线用户的列表
		userIndex, _ := strconv.Atoi(this.APUsersList[0]["userIndex"])
		// 获取当前需要处理的用户
		user := this.APUsersList[userIndex]

		if user["userStatus"] == "1" { // 如果为 1：未认证  需要认证走认证
			this.Action = "00000020"
			this.APStatusCode = 8
			this.logApPrint("##########未认证，需要走登录认证流程，开始认证")
		} else if user["userStatus"] == "2" { // 否则2:不需要认证 走记账包
			// 然后开始记录当前时间为用户在线时间 清空离线时间
			if AP_User_Reline_Bool {
				this.AP_User_OnLine_Time = time.Now().Unix()
				this.AP_User_OffLine_Time = 0
			}
			this.Action = "00000028"
			this.APStatusCode = 5
			this.logApPrint("########## 不需要认证或认证完成，开始走记账包流程")
		} else {
			this.logApErrPrint("########## 异常，既不走认证完成也不走未认证")
		}
		return
	} else { // 继续上线下个用户

		this.APUsersList[0]["userIndex"] = strconv.Itoa(userIndex + 1)
		// print "继续上线第[%d]个用户" % this.APUsersList[0]
	}
	// 设置 下次发送走 777
	this.APStatusCode = 7
	// 准备开始 下个 用户状态数据请求
	this.Action = "0000001e"
	// time.sleep(WAITTIME)
}

//  Action：32 终端授权请求 AC -> AP（Station Access Request）
func (this *AP) __stationAccessRequest() {

	// AP 上线用户的列表
	userIndex, _ := strconv.Atoi(this.APUsersList[0]["userIndex"])
	// 获取当前需要处理的用户
	user := this.APUsersList[userIndex]

	//  如果登录成功 比较 返回的用户IP是否与发送的一致
	if user["userIP"] != hex_to_ip(this.rsDataDict["0032"]) {
		this.logApErrPrint(fmt.Sprintf("err:[%s]登录后接收到到错误返回的IP地址[%s]", user["userIP"], hex_to_ip(this.rsDataDict["0032"])))
		return
	}

	// 如果校验成功，发送收到结果
	// 设置 下次发送走 777
	this.APStatusCode = 7
	// 准备开始 下个 用户认证
	this.Action = "00000021"
	// time.sleep(0.5)
}

// Action：33 终端授权响应 AP -> AC
func (this *AP) __stationAccessResponse() {

	sMsg := this.tlv.packDisc(this.updateHeaderMsg(), this.updateDataMsg())
	this.sendMsg(sMsg)

	// AP 上线用户的列表
	userIndex, _ := strconv.Atoi(this.APUsersList[0]["userIndex"])

	// 根据返回结果更新用户状态 1：未认证 2：认证成功
	this.APUsersList[userIndex]["userStatus"] = "2"

	// 如果用户已经没有了，则上线完成
	if userIndex == len(this.APUsersList)-1 {
		// 重置索引状态 为1 方便下次取用户
		this.APUsersList[0]["userIndex"] = "1"
		this.logApPrint(fmt.Sprintf("#########一共 %d 个用户认证上线完成（认证完成）", userIndex))
		// time.sleep(this.HeartTime)
		// 认证完成后开始 走 记账包流程
		this.Action = "00000028"
		this.APStatusCode = 5
		if AP_User_Reline_Bool {
			// 然后开始记录当前时间为用户在线时间 清空离线时间
			this.AP_User_OnLine_Time = time.Now().Unix()
			this.AP_User_OffLine_Time = 0
		}
		return
	} else { // 继续上线下个用户
		this.APUsersList[0]["userIndex"] = strconv.Itoa(userIndex + 1)
	}
	// 设置 下次发送走 888
	this.APStatusCode = 8
	// 准备开始 下个 用户认证
	this.Action = "00000020"
	// time.sleep(0.5)
}

// Action：40 终端账单请求（Station Bill Request）
func (this *AP) __newStationBillRequest() {
	// print "Action 40 终端账单请求"

	// 第一次遍历记账包需要更新计费 和 等待剩余计费时间不足心跳时间的部分：
	if this.APUsersList[0]["userIndex"] == "1" {
		// 然后更新对用中的用户并进行计费更新操作
		this.APUsersList = this.APUsersObject.updateUsersListBill(this.APUsersList)
		// 记账时间小于心跳时间 且记账时间不等于0 则等待
		if this.HeartTime > this.BillWaitTime && this.BillWaitTime > 0 {
			time.Sleep(time.Duration(this.BillWaitTime) * time.Second)
		}
	}
	sMsg := this.tlv.packDisc(this.updateHeaderMsg(), this.updateDataMsg())
	this.sendMsg(sMsg)
	// 设置下个code 和Action
	this.APStatusCode = 8
	// AP 动作 Action 41 进行账单响应
	this.Action = "00000029"
}

// Action：41 终端账单响应（Station Bill Response）
func (this *AP) __newStationBillResponse() {
	// print "Action：41 终端账单响应"
	// print this.rsDataDict

	// AP 上线用户的列表
	userIndex, _ := strconv.Atoi(this.APUsersList[0]["userIndex"])

	// 如果用户账单已经已经遍历完成，则等待下次发送账单
	if userIndex == len(this.APUsersList)-1 {
		// 重置索引状态 为1 方便下次取用户
		this.APUsersList[0]["userIndex"] = "1"
		this.logApPrint(fmt.Sprintf("#########一共 %d 个用户更新记账完成", userIndex))

		// 重置状态
		this.BillWaitTime = this.BILL_WAIT_TIME

		// 上线完成后走心跳 由心跳判断是否等待 再回来，防止账单时间大于心跳时间 掉线
		this.Action = "00000028"
		this.APStatusCode = 5
		return
	} else { // 继续上线下个用户
		this.APUsersList[0]["userIndex"] = strconv.Itoa(userIndex + 1)
	}
	this.APStatusCode = 7
	// 准备开始 下个 用户记账请求
	this.Action = "00000028"
	// time.sleep(WAITTIME)
}

// Action：38 终端离开请求（Station Left Request）
func (this *AP) __stationLeftRequest() {
	// print "Action 38 端离开请求"

	// AP 上线用户的列表
	userIndex, _ := strconv.Atoi(this.APUsersList[0]["userIndex"])
	// 获取当前需要处理的用户
	user := this.APUsersList[userIndex]

	// 如果用户未上线 则不下线， 如果上线了则 下先 继续下个
	if user["userStatus"] == "0" { // 如果用户未上线
		userIndex += 1
		this.APUsersList[0]["userIndex"] = strconv.Itoa(userIndex)
	}
	if user["userStatus"] == "2" { // 如果用户已经上线
		//pass
	} else if user["userStatus"] == "1" { // 继续 AC action 授权 32
		//pass
	}

	sMsg := this.tlv.packDisc(this.updateHeaderMsg(), this.updateDataMsg())
	this.sendMsg(sMsg)
	// 设置下个code 和Action
	this.APStatusCode = 8
	// AP 动作 Action 39 下线响应
	this.Action = "00000027"
}

// Action：39 终端离开响应（Station Left Response）
func (this *AP) __stationLeftResponse() {
	// print "Action：39 终端离开响应"
	// print this.rsDataDict

	// AP 离线用户的列表
	userIndex, _ := strconv.Atoi(this.APUsersList[0]["userIndex"])
	// 清除用户的状态
	this.APUsersList[userIndex]["userStatus"] = "0"

	// 如果用户已经没有了，则离线完成
	if userIndex == len(this.APUsersList)-1 {
		// 重置索引状态 为1 方便下次取用户
		this.APUsersList[0]["userIndex"] = "1"
		this.logApPrint(fmt.Sprintf("##########一共 %d 个用户下线(left)完成", userIndex))
		// time.sleep(this.HeartTime)

		// 离线完成后走 心跳流程 （等待一定时间后重新开始上线）
		//userIndex, _ := strconv.Atoi(this.APUsersList[0]["userIndex"])

		// if this.APUsersList[userIndex]["userStatus"] == 0:  // 如果为 0 下线完成
		// 清除用户的状态 走心跳 开始倒计时
		this.APUsersList = []map[string]string{}
		this.APUsersObject = nil
		this.Action = "00000000"
		this.APStatusCode = 5
		//// 设置用户离线时间 开始记时离线时间 清理在西安时间
		if AP_User_Reline_Bool {
			this.AP_User_OnLine_Time = 0
			this.AP_User_OffLine_Time = time.Now().Unix()
		}
		return
	} else { // 继续离线下个用户
		this.APUsersList[0]["userIndex"] = strconv.Itoa(userIndex + 1)
	}
	// 设置 下次发送走 777
	this.APStatusCode = 7
	// 准备开始 下个 用户状态数据请求
	this.Action = "00000026"
	// time.sleep(WAITTIME)
}

//  Action：36 终端离开请求 AC -> AP（Station Drop http Request）
func (this *AP) __stationDropRequest() {

	userIndex, _ := strconv.Atoi(this.APUsersList[0]["userIndex"])
	//  如果登注销功 比较 返回的用户IP是否与发送的一致
	if this.APUsersList[userIndex]["userIP"] != hex_to_ip(this.rsDataDict["0032"]) {
		this.logApErrPrint(fmt.Sprintf("err:[%s]注销后接收到错误返回的IP地址[%s]", this.APUsersList[userIndex]["userIP"], hex_to_ip(this.rsDataDict["0032"])))
		return
	}
	// 如果校验成功，发送收到结果
	// 设置 下次发送走 777
	this.APStatusCode = 7
	// 给AC返回响应信息
	this.Action = "00000025"
	// time.sleep(0.5)
}

// Action：37 踢终端响应 AP -> AC（（Drop Station Response）
func (this *AP) __stationDropResponse() {

	sMsg := this.tlv.packDisc(this.updateHeaderMsg(), this.updateDataMsg())
	this.sendMsg(sMsg)

	// AP 上线用户的列表
	userIndex, _ := strconv.Atoi(this.APUsersList[0]["userIndex"])

	// 根据返回结果更新用户状态 1：已离线 2：认证成功
	this.APUsersList[userIndex]["userStatus"] = "1"

	// 如果用户已经没有了，则下线完成
	if userIndex == len(this.APUsersList)-1 {

		// 重置索引状态 为1 方便下次取用户
		this.APUsersList[0]["userIndex"] = "1"
		// print "一共 %d 个用户离线完成" % userIndex
		this.logApPrint(fmt.Sprintf("#######一共 %d 个用户注销(drop)完成", userIndex))
		// time.sleep(this.HeartTime)
		// 下线完成后开始 走 踢用户下线流程
		// exit(0)
		// this.Action = "00000026"
		// this.APStatusCode = 7
		// return
		//////////
		// 清除用户的状态 走心跳 开始倒计时
		this.APUsersList = []map[string]string{}
		this.APUsersObject = nil
		this.Action = "00000000"
		this.APStatusCode = 5
		//// 设置用户离线时间 开始记时离线时间 清理在线时间
		if AP_User_Reline_Bool {
			this.AP_User_OnLine_Time = 0
			this.AP_User_OffLine_Time = time.Now().Unix()
		}
		return
	} else { // 继续离线下个用户
		this.APUsersList[0]["userIndex"] = strconv.Itoa(userIndex + 1)
	}

	// 设置 下次发送走 888
	this.APStatusCode = 8
	// 准备开始 下个 用户离线
	this.Action = "00000024"
	// time.sleep(0.5)
}

// portal http 登录操作 如果失败一直重试
func (this *AP) __portalHttpLogin() bool {
	// print "portal http " params  retry=false
	// 用户的列表
	userIndex, _ := strconv.Atoi(this.APUsersList[0]["userIndex"])
	user := this.APUsersList[userIndex]

	// 调用AC 登录接口进行登录进行测试
	// http://172.16.0.5:9009/login?wlanacname=&wlanacip=172.16.0.5&wlanuserip=172.16.2.2&wlanusermac=b0:e2:35:f7:a7:aa&apid=153dd138-1dd2-11b2-b67b-2344671c0f08&apmac=00:19:3B:0D:3F:84&orgid=6D95A751-72DB-4E99-BAC9-DB7DFD97BDC7&brcid=0&ssid=2ac_lizhen&wlanuserfirsturl=111&username=111111&userpasswd=111111
	loginUrl := fmt.Sprintf("http://%s?wlanacname=&wlanacip=%s&wlanuserip=%s&wlanusermac=%s&apid=%s&apmac=%s&orgid=%s&brcid=0&ssid=%s&wlanuserfirsturl=111&username=%s&userpasswd=%s",
		Portal_Login_URL, AC_IP, user["userIP"], strings.Replace(user["userMac"], "-", ":", -1), hex_to_str(this.DEV_ID),
		hex_to_str(this.AP_MAC), ORGID, this.AP_SSID, Portal_Login_Account, Portal_Login_Passwd)
	// print loginUrl
	// 进行登录
	//this.logApPrint("#######用户开始进行登录")
	c := &http.Client{
		Timeout: time.Duration(Portal_Http_Timeout) * time.Second,
	}

	resp, err := c.Get(loginUrl)
	if !this.check(fmt.Sprintf("http login request error:[%s]异常的login  ", user["userIP"]), err) {
		// 失败 重新登录
		// 没有登录成功 先走一次心跳
		this.isPortalHeartStatus = true
		// 走心跳
		this.APStatusCode = 5
		return false
	}

	defer resp.Body.Close()
	body, err1 := ioutil.ReadAll(resp.Body)
	if !this.check(fmt.Sprintf("http login request Body [%s]异常 ", user["userIP"]), err1) {
		// 失败 重新登录
		// 没有登录成功 先走一次心跳
		this.isPortalHeartStatus = true
		// 走心跳
		this.APStatusCode = 5
		return false
	}
	//fmt.Println(string(body))
	var rsjson map[string]interface{}
	err2 := json.Unmarshal(body, &rsjson)
	if !this.check(fmt.Sprintf("http login response rsjson error [%s]异常的loginResult   ", user["userIP"]), err2) { //
		// 失败 重新登录
		// 没有登录成功 先走一次心跳
		this.isPortalHeartStatus = true
		// 走心跳
		this.APStatusCode = 5
		return false
	}

	//fmt.Println("1", rsjson["errorCode"])
	//fmt.Println("2", rsjson["errorMessage"])
	errorCode := int(rsjson["errorCode"].(float64))

	// 检查返回的登录结果
	if errorCode != 0 {
		// print rs.status_code
		this.logApErrPrint(fmt.Sprintf("err:[%v]错误的login结果：%v", user["userIP"], string(body)))
		// time.sleep(0.5)
		// return this.__portalHttpLogin()
		// 没有登录成功 先走一次心跳
		this.isPortalHeartStatus = true

		time.Sleep(time.Duration(Portal_LoginErr_WaitTime) * time.Second)
		// 如果登录失败一次后等待一个登录间隔再次尝试
		if errorCode == 10 || errorCode == 12 || errorCode == 14 {
			// 扔掉失败的失败的包
			this.recvMsg(true)
		}
		return false
	}
	this.logApPrint("###########portal(http)返回用户登录成功！！")
	// 登录成功 重置心跳状态
	this.isPortalHeartStatus = false
	return true
}

// portal http 注销操作 并返回结结果
func (this *AP) __portalHttpLogout() bool {
	// 用户的列表获取
	userIndex, _ := strconv.Atoi(this.APUsersList[0]["userIndex"])
	user := this.APUsersList[userIndex]

	// 调用AC退出进行logout进行测试
	// http://172.16.0.5:9009/logout?wlanacname=&wlanacip=172.16.0.5&wlanuserip=172.16.2.2&wlanusermac=b0:e2:35:f7:a7:aa&apid=153dd138-1dd2-11b2-b67b-2344671c0f08&apmac=00:19:3B:0D:3F:84&orgid=6D95A751-72DB-4E99-BAC9-DB7DFD97BDC7&brcid=0&ssid=2ac_lizhen&wlanuserfirsturl=111&username=111111&userpasswd=111111
	logoutUrl := fmt.Sprintf("http://%s?wlanacname=&wlanacip=%s&wlanuserip=%s&wlanusermac=%s&apid=%s&apmac=%s&orgid=%s&brcid=0&ssid=%s&wlanuserfirsturl=111&username=%s&userpasswd=%s",
		Portal_Logout_URL, AC_IP, user["userIP"], strings.Replace(user["userMac"], "-", ":", -1), hex_to_str(this.DEV_ID),
		hex_to_str(this.AP_MAC), ORGID, this.AP_SSID, Portal_Login_Account, Portal_Login_Passwd)

	this.logApPrint("#######用户开始进行登录")
	c := &http.Client{
		Timeout: time.Duration(Portal_Http_Timeout) * time.Second,
	}

	resp, err := c.Get(logoutUrl)
	if !this.check(fmt.Sprintf("http logout error [%s]异常: ", user["userIP"]), err) {
		// 失败 重新登录
		// 没有登录成功 先走一次心跳
		this.isPortalHeartStatus = true
		// 走心跳
		this.APStatusCode = 5
		return false
	}

	defer resp.Body.Close()
	body, err1 := ioutil.ReadAll(resp.Body)
	if !this.check(fmt.Sprintf("http logout request Body [%s]异常:  ", user["userIP"]), err1) {
		// 失败 重新登录
		// 没有注销成功 先走一次心跳
		this.isPortalHeartStatus = true
		// 走心跳
		this.APStatusCode = 5
		return false
	}
	//fmt.Println(string(body))
	var rsjson map[string]interface{}
	err2 := json.Unmarshal(body, &rsjson)
	if !this.check(fmt.Sprintf("http login response rsjson error):[%s]异常的logoutResult   ", user["userIP"]), err2) { //
		// 失败 重新登录
		// 没有注销成功 先走一次心跳
		this.isPortalHeartStatus = true
		// 走心跳
		this.APStatusCode = 5
		return false
	}

	//fmt.Println("1", rsjson["errorCode"])
	//fmt.Println("2", rsjson["errorMessage"])
	errorCode := int(rsjson["errorCode"].(float64))

	// 14 已经不在线了
	if errorCode != 0 || errorCode == 14 {
		// print rs.status_code
		this.logApErrPrint(fmt.Sprintf("err:[%v]错误的logout结果：%v", user["userIP"], string(body)))
		// 没有logout成功 先走一次心跳
		this.isPortalHeartStatus = true
		// 走心跳
		this.APStatusCode = 5
		return false
	}
	// 注销成功 重置心跳状态
	this.isPortalHeartStatus = false
	return true
}

////// 发送和接收UDP消息
// 发送并接收数据 code 1 3 5
func (this *AP) sendMsg(data string) {
	// code AP 向AC发送请求
	//fmt.Println("data=", data)
	//这是发送消息对列死亡时间
	this.udpCliSock.SetReadDeadline(time.Now().Add(time.Duration(WAITTIME) * time.Second))
	_, err := this.udpCliSock.Write(hex_to_byte(data))
	// 如果发送失败则打印错误，否则打印发送成功
	if this.check("Send messge error!", err) {
		// 只记录 从心跳后开始的
		// this.logApPrint("send messge success!")
		if this.APStatusCode >= 0 {
			this.logApPrint("Send messge success!")
		}
	}
}

// 接收消息  code 2 4 6        lost=false 表示异常消息可以仍掉接收结果
func (this *AP) recvMsg(lost bool) bool {
	//fmt.Println("recv-1111")
	buff := make([]byte, BUFSIZE)

	//这是接收消息对列死亡时间
	this.udpCliSock.SetReadDeadline(time.Now().Add(time.Duration(WAITTIME) * time.Second))
	_, err := this.udpCliSock.Read(buff[0:])

	//fmt.Println("recv-222")
	var errStr string
	if lost {
		errStr = "No packet recved - lost=true,err="
	} else {
		errStr = "No packet recved ,err="
	}
	//检查接收错误的
	if !this.check(errStr, err) {
		return false
	}

	//this.logApPrint("Receive data success! ")

	//fmt.Println("recv-555")
	// print "recv-ccc"
	//fmt.Println("buff-byte=", string(buff))
	//fmt.Println("buff-hex=", byte_to_hex(buff))
	// print "R:",tlv_unpack(rcData)

	//	var rsHeaderCode int
	//	var rsErrorCode string
	//	var rsDataDict map[string]string

	//fmt.Println("recv-aaa")
	// TLV进行解包
	rsHeaderCode, rsErrorCode, rsDataDict := this.tlv.unpackDict(byte_to_hex(buff))

	//fmt.Println("recv-bbb")
	//this.logApPrint(fmt.Sprintf("Receive data success! return rsHeaderCode=%d ,rsErrorCode=%v,rsDataDict=%v ", rsHeaderCode, this.rsErrorCode, rsDataDict))

	// 只记录 从心跳开始的
	// this.logApPrint("Receive data success! ")
	// 如果需要丢掉
	if lost {
		this.logApPrint(fmt.Sprintf("Receive data success! --lost rsDataDict:%v", rsDataDict))
		return true
	}
	this.rsHeaderCode, this.rsErrorCode, this.rsDataDict = rsHeaderCode, rsErrorCode, rsDataDict
	if this.APStatusCode >= 1 {
		this.logApPrint("Receive data success!") // ;print this.rsDataDict
	}
	// 判断返回码是否正确
	if this.rsHeaderCode != this.APStatusCode {
		//fmt.Println("recv-ddd")
		//如果返回吗不正确，则从头开始
		if this.APStatusCode <= 4 { // 1-4 出现错误 重新重1开始
			this.APStatusCode = 1
			//fmt.Println("recv-eeee")
			return false
		} else if this.rsHeaderCode == 6 { // 如果是心跳继续心跳
			//fmt.Println("recv-ffff")
			return true
		} else if this.rsHeaderCode == 7 || this.rsHeaderCode == 8 {
			r := this.rsDataDict["0001"]
			// 忽略 是AC提示来要更新 配置 必须给回应要不然一直发更新配置
			// {"Code": 7, "0001": "00000005", "SeqNum": 0, "PayLoadType": 2, "SessionID": "00000001","0014": "00000000"}
			if r == "00000005" {
				this.logApPrint("接收到到AC配置更新请求，准备更新配置")
				// rsHeaderCode,rsErrorCode,rsDataDict =this.rsHeaderCode, this.rsErrorCode, this.rsDataDict
				// 丢掉上请求的一个结果
				this.recvMsg(true)
				// 走更新配置流程
				this.APStatusCode = 5
				this.Action = "00000001"
				// fmt.Println( "recv-iiii")
				// this.rsHeaderCode, this.rsErrorCode, this.rsDataDict = rsHeaderCode,rsErrorCode,rsDataDict
				return true
			} else if r == "00000020" || r == "00000021" || r == "00000024" || r == "00000025" { // 忽略 因为他发的是反的 自动忽略
				// fmt.Println( "recv-gggg")
				return true
			}
		}
		this.logApErrPrint(fmt.Sprintf("err:警告!错误的返回值headerCode(%v) != APStatusCode(%v),接串了重新走向心跳！！！rsDataDict=%v", this.rsHeaderCode, this.APStatusCode, this.rsDataDict))
		// print this.rsDataDict
		// this.sendHeaderDict["sessionID"] = this.rsDataDict["SessionID"]
		// 如果是 5-8  心跳错误 或 AP状态错误重新走
		this.APStatusCode = 5
		// this.Action = "00000000"
		// fmt.Println( "recv-jjjj")
		return false
	}
	//fmt.Println("recv-KKKK") //,this.sendHeaderDict["sessionID"] ,this.rsDataDict["SessionID"]
	// this.APStatusCode=4 时 系统分配sessionID 号：
	if this.APStatusCode == 4 {
		if this.sendHeaderDict["sessionID"] != this.rsDataDict["SessionID"] {
			this.sendHeaderDict["sessionID"] = this.rsDataDict["SessionID"]
		} else { //如果没有分配则
			this.logApErrPrint("err:警告!分配错误的sesssion")
		}
	}
	return true
}

// 更新消息的head
func (this *AP) updateHeaderMsg() map[string]string {
	// 拼接头部分
	//    协议头如下图所示
	//        0                   1                   2                   3
	//        0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
	//       +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	//       |    Version    |      Type     |             Length            |
	//       +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	//       |                           Session ID                          |
	//       +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	//	   |          SEQ NUM              |     Code      | Payload Type  |
	//       +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	// 01:00:00:cf:   Version:1 type:0 length:207
	// 00:00:00:01:   sessionID
	// 27:dd:07:02:   SEQ Num   code  payload Type
	// 把sessionID SEQ NUM Code Payload Type 看作msg前面的一部分：
	// this.sendHeaderDict["sessionID"] = ip_to_Hex("172.16.0.61")
	// 断开后 session 需要重新获取

	if this.APStatusCode == 1 {
		this.sendHeaderDict["sessionID"] = int_to_hex(0, 8) // sessionID 0的情况下发现使用自动分配的
	}
	this.sendHeaderDict["seqNum"] = int_to_hex(int64(this.counter), 4)
	// 计数器加1   如果大于 65535 则 重新开始  否则会越界
	if this.counter < 65535 {
		this.counter += 1
	} else {
		this.counter = 1
	}
	// print this.counter
	// 33  AP->AC 终端授权响应（Station Access Response））登录结果返回需要是8
	// 37  AP->AC 踢终端响应（Drop Station Response））登录结果返回需要是8
	if this.APStatusCode == 7 && (this.Action == "00000021" || this.Action == "00000025") {
		this.sendHeaderDict["code"] = int_to_hex(8, 2)
	} else { // 其他的正常
		this.sendHeaderDict["code"] = int_to_hex(int64(this.APStatusCode), 2) // Code : 1 表示 Discover Request
	}

	return this.sendHeaderDict
}

func (this *AP) updateDataMsg() map[string]string {
	// 要发送的
	tmpDict := make(map[string]string)

	// 根据AP发送状态进行 组合数据
	switch this.APStatusCode {
	case 1:
		tmpDict = this.setValue(tmpDict)
	case 3: // 1和3的内容一只
		tmpDict = this.setValue(tmpDict)
	case 5: // 5 的状态只有 DEV_ID 和 AP_IP
		tmpDict["0001"] = this.DEV_ID // DEV ID
		tmpDict["0007"] = this.AP_IP  // Device Local Address
	case 7: // 7 表示开始发送给ACtion信息
		tmpDict["0001"] = this.Action // Action
		switch this.Action {
		// code 7的格式
		case "00000001":
			tmpDict["0002"] = this.DEV_ID                 // DEV ID
			tmpDict["0003"] = this.AP_MAC                 // Device Hardware Address
			tmpDict["0004"] = this.AP_IP                  // Device Local Address
			tmpDict["0005"] = this.DEV_BoardName          // Board Name
			tmpDict["0011"] = this.DEV_ProductName        // Product Name
			tmpDict["0006"] = this.DEV_FirmwareVersion    // Firmware Version
			tmpDict["0010"] = this.DEV_CountryCode        // Country Code
			tmpDict["0007"] = str_to_hex(this.Config_MD5) // BRC ID str_to_hex(this.Config_MD5,32)
		case "00000003", "00000005": // 存在配置文件
			tmpDict["0002"] = this.DEV_ID                                   // DEV ID
			tmpDict["0003"] = this.AP_MAC                                   // Device Hardware Address
			tmpDict["0005"] = this.DEV_BoardName                            // Board Name
			tmpDict["0011"] = this.DEV_ProductName                          // Product Name
			tmpDict["0006"] = this.DEV_FirmwareVersion                      // Firmware Version
			tmpDict["000d"] = int_to_hex(int64(this.Config_Block_Index), 8) // Firmware Version int_to_Hex(this.Config_Block_Index,8)
		case "00000006": // 6 更新配置响应（Update Configuration Response）
			tmpDict["0002"] = this.DEV_ID // DEV ID
			tmpDict["0003"] = this.AP_MAC // Device Hardware Address
		case "00000007": // 7 设备状态数据请求（Device State Data request）
			tmpDict["0002"] = this.DEV_ID // DEV ID
			tmpDict["0003"] = this.AP_MAC // Device Hardware Address
			// tmpDict["001e"] = this.DataCompress  // Data Compress Class： 数据压缩类型
			// tmpDict["000f"] = this.DeviceStateData  // Device State Data: 设备状态数据
		default:
			index, _ := strconv.Atoi(this.APUsersList[0]["userIndex"])
			user := this.APUsersList[index]
			switch this.Action {
			case "0000001e": // 30 新终端请求（New Station Request）
				tmpDict["0002"] = this.DEV_ID                                   // DEV ID
				tmpDict["0003"] = this.AP_MAC                                   // Device Hardware Address
				tmpDict["0032"] = ip_to_hex(user["userIP"])                     // Station IP Address： 终端IP地址
				tmpDict["0034"] = strings.Replace(user["userMac"], "-", "", -1) // Station Hardware Address 6字节
				tmpDict["0051"] = int_to_hex(int64(IF_TYPE), 8)                 //int_to_Hex(IF_TYPE,8)
				tmpDict["0035"] = int_to_hex(int64(this.AP_PHY), 8)             //int_to_Hex(this.AP_PHY ,8)
				tmpDict["0036"] = int_to_hex(int64(this.AP_VAP), 8)             //int_to_Hex(this.AP_VAP,8)
				tmpDict["0037"] = str_to_hex(this.AP_SSID)                      //str_to_Hex(this.AP_SSID)
				tmpDict["0038"] = str_to_hex(InterfaceName)                     //str_to_Hex(InterfaceName)
			case "00000021": // 33  AP->AC 终端授权响应（Station Access Response）
				tmpDict["0002"] = this.DEV_ID                                   // DEV ID
				tmpDict["0003"] = this.AP_MAC                                   // Device Hardware Address
				tmpDict["0032"] = ip_to_hex(user["userIP"])                     // Station IP Address： 终端IP地址
				tmpDict["0034"] = strings.Replace(user["userMac"], "-", "", -1) // Station Hardware Address
				tmpDict["0051"] = int_to_hex(int64(IF_TYPE), 8)
				tmpDict["0035"] = int_to_hex(int64(this.AP_PHY), 8)
				tmpDict["0036"] = int_to_hex(int64(this.AP_VAP), 8)
				tmpDict["004e"] = int_to_hex(0, 8) //Station Error Code
			case "00000028": // 40 终端账单请求（Station Bill Request）
				tmpDict["0002"] = this.DEV_ID                                   // DEV ID
				tmpDict["0003"] = this.AP_MAC                                   // Device Hardware Address
				tmpDict["0032"] = ip_to_hex(user["userIP"])                     // Station IP Address： 终端IP地址
				tmpDict["0034"] = strings.Replace(user["userMac"], "-", "", -1) // Station Hardware Address 6字节
				tmpDict["0051"] = int_to_hex(int64(IF_TYPE), 8)
				tmpDict["0035"] = int_to_hex(int64(this.AP_PHY), 8)
				tmpDict["0036"] = int_to_hex(int64(this.AP_VAP), 8)
				// 注意时间需要转换 用户默认记录的是开始时间，session时间需要每次用当前时间减去开始时间 单位秒
				tmpDict["0047"] = int_to_hex(time.Now().Unix()-hex_to_int(user["userBillSession"]), 16) // Station - Bill - Session = 4094
				tmpDict["0048"] = user["userBillInputBytes"]                                            // Station - Bill - Input - Bytes = 14484238
				tmpDict["0049"] = user["userBillInputPackets"]                                          // Station - Bill - Input - Packets = 23022
				tmpDict["004a"] = user["userBillOutputBytes"]                                           // Station - Bill - Output - Bytes = 4535823
				tmpDict["004b"] = user["userBillOutputPackets"]                                         // Station - Bill - Output - Packets = 36855
				tmpDict["004c"] = user["userBillInputRate"]                                             // Station - Bill - Input - Rate = 741
				tmpDict["004d"] = user["userBillOutRate"]                                               // Station - Bill - Output - Rate = 881
			case "00000025": // 37  AP->AC 踢终端响应（Drop Station Response）
				tmpDict["0002"] = this.DEV_ID                                   // DEV ID
				tmpDict["0003"] = this.AP_MAC                                   // Device Hardware Address
				tmpDict["0032"] = ip_to_hex(user["userIP"])                     // Station IP Address： 终端IP地址
				tmpDict["0034"] = strings.Replace(user["userMac"], "-", "", -1) // Station Hardware Address 6字节
				tmpDict["0051"] = int_to_hex(int64(IF_TYPE), 8)
				tmpDict["0035"] = int_to_hex(int64(this.AP_PHY), 8)
				tmpDict["0036"] = int_to_hex(int64(this.AP_VAP), 8)
				// 注意时间需要转换 用户默认记录的是开始时间，session时间需要每次用当前时间减去开始时间 单位秒
				tmpDict["004e"] = int_to_hex(0, 8)                                                      // Station Error Code
				tmpDict["0047"] = int_to_hex(time.Now().Unix()-hex_to_int(user["userBillSession"]), 16) // Station - Bill - Session = 4094
				tmpDict["0048"] = user["userBillInputBytes"]                                            // Station - Bill - Input - Bytes = 14484238
				tmpDict["0049"] = user["userBillInputPackets"]                                          // Station - Bill - Input - Packets = 23022
				tmpDict["004a"] = user["userBillOutputBytes"]                                           // Station - Bill - Output - Bytes = 4535823
				tmpDict["004b"] = user["userBillOutputPackets"]                                         // Station - Bill - Output - Packets = 36855
				tmpDict["004c"] = user["userBillInputRate"]                                             // Station - Bill - Input - Rate = 741
				tmpDict["004d"] = user["userBillOutRate"]                                               // Station - Bill - Output - Rate = 881
			case "00000026": // Action：38 终端离开请求（Station Left Request）
				tmpDict["0002"] = this.DEV_ID                                   // DEV ID
				tmpDict["0003"] = this.AP_MAC                                   // Device Hardware Address
				tmpDict["0032"] = ip_to_hex(user["userIP"])                     // Station IP Address： 终端IP地址
				tmpDict["0034"] = strings.Replace(user["userMac"], "-", "", -1) // Station Hardware Address 6字节
				tmpDict["0051"] = int_to_hex(int64(IF_TYPE), 8)
				tmpDict["0035"] = int_to_hex(int64(this.AP_PHY), 8)
				tmpDict["0036"] = int_to_hex(int64(this.AP_VAP), 8)
				// 注意时间需要转换 用户默认记录的是开始时间，session时间需要每次用当前时间减去开始时间 单位秒
				tmpDict["0047"] = int_to_hex(time.Now().Unix()-hex_to_int(user["userBillSession"]), 16) // Station - Bill - Session = 4094
				tmpDict["0048"] = user["userBillInputBytes"]                                            // Station - Bill - Input - Bytes = 14484238
				tmpDict["0049"] = user["userBillInputPackets"]                                          // Station - Bill - Input - Packets = 23022
				tmpDict["004a"] = user["userBillOutputBytes"]                                           // Station - Bill - Output - Bytes = 4535823
				tmpDict["004b"] = user["userBillOutputPackets"]                                         // Station - Bill - Output - Packets = 36855
			default:
				this.logApErrPrint(fmt.Sprintf("err:错误的Action[1]=%v", this.Action))
			}
		}
	default:
		this.logApErrPrint(fmt.Sprintf("err:意外退出：错误的APcode[2]=%v", this.APStatusCode))
		// exit(0)
	}
	this.sendDataDict = tmpDict
	return this.sendDataDict
}

// code 1-6 的数据格式 公共设置部分
func (this *AP) setValue(tmpDict map[string]string) map[string]string {
	tmpDict["0001"] = this.DEV_ID              // DEV ID
	tmpDict["0002"] = this.DEV_Class           // Device Class
	tmpDict["0003"] = this.DEV_Vendor          // Device Vendor
	tmpDict["0004"] = this.AP_MAC              // Device Hardware Address
	tmpDict["0007"] = this.AP_IP               // Device Local Address
	tmpDict["0009"] = this.DEV_BoardName       // Board Name
	tmpDict["000c"] = this.DEV_ProductName     // Product Name
	tmpDict["000a"] = this.DEV_FirmwareVersion // Firmware Version
	tmpDict["000b"] = this.DEV_CountryCode     // Country Code
	tmpDict["000d"] = str_to_hex(ORGID)        // ORG ID
	tmpDict["000e"] = this.DEV_BRCID           // BRC ID
	return tmpDict
}

//////工具类
// 获取字符串md5
func (this *AP) getStrMD5(strData string) string {
	myMd5 := md5.New()
	myMd5.Write([]byte(strData))
	cipherStr := myMd5.Sum(nil)
	return hex.EncodeToString(cipherStr)
}

//用于检查异常异常错误，并输出到文件
func (this *AP) check(msg string, e error) bool {
	if e != nil {
		//panic(e)
		//this.logApErrPrint(msg)
		this.logApErrPrint(msg + error.Error(e))
		return false
	} else {
		//fmt.Println("echo err OK")
		return true
	}
}

// 写错误日志到日志
func (this *AP) logApErrPrint(str string) {
	writeFile(this.logApPrint(str), fmt.Sprintf("err-%d-%s.log", this.AP_Index, hex_to_str(this.AP_MAC)))
}

// 打印debug日志
//func (this *AP) logApPrint(msg string, isdebug bool) string {
//	// 如果没有开启debug日志打印  且 是debug日志 都不打印
//	if DEBUG_LOG_BOOL && isdebug:
//	    return
// 打印debug日志
func (this *AP) logApPrint(msg string) string {
	var pStr string
	if this.logStatus {
		if len(this.rsDataDict) > 0 {
			if this.APStatusCode == 7 || this.APStatusCode == 8 {
				if len(this.APUsersList) > 0 {
					index, _ := strconv.Atoi(this.APUsersList[0]["userIndex"])
					userip, ok := this.APUsersList[index]["userIP"]
					if ok { //this.Action in ("00000028","00000029"):
						pStr = fmt.Sprintf("runtime:%d-1-PID:%d-APID:[%d]-sessionid:%d-status:%d-APAction:%s-clientIP:%s-%s ",
							(time.Now().Unix() - this.startTime.Unix()), os.Getpid(), this.AP_Index, hex_to_int(this.rsDataDict["SessionID"]),
							this.APStatusCode, this.Action, userip, msg)
					}
				} else {
					pStr = fmt.Sprintf("runtime:%d-2-PID:%d-APID:[%d]-sessionid:%d-status:%d-APAction:%s-%s ",
						(time.Now().Unix() - this.startTime.Unix()), os.Getpid(), this.AP_Index, hex_to_int(this.rsDataDict["SessionID"]),
						this.APStatusCode, this.Action, msg)
				}
			} else {
				pStr = fmt.Sprintf("runtime:%d-3-PID:%d-APID:[%d]-sessionid:%d-status:%d-%s ",
					(time.Now().Unix() - this.startTime.Unix()), os.Getpid(), this.AP_Index, hex_to_int(this.rsDataDict["SessionID"]),
					this.APStatusCode, msg)
			}
		} else if this.APStatusCode != 0 {
			pStr = fmt.Sprintf("runtime:%d-4-PID:%d-APID:[%d]-sessionid:%d-status:%d-APAction:%v-%v ",
				(time.Now().Unix() - this.startTime.Unix()), os.Getpid(), this.AP_Index, hex_to_int(this.sendHeaderDict["sessionID"]),
				this.APStatusCode, this.Action, msg)
		} else {
			pStr = fmt.Sprintf("runtime:%d-5-PID:%d-APID:[%d]-sessionid:%d-status:%d-%s ", (time.Now().Unix() - this.startTime.Unix()),
				os.Getpid(), this.AP_Index, hex_to_int(this.sendHeaderDict["sessionID"]), this.APStatusCode, msg)
		}
	}
	//time.Sleep(1 * 1e9)
	logPrint(pStr)
	return pStr
}
