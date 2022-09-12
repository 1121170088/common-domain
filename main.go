package main

import (
	"bufio"
	"flag"
	"fmt"
	"gopkg.in/yaml.v3"
	"io"
	"log"
	"os"
	"reflect"
	"strings"
)

var (
	sourceFile string
	dstFile string
	tree map[byte] *node = make(map[byte] *node)
	y bool
	domaims []string
	format string
)

type node struct{
	end bool
	folw map[byte] *node
}

func init()  {
	flag.StringVar(&sourceFile, "s", "", "source file containing domain list to be handled, one domain one line")
	flag.StringVar(&dstFile, "d", "", "output file")
	flag.BoolVar(&y, "y", false, "output yaml format")
	flag.StringVar(&format, "f", "", "format domain")
	flag.Parse()
}

func reverse(s interface{}) {
	n := reflect.ValueOf(s).Len()
	swap := reflect.Swapper(s)
	for i, j := 0, n-1; i < j; i, j = i+1, j-1 {
		swap(i, j)
	}
}
func hasDomain(domain string) (has bool) {
	bytes := []byte(domain)
	reverse(bytes)
	var preNode *node = nil
	var ok bool = false
	has = true
	for _, b := range bytes {
		if preNode == nil {
			preNode, ok = tree[b]
			if !ok {
				preNode = &node{
					end:  false,
					folw: make(map[byte] *node),
				}
				tree[b] = preNode
				has = false
			}
		} else {
			preNode2, ok := preNode.folw[b]
			if !ok {
				preNode2 = &node{
					end:  false,
					folw: make(map[byte] *node),
				}
				preNode.folw[b]= preNode2
				has = false
			}
			preNode = preNode2
		}

	}
	if preNode != nil {
		if !preNode.end {
			preNode.end = true
			has = false
		}
	}
	return
}

func spellDomain(m map[byte] *node, bytes *[]byte)  {
	for k, v := range m {
		*bytes = append(*bytes, k)
		if v.end {
			bytes2 := make([]byte, len(*bytes))
			copy(bytes2, *bytes)
			*bytes = (*bytes)[:len(*bytes) - 1]
			reverse(bytes2)
			domain := strings.Trim(string(bytes2), "\x00")
			if format != "" {
				domain = fmt.Sprintf(format, domain)
			}
			domaims = append(domaims, domain)
			continue
		} else {
			folw := v.folw
			spellDomain(folw, bytes)
			*bytes = (*bytes)[:len(*bytes) - 1]
		}

	}
}

func main()  {

	//sourceFile = "list.txt"
	//dstFile = "1.txt"

	if sourceFile == "" || dstFile == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	sf, err := os.OpenFile(sourceFile, os.O_APPEND|os.O_RDWR, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
	defer sf.Close()

	rd := bufio.NewReader(sf)

	for {
		line, err := rd.ReadString('\n')
		if err != nil || err == io.EOF {
			break
		}
		line = strings.Trim(line, "\n")
		line = strings.Trim(line, "\r")
		line = strings.Trim(line, " ")
		hasDomain(line)
	}
	domaims = make([]string, 0)
	for k, v := range tree {
		bytes := make([]byte, 1)
		bytes = append(bytes, k)
		folw := v.folw
		spellDomain(folw, &bytes)
	}


	if domaims != nil {
		if y {
			m := make(map[string] []string)
			m["payload"] = domaims
			bytes, err := yaml.Marshal(m)
			if err != nil {
				log.Fatal(err)
			}
			err = os.WriteFile(dstFile, bytes, os.ModePerm)
			if err != nil {
				log.Fatal(err)
			}
		} else {
			f, err := os.OpenFile(dstFile,  os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
			if err != nil {
				log.Fatal(err)
			}
			defer f.Close()

			for _, d := range domaims {
				io.WriteString(f, d + "\n")
			}
		}

	}
}
