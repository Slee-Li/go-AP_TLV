package main

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

//用于检查异常
func check(e error) {
	if e != nil {
		panic(e)
	}
}

//用于统一日志的的输出接口
func logPrint(str string) string {
	str = time.Now().Format("[2006-01-02 15:04:05.0000]: ") + str
	fmt.Println(str)
	return fmt.Sprintln(str)
}

//错误日志输出到前台也要到错误日志页面
func logErrPrint(str, file string) {
	writeFile(logPrint("err-"+str), file)
}

//用于将日志输出到日志文件
func writeFile(bufferStr, fileName string) {
	f, _ := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	defer f.Close()
	//f.Write([]byte(bufferStr))
	w := bufio.NewWriter(f)
	fmt.Fprintln(w, bufferStr)
	w.Flush()
}

//字符串转16进制
func str_to_hex(rsStr string) string {
	//	str := fmt.Sprintf("%x", rsStr)  // 等价方法
	//	fmt.Println(str)
	return hex.EncodeToString([]byte(rsStr))
}

//16进制转字符串
func hex_to_str(hexStr string) string {
	str, _ := hex.DecodeString(hexStr)
	return string(str)
}

//10进制转16进制
func int_to_hex(rsInt int64, length int) string {
	// ftm_str, _ := strconv.ParseInt(str, 10, 16) //等价，但是不能控制位数
	fmt_str := fmt.Sprintf("%%0.%dx", length)
	return fmt.Sprintf(fmt_str, rsInt)
}

//16进制转10进制
func hex_to_int(str string) int64 {
	i, _ := strconv.ParseInt(str, 16, 64)
	return i
}

//IP地址转16进制
func ip_to_hex(ip string) string {
	var rsIP string
	for _, i := range strings.Split(ip, ".") {
		v, _ := strconv.Atoi(i)
		rsIP += int_to_hex(int64(v), 2)
	}
	return rsIP
}

//16进制转IP地址
func hex_to_ip(str string) string {
	var rsIP []string
	re, _ := regexp.Compile(".{2}")
	for _, i := range re.FindAllString(str, -1) {
		// fmt.Println(i)
		v := strconv.FormatInt(hex_to_int(i), 10)
		rsIP = append(rsIP, v)
	}
	return strings.Join(rsIP, ".")
}

//16进制转字节流
func hex_to_byte(hStr string) []byte {
	var rsByte []byte
	re, _ := regexp.Compile(".{2}")

	for _, i := range re.FindAllString(hStr, -1) {
		// fmt.Println(i)
		v, _ := strconv.ParseUint(i, 16, 8)
		rsByte = append(rsByte, uint8(v))
	}
	return rsByte
}

//字节流转16进制字符串
func byte_to_hex(rsByte []byte) string {
	var hStr string
	for _, i := range rsByte {
		//hStr += strconv.FormatUint(uint64(i), 16)
		//必须是0为的16进制
		hStr += fmt.Sprintf("%0.2x", i)
	}
	return hStr
}

////转换10进制到16进制 如果转换后的16进制是奇数位则修改成偶数位
//func hexify(number {
//	/*
//	Convert integer to hex string representation, e.g. 12 to "000C"
//	*/

//}
// 平方函数
func SqrtXY(x int, y int) int {
	if y <= 0 {
		return 1
	} else if y == 1 {
		return x
	} else {
		y -= 1
		return x * SqrtXY(x, y)
	}

}

// 将十进制转16进制的MAC地址 共8组 length每组表示2位即255 str字符串格式 %d 中变化的组数，splitStr分割符号 - 或：
func int_to_mac(testNum int, length int, indexStr string, splitStr string, isUpper bool) string {
	//计算默认数量
	sum := SqrtXY(256, length)
	var macStr string
	//计算testnum的范围
	if testNum < sum {
		switch length {
		case 1:
			macStr = fmt.Sprintf(indexStr, fmt.Sprintf("%02X", testNum))
		case 2:
			macStr = fmt.Sprintf(indexStr, fmt.Sprintf("%04X", testNum))
		case 3:
			macStr = fmt.Sprintf(indexStr, fmt.Sprintf("%06X", testNum))
		default:
			fmt.Printf("[MAC1]length(%d)只能是1-3之间\n", length)
			return ""
		}
	} else {
		fmt.Printf("[MAC2]测试的数字%d不能大于%d\n", testNum, sum)
		return ""
	}
	//判断mac地址是否需要大小写
	if isUpper {
		macStr = strings.ToUpper(macStr)
	} else {
		macStr = strings.ToLower(macStr)
	}
	//添加分割符 - or :
	re, _ := regexp.Compile(".{2}")
	macStr = strings.Join(re.FindAllString(macStr, -1), splitStr)
	return macStr
}

