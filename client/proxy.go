package client

import (
	"log"
	"os"
	"path/filepath"
	"strings"
)

var (
	proxyNum = 0
)

func getProxyFromList() string {
	dir := filepath.Dir(os.Args[0])
	fileName := filepath.Join(dir, "proxy.list")
	buf, err := os.ReadFile(fileName)
	if err != nil {
		log.Println("Error load proxy list:", err)
		return ""
	}
	var list []string
	for _, proxy := range strings.Split(string(buf), "\n") {
		proxy = strings.TrimSpace(proxy)
		if proxy != "" {
			list = append(list, proxy)
		}
	}
	if proxyNum >= len(list) {
		proxyNum = 0
	}
	if len(list) == 0 {
		return ""
	}
	proxyHost := ""
	proxyHost = list[proxyNum]
	if !strings.HasPrefix(proxyHost, "http") && !strings.HasPrefix(proxyHost, "socks") {
		proxyHost = "//" + proxyHost
	}
	proxyNum++
	return proxyHost
}
