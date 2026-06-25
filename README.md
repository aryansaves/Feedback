Hello ! this is feedback   
A feedback to your API servers  
Run benchmarks against your server   
Load test it and return with stats  
  
how to run it u ask ? 😏  
  
clone the repo :  
`git clone https://github.com/aryansaves/Feedback`  
  
build the binary :  
`go build tester.go`  
  
test with the mock server provided i.e. main.go  
  
Once the binary is created use `./tester {arguement} ...` to run the cli  
  
flags : 
-u : provide the singular URL directly 
for e.g. :  `./tester -u http://localhost:3000/test`

-f : provide the url in a file
for e.g. : `./tester -f filename.txt`

-n : number of requests
for e.g. : `./tester -u http://localhost:3000/test -n 100` defaults to "10" 
implements round robin when multiple urls in the file provided

-c : concurrency scale
for e.g. : `./tester -f filename.txt -c 10` defaults to "5"

![Project Diagram](example.png)

@all the code here in this repo was handwritten 😜