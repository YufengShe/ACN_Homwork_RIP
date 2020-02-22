package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"time"
)


//救救Golang fmt.Scanf()可怜的输入机制吧
func Scanf(a *string) {
	reader := bufio.NewReader(os.Stdin)
	data, _, _ := reader.ReadLine()
	*a = string(data)
}

//string转int
func String2int(str string) int{

	intvar, err := strconv.Atoi(str)
	if err != nil {
		log.Panic(err)
		return -1
	} else {
		return intvar
	}
}


//初始化路由 包括路由表ID、路由表端口、邻居节点端口、路由表信息等等
func CreateRouter(routerId int, listenPort int, adjcents []int) *Router{

	router := &Router{
		RouterId:routerId,
		ListenPort:listenPort,
		SendPort: listenPort + 100,
		AdjPorts:adjcents,
		RouterTable: []RouterTableItem{},
		ActiveFlag: map[int]int{},
		PriorityRouters: []int{},
		RefusedNode: []int{},
		//Conn:nil,
	}

	//初始化自己的路由表
	routeRule0 := RouterTableItem{
		DestPort:listenPort,
		Hops:0,
		NextPort:listenPort,
		Path: []int{listenPort},
	}

	router.RouterTable = append(router.RouterTable, routeRule0)


	//初始化ActiveFlag
	for _, v := range adjcents {
		router.ActiveFlag[v] = 1
	}

	return router
}

//开启监听端口
func (router *Router) Listening() {

	//构建server结构体变量
	serverInfo := ServerInfo{
		IpAddress:"127.0.0.1",
		Port:router.ListenPort,
		BuffSize: 8192,
	}

	//获取udp地址
	address := serverInfo.IpAddress + ":" + strconv.Itoa(serverInfo.Port)
	laddr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	//绑定路由器监听端口
	conn, err := net.ListenUDP("udp", laddr)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	defer conn.Close()

	//开始监听
	for  {
		//创建接收缓冲区
		buff := make([]byte, serverInfo.BuffSize)
		length, rAddr, err := conn.ReadFromUDP(buff)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		//从缓冲区取出真正的传输数据、解析数据来源
		realData := buff[:length]
		//rIPAddr := rAddr.IP
		rPort := rAddr.Port

		//接收到邻居新消息 将该邻居的活跃值设置为1
		router.ActiveFlag[rPort-100] = 1 //发送端口-100 = 接收端口

		//处理数据
		//fmt.Println("rport:",rPort)
		router.ProcessMsg(rPort-100, realData)

	}


}

//发送数据
func (router *Router) SendMsg(rIPAddr string, rPort int, msg []byte) {

	//本地发送信息
	cInfo := ClientInfo{
		IpAddress: "127.0.0.1",
		Port: router.SendPort,
	}

	//获取本地地址
	laddress :=  cInfo.IpAddress + ":" + strconv.Itoa(cInfo.Port)
	laddr, err := net.ResolveUDPAddr("udp", laddress)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	//获取远程服务器地址
	raddress := rIPAddr + ":" + strconv.Itoa(rPort)
	raddr, err := net.ResolveUDPAddr("udp", raddress)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	//发送udp数据包
	conn, err := net.DialUDP("udp", laddr, raddr)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	defer conn.Close()

	_, err = conn.Write(msg)

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

}


//每隔六秒发送数据路由信息给邻居节点
func (router *Router)SendRouterTable(){

	//发送路由数据给各邻居节点
	for {
		for _, adjcent := range router.AdjPorts {

			//毒性逆转生成分享给该节点的路由表信息
			table := router.PoisonReverse(adjcent)

			//序列化RouterTable
			tjson, err := json.Marshal(table)
			if err != nil {
				log.Panic(err)
			}

			//封装发送数据并序列化
			msg := &MsgStruct{
				Type: RouteTableMsg,
				Msg: tjson,
			}

			jsonbytes, err := json.Marshal(msg)
			if err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}

			router.SendMsg(Local, adjcent, jsonbytes)
		}

		time.Sleep(6*time.Second)
	}

	return
}



