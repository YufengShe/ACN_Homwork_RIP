package main

//ip、port、buffsize info for udp server
type ServerInfo struct {
	IpAddress string
	Port int
	BuffSize int	//max buff size to receive data
}


//ip、port info for udp client
type ClientInfo struct {
	IpAddress string
	Port int
}

//传统RTP路由表 -> 包含目标地址、最小跳数和下一跳地址
type RouterTableItem struct {
	DestPort int `json:"des_port"`
	Hops int `json:"hops"`
	NextPort int `json:"next_port"`
	Path []int `json:"path"`  //此路由信息条目是从哪个节点学习得来的
}


//路由器结构体
type Router struct {
	RouterId int `json:"router_id"`    //id
	ListenPort int `json:"my_port"`		//路由器的监听端口
	SendPort   int `json:"send_port"`    //路由器的发送端口
	AdjPorts []int `json:"adj_ports"`  // 邻居节点的监听端口
	RouterTable []RouterTableItem `json:"router_table"` //路由表
	ActiveFlag map[int]int `json:"active_flag"` //活跃度监视flag
	PriorityRouters []int `json:"priority_routers"`  //优先节点
	RefusedNode []int `json:"refused_node"` //拒绝节点


}
//消息信封 里面可封装路由表信息和数据包信息
type MsgStruct struct {
	Type string `json:"type"`
	Msg []byte `json:"msg"`
}

//数据包结构体
type DataPacket struct {
	NextPort int `json:"next_port"`
	Path []int 	`json:"path"`
	Data []byte `json:"data"`
}

const (
	Local = "127.0.0.1"
	RouteTableMsg = "RoutingTable"
	DataPacketMsg = "DataPacket"
	PacketContent = "DataPacket"
)