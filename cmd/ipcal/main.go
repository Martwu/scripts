package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

func main() {
	pFlag := flag.String("p", "", "CIDR或带掩码的IP (e.g. 192.168.24.0/24)")
	nFlag := flag.String("n", "", "掩码长度或点分十进制掩码 (e.g. 24 or 255.255.255.0)")
	cFlag := flag.String("c", "", "IP地址 (e.g. 192.168.1.1)")
	aFlag := flag.Int("a", 0, "所需可用IP数量")
	flag.Parse()

	switch {
	case *pFlag != "":
		processPFlag(*pFlag)
	case *nFlag != "":
		processNFlag(*nFlag)
	case *cFlag != "" && *aFlag > 0:
		processCAFlag(*cFlag, *aFlag)
	default:
		fmt.Println("请使用以下参数：")
		fmt.Println("  -p <CIDR>      : 计算网络信息 (e.g. -p 192.168.24.0/24)")
		fmt.Println("  -n <掩码>       : 转换掩码格式 (e.g. -n 24 或 -n 255.255.255.0)")
		fmt.Println("  -c <IP> -a <数量>: 计算满足可用IP数量的子网 (e.g. -c 192.168.1.1 -a 10)")
		os.Exit(1)
	}
}

func processPFlag(input string) {
	_, ipNet, err := parseCIDR(input)
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		return
	}

	networkIP := ipNet.IP
	broadcastIP := calculateBroadcast(ipNet)
	firstIP := nextIP(networkIP)
	lastIP := previousIP(broadcastIP)

	printResults(
		ipNet.IP.String(),
		broadcastIP.String(),
		firstIP.String(),
		lastIP.String(),
		ipNet.Mask.String(),
	)
}

func processNFlag(input string) {
	if maskLen, err := strconv.Atoi(input); err == nil {
		if maskLen < 0 || maskLen > 32 {
			fmt.Println("错误: 掩码长度必须在0-32之间")
			return
		}
		mask := net.CIDRMask(maskLen, 32)
		fmt.Printf("%s\n", net.IP(mask).To4().String())
	} else if strings.Contains(input, ".") {
		mask := net.IPMask(net.ParseIP(input).To4())
		if mask == nil {
			fmt.Println("错误: 无效的子网掩码")
			return
		}
		ones, _ := mask.Size()
		fmt.Printf("%d\n", ones)
	} else {
		fmt.Println("错误: 无效的输入格式")
	}
}

func processCAFlag(ipStr string, minAvailable int) {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		fmt.Println("错误: 无效的IP地址")
		return
	}

	minTotalIPs := minAvailable + 2
	maskLen := 32
	var networkIP net.IP

	for maskLen >= 0 {
		mask := net.CIDRMask(maskLen, 32)
		networkIP = ip.Mask(mask)

		totalIPs := 1 << (32 - maskLen)
		if totalIPs >= minTotalIPs {
			break
		}
		maskLen--
	}

	if maskLen < 0 {
		fmt.Println("错误: 找不到满足条件的子网")
		return
	}

	ipNet := &net.IPNet{IP: networkIP, Mask: net.CIDRMask(maskLen, 32)}
	broadcastIP := calculateBroadcast(ipNet)
	firstIP := nextIP(networkIP)
	lastIP := previousIP(broadcastIP)

	printResults(
		networkIP.String(),
		broadcastIP.String(),
		firstIP.String(),
		lastIP.String(),
		net.IP(net.CIDRMask(maskLen, 32)).String(),
	)
}

func parseCIDR(input string) (net.IP, *net.IPNet, error) {
	if strings.Contains(input, "/") {
		parts := strings.Split(input, "/")
		if len(parts) != 2 {
			return nil, nil, fmt.Errorf("无效的输入格式")
		}

		if strings.Contains(parts[1], ".") {
			ip := net.ParseIP(parts[0])
			if ip == nil {
				return nil, nil, fmt.Errorf("无效的IP地址")
			}
			mask := net.IPMask(net.ParseIP(parts[1]).To4())
			if mask == nil {
				return nil, nil, fmt.Errorf("无效的子网掩码")
			}
			_, bits := mask.Size()
			if bits != 32 {
				return nil, nil, fmt.Errorf("IPv4掩码必须是32位")
			}
			return ip, &net.IPNet{IP: ip.Mask(mask), Mask: mask}, nil
		}
	}

	_, ipNet, err := net.ParseCIDR(input)
	if err != nil {
		return nil, nil, fmt.Errorf("解析CIDR失败: %v", err)
	}
	return ipNet.IP, ipNet, nil
}

func calculateBroadcast(ipNet *net.IPNet) net.IP {
	broadcast := make(net.IP, len(ipNet.IP))
	copy(broadcast, ipNet.IP)
	mask := ipNet.Mask

	for i := 0; i < len(ipNet.IP); i++ {
		broadcast[i] |= ^mask[i]
	}
	return broadcast
}

func nextIP(ip net.IP) net.IP {
	next := make(net.IP, len(ip))
	copy(next, ip)
	for i := len(next) - 1; i >= 0; i-- {
		next[i]++
		if next[i] != 0 {
			break
		}
	}
	return next
}

func previousIP(ip net.IP) net.IP {
	prev := make(net.IP, len(ip))
	copy(prev, ip)
	for i := len(prev) - 1; i >= 0; i-- {
		prev[i]--
		if prev[i] != 255 {
			break
		}
	}
	return prev
}

func printResults(network, broadcast, first, last, mask string) {
	fmt.Println("网络IP:     ", network)
	fmt.Println("广播IP:     ", broadcast)
	fmt.Println("第一个可用IP:", first)
	fmt.Println("最后一个可用IP:", last)
	fmt.Println("子网掩码:    ", mask)
}