//路由表更新
func (router *Router)RouterTableUpdate(adjPort int, adjcentTable []RouterTableItem) {

	//本节点的路由表
	myTable := router.RouterTable

	//增加通往新地址的路由表项 并更新已有的路由表项
	for _, adjTableItem := range adjcentTable {

		//目的地是本节点时 无需记录其余路径
		if adjTableItem.DestPort == router.ListenPort {
			continue
		}

		//防止重复环路出现
		if router.CirclePath(adjTableItem.Path) {
			continue
		}

		//判断该路由表项是否已经存在
		flag := 0

		//邻居节点的路径
		compAdjPath := append(adjTableItem.Path, router.ListenPort)
		//更新路由表
		for index, myTableItem := range myTable {


			//更新相同表项(从同一邻居处学习得到的同一路径)
			if adjTableItem.DestPort == myTableItem.DestPort &&  ComIntArry(compAdjPath, myTableItem.Path)  {
				flag = 1
				var hops int
				if adjTableItem.Hops == 16 {
					hops = 16
				} else {
					hops = adjTableItem.Hops + 1
				}
				newTableItem := RouterTableItem{
					DestPort:myTableItem.DestPort,
					Hops:hops,
					NextPort:adjPort,
					Path:compAdjPath,
				}
				myTable[index] = newTableItem
			}

		}


		//该路由表项是一个新表项
		if flag == 0 {
			//add
			var hops int
			if adjTableItem.Hops == 16 {
				hops = 16
			} else {
				hops = adjTableItem.Hops + 1
			}
			newItem := RouterTableItem{
				DestPort:adjTableItem.DestPort,
				Hops:hops,
				NextPort:adjPort,
				Path:compAdjPath,
			}
			myTable = append(myTable, newItem)
		}
	}

	router.RouterTable = myTable

}

//打印路由表
func (router *Router)PrintTable() {

	fmt.Println("-------------------------------------------------------------------")
	fmt.Print("|Des   ")
	fmt.Print("|Hops  ")
	fmt.Print("|Next  ")
	fmt.Println("|Path")
	fmt.Println("-------------------------------------------------------------------")


	for i:=0; i<len(router.RouterTable); i++ {
		if router.RouterTable[i].Hops != 16 {
			routerItem := router.RouterTable[i]
			hops := strconv.Itoa(routerItem.Hops)

			//打印desPort
			fmt.Print("|")
			fmt.Print(routerItem.DestPort)
			fmt.Print("  ")

			//打印Hops
			fmt.Print("|")
			fmt.Print(hops)
			for j:=0; j<6-len(hops); j++ {
				fmt.Print(" ")
			}

			//打印Next
			fmt.Print("|")
			fmt.Print(routerItem.NextPort)
			fmt.Print("  ")

			//打印From
			fmt.Print("|")
			len_path := len(routerItem.Path)
			for i := len_path-1; i>=0; i--{
				fmt.Print(routerItem.Path[i])
				fmt.Print(" ")
			}
			fmt.Println("")
		}
		}


	fmt.Println("-------------------------------------------------------------------")
}

//处理监听到的数据
func (router *Router)ProcessMsg(rport int, jsonbytes []byte) {

	//反序列化json
	msg := new(MsgStruct)
	err := json.Unmarshal(jsonbytes, msg)
	if err != nil {
		log.Panic(err)
	}

	//判断msg类型并处理
	if msg.Type == RouteTableMsg {
		//获取邻居路由表信息并反序列化
		tjson := msg.Msg
		table := new([]RouterTableItem)
		err := json.Unmarshal(tjson, table)
		if err != nil {
			log.Panic(err)
		}

		//更新自己的路由表
		//fmt.Println(*table)
		router.RouterTableUpdate(rport, *table)
	} else if msg.Type == DataPacketMsg {

		//解析packet并转发数据包
		router.TranversePack(rport, msg.Msg)

	} else {
		fmt.Println("Wrong Msg Type")
	}
	return
}

//路由表中nextHop为nextPort的路由表项设置为不可达(16)
func (router *Router) DeleteInfo(nextPort int) {

	table := router.RouterTable

	for index, routerItem := range table {
		if routerItem.NextPort == nextPort {

			 newItem := RouterTableItem{
			 	DestPort:routerItem.DestPort,
			 	Hops:16,
			 	NextPort:nextPort,
			 	Path:routerItem.Path,
			 }

			 table[index] = newItem
		}
	}
	router.RouterTable = table
	return
}




//设置一个timer计时器 计时器会查询各邻居节点的活跃flag
func (router *Router)Timer()  {

	for {
		time.Sleep(18*time.Second)

		//检查邻居及节点活跃检测变量
		for adjcent, active := range router.ActiveFlag {
			if active == 1 { //为1说明前十八秒活跃过 重置为0
				router.ActiveFlag[adjcent] = 0
			} else { //为0 判定邻居节点已死 有事烧纸

				//将通过邻居节点到达的路由信息设为16
				router.DeleteInfo(adjcent)
			}
		}
	}
}

//毒性逆转 创造发送给邻居adjcent的路由表
func (router *Router)PoisonReverse(adjcent int) []RouterTableItem{
	table := router.RouterTable
	newtable := []RouterTableItem{}
	for _, routerItem := range table {
		//将来源为该节点的路由信息变为16后传回去
		if routerItem.NextPort == adjcent {

			item := RouterTableItem{
				DestPort:routerItem.DestPort,
				Hops:16,
				NextPort:routerItem.NextPort,
				Path:routerItem.Path[:len(routerItem.Path)-1],
			}

			newtable = append(newtable, item)
		} else {
			newtable = append(newtable, routerItem)
		}
	}


	return  newtable

}