// 10进制索引号转IP格式
func index_to_ip(index int, length int, indexStr string) string {
	//计算默认数量
	sum := SqrtXY(256, length)
	var ipStr string
	//计算testnum的范围
	if index < sum {
		switch length {
		case 1:
			ipStr = fmt.Sprintf(indexStr, index%sum)
		case 2:
			ipStr = fmt.Sprintf(indexStr, index/SqrtXY(256, length-1), index%SqrtXY(256, length-1))
		case 3:
			ipStr = fmt.Sprintf(indexStr, index/SqrtXY(256, length-1), index/SqrtXY(256, length-2), index%SqrtXY(256, length-2))
		default:
			fmt.Printf("[IP1]length(%d)只能是1-3之间\n", length)
			return ""
		}
	} else {
		fmt.Printf("[IP2]测试的数字%d不能大于%d\n", index, sum)
		return ""
	}
	return ipStr
}

//不需要执行命令的结果与成功与否，执行命令马上就返回
func exec_shell_nowait_result(command string) {
	//处理启动参数，通过空格分离 如：setsid /home/luojing/gotest/src/test_main/iwatch/test/while_little &
	//command_name_and_args := strings.FieldsFunc(command, splite_command)
	//cmd := exec.Command("/bin/bash", "-c", command)
	cmd := exec.Command("/bin/bash", "-c", command)
	//开始执行c包含的命令，但并不会等待该命令完成即返回
	check(cmd.Start())
	logPrint(fmt.Sprintf("run cmd: %s", command))
	return
}

//不需要执行命令的结果与成功与否，执行命令马上就返回
func exec_shell_wait_result(command string) {
	//处理启动参数，通过空格分离 如：setsid /home/luojing/gotest/src/test_main/iwatch/test/while_little &
	//command_name_and_args := strings.FieldsFunc(command, splite_command)
	//cmd := exec.Command("/bin/bash", "-c", command)
	cmd := exec.Command("/bin/bash", "-c", command)
	//开始执行c包含的命令，但会等待该命令完成即返回
	check(cmd.Run())
	logPrint(command)
	time.Sleep(7)
	return
}

///////////////TLV 类型
type TLV struct {
	CODE_TAGS   []string
	ACTION_TAGS []string
	tags        map[string]string
	tagsList    []string
	//tlv_str     string
	//tags_len []int
	isAction bool //判断是否是action类型的tags
}

//func T16L16VPackStr(t uint16, v string) []byte {
//	s := make([]byte, 256)
//	s = append(s, uint8(t>>8))
//	s = append(s, uint8(t))
//}

