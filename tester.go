package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptrace"
	"sync"
	"time"
)
type flagd struct {
	n int
	c int
	url string
}

type metrics struct {
	ttfb time.Duration
	ttlb time.Duration
	totaltime time.Duration
}

func fetch(f flagd, wg *sync.WaitGroup, met chan metrics){
	defer wg.Done()
	for range f.n/f.c {
	req, err := http.NewRequest("GET", f.url, nil)
	if err != nil {
		log.Fatalf("fetch error %v", err)
	}

	var appst, start, firstByte, lastByte time.Time

	trace := &httptrace.ClientTrace{
		WroteRequest: func(info httptrace.WroteRequestInfo){
			start = time.Now()
		},
		GotFirstResponseByte: func(){
			firstByte = time.Now()
		},
	}
	appst = time.Now()
	req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))
	
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("Network execution error: %v", err)
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("read error %v", err)
	} 
	fmt.Printf(string(body))
	lastByte = time.Now()
	ttfb := firstByte.Sub(start)
	ttlb := lastByte.Sub(start)
	totaltime := time.Since(appst)
	met <- metrics{ttfb : ttfb, ttlb : ttlb, totaltime: totaltime}
	} 
	
}
func mn(){
	url := flag.String("u","","URL to test")
	n:= flag.Int("n", 10, "number of requests")
	c:= flag.Int("c", 5, "number of concurrent requests")
	flag.Parse()
	f := flagd{url: *url, c : *c, n : *n}

	var wg sync.WaitGroup
	var metr metrics
	ch := make(chan metrics)
	var ttfbList []time.Duration
	var ttlbList []time.Duration
	var ttList []time.Duration
	for range *c{
		wg.Add(1)
		go fetch(f, &wg, ch)
	}
	for range *c{
		metr = <-ch
		ttfbList = append(ttfbList, metr.ttlb)
		ttlbList = append(ttlbList, metr.ttfb)
		ttList = append(ttList, metr.totaltime)
	}
	maxTTFB, minTTFB := ttfbList[0], ttfbList[0]
	maxTTLB, minTTLB := ttlbList[0], ttlbList[0]
	maxTT, minTT := ttList[0], ttList[0]
	for i:=0; i < len(ttList); i++ {
		maxTTFB = max(maxTTFB, ttfbList[i])
		maxTTLB = max(maxTTLB, ttlbList[i])
		maxTT = max(maxTT, ttList[i])
		minTTFB = min(minTTFB, ttfbList[i])
		minTTLB = min(minTTLB, ttlbList[i])
		minTT = min(minTT, ttList[i])
		
	}
	
	wg.Wait()
	
}