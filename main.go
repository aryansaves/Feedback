package main
import (
	"fmt"
	"net/http"
	"log"
)
func handler(w http.ResponseWriter, r *http.Request){
	fmt.Fprintf(w, "I love %s\n", r.URL.Path[1:])
	fmt.Fprintln(w, r.Header)
}
func server(){
	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(":3000", nil))
}