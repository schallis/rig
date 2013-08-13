package main

import (
	"encoding/json"
	"fmt"
	"github.com/gocardless/rig"
	"github.com/gorilla/mux"
	"log"
	"net"
	"net/http"
	"strings"
)

type RouteHandler func(*Server, http.ResponseWriter, *http.Request, map[string]string) error

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

	mapRoutes := map[string][]map[string]RouteHandler{
		"GET": {
			{"/version": getVersion},
			{"/resolve": getResolve},
		},
		"POST": {
			{"/{stack:.*}/{service:.*}/{process:.*}/start": postProcessStart},
			{"/{stack:.*}/{service:.*}/{process:.*}/stop": postProcessStop},
			{"/{stack:.*}/{service:.*}/{process:.*}/tail": postProcessTail},
			{"/{stack:.*}/{service:.*}/start": postServiceStart},
			{"/{stack:.*}/{service:.*}/stop": postServiceStop},
			{"/{stack:.*}/restart": postStackRestart},
			{"/{stack:.*}/start": postStackStart},
			{"/{stack:.*}/stop": postStackStop},
			{"/{stack:.*}/tail": postStackTail},
		},
	}

	for method, routes := range mapRoutes {
		for _, mapRoute := range routes {
			for route, handlerFunc := range mapRoute {
				registerRoute(srv, r, method, route, handlerFunc)
			}
		}
	}

	return r, nil
}

func registerRoute(srv *Server, r *mux.Router, method, route string, handlerFunc RouteHandler) {
	log.Printf("Registring %s %s", method, route)
	f := func(w http.ResponseWriter, r *http.Request) {
		if err := handlerFunc(srv, w, r, mux.Vars(r)); err != nil {
			httpError(w, err)
		}
	}

	r.Path(route).Methods(method).HandlerFunc(f)
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
	d := buildDescriptor(vars)

	if err := srv.StartProcess(d); err != nil {
		return err
	}

	return nil
}

func postProcessStop(srv *Server, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	if vars == nil {
		return fmt.Errorf("Missing parameter")
	}
	d := buildDescriptor(vars)

	if err := srv.StopProcess(d); err != nil {
		return err
	}

	return nil
}

func postProcessTail(srv *Server, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	if vars == nil {
		return fmt.Errorf("Missing parameter")
	}
	d := buildDescriptor(vars)

	w.Header().Set("Content-Type", "application/json")

	subCh := make(chan ProcessOutputMessage)
	sub, err := srv.TailProcess(d, subCh)
	if err != nil {
		return err
	}
	for {
		select {
		case msg := <-sub.msgCh:
			b, err := json.Marshal(msg)
			if err != nil {
				return err
			}
			w.Write(b)
			w.(http.Flusher).Flush()
		}
	}

	return nil
}

func postServiceStart(srv *Server, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	if vars == nil {
		return fmt.Errorf("Missing parameter")
	}
	d := buildDescriptor(vars)

	if err := srv.StartService(d); err != nil {
		return err
	}

	return nil
}

func postServiceStop(srv *Server, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	if vars == nil {
		return fmt.Errorf("Missing parameter")
	}
	d := buildDescriptor(vars)

	if err := srv.StopService(d); err != nil {
		return err
	}

	return nil
}

func postStackStart(srv *Server, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	if vars == nil {
		return fmt.Errorf("Missing parameter")
	}
	d := buildDescriptor(vars)

	if err := srv.StartStack(d); err != nil {
		return err
	}

	return nil
}

func postStackStop(srv *Server, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	if vars == nil {
		return fmt.Errorf("Missing parameter")
	}
	d := buildDescriptor(vars)

	if err := srv.StopStack(d); err != nil {
		return err
	}

	return nil
}

func postStackTail(srv *Server, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	return nil
}

func buildDescriptor(vars map[string]string) *rig.Descriptor {
	return &rig.Descriptor{
		Stack:   vars["stack"],
		Service: vars["service"],
		Process: vars["process"],
	}
}
