/*
 interface for a server pool configuration
*/

package Config

import (
	"net/http"

	"balango/internal/Instance"
)

/******************************************************** INTERFACE FOR A SERVER POOL CONFIGURATION *************************************************/

type Config interface {

	//function to get the next server
	Next() *Instance.ServerRoute
	//function to build the server pool
	BuildPool(string)
	//function to balance the load
	LoadBalance(http.ResponseWriter, *http.Request)
	//function for health check
	HealthCheck()
}