//防止出现环
func (router *Router)CirclePath(path []int) bool {
	my := router.ListenPort

	for _, pathNode := range  path {
		if pathNode == my {
			return true
		}
	}

	return false
}



//比较两个[]int是否一致
func ComIntArry(a, b []int)  bool{

	len_a := len(a)
	len_b := len(b)

	if len_a != len_b {
		return false
	}

	for index, elem := range a {
		if elem != b[index] {
			return false
		}
	}

	return true
}

func (router *Router) FindPath(desPort int) (RouterTableItem, bool) {
	//选择基于策略的最短路由项
	sPathItem := RouterTableItem{
		DestPort:desPort,
		NextPort:desPort,
		Hops:16,
		Path: []int{},
	}

	for _, tableItem := range router.RouterTable {

		//筛选到达目的地的路由表项
		if tableItem.DestPort != desPort {
			continue
		}

		//该路径当前是否可达
		if tableItem.Hops == 16 {
			continue
		}

		//是否满足必经节点策略（每个必经节点列表中的节点都要在路径中出现）
		flag_path := 1
		for _, prinode := range router.PriorityRouters {
			flag := 0
			for _, node := range tableItem.Path {
				if prinode == node {
					flag = 1
					break
				}
			}

			if flag == 0 {
				flag_path = 0
				break
			}
		}

		if flag_path == 0 {
			continue
		}

		//是否满足拒绝节点规则
		if len(router.RefusedNode) != 0 {

			flag_path = 1

			refused := router.RefusedNode[0]
			for _, node := range tableItem.Path {
				if node == refused {
					flag_path = 0
					break
				}
			}
			if flag_path == 0 {
				continue
			}
		}

		//找到合适路由表项 选择其中的最优路径
		if tableItem.Hops < sPathItem.Hops {
			sPathItem = tableItem
		}
	}

	//返回结果
	if sPathItem.Hops < 16 {
		return sPathItem, true
	} else {
		return sPathItem, false
	}

}

//基于策略构造并发送数据包
func (router *Router)SendPacket(desPort int)  {

	//寻找路由路径
	routerItem, flag := router.FindPath(desPort)
	if flag == false {
		fmt.Println("无法找到满足Priority和Refused列表的路径!")
		return
	} else {
		fmt.Println("找到满足策略的路径：", routerItem.Path)
	}

	//构造发送路径
	var spath []int
	for _, v := range routerItem.Path {
		spath = append(spath,v)
	}

	//构造发送数据包
	nextport := routerItem.NextPort
	packet := &DataPacket{
		NextPort: nextport,
		Path: spath[:len(spath)-1],
		Data: []byte("Data Packet"),
	}

	//序列化数据包
	packbytes, err := json.Marshal(packet)
	if err != nil {
		log.Panic(err)
	}

	//封装成发送消息并序列化
	msg := &MsgStruct{
		Type:DataPacketMsg,
		Msg: packbytes,
	}
	jsonbytes, err := json.Marshal(msg)
	if err != nil {
		log.Panic(err)
	}

	//发送消息
	router.SendMsg(Local, nextport, jsonbytes)
	fmt.Printf("向%d发送目的地为%d的数据包..............\n", nextport, desPort)
}

//基于策略的数据包转发
func (router *Router)TranversePack(preport int, packetbytes []byte)  {

	//反序列化DataPacket
	packet := new(DataPacket)
	err := json.Unmarshal(packetbytes, packet)
	if err != nil {
		log.Panic(err)
	}

	if packet.NextPort != router.ListenPort {
		fmt.Println("Wrong Packet Received: ", packet)
		return
	} else {
		//判断自己是否为目标节点
		desport := packet.Path[0]
		if desport == router.ListenPort {
			fmt.Println("成功接收到数据包: ", string(packet.Data))
			return
		}

		//否则转发
		//从路径中去除本节点 构造新的数据包
		spath := packet.Path[:len(packet.Path)-1]
		nextport := spath[len(spath)-1]
		datapacket := &DataPacket{
			Path:spath,
			NextPort:nextport,
			Data: packet.Data,
		}
		packbytes,err := json.Marshal(datapacket)
		if err != nil {
			log.Panic(err)
		}

		//封装成发送消息并序列化
		msg := &MsgStruct{
			Type:DataPacketMsg,
			Msg: packbytes,
		}
		jsonbytes, err := json.Marshal(msg)
		if err != nil {
			log.Panic(err)
		}

		//向下一条节点转发数据包
		router.SendMsg(Local, nextport, jsonbytes)
		fmt.Printf("向%d转发从%d来的数据包\n", nextport, preport)
	}
}