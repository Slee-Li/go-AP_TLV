package main

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/widuu/goini"
)

var wg sync.WaitGroup //定义一个同步等待的组
const VERSION = "V1.0.2"

//加载配置文件中的参数
var conf = goini.SetConfig("ap.ini")

//配置文件加载全局参数
var (
	//启动AP的数量       最大255*255 建议一次100个   如果传入参数则失效
	AP_NUM, _ = strconv.Atoi(conf.GetValue("AP", "AP_NUM"))
	//启动AP开始的Index  每个AP的Index 必须唯一  最大255*255  如果传入参数则失效
	AP_Begin_Index, _ = strconv.Atoi(conf.GetValue("AP", "AP_Begin_Index"))
	//AP 每秒上线数量   默认0的情况下表示不需要暂停
	AP_PER_NUM, _ = strconv.Atoi(conf.GetValue("AP", "AP_PER_NUM"))
	//每个AP下用户的数量 不能大于255个
	AP_USER_NUM, _ = strconv.Atoi(conf.GetValue("AP", "AP_USER_NUM"))
	// 设置AP掉线后是否自动重新上线  默认不开启，方便长时间测试后查看进程的状态
	AP_Reline_Status, _ = strconv.ParseBool(conf.GetValue("AP", "AP_Reline_Status"))
	// AP 重新上线间隔是时间  (注意不是用户)
	AP_Reline_WaitTime, _ = strconv.Atoi(conf.GetValue("AP", "AP_Reline_WaitTime"))
	// 初始化AP的 的IP头  MAC头 和 设备ID 头
	IPStr      = conf.GetValue("AP", "IPStr") //"172.%d.%d.1"
	MACStr     = conf.GetValue("AP", "MACStr")
	DEVIDStr   = conf.GetValue("AP", "DEVIDStr")
	BUFSIZE, _ = strconv.Atoi(conf.GetValue("AP", "BUFSIZE"))
	// 连接AC的参数
	//AC_IP   = "47.100.64.143"
	//AC_Port = "10086"
	//ORGID   = "3AB67B2B-B81F-430E-9917-A298E970DB24"
	AC_IP   = conf.GetValue("AP", "AC_IP")   //"192.68.88.18"
	AC_Port = conf.GetValue("AP", "AC_Port") //"10086"
	ORGID   = conf.GetValue("AP", "ORGID")   //"3F7375DE-F58C-405D-AF2B-A41651FEDBC0"

	//AP关联的配置文件 （默认读取）以下参数需要保持与数据aps_config_interface 中对应的字段
	IF_TYPE, _    = strconv.Atoi(conf.GetValue("AP", "IF_TYPE")) // 0：有线  1：无线 （基本上都是无线） 直接写死1
	InterfaceName = conf.GetValue("AP", "InterfaceName")         //"eth0"
	// 发送消息和节后消息等待时间 (正常3秒，冗余1秒 共4)
	WAITTIME, _ = strconv.Atoi(conf.GetValue("AP", "WAITTIME"))

	// AP 硬件型号
	DEV_BoardName = conf.GetValue("AP", "DEV_BoardName") //"FWBD-2800
	// AP 产品名称
	DEV_ProductName = conf.GetValue("AP", "DEV_ProductName") //"Netshell"
	// 设备类型 设备类型整数值，取值如下：1 – AP	2 – AC 3 – ROUTER 4 – SWITCH 5 – NAC 6 – PLC MASTER 7 – PLC SLAVE 8 – CPE
	DEV_Class, _ = strconv.ParseInt(conf.GetValue("AP", "DEV_Class"), 10, 10)
	// 厂商ID
	DEV_Vendor, _ = strconv.ParseInt(conf.GetValue("AP", "DEV_Vendor"), 10, 10)
	// 固件版本
	DEV_FirmwareVersion = conf.GetValue("AP", "DEV_FirmwareVersion")
	// 国家代码
	DEV_CountryCode = conf.GetValue("AP", "DEV_CountryCode")
	// BRC ID
	DEV_BRCID = conf.GetValue("AP", "DEV_BRCID")

	/////////AC 登录/注销 认证相关信息 是否登录目前是使用portal的配置
	// 登录账号密码
	Portal_Login_Account = conf.GetValue("AP", "Portal_Login_Account")
	// 登录账号密码
	Portal_Login_Passwd = conf.GetValue("AP", "Portal_Login_Passwd")
	// 登录URL地址  %s:9009/login
	Portal_Login_URL = strings.Replace(conf.GetValue("AP", "Portal_Login_URL"), "%s", AC_IP, -1)
	// 注销URL地址  %s:9009/logout
	Portal_Logout_URL = strings.Replace(conf.GetValue("AP", "Portal_Logout_URL"), "%s", AC_IP, -1)
	// http 请求超时时间  (正常3秒，冗余0.5秒)
	Portal_Http_Timeout, _ = strconv.ParseInt(conf.GetValue("AP", "Portal_Http_Timeout"), 10, 10)
	// portal 如果在登录中等待下次重试时间 1  注意考虑 Portal_Http_Timeout 的时间
	Portal_LoginErr_WaitTime, _ = strconv.ParseInt(conf.GetValue("AP", "Portal_LoginErr_WaitTime"), 10, 10)
	// portal 虚拟用户上线后 在开始登录前等待的时间 (模拟终端上线弹出portal后 到 点击登录按钮的时间间隔)
	Portal_Online_To_Login_WaitTime, _ = strconv.ParseInt(conf.GetValue("AP", "Portal_Online_To_Login_WaitTime"), 10, 10)

	///////// 设置AP中的用户是否自动上下线
	// 设置虚拟终端用户是否需要上线或显现操作
	AP_User_Reline_Bool, _ = strconv.ParseBool(conf.GetValue("AP", "AP_User_Reline_Bool")) //True: 上线一算时间后下线，False ：上线后一直在线不下
	// 设置虚拟终端在上线后是否需要根据返回结果进行portal认证 False 表示 正常流程
	AP_User_Ignore_Portal_Bool, _ = strconv.ParseBool(conf.GetValue("AP", "AP_User_Ignore_Portal_Bool"))
	// 设置 忽略portal后直下线间隔时间    （不能大于心跳）
	AP_User_Ignore_Portal_Offline_WaitTime, _ = strconv.ParseInt(conf.GetValue("AP", "AP_User_Ignore_Portal_Offline_WaitTime"), 10, 10)
	// 下线方式 drop 还是left 正常是 left  表示用户离开了
	AP_User_Is_Left, _ = strconv.ParseBool(conf.GetValue("AP", "AP_User_Is_Left"))
	// 用户上线后最大在线时间后开下线用户  (请不要小于记账时间)
	AP_User_Online_WaitTime, _ = strconv.ParseInt(conf.GetValue("AP", "AP_User_Online_WaitTime"), 10, 10)
	// 用户下线后再次 上线间隔时间    不要小于心跳时间
	AP_User_Reline_WaitTime, _ = strconv.ParseInt(conf.GetValue("AP", "AP_User_Reline_WaitTime"), 10, 10)
)

func work(wg *sync.WaitGroup, index int) {
	// 将设备ID补全
	dev_id := fmt.Sprintf(DEVIDStr, int_to_hex(int64(index), 6)) // 设备ID使用16进制
	ap_mac := int_to_mac(index, 2, MACStr, "-", true)
	ap_ip := index_to_ip(index, 2, IPStr)

	// 调用AP属性
SIGN:
	ap := NewAP(ap_ip, ap_mac, dev_id, index)

	//logPrint(fmt.Sprintf("######[PROCESS]######启动第【%d】个虚拟AP成功！", index))
	// 等待AP对象都创建成功后开始，减少并发时CPU
	time.Sleep(time.Duration(WAITTIME) * time.Second)
	//  wg.Done()   // test
	// return		// test
	ap.start()

	// 如果AP退出后 判断是否重新启动
	if AP_Reline_Status {
		// 等待指定时间后重新上线
		logErrPrint(fmt.Sprintf("AP [%d] 等待重新上线... wait AP reline", index), fmt.Sprint("err-%d-%s.log", index, ap_mac))
		time.Sleep(time.Duration(AP_Reline_WaitTime))
		logErrPrint(fmt.Sprintf("AP [%d] 等待 %d 秒后重新上线成功, AP reline OK", index, AP_Reline_WaitTime), fmt.Sprint("err-%d-%s.log", index, ap_mac))
		goto SIGN
	} else {
		// 退出当前线程
		wg.Done()
		return
	}
}

func main() {
	//	fmt.Println(AC_IP)
	//	os.Exit(0)
	argsList := os.Args[1:]
	// fmt.Println(argsList, len(argsList))
	if len(argsList) == 1 && strings.ToLower(argsList[0]) == "-v" {
		logPrint(fmt.Sprintf("当前版本号: %v", VERSION))
		return
	} else if len(argsList) == 2 {
		AP_Begin_Index, _ = strconv.Atoi(argsList[0])
		AP_NUM, _ = strconv.Atoi(argsList[1])
		logPrint(fmt.Sprintf("初始化参数:AP_Begin_Index=%v ,AP_NUM=%v", AP_Begin_Index, AP_NUM))
	} else if len(argsList) == 0 {
		logPrint(fmt.Sprintf("默认参数:AP_Begin_Index=%v ,AP_NUM=%v", AP_Begin_Index, AP_NUM))
	} else {
		logPrint(fmt.Sprintf("err:参数错误:AP_Begin_Index ,AP_NUM ,参数不生效！！！"))
		return
	}

	logPrint(fmt.Sprintf("当前版本号: %v", VERSION))
	cpu_num := runtime.NumCPU()
	logPrint(fmt.Sprintf("当前CPU核心数: %v", cpu_num))

	//如果是0个不暂停
	if AP_PER_NUM == 0 {
		logPrint(fmt.Sprintf("启动 AP索引从%d到%d,一共模拟 %d 个AP,直接全部启动", AP_Begin_Index, AP_Begin_Index+AP_NUM-1, AP_NUM))
	} else {
		logPrint(fmt.Sprintf("启动 AP索引从%d到%d,一共模拟 %d 个AP,每秒启动%d个AP", AP_Begin_Index, AP_Begin_Index+AP_NUM-1, AP_NUM, AP_PER_NUM))
	}
	runtime.GOMAXPROCS(cpu_num) //限制运行同时使用的CPU核心数，默认直接获取系统的逻辑核心数

	// 记录第一轮并发AP的开始时间
	stop_time := time.After(time.Second)
	for i := AP_Begin_Index; i < AP_Begin_Index+AP_NUM; i++ {
		wg.Add(1) //为同步等待组增加一个成员
		go work(&wg, i)
		//判断是否达到每秒启动数量
		if AP_PER_NUM != 0 && i%AP_PER_NUM == 0 {
			//更新下一轮并发AP的开始时间
			<-stop_time
			//logPrint("333333")
			stop_time = time.After(time.Second)
		}
		//logPrint("22222")
	}
	//logPrint("44444")

	//阻塞等待所有组内线程都执行完毕退出
	wg.Wait()
	logPrint("所有AP已运行结束！！！")

}
