package main

import (
	"flag"
)

func main() {
	var configPath string
	var varsPath string

	flag.StringVar(&configPath, "config", "proxy.json", "Path to JSON config for Proxies. \nSee example.json for format.")
	flag.StringVar(&varsPath, "vars", "", "Path to JSON config for Variables, \nFile Format { \"key1\": \"value1\", \"key2\": \"value2\" ...} \nSee example.vars.json for format. ")
	flag.Parse()

	proxy := ProxyConfig{}
	proxy.LoadConfig(configPath, varsPath)
	proxy.Run()
}
