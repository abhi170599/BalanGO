/*
Adaptive Optimization configuration for server pool
*/

package Adaptive

import (
	"context"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"
	"time"

	"balango/internal/Instance"
)

//request context
const (
	Attempts int = iota
	Retry
)

/************************************************ ADAPTIVE CONFIGURATION FOR SERVER POOL ***************************************************/

type AdaptivePool struct {
	servers []*Instance.ServerRoute
	size    int
	lock    sync.RWMutex
	status  *Status
}

/*
Structures to hold the status of a instance
*/

type ServerInfo struct {
	activeConnections float64
	meanResponseTime  float64
	lastConnections   float64
}

//structure to hold the status of servers in the pool
type Status struct {
	lock      sync.RWMutex
	statusMap map[*Instance.ServerRoute]ServerInfo
}

//function to init the staus
func (s *AdaptivePool) init() {

	s.lock.Lock()
	for _, server := range s.servers {
		s.status.statusMap[server] = ServerInfo{0.0, 0.0, 0.0}
	}
	s.lock.Unlock()
}

//function to update the activeConnection of server
func (s *Status) updateConnections(conn float64, server *Instance.ServerRoute) {
	server.Lock.Lock()
	var stat = s.statusMap[server]
	stat.activeConnections = conn
	s.statusMap[server] = stat
	server.Lock.Unlock()
}

//function to update the meanResponseTime
func (s *Status) updateResonseTime(respTime float64, server *Instance.ServerRoute) {
	server.Lock.Lock()
	var stat = s.statusMap[server]
	stat.meanResponseTime = respTime
	s.statusMap[server] = stat
	server.Lock.Unlock()
}

//function to update the lastConnections
func (s *Status) updateLastConnections(conn float64, server *Instance.ServerRoute) {
	server.Lock.Lock()
	var stat = s.statusMap[server]
	stat.lastConnections = conn
	s.statusMap[server] = stat
	server.Lock.Unlock()
}

//function to calculte the fitness of a server
func (s *Status) fitness(server *Instance.ServerRoute) (fitness float64) {
	s.lock.Lock()
	fitness = 0.35*s.statusMap[server].activeConnections +
		0.35*s.statusMap[server].meanResponseTime +
		0.30*s.statusMap[server].lastConnections
	s.lock.Unlock()
	return fitness
}

/*
PRIORITY QUEUE RELATED FUNCTIONS
*/

//function to check if a node is the leaf node
func (s *AdaptivePool) isLeaf(index int) (leaf bool) {
	//s.lock.Lock()
	if index >= (s.size/2) && index <= s.size {
		leaf = true
	}
	leaf = false
	//s.lock.Unlock()
	return leaf
}

//function to get the parent of a node
func (s *AdaptivePool) parent(index int) int {
	return (index - 1) / 2
}

//function to get the left child of a node
func (s *AdaptivePool) leftChild(index int) int {
	return 2*index + 1
}

//function to get the right child of a node
func (s *AdaptivePool) rightChild(index int) int {
	return 2*index + 2
}

//function to a server instance to the pool
func (s *AdaptivePool) add(server *Instance.ServerRoute) {
	s.lock.Lock()
	s.servers = append(s.servers, server)
	s.size++
	s.Up(s.size - 1)
	s.lock.Unlock()
}

//funcion to swap two nodes
func (s *AdaptivePool) swap(first, second int) {
	//s.lock.Lock()

	temp := s.servers[first]
	s.servers[first] = s.servers[second]
	s.servers[second] = temp
	//s.lock.Unlock()
}

//function to pass the node upwards in a heap
func (s *AdaptivePool) Up(index int) {

	//s.lock.Lock()
	for s.status.fitness(s.servers[index]) < s.status.fitness(s.servers[s.parent(index)]) {
		s.swap(index, s.parent(index))
	}
	//s.lock.Unlock()
}

//function to pass the node downwards in a heap
func (s *AdaptivePool) Down(current int) {

	//s.lock.Lock()
	if s.isLeaf(current) {
		return
	}
	smallest := current
	leftChildIndex := s.leftChild(current)
	rightChildIndex := s.rightChild(current)

	//if current is smallest then return
	if leftChildIndex < s.size && s.status.fitness(s.servers[leftChildIndex]) < s.status.fitness(s.servers[current]) {
		smallest = leftChildIndex
	}
	if rightChildIndex < s.size && s.status.fitness(s.servers[rightChildIndex]) < s.status.fitness(s.servers[current]) {
		smallest = rightChildIndex
	}

	if smallest != current {
		s.swap(current, smallest)
		s.Down(smallest)
	}
	//s.lock.Lock()
}