// 初始化TLV类型  的默认参数变量
func (t *TLV) initParams() {
	// code 1-6 使用的字典码
	t.CODE_TAGS = []string{
		/////////////// header
		"0200", // : "header version And type"   //header 部分 Version：表示协议版本，当前值为2。Type：表示数据加密模式，0 - 表示不加密 1 - 表示加密协商阶段 2 – 表示加密通信阶段
		"0100", // :  AC response header
		//////////////// data
		"0001", // : "Device ID" 设备唯一Id，最大长度不超过63字节。值类型为字符串
		"0002", // "Device Class"设备类型整数值，取值如下：1 – AP	2 – AC 3 – ROUTER 4 – SWITCH 5 – NAC 6 – PLC MASTER 7 – PLC SLAVE 8 – CPE
		"0003", // : "Device Vendor",           // 设备厂商ID
		"0004", // : "Device Hardware Address", // 设备硬件地址（MAC），值为字符串，最大长度为17字节
		"0007", // : "Device Local Address",    // 设备本地IPv4地址，类型为4字节的整数
		"0009", // : "Board Name",          // 硬件型号，类型为字符串，最大长度127
		"000c", // : "Product Name",        // 产品型号，类型为字符串，最大长度127
		"000a", // : "Firmware Version",    // 固件版本号，类型为字符串，最大长度127字节
		"000b", // : "Country Code",    // 国家代码，类型为字符串，最大长度2字节
		"000d", // : "ORG ID",         // 组织唯一ID，类型为字符串，最大长度63字节
		"000e", // : "BRC ID",         // 组织分部唯一ID，类型为字符串，最大长度63字节
		"0014", // : "Error Code",     // 错误码，类型为4字节正整数     -----
		"0028", // : "Echo Time",      // Echo Request间隔时间，类型为4字节正整数，单位秒
		"0029", // : "Echo Retry MAX", // Echo Request无响应次数阀值，类型为4字节正整数
		"002a", // : "Session Flag",   // 会话标识，类型为4字节正整数，值含义如下：	1 – OPEN 2 – RECOVER 3 – REST
		"002b", // : "Session Wait Time",  // 会话等待时长，类型为4字节正整数，单位秒
		"0008", // : "UNIX Timestamp",   // UNIX时间戳，类型为8字节正整数，单位秒
		"0015", // : "AC Name",       // AC名字，类型为UTF-8的字符串，最大长度为128字节
		"000f",
		"0012",
		"0000", //	结束标志
	}

	// code 7-8  action 使用的字典码
	t.ACTION_TAGS = []string{
		//// header
		"0200", // : "header version And type"   //header 部分 Version：表示协议版本，当前值为2。Type：表示数据加密模式，0 - 表示不加密 1 - 表示加密协商阶段 2 – 表示加密通信阶段
		"0100", // :  AC response header
		// data
		"0001", //: Action： length :8 报文动作，类型为无符号32位整型
		"0002", //: "Device ID" 设备唯一Id，最大长度不超过63字节。值类型为字符串
		"0003", // : "Device Hardware Address", // 设备硬件地址（MAC），值为字符串，最大长度为17字节
		"0004", // : "Device Local Address",    // 设备本地IPv4地址，类型为4字节的整数
		"0005", // : "Board Name",          // 硬件型号，类型为字符串，最大长度127
		"0006", // : "Firmware Version",    // 固件版本号，类型为字符串，最大长度127字节
		"0007", // " "Configuration MD5" : 配置文件的MD5值，类型为字符串，长度为32字节
		"0008", // "  Configuration Compress Class: 配置文件压缩类型，类型为无符号32位整型
		"0009", // "  Configuration Status: 配置文件状态，类型为无符号32位整型，值含义见下表
		"000a", // "  Configuration Size: 配置文件长度，类型为无符号32位整型，单位字节
		"000b", // " Configuration Content: 配置文件内容，类型为二进制，最大长度为65291字节
		"000c", // " Configuration Blocks: 配置文件分块总数，类型为无符号32位整型
		"000d", // " Configuration Block Index: 配置文件分块序号，类型无符号32位整型
		"000e", // " Device State Data Interval : 设备状态数据上报间隔，类型无符号32位整型，单位秒
		"000f", // "Device State Data: 设备状态数据，类型为二进制，最大长度为65291字节
		"0010", // : "Country Code",    // 国家代码，类型为字符串，最大长度2字节
		"0011", // : "Product Name",        // 产品型号，类型为字符串，最大长度127
		"0012", // : Soft Version: 软件版本，类型为无符号32位整型
		"0013", // : Soft Version U64: 软件版本，类型为无符号64位整型
		"0014", // : Firmware MD5 : 配置文件的MD5值，类型为字符串，长度为32字节
		"0015", // : Firmware Size: 固件大小，类型为无符号32位整型，单位字节
		"0016", // : Firmware Name: 固件名字，类型为字符串，最大长度为127字节
		"0017", // : Firmware URL: 固件下载的URL，类型为字符串，最大长度为255字节
		"0018", // : Firmware Status: 固件状态，类型为无符号32位整型，值含义见下表
		"0019", // : Update Firmware Result Code: 更新固件结果码，类型为无符号32位整型，值含义见下表
		"001a", // : "ORG ID",         // 组织唯一ID，类型为字符串，最大长度63字节
		"001b", // : "BRC ID",         // 组织分部唯一ID，类型为字符串，最大长度63字节
		"001e", // : Data Compress Class： 数据压缩类型，类型为无符号32位整型，值含义见下表
		"001f", // : Request ID U32: 请求ID，类型为无符号32位整型
		"0020", // : Request ID U64: 请求ID，类型为无符号64位整型
		"0028", // : AC Name: AC名字，类型为字符串，最大长度128字节
		"0029", // : Command: 待执行命令字符串，类型为字符串，最大长度为256字节
		"002a", // : Command Call ID： 命令呼叫ID，类型为字符串，最大长度64字节
		"002b", // : Command Execute Method： 命令执行方法，类型为无符号32位整型，值含义见下表
		"002c", // : Command Execute Result Code： 命令执行结果码，类型为无符号32位整型
		"002d", // : Command Result String： 命令执行结果字符串，类型为字符串，最大长度65291字节

		"0032", // : Station IP Address： 终端IP地址，长度4字节
		"0034", // : Station Hardware Address： 终端MAC地址，长度6字节
		"0035", // : Station Calling PHY： 终端呼叫设备的物理接口序号，类型无符号32位整型
		"0036", // : Station Calling VAP： 终端呼叫设备的虚拟接口序号，类型无符号32位整型
		"0037", // : Station Calling SSID： 终端呼叫设备的SSID，类型字符串，最大长度32字节
		"0038", // : Station Calling Interface Name： 终端呼叫设备的接口名字，类型字符串，最大长度16字节
		"0039", // : Calling Station ID： 主叫设备ID，类型为字符串，最大长度128字节
		"003a", // : Called Station ID： 被叫设备ID，类型为字符串，最大长度128字节
		"003b", // : Station Flag： 终端标志，类型无符号32位整型，值含义见下表
		"003c", // : Station User Name： 终端用户名，类型字符串，最大长度128字节

		"0041", // : Station Quota Input Rate： 终端输入速率，类型无符号64位整型，单位字节每秒
		"0042", // : Station Quota Output Rate：终端输出速率，类型无符号64位整型，单位字节每秒
		"0043", // : Station Quota Bytes：终端流量总配额，类型无符号64位整型，单位字节
		"0044", // : Station Quota Session：终端时长总配额，类型无符号64位整型，单位秒
		"0045", //
		"0046", //
		"0047", // : Station Bill Session：终端账单之使用时长，类型无符号64位整型，单位秒
		"0048", // : Station Bill Input Bytes： 终端账单之输入流量，类型无符号64位整型，单位字节
		"0049", // : Station Bill Input Packets：终端账单之输入数据包，类型无符号64位整型，单位个
		"004a", // : Station Bill Output Bytes：终端账单之输出流量，类型无符号64位整型，单位字节
		"004b", // : Station Bill Output Packets：终端账单之输出数据包，类型无符号64位整型，单位个
		"004c", // : Station Bill Input Rate： 终端账单之输入速率，类型无符号64位整型，单位字节每秒
		"004d", // : Station Bill Output Rate：终端账单之输出速率，类型无符号64位整型，单位字节每秒
		"004e", // : Station Error Code： 终端错误码，类型无符号32位整型，值含义见下表
		"004f", // : Station Calling Old PHY： 终端呼叫设备的原物理接口序号，类型无符号32位整型
		"0050", // : Station Calling Old VAP： 终端呼叫设备的原虚拟接口序号，类型无符号32位整型
		"0051", // : Station Calling Interface Type：终端呼叫设备接口类型，类型无符号32位整型
		"0052", // : Station Calling Old Interface Type： 终端呼叫设备原接口类型，类型无符号32位整型
		"0053", // : Station Authenticate Type： 终端认证类型，类型无符号32位整型
		"0059",
		"0000", //	结束标志
	}
	//
	//t.tags = make(map[string]string)
	//	t.tlvStr = ""

}

