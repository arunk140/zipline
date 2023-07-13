package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"regexp"
	"strings"
	"sync"
)

type Forward struct {
	Dst  string `json:"dst"`
	Src  string `json:"src"`
	Type string `json:"type"`

	Label   string `json:"label"`
	Silent  bool   `json:"silent"`
	Log     string `json:"log"`
	Disable bool   `json:"disable"`
}

type ProxyConfig struct {
	Forward  []Forward `json:"forward"`
	Silent   bool      `json:"silent"`
	Disable  bool      `json:"disable"`
	filepath string
	varspath string
	raw      string
}

type Vars map[string]string

var ValidForwardTypes = [...]string{"tcp", "udp", "http", "https"}

func (p *ProxyConfig) LoadConfig(filepath string, varspath string) (*ProxyConfig, error) {
	p.filepath = filepath
	bt, err := os.ReadFile(filepath)
	if err != nil {
		log.Fatal(err)
	}

	str := string(bt)

	if varspath != "" {
		p.varspath = varspath

		variables := new(Vars)

		raw, err := os.ReadFile(varspath)
		if err != nil {
			log.Fatal("Failed to Open Variables File, Format { \"key1\": \"value1\", \"key2\": \"value2\" ...} ", varspath)
		}
		err = json.Unmarshal(raw, &variables)
		if err != nil {
			log.Fatal("Failed to Parse Variables, Format { \"key1\": \"value1\", \"key2\": \"value2\" ...} ", varspath)
		}
		for k, v := range *variables {
			str = strings.ReplaceAll(str, fmt.Sprintf("{{%s}}", k), v)
		}
	}

	if strings.Contains(str, "{{") {
		log.Println("Unknown Variables found in Config")
		re := regexp.MustCompile(`\{\{([^\}]+)\}\}`)
		matches := re.FindAllStringSubmatch(str, -1)
		for _, m := range matches {
			log.Println(m[0])
		}
		log.Fatalln("")
	}

	p.raw = str
	err = json.Unmarshal([]byte(str), &p)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	return p, nil
}

func (p *ProxyConfig) Run() {
	var wg sync.WaitGroup

	if p.Disable {
		return
	}

	if p.Silent {
		log.SetOutput(io.Discard)
	}

	for _, f := range p.Forward {
		if f.Disable {
			log.Println(f.Label, "Proxy Disabled", "Src: "+f.Src, "Dst: "+f.Dst, "Type:"+f.Type)
			continue
		}
		log.Println(f.Label, "Proxy Running on "+f.Src, "Forwarding to "+f.Dst, "Type:"+f.Type)
		wg.Add(1)
		go handleForward(f, &wg)
	}

	wg.Wait()
}

func (f *Forward) UnmarshalJSON(data []byte) error {
	type Alias Forward
	err := json.Unmarshal(data, (*Alias)(f))
	if err != nil {
		return err
	}

	if f.Type == "http" && f.Src == "" {
		f.Src = ":80"
	}
	if f.Type == "https" && f.Src == "" {
		f.Src = ":443"
	}

	if f.Src == f.Dst {
		log.Println("Error Parsing Config - src")
		log.Fatalln("Invalid Source cannot be same as Destination - ", f.Src, f.Dst, f.Type)
	}

	foundMatch := false

	if f.Type == "" || f.Type == "http" || f.Type == "https" {
		f.Type = "tcp"
		foundMatch = true
	}

	if !foundMatch {
		for _, k := range ValidForwardTypes {
			if k == strings.ToLower(f.Type) {
				foundMatch = true
				break
			}
		}
		if !foundMatch {
			log.Println("Error Parsing Config - type")
			log.Fatalln("Invalid Type for Forward Config - ", f.Src, f.Dst, f.Type)
		}
	}

	_, _, err = net.SplitHostPort(f.Src)
	if err != nil {
		log.Println("Error Parsing Config - src")
		log.Fatalln("Invalid Source for Forward Config - ", f.Src, f.Dst, f.Type)

		log.Fatalln(err)
		return err
	}

	_, _, err = net.SplitHostPort(f.Dst)
	if err != nil {
		log.Println("Error Parsing Config - dst")
		log.Fatalln("Invalid Destination for Forward Config - ", f.Src, f.Dst, f.Type)
		log.Fatalln(err)
		return err
	}

	if f.Label == "" {
		f.Label = "Unlabelled"
	}

	return nil
}

func handleForward(f Forward, wg *sync.WaitGroup) {
	ln, err := net.Listen(string(f.Type), f.Src)
	if err != nil {
		wg.Done()
		log.Println(err)
		return
	}

	for {
		srcConn, err := ln.Accept()
		if err != nil {
			wg.Done()
			log.Println(err)
			return
		}
		go handleConnection(srcConn, f)
	}
}

func handleConnection(srcConn net.Conn, f Forward) {
	dstConn, err := net.Dial(string(f.Type), f.Dst)
	if err != nil {
		if !f.Silent {
			log.Printf("Error: Failed: %v", err)
		}
		defer srcConn.Close()
		return
	}
	go func() {
		defer dstConn.Close()
		defer srcConn.Close()
		if !f.Silent {
			log.Println(f.Src, "->", f.Dst)
		}
		ioCopy(dstConn, srcConn, f)
	}()
	go func() {
		defer dstConn.Close()
		defer srcConn.Close()
		if !f.Silent {
			log.Println(f.Dst, "<-", f.Src)
		}
		ioCopy(srcConn, dstConn, f)
	}()
}

func ioCopy(dst net.Conn, src io.Reader, f Forward) (written int64, err error) {
	if f.Log == "" {
		return io.Copy(dst, src)
	}
	fileName := fmt.Sprintf(f.Log)

	logfile, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return -1, err
	}
	defer logfile.Close()

	mWriter := io.MultiWriter(logfile, dst)

	return io.Copy(mWriter, src)
}
