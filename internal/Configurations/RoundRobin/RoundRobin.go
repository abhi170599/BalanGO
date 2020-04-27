/*
  RoundRobin Coniguration for the server Pool
*/

package RoundRobin

import (
	"context"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync/atomic"
	"time"

	"balango/internal/Instance"
)

//request context
const (
	Attempts int = iota
	Retry
)

/*************************************************** ROUND ROBIN CONFIGURATION ************************************************/

type RoundRobinPool struct {
	servers []*Instance.ServerRoute
	current uint64
}

/*
 function to add a server instance
*/
func (p *RoundRobinPool) addServer(server *Instance.ServerRoute) {
	p.servers = append(p.servers, server)
}

/*
function to update the current index
*/
func (p *RoundRobinPool) nextIndex() int {
	return int(atomic.AddUint64(&p.current, uint64(1)) % uint64(len(p.servers)))
}

/*
function to get the next alive server
*/
func (p *RoundRobinPool) Next() *Instance.ServerRoute {

	//loop the pool to get the next active server
	next := p.nextIndex()
	l := len(p.servers) + next //start from next and move a full cycle

	for i := next; i < l; i++ {
		idx := i % len(p.servers)
		if p.servers[idx].IsAlive() {
			if i != next {
				atomic.StoreUint64(&p.current, uint64(idx))
			}
			return p.servers[idx]
		}
	}
	return nil
}

/*
function to update the health status of a server instance
*/
func (p *RoundRobinPool) updateStatus(serverUrl *url.URL, alive bool) {
	for _, b := range p.servers {
		if b.URL.String() == serverUrl.String() {
			b.SetAlive(alive)
			break
		}
	}
}

/*
function to get the attempts from context
*/
func getAttemptsFromContext(r *http.Request) int {
	if attempts, ok := r.Context().Value(Attempts).(int); ok {
		return attempts
	}
	return 1
}

/*
function to get  retry from context
*/
func getRetryFromContext(r *http.Request) int {
	if retry, ok := r.Context().Value(Retry).(int); ok {
		return retry
	}
	return 0
}

/*
Health Check method
*/
func (p *RoundRobinPool) HealthCheckUp() {
	for _, b := range p.servers {
		status := "up"
		alive := p.isServerAlive(b.URL)
		b.SetAlive(alive)
		if !alive {
			status = "down"
		}
		log.Printf("%s [%s]\n", b.URL, status)
	}
}

func (p *RoundRobinPool) isServerAlive(u *url.URL) bool {

	timeout := 2 * time.Second
	conn, err := net.DialTimeout("tcp", u.Host, timeout)
	if err != nil {
		log.Println("site unreachable, error: ", err)
		return false
	}
	_ = conn.Close()
	return true
}

//func to run healthCheck routine for every 2 min
func (p *RoundRobinPool) HealthCheck() {

	t := time.NewTicker(time.Minute * 2)
	for {
		select {
		case <-t.C:
			log.Println("starting health check...")
			p.HealthCheckUp()
			log.Println("Health check completed")
		}
	}
}

/*
function to balance the load
*/
func (p *RoundRobinPool) LoadBalance(w http.ResponseWriter, r *http.Request) {
	attempts := getAttemptsFromContext(r)
	if attempts > 3 {
		log.Printf("%s(%s) Max attempts reached, terminating\n", r.RemoteAddr, r.URL.Path)
		http.Error(w, "Service not available", http.StatusServiceUnavailable)
		return
	}

	peer := p.Next()
	if peer != nil {
		peer.ReverseProxy.ServeHTTP(w, r)
		return
	}
	http.Error(w, "Service not available", http.StatusServiceUnavailable)
}

/*
Function to build the server pool
*/
func (p *RoundRobinPool) BuildPool(serverList string) {

	if len(serverList) == 0 {
		log.Fatal("Please provide one or more backends to load balance")

	}

	tokens := strings.Split(serverList, ",")
	for _, tok := range tokens {
		serverUrl, err := url.Parse(tok)
		if err != nil {
			log.Fatal(err)
		}

		proxy := httputil.NewSingleHostReverseProxy(serverUrl)
		proxy.ErrorHandler = func(writer http.ResponseWriter, request *http.Request, e error) {
			log.Printf("[%s] %s\n", serverUrl.Host, e.Error())
			retries := getRetryFromContext(request)
			if retries < 3 {
				select {
				case <-time.After(10 * time.Millisecond):
					ctx := context.WithValue(request.Context(), Retry, retries+1)
					proxy.ServeHTTP(writer, request.WithContext(ctx))
				}
				return
			}

			//after 3 retries, mark the backend down
			p.updateStatus(serverUrl, false)

			//if the same request routing for few attempts with different backends, increase the count of attempts
			attempts := getAttemptsFromContext(request)
			log.Printf("%s(%s) Attempting retry %d\n", request.RemoteAddr, request.URL.Path, attempts)
			ctx := context.WithValue(request.Context(), Attempts, attempts+1)
			p.LoadBalance(writer, request.WithContext(ctx))
		}

		p.addServer(&Instance.ServerRoute{
			URL:          serverUrl,
			Alive:        true,
			ReverseProxy: proxy,
		})
		log.Printf("Configured Server: %s\n", serverUrl)
	}
}

/*
function to create a new RoundRobinPool
*/
func NewRRPool() *RoundRobinPool {

	instances := make([]*Instance.ServerRoute, 0)
	pool := RoundRobinPool{
		servers: instances,
		current: 0,
	}
	return &pool
}
