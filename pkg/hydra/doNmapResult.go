package hydra

import (
	"bytes"
	"fmt"
	"github.com/antchfx/xmlquery"
	"github.com/hktalent/scan4all/pkg"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
)

// 弱口令检测
func CheckWeakPassword(ip, service string, port int, wg *sync.WaitGroup) {
	defer wg.Done()
	// 在弱口令检测范围就开始检测，结果....
	service = strings.ToLower(service)
	if pkg.Contains(ProtocolList, service) {
		//log.Println("start CheckWeakPassword ", ip, ":", port, "(", service, ")")
		wg.Add(1)
		Start(ip, port, service, wg)
	}
}

func GetAttr(att []xmlquery.Attr, name string) string {
	for _, x := range att {
		if x.Name.Local == name {
			return x.Value
		}
	}
	return ""
}

func DoParseXml(s string, wg *sync.WaitGroup, bf *bytes.Buffer) {
	defer wg.Done()
	doc, err := xmlquery.Parse(strings.NewReader(s))
	if err != nil {
		log.Println("DoParseXml： ", err)
		return
	}
	var enableEsSv = pkg.GetVal("enableEsSv")
	m1 := make(map[string][][]string)
	for _, n := range xmlquery.Find(doc, "//host") {
		x1 := n.SelectElement("address").Attr[0].Value
		ps := n.SelectElements("ports/port")
		for _, x := range ps {
			if "open" == x.SelectElement("state").Attr[0].Value {
				ip := x1
				szPort := GetAttr(x.Attr, "portid")
				port, _ := strconv.Atoi(szPort)
				service := GetAttr(x.SelectElement("service").Attr, "name")
				//bf.Write([]byte(fmt.Sprintf("%s:%s\n", ip, szPort)))
				bf.Write([]byte(fmt.Sprintf("http://%s:%s\n", ip, szPort)))
				wg.Add(1)
				go CheckWeakPassword(ip, service, port, wg)
				// 存储结果到其他地方
				//x9 := AuthInfo{IPAddr: ip, Port: port, Protocol: service}
				if "true" == enableEsSv {
					var xx09 = [][]string{}
					if a1, ok := m1[ip]; ok {
						xx09 = a1
					}
					m1[ip] = append(xx09, []string{szPort, service})
				}
				fmt.Printf("%s\t%d\t%s\n", ip, port, service)
			}
		}
	}
	if "true" == enableEsSv {
		if 0 < len(m1) {
			for k, x := range m1 {
				pkg.SendAData[[]string](k, x, "nmap")
			}
		}
	}
}

func DoNmapRst(wg *sync.WaitGroup, bf *bytes.Buffer) {
	defer wg.Done()
	if x1, ok := pkg.TmpFile[pkg.Naabu]; ok {
		for _, x := range x1 {
			defer func(r *os.File) {
				r.Close()
				os.RemoveAll(r.Name())
			}(x)
			b, err := ioutil.ReadFile(x.Name())
			if nil == err && 0 < len(b) {
				//fmt.Println("read nmap xml file ok: ", len(b))
				wg.Add(1)
				DoParseXml(string(b), wg, bf)
			} else {
				log.Println("ioutil.ReadFile(x.Name()): ", err)
			}
		}
	} else {
		log.Println("check weak passwd: not find nmap tmp*.xml file")
	}
}
