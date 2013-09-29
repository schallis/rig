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
			{"/list": getList},
			{"/ps": getPs},
			{"/resolve": getResolve},
			{"/version": getVersion},
		},
		"POST": {
			{"/{stack:.*}/{service:.*}/{process:.*}/start": postProcessStart},
			{"/{stack:.*}/{service:.*}/{process:.*}/stop": postProcessStop},
			{"/{stack:.*}/{service:.*}/{process:.*}/tail": postProcessTail},
			{"/{stack:.*}/{service:.*}/start": postServiceStart},
			{"/{stack:.*}/{service:.*}/stop": postServiceStop},
			{"/{stack:.*}/{service:.*}/tail": postServiceTail},
			{"/{stack:.*}/restart": postStackRestart},
			{"/{stack:.*}/start": postStackStart},
			{"/{stack:.*}/stop": postStackStop},
			{"/{stack:.*}/tail": postStackTail},
			{"/config/reload": postConfigReload},
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

func getList(srv *Server, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	stacks := make(map[string]map[string][]string)
	for stackName, s := range srv.Stacks {
		stacks[stackName] = make(map[string][]string)
		for serviceName, svc := range s.Services {
			processes := []string{}
			for processName, _ := range svc.Processes {
				processes = append(processes, processName)
			}
			stacks[stackName][serviceName] = processes
		}
	}

	b, err := json.Marshal(stacks)
	if err != nil {
		return err
	}
	writeJSON(w, b)

	return nil
}

func getPs(srv *Server, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	stacks := make(map[string]map[string][]*rig.ApiProcess)
	for stackName, s := range srv.Stacks {
		stacks[stackName] = make(map[string][]*rig.ApiProcess)
		for serviceName, svc := range s.Services {
			processes := []*rig.ApiProcess{}
			for _, p := range svc.Processes {
				if p.Process != nil {
					apiProcess := &rig.ApiProcess{
						Name:   p.Name,
						Pid:    p.Process.Pid,
						Status: int(p.Status),
					}
					processes = append(processes, apiProcess)
				}
			}
			stacks[stackName][serviceName] = processes
		}
	}

	b, err := json.Marshal(stacks)
	if err != nil {
		return err
	}
	writeJSON(w, b)

	return nil
}

func getResolve(srv *Server, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	if vars == nil {
		return fmt.Errorf("Missing parameter")
	}

	if err := r.ParseForm(); err != nil {
		return err
	}

	descriptor := r.Form.Get("descriptor")
	pwd := r.Form.Get("pwd")

	d, err := srv.Resolve(descriptor, pwd)
	if err != nil {
		return err
	}

	b, err := json.Marshal(d)
	if err != nil {
		return err
	}
	writeJSON(w, b)
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

	subCh := make(chan rig.ProcessOutputMessage)
	srv.TailProcess(d, subCh)

	for {
		select {
		case msg := <-subCh:
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

func postServiceTail(srv *Server, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	if vars == nil {
		return fmt.Errorf("Missing parameter")
	}
	d := buildDescriptor(vars)

	w.Header().Set("Content-Type", "application/json")

	subCh := make(chan rig.ProcessOutputMessage)
	srv.TailService(d, subCh)

	for {
		select {
		case msg := <-subCh:
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
	if vars == nil {
		return fmt.Errorf("Missing parameter")
	}
	d := buildDescriptor(vars)

	w.Header().Set("Content-Type", "application/json")

	subCh := make(chan rig.ProcessOutputMessage)
	srv.TailStack(d, subCh)

	for {
		select {
		case msg := <-subCh:
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

func postConfigReload(srv *Server, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	err := srv.ReloadConfig()
	if err != nil {
		return err
	}

	return nil
}

func buildDescriptor(vars map[string]string) *rig.Descriptor {
	return &rig.Descriptor{
		Stack:   vars["stack"],
		Service: vars["service"],
		Process: vars["process"],
	}
}
