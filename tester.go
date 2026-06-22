package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
)
type flagd struct {
	n int
	c int
	url string
	success int
	failure int
}
type sf struct {
	suc int
	fail int
}
func fetch(f flagd, ch chan sf){
	for range f.n/f.c {
	resp, err := http.Get(f.url)
	if err != nil {
		log.Fatalf("fetch error %v", err)
		f.failure += 1
	}
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("read error %v", err)
		f.failure += 1
	} 
	f.success += 1
	fmt.Printf(string(body))
	fmt.Println(resp.Status)
	} 
	ch <- sf{suc: f.success, fail: f.failure}
}
func main(){
	url := flag.String("u","","URL to test")
	n:= flag.Int("n", 10, "number of requests")
	c:= flag.Int("c", 5, "number of concurrent requests")
	flag.Parse()
	f := flagd{url: *url, c : *c, n : *n}
	ch := make(chan sf)
	var res sf
	var pass, fail int
	for range *c{
		go fetch(f, ch)
	}
	for range *c{
		res = <- ch
		pass += res.suc
		fail += res.fail
	}
	fmt.Printf("failures : %d\n",fail)
	fmt.Printf("success : %d\n",pass)
}