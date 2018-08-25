package main

import (
	//"fmt"
	"strings"
	"time"
)

// AP 用户类类型
type APUsers struct {
	APUsersList []map[string]string
}

// APUser初始化函数
func NewAPUsers(APIndex int, APIP string, APMAC string) *APUsers {
	au := new(APUsers)
	// 生成的总个数 不能大于255
	// 开始的索引
	startIndex := 1

	// 0表示初始状态
	au.APUsersList = append(au.APUsersList, map[string]string{"userIndex": "0"})
	userDict := map[string]string{}

	//fmt.Printf("APIP=%v\n", APIP)
	//fmt.Printf("APIP[:(len(APIP)-1)]=%v\n", APIP[:(len(APIP)-1)])

	//ip和mac 格式化字符串
	ipstr := APIP[:(len(APIP)-1)] + "%d"
	macstr := strings.Replace(APMAC[:len(APMAC)-2], "-", "", -1) + "%s"
	//fmt.Printf("ipstr=%v,macstr=%v\n", ipstr, macstr)

	//生成客户端装入列表
	for i := APIndex + startIndex; i < (APIndex + startIndex + AP_USER_NUM); i++ {
		//fmt.Printf("i=%v,userIP=%v\n", i, index_to_ip(i-APIndex, 1, ipstr))
		// 获取User IP 2组 共 256 * 256 个
		userDict["userIP"] = index_to_ip(i-APIndex, 1, ipstr)
		userDict["userMac"] = int_to_mac(i-APIndex, 1, macstr, "-", true)
		// 记录用户的状态 0: 初始化状态， 1：上线状态（未使用） 2：计费状态   3：离线状态（未使用） 4：需要重新上线状态（未使用）
		userDict["userStatus"] = "0"
		// 计费相关状态  初始化没有值
		userDict["userBillSession"] = int_to_hex(time.Now().Unix(), 16) // Station - Bill - Session = 4094 表示用户开始的时间
		userDict["userBillInputBytes"] = int_to_hex(0, 16)              // Station - Bill - Input - Bytes = 14484238
		userDict["userBillInputPackets"] = int_to_hex(0, 16)            // Station - Bill - Input - Packets = 23022
		userDict["userBillOutputBytes"] = int_to_hex(0, 16)             // Station - Bill - Output - Bytes = 4535823
		userDict["userBillOutputPackets"] = int_to_hex(0, 16)           // Station - Bill - Output - Packets = 36855
		userDict["userBillInputRate"] = int_to_hex(0, 16)               // Station - Bill - Input - Rate = 741
		userDict["userBillOutRate"] = int_to_hex(0, 16)                 // Station - Bill - Output - Rate = 881
		au.APUsersList = append(au.APUsersList, userDict)
		userDict = map[string]string{}
	}

	return au
}

// 获取最新的用户列表
func (this *APUsers) getAPUsersList() []map[string]string {
	return this.APUsersList
}

// 更新用户列表状态
func (this *APUsers) setAPUsersList(userList []map[string]string) {
	this.APUsersList = userList
}

// 更新用户的计费状态
func (this *APUsers) updateUsersListBill(userList []map[string]string) []map[string]string {
	this.APUsersList = userList

	//获得计费索引
	//userIndex := userList[0]["userIndex"]
	// 增量
	var add int64
	add = 1
	// 将索引 记录上
	userList = []map[string]string{{"userIndex": userList[0]["userIndex"]}}

	// 更新用户的状态
	for _, user := range this.APUsersList[1:] {
		// 计费相关状态  初始化没有值
		user["userBillInputBytes"] = int_to_hex(hex_to_int(user["userBillInputBytes"])+1024*1024*add, 16) // Station - Bill - Input - Bytes = 14484238

		user["userBillInputPackets"] = int_to_hex(hex_to_int(user["userBillInputPackets"])+1024*add, 16) // Station - Bill - Input - Packets = 23022

		user["userBillOutputBytes"] = int_to_hex(hex_to_int(user["userBillOutputBytes"])+1024*1024*add, 16) // Station - Bill - Output - Bytes = 4535823

		user["userBillOutputPackets"] = int_to_hex(hex_to_int(user["userBillOutputPackets"])+1024*add, 16) // Station - Bill - Output - Packets = 36855

		user["userBillInputRate"] = int_to_hex(hex_to_int(user["userBillInputRate"])+1024*add, 16) // Station - Bill - Input - Rate = 741

		user["userBillOutRate"] = int_to_hex(hex_to_int(user["userBillOutRate"])+1024*add, 16) // Station - Bill - Output - Rate = 881

		userList = append(userList, user)
	}
	// 将更新后的APUsersList
	this.APUsersList = userList
	return this.APUsersList
}
