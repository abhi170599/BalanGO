/*
  Definition and declaration of an individual backend instances
  @Created: 23/04/2020
  @Last Update:23/04/2020
*/

package Instance

import (
	"net/http/httputil"
	"net/url"
	"sync"
)

/***************** A SINGLE SERVER INSTANCE *******************/

type ServerRoute struct {
	URL          *url.URL
	Alive        bool
	Lock         sync.RWMutex
	ReverseProxy *httputil.ReverseProxy
}

/***************** UTILITY FUNCTIONS FOR AN INSTANCE *****************/

//function to set the backend health status to alive
func (s *ServerRoute) SetAlive(alive bool) {
	s.Lock.Lock()
	s.Alive = alive
	s.Lock.Unlock()
}

//function to query the health of an intaance
func (s *ServerRoute) IsAlive() (alive bool) {
	s.Lock.Lock()
	alive = s.Alive
	s.Lock.Unlock()
	return alive
}
