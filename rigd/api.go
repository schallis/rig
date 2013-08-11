package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net"
	"net/http"
	"strings"
)

func ListenAndServe(addr string, srv *Server) error {
	log.Printf("Listening for HTTP on %s\n", addr)

	r, err := makeRouter(srv)
	if err != nil {
		return err
	}

	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	httpServer := http.Server{Addr: addr, Handler: r}
	return httpServer.Serve(l)
}

func makeRouter(srv *Server) (*mux.Router, error) {
	r := mux.NewRouter()

	mapRoutes := map[string]map[string]func(*Server, http.ResponseWriter, *http.Request, map[string]string) error{
		"GET": {
			"/version": getVersion,
			"/resolve": getResolve,
		},
		"POST": {
			"/{stack:.*}/{service:.*}/{process:.*}/start": postProcessStart,
			"/{stack:.*}/{service:.*}/{process:.*}/stop":  postProcessStop,
			"/{stack:.*}/{service:.*}/start":              postServiceStart,
			"/{stack:.*}/{service:.*}/stop":               postServiceStop,
			"/{stack:.*}/start":                           postStackStart,
			"/{stack:.*}/stop":                            postStackStop,
			"/{stack:.*}/tail":                            postStackTail,
			"/{stack:.*}/restart":                         postStackRestart,
		},
	}

	for method, routes := range mapRoutes {
		for route, handlerFunc := range routes {
			currentRoute := route
			currentMethod := method
			currentHandlerFunc := handlerFunc
			f := func(w http.ResponseWriter, r *http.Request) {
				if err := currentHandlerFunc(srv, w, r, mux.Vars(r)); err != nil {
					httpError(w, err)
				}
			}

			r.Path(currentRoute).Methods(currentMethod).HandlerFunc(f)
		}
	}

	return r, nil
}

func httpError(w http.ResponseWriter, err error) {
	if strings.HasPrefix(err.Error(), "No such") {
		http.Error(w, err.Error(), http.StatusNotFound)
	} else if strings.HasPrefix(err.Error(), "Bad parameter") {
		http.Error(w, err.Error(), http.StatusBadRequest)
	} else if strings.HasPrefix(err.Error(), "Impossible") {
		http.Error(w, err.Error(), http.StatusNotAcceptable)
	} else {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func writeJSON(w http.ResponseWriter, b []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}

func getResolve(srv *Server, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	return nil
}

func getVersion(srv *Server, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	v := srv.Version()
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	writeJSON(w, b)
	return nil
}

func postStackRestart(srv *Server, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	return nil
}

func postProcessStart(srv *Server, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	if vars == nil {
		return fmt.Errorf("Missing parameter")
	}
	stack := vars["stack"]
	service := vars["service"]
	process := vars["process"]

	if err := srv.StartProcess(stack, service, process); err != nil {
		log.Println(err)
	}

	return nil
}

func postProcessStop(srv *Server, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	if vars == nil {
		return fmt.Errorf("Missing parameter")
	}
	stack := vars["stack"]
	service := vars["service"]
	process := vars["process"]

	if err := srv.StopProcess(stack, service, process); err != nil {
		log.Println(err)
	}

	return nil
}

func postServiceStart(srv *Server, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	if vars == nil {
		return fmt.Errorf("Missing parameter")
	}
	stack := vars["stack"]
	service := vars["service"]

	if err := srv.StartService(stack, service); err != nil {
		log.Println(err)
	}

	return nil
}

func postServiceStop(srv *Server, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	if vars == nil {
		return fmt.Errorf("Missing parameter")
	}
	// stack := vars["stack"]
	// service := vars["service"]

	// if err := srv.StopService(stack, service); err != nil {
	// 	log.Println(err)
	// }

	return nil
}

func postStackStart(srv *Server, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	if vars == nil {
		return fmt.Errorf("Missing parameter")
	}
	stack := vars["stack"]

	if err := srv.StartStack(stack); err != nil {
		log.Println(err)
	}

	return nil
}

func postStackStop(srv *Server, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	if vars == nil {
		return fmt.Errorf("Missing parameter")
	}
	// stack := vars["stack"]

	// if err := srv.StopStack(stack); err != nil {
	// 	log.Println(err)
	// }

	return nil
}

func postStackTail(srv *Server, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	return nil
}
