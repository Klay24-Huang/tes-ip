package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"
	"time"
)

const (
	port          = ":7071"
	limitDuration = 60 * time.Second
	limitCount    = 60
)

// Change the the map to hold values of the type visitor.
var visitors = make(map[string]*visitor)
var mu sync.RWMutex

func main() {
	cleanupBackground()
	mux := http.NewServeMux()
	getCountHandlerFunc := http.HandlerFunc(getCountHandler)
	mux.Handle("/", limitIpMiddleware(getCountHandlerFunc))

	log.Printf("Listening on %s...", port)
	http.ListenAndServe(port, mux)
}

func getCountHandler(w http.ResponseWriter, r *http.Request) {
	ip, _ := getIp(r)
	visitor := getVisitor(ip)
	w.Write([]byte(fmt.Sprintf("%v", visitor.count)))
}

// 紀錄次數
type visitor struct {
	count      int
	recordTIme time.Time
}

func (v *visitor) allow() bool {
	mu.Lock()
	defer mu.Unlock()
	if v.count < limitCount {
		v.count++
		return true
	}

	return false
}

// 在背景清除過期ip資料
func cleanupBackground() {
	go cleanupVisitors()
}

func getVisitor(ip string) *visitor {
	mu.Lock()
	defer mu.Unlock()

	v, exists := visitors[ip]
	if !exists {
		visitor := visitor{0, time.Now()}
		visitors[ip] = &visitor
		return &visitor
	}
	return v
}

// 每100毫秒清除過期ip紀錄
func cleanupVisitors() {
	for {
		time.Sleep(100 * time.Millisecond)

		mu.Lock()
		for ip, v := range visitors {
			if time.Since(v.recordTIme) > limitDuration {
				delete(visitors, ip)
				fmt.Printf("delete, ip:%s \n", ip)
			}
		}
		mu.Unlock()
	}
}

func getIp(r *http.Request) (ip string, err error) {
	ip, _, err = net.SplitHostPort(r.RemoteAddr)
	return
}

func limitIpMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// allow all CORS
		w.Header().Set("Content-Type", "text/html; charset=ascii")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type,access-control-allow-origin, access-control-allow-headers")
		ip, err := getIp(r)
		if err != nil {
			log.Print(err.Error())
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		visitor := getVisitor(ip)
		if !visitor.allow() {
			http.Error(w, "Error", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}
