#test git command
package main

import (
	"fmt"
	"os"
	"strings"
)

func main()  {

	//读取命令行参数
	args := os.Args
	routerId := String2int(args[1])
	listenPort := String2int(args[2])
	adjArgs := args[3:]
	adjcents := []int{}

	for _, adj := range adjArgs{
		//string2int
		adjcent := String2int(adj)
		adjcents = append(adjcents, adjcent)
	}


	//初始化路由器
	router := CreateRouter(routerId, listenPort, adjcents)
	fmt.Println(router.RouterTable)

	//开启路由器监听端口
	go router.Listening()
	fmt.Println("路由器开始监听--------------")

	//定时发送路由表
	go router.SendRouterTable()


	//计时器线程
	go router.Timer()

	var input string
	for {
			Scanf(&input)
			if input == "Q" {
				break
			} else if input == "T" {
				router.PrintTable()
			} else if input[0] == 'S' {
				//获取目的地端口
				nodes := strings.Split(input, " ")
				node := nodes[1]
				desport := String2int(node)
				//发送数据包
				router.SendPacket(desport)

			} else if input == "A" {
				fmt.Println(router.ActiveFlag)
			} else if input == "N" {
				fmt.Println("Activity of Adjcents :")
				fmt.Println(router.ActiveFlag)
			} else if input[0] == 'P' {
				//获取必经路由节点
				nodes := strings.Split(input, " ")
				nodes = nodes[2:]

				//转换为整数
				var priports []int
				for _, v := range nodes {
					node := String2int(v)
					priports = append(priports, node)
				}

				//保存至路由器设置中
				router.PriorityRouters = priports
				fmt.Println("PriorityRouters 设置成功: ", router.PriorityRouters)

			}else if input[0] == 'R' {
				//获取拒绝节点
				nodes := strings.Split(input, " ")
				nodes = nodes[1:]

				//转换为整数
				var refusePorts []int
				for _, v := range nodes {
					node := String2int(v)
					refusePorts = append(refusePorts, node)
				}

				//保存至路由器设置中
				router.RefusedNode = refusePorts
				fmt.Println("RefusedRouters 设置成功: ", router.RefusedNode)

			} else {
				fmt.Println("no Key")
			}
		}


}