//function to get the next alive server instance
func (s *AdaptivePool) Next() (server *Instance.ServerRoute) {

	s.lock.Lock()
	extracted_servers := make([]*Instance.ServerRoute, 0)
	//get the next minimum
	server = s.getNext()
	for server.IsAlive() != true && s.size > 0 {
		extracted_servers = append(extracted_servers, server)
		server = s.getNext()
	}

	if s.size == 0 {
		server = nil
	} else {
		for _, ser := range extracted_servers {
			s.add(ser)
		}
	}
	s.lock.Unlock()
	return server
}

//function to get the next server
func (s *AdaptivePool) getNext() *Instance.ServerRoute {
	//s.lock.Lock()
	top := s.servers[0]
	s.servers[0] = s.servers[s.size-1]
	s.servers = s.servers[:(s.size)-1]
	s.size--
	s.Down(0)
	//s.lock.Unlock()
	return top
}

//function to update the health status of a server instance
func (s *AdaptivePool) updateStatus(serverUrl *url.URL, alive bool) {
	for _, b := range s.servers {
		if b.URL.String() == serverUrl.String() {
			b.SetAlive(alive)
			break
		}
	}
}

//function to get the attempts for a request
func getAttemptsFromContext(r *http.Request) int {
	if attempts, ok := r.Context().Value(Attempts).(int); ok {
		return attempts
	}
	return 1
}

//function to get  retry from context
func getRetryFromContext(r *http.Request) int {
	if retry, ok := r.Context().Value(Retry).(int); ok {
		return retry
	}
	return 0
}

//isAlive checks weather a backend is Alive by extablishing a TCP connection
func (s *AdaptivePool) isServerAlive(u *url.URL) bool {

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
func (s *AdaptivePool) HealthCheck() {

	t := time.NewTicker(time.Minute * 2)
	for {
		select {
		case <-t.C:
			log.Println("starting health check...")
			s.HealthCheckUp()
			log.Println("Health check completed")
		}
	}
}

//Health Checkup Routine
func (s *AdaptivePool) HealthCheckUp() {
	for _, b := range s.servers {
		status := "up"
		alive := s.isServerAlive(b.URL)
		b.SetAlive(alive)
		if !alive {
			status = "down"
		}
		log.Printf("%s [%s]\n", b.URL, status)
	}
}

/*
funtion to balance the load
*/

func (s *AdaptivePool) LoadBalance(w http.ResponseWriter, r *http.Request) {

	attempts := getAttemptsFromContext(r)
	if attempts > 3 {
		log.Printf("%s(%s) Max attempts reached, terminating\n", r.RemoteAddr, r.URL.Path)
		http.Error(w, "Service not available", http.StatusServiceUnavailable)
		return
	}
	//fmt.Println(status.statusMap)

	peer := s.Next()

	if peer != nil {
		active := s.status.statusMap[peer].activeConnections
		resp := s.status.statusMap[peer].meanResponseTime
		last := s.status.statusMap[peer].lastConnections
		s.status.updateConnections(active+1.0, peer)
		start := time.Now()
		peer.ReverseProxy.ServeHTTP(w, r)

		elapsed := float64(time.Since(start))

		//update mean response time
		sum_mean := last * resp
		mean_resp := (sum_mean + elapsed) / (last + 1)

		//update status of the peer and add it to the pool again
		s.status.updateLastConnections(last+1.0, peer)
		s.status.updateResonseTime(mean_resp, peer)
		s.status.updateConnections(active, peer)

		s.add(peer)

		return
	}
	http.Error(w, "Service not available", http.StatusServiceUnavailable)
}

func (s *AdaptivePool) BuildPool(serverList string) {

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
			s.updateStatus(serverUrl, false)

			//if the same request routing for few attempts with different backends, increase the count of attempts
			attempts := getAttemptsFromContext(request)
			log.Printf("%s(%s) Attempting retry %d\n", request.RemoteAddr, request.URL.Path, attempts)
			ctx := context.WithValue(request.Context(), Attempts, attempts+1)
			s.LoadBalance(writer, request.WithContext(ctx))
		}

		s.add(&Instance.ServerRoute{
			URL:          serverUrl,
			Alive:        true,
			ReverseProxy: proxy,
		})
		log.Printf("Configured Server: %s\n", serverUrl)
	}
	s.init()
}

func NewAdaptivePool() *AdaptivePool {

	statusMap := make(map[*Instance.ServerRoute]ServerInfo)
	status := &Status{statusMap: statusMap}
	instances := make([]*Instance.ServerRoute, 0)

	pool := AdaptivePool{
		servers: instances,
		size:    0,
		status:  status,
	}

	return &pool
}