// 初始化TLV类型 initTLV函数  isAction true 表示action类型，false表示code类型
func NewTLV() *TLV {
	t := new(TLV)
	//初始化默认参数
	t.initParams()
	//初始化默认 的类型
	t.isAction = true
	t.autoChangeCode(false)
	return t
}

// 根据code自动切换适合的TLV格式  isAction 是否是 7 8类型
func (t *TLV) autoChangeCode(isAction bool) {
	//如果无变化则直接退出
	if t.isAction == isAction { //无变化直接退出
		return
	}
	//更新当前的action
	t.isAction = isAction
	// 有变化需要判断初始化 TLV的类型
	if t.isAction {
		t.tagsList = t.ACTION_TAGS
	} else {
		t.tagsList = t.CODE_TAGS
	}
}

func (t *TLV) build(data_dict map[string]string) string {
	//length := 4
	tlv_string := ""

	// 循环匹配tag
	for _, tag := range t.tagsList {
		// 如果tag匹配错误，跳过本次，继续匹配下个
		value, ok := data_dict[tag]
		if !ok {
			continue
		}
		//如果匹配成功，则开始计算
		// fmt.Printf("tag:%v value:%v\n" , tag,value)

		//16进制字符串必须是偶数，否则错误
		if len(value)%2 == 1 {
			fmt.Println("Invalid value length - the length must be even", value)
			//error.Error("Invalid value length - the length must be even")
			return ""
		}

		//		fmt.Println(tlv_string, strings.ToUpper(tag), int_to_hex(len(value)/2, 4), strings.ToUpper(value))
		tlv_string = tlv_string + strings.ToUpper(tag) + int_to_hex(int64(len(value)/2+len(tag)), 4) + strings.ToUpper(value)
	}
	return tlv_string
}

// 进行包拼接操作 最终拼接为16进制
func (t *TLV) packDisc(headerDict map[string]string, dataDict map[string]string) string {
	code := headerDict["code"] // Code : 1 表示 Discover Request
	// 自动判断tlv类型字典
	c := hex_to_int(code)
	if c == 7 || c == 8 {
		t.autoChangeCode(true)
	} else {
		t.autoChangeCode(false)
	}

	// print len(self.tags),dataDict
	msgData := t.build(dataDict)
	// print msgData
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
	headerTag := headerDict["header_tag"]
	sessionID := headerDict["sessionID"]
	seqNum := headerDict["seqNum"]
	payLoadType := headerDict["payLoadType"]
	headData := sessionID + seqNum + code + payLoadType
	// print headData
	headDisc := map[string]string{headerTag: headData + msgData}
	msgHex := t.build(headDisc)
	// print msgHex
	// 最后将16进制转二进制发送
	//return hex_to_byte(msgHex)
	return msgHex
}

func (t *TLV) parse(tlv_string string) (parsed_data map[string]string) {
	lengthLen := 2 * 2

	parsed_data = make(map[string]string)
	tag_length := 4

	//fmt.Println(t.tagsList)

	for i := 0; i < len(tlv_string); {
		tag_found := false
		for _, tag := range t.tagsList {
			//fmt.Println(len(tlv_string), i, i+tag_length, tlv_string[i:i+tag_length], tag, tlv_string[i:i+tag_length] == tag)

			// 遇到结束标志符号 则不在检查
			if tag == "0000" && string(tlv_string[i:i+tag_length]) == "0000" && !tag_found {
				goto OK
			}
			if string(tlv_string[i:i+tag_length]) == tag {
				value_length := hex_to_int(tlv_string[i+tag_length:i+tag_length+lengthLen]) - 4
				value_start_position := i + tag_length + lengthLen
				value_end_position := i + tag_length + lengthLen + int(value_length*2)
				//fmt.Println(value_start_position, len(tlv_string), value_end_position, value_length)

				if value_end_position > len(tlv_string) {
					fmt.Println("Parse error,value_end_position > len(tlv_string)")
					goto OK
					//raise ValueError("Parse error: tag " + tag + " declared data of length " + str(value_length) + ", but actual data length is " + str(int(len(tlv_string[value_start_position - 1:-1]) / 2)))
				}
				value := tlv_string[value_start_position:value_end_position]
				//fmt.Printf("tag:%v  value: %v\n", tag, value)
				parsed_data[tag] = value

				i = value_end_position
				tag_found = true
				if value_end_position == len(tlv_string) {
					goto OK
				}
				// 找到标签后继续下个轮循 否则会遇到0000结束符号
				break
			}

		}
		if !tag_found {
			msg := "Unknown tag found: " + tlv_string[i:]
			// print tlv_string
			//raise ValueError(msg)
			time.Sleep(300 * time.Millisecond)
			fmt.Println(msg)
		}
	}

OK:
	return parsed_data
}

