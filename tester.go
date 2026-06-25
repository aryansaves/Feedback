package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httptrace"
	"os"
	"sync"
	"time"
)
type flagd struct {
	n int
	c int
	urls [] string
}

type metrics struct {
	ttfb time.Duration
	ttlb time.Duration
	totaltime time.Duration
	failed bool
}
type failure struct{
	failr int
}

func fetch(f flagd, id int, client *http.Client, wg *sync.WaitGroup, met chan metrics, chf chan failure){
	defer wg.Done()
	var failt int 
	for i:= range f.n/f.c {
		url := f.urls[(id + i) % len(f.urls)]
		m, err := doRequest(url, client)
		if err != nil {
			failt++
			met <- metrics{failed: true}
			continue
		}
		met <- m
	}
	chf <- failure{failr : failt}
}
func doRequest(url string, client *http.Client) (metrics, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Printf("request error")
		return metrics{}, err
	}
	var start, firstByte, lastByte time.Time
	trace := &httptrace.ClientTrace{
	WroteRequest: func(info httptrace.WroteRequestInfo) {
			start = time.Now()
		},
		GotFirstResponseByte: func() {
			firstByte = time.Now()
		},
	}
	req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))
	appst := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Network execution error: %v", err)
		return metrics{}, err
	}
	resp.Body.Close()
	lastByte = time.Now()
	ttfb := firstByte.Sub(start)
	ttlb := lastByte.Sub(start)
	return metrics{ttfb : ttfb, ttlb: ttlb, totaltime: time.Since(appst)}, nil
}
func main(){
	url := flag.String("u","","URL to test")
	n:= flag.Int("n", 10, "number of requests")
	c:= flag.Int("c", 5, "number of concurrent requests")
	filename := flag.String("f","","File to scan")
	flag.Parse()
	var links[] string
	if *filename != "" {
	content, err := os.Open(*filename)
	if err != nil {
		log.Fatalf("Failed to read file: %s", err)
	}
	defer content.Close()
	var lines []string
	scanner := bufio.NewScanner(content)
	for scanner.Scan(){
		lines = append(lines, scanner.Text())
	}
	for _, value := range lines{
		links = append(links, value)
	}
	}
	if *url != "" {
    links = append(links, *url)
	}
	if len(links) == 0 {
    log.Fatal("provide a URL with -u or a file with -f")
	}
	f := flagd{urls : links , c : *c, n : *n}
	if *n < *c {
		fmt.Printf("try better lil bro 😂\n")
		return
	}
	var wg sync.WaitGroup
	var metr metrics
	ch := make(chan metrics)
	chf := make (chan failure)
	var ttfbList []time.Duration
	var ttlbList []time.Duration
	var ttList []time.Duration
	var failurecount failure
	client := &http.Client{
    Transport: &http.Transport{
        MaxIdleConns:        f.c + 10,
        MaxIdleConnsPerHost: f.c + 10,
        IdleConnTimeout:     90 * time.Second,
    },
}
	start := time.Now()
	for i := range *c{
		wg.Add(1)
		go fetch(f, i, client, &wg, ch, chf)
	}
	for range *n{
		metr = <-ch
  		if metr.failed {
        continue
    	}
		ttfbList = append(ttfbList, metr.ttlb)
		ttlbList = append(ttlbList, metr.ttfb)
		ttList = append(ttList, metr.totaltime)
	}
	for range *c{
		f := <- chf
		failurecount.failr += f.failr
	}
	elapsed := time.Since(start)
	maxTTFB, minTTFB := ttfbList[0], ttfbList[0]
	maxTTLB, minTTLB := ttlbList[0], ttlbList[0]
	maxTT, minTT := ttList[0], ttList[0]
	var sumttfb, sumttlb, sumtt time.Duration
	for i:=0; i < len(ttList); i++ {
		maxTTFB = max(maxTTFB, ttfbList[i])
		maxTTLB = max(maxTTLB, ttlbList[i])
		maxTT = max(maxTT, ttList[i])
		minTTFB = min(minTTFB, ttfbList[i])
		minTTLB = min(minTTLB, ttlbList[i])
		minTT = min(minTT, ttList[i])
		sumttfb += ttfbList[i]
		sumttlb += ttlbList[i]
		sumtt += ttList[i]
	}
	meanttfb := sumttfb/time.Duration(len(ttfbList))
	meanttlb := sumttlb/time.Duration(len(ttlbList))
	meantt := sumtt/time.Duration(len(ttList))
	fmt.Println()
	fmt.Printf("Results : \n")
	fmt.Println()
	fmt.Printf("Total Requests..................................: %d\n",*n)
	fmt.Printf("Failed Requests.................................: %d\n",failurecount.failr)
	fmt.Printf("Request/Second..................................: %.3f\n",float64(*n)/elapsed.Seconds())
	fmt.Println()
	fmt.Printf("Total Requests Time (s) (Min, Max, Mean)........: %.5f, %.5f, %.5f\n",minTT.Seconds(), maxTT.Seconds(), meantt.Seconds())
	fmt.Printf("Total to First Byte (s) (Min, Max, Mean)........: %.5f, %.5f, %.5f\n",minTTFB.Seconds(), maxTTFB.Seconds(), meanttfb.Seconds())
	fmt.Printf("Total to Last Byte  (s) (Min, Max, Mean)........: %.5f, %.5f, %.5f\n",minTTLB.Seconds(), maxTTLB.Seconds(), meanttlb.Seconds())
	
	wg.Wait()
}
