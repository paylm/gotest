package main

import (
	"fmt"
	"strings"
)

func stx(s string, b []byte, ln int) {
	fmt.Println(s)
	fmt.Println(string(b[:len(b)]), ln, cap(b))
	fmt.Println(strings.Compare(s, string(b)))
	sb := string(b[:ln])
	fmt.Println(strings.Compare(s, sb))
	fmt.Println(len(sb), len(s))
	fmt.Println((s == sb))
}

func main() {
	fmt.Println("vim-go")
	s := "ack:6"
	b := []byte{'a', 'c', 'k', ':', '6'}
	c := string(b[:])
	fmt.Println(strings.Compare(s, c))
	//fmt.Println(strings.Compare(s, "ack:6"))
	//fmt.Println(len(c))

	cs := make([]byte, 512)
	cs[0] = 'a'
	cs[1] = 'c'
	cs[2] = 'k'
	fmt.Println(len(cs))
	stx("ack", cs, 3)
	//stx("acc", cs, len(cs))
}