func (t *TLV) unpackDict(msg string) (int, string, map[string]string) {
	//fmt.Println("unpack,hexmsg=", msg)
	rsDataDict := make(map[string]string)
	// 将data取出来
	rsData := t.parse(msg)["0100"]
	offset := 0
	// print rsData
	rsDataDict["SessionID"] = rsData[offset : offset+8]
	offset += 8
	// print "SessionID",rsDataDict["SessionID"]

	rsDataDict["SeqNum"] = string(hex_to_int(rsData[offset : offset+4]))
	offset += 4

	c := int(hex_to_int(rsData[offset : offset+2]))
	rsDataDict["Code"] = string(c)
	offset += 2

	rsDataDict["PayLoadType"] = string(hex_to_int(rsData[offset : offset+2]))
	offset += 2
	// print rsDataDict
	//fmt.Println("isaction c=", c)
	// 自动判断tlv类型字典
	if c == 7 || c == 8 {
		t.autoChangeCode(true)
	} else {
		t.autoChangeCode(false)
	}

	// 把data 中header部分 把sessionID SEQ NUM Code Payload Type  和 msg取出
	// print rsData[offset:offset + 2]
	aa := t.parse(rsData[offset:])
	// 合并字典
	for k, v := range aa {
		rsDataDict[k] = v
	}
	//	rsDataDict = dict(rsDataDict, **aa)

	// print rsData
	// print rsDataDict
	// 最后将解码的header中响应code 、data中的解码  和 解码字典
	_, ok := rsDataDict["0014"]
	if !ok {
		rsDataDict["0014"] = "00000000"
	}
	return c, string(rsDataDict["0014"]), rsDataDict
}

//func main() {
//	writeFile("11111\n", "test.log")
//	str := "V4.8.1_BETA r121-4bca85b"
//	str_hex := str_to_hex(str)
//	fmt.Println(str_hex)
//	fmt.Println(hex_to_str(str_hex))

//	fmt.Println(int_to_hex(255, 8))
//	fmt.Println(ip_to_hex("192.168.90.1"))
//	fmt.Println(hex_to_ip(ip_to_hex("192.168.90.1")))
//	fmt.Println(int_to_mac(65535, 2, "0000%s00", "-", true))
//	fmt.Println(index_to_ip(256, 2, "122.%d.%d.1"))

//	tlv := NewTLV(true)
//	//testHexStr := "0001002845453235384542382d413343452d343046332d423046322d4238413545363736303030310014000800000000"
//	testHexStr := "0008000800000000000a000800001218000c0008000000050009000800000003000700246266656631303133346134336462323439623062386665393139663461343264"
//	fmt.Println(tlv.parse(testHexStr))
//	fmt.Println(tlv.build(tlv.parse(testHexStr)))
//	//	fmt.Println(tlv.packDisc())

//}

//type AP struct {
//	GO_Begin_Index int
//	GO_Sum         int
//	GO_PER_NUM     int
//	GO_Pause_Time  float64
//}

//// 初始化AP参数
//func newAP(index, sum, per_num) *AP {

//	ap := new(AP)
//	//暂停时间格式
//	ap.GO_Pause_Time = 0.1111
//	//如果是0个不暂停
//	if AP_PER_NUM == 0 {
//		ap.GO_Pause_Time = 0
//	} else {
//		ap.GO_Pause_Time = 1.00000 / float64(AP_PER_NUM)
//	}
//	return ap
//}

////并发函数
//func (this, *Demo) Go() {
//	this.input <- time.Now().Format("2006-01-02 15:04:05")
//	time.Sleep(time.Millisecond * 500)
//	<-this.goroutine_cnt
//}
