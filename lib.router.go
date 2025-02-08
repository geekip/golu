package main

import (
	"log"
	"net/http"
	"strings"
	"sync"

	router "github.com/geekip/mux"
	lua "github.com/yuin/gopher-lua"
)

var (
	routerInstance *router.Mux
	routerOnce     sync.Once
)

func routerLoader(L *lua.LState) int {
	routerOnce.Do(func() {
		routerInstance = router.New()
	})

	api := map[string]lua.LGFunction{
		"listen":    apiListen,
		"handle":    apiHandle,
		"serveFile": apiServeFile,
		"serveDir":  apiServeDir,
		"all":       apiMethod("*"),
		"connect":   apiMethod(http.MethodConnect),
		"delete":    apiMethod(http.MethodDelete),
		"get":       apiMethod(http.MethodGet),
		"head":      apiMethod(http.MethodHead),
		"options":   apiMethod(http.MethodOptions),
		"patch":     apiMethod(http.MethodPatch),
		"post":      apiMethod(http.MethodPost),
		"put":       apiMethod(http.MethodPut),
		"trace":     apiMethod(http.MethodTrace),
	}
	L.Push(L.SetFuncs(L.NewTable(), api))
	return 1
}

func apiListen(L *lua.LState) int {
	addr := L.CheckString(1)
	log.Printf("Starting server on %s\n", addr)
	if err := http.ListenAndServe(addr, routerInstance); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
	return 0
}

func apiServeDir(L *lua.LState) int {
	prefix := L.CheckString(1)
	path := L.CheckString(2)
	routerInstance.HandlerFunc(prefix, func(w http.ResponseWriter, req *http.Request) {
		params := router.Params(req)
		basePath := strings.TrimSuffix(req.URL.Path, params["*"])
		http.StripPrefix(basePath, http.FileServer(http.Dir(path))).ServeHTTP(w, req)
	})
	return 0
}

func apiServeFile(L *lua.LState) int {
	prefix := L.CheckString(1)
	path := L.CheckString(2)
	routerInstance.HandlerFunc(prefix, func(w http.ResponseWriter, req *http.Request) {
		http.ServeFile(w, req, path)
	})
	return 0
}

func apiHandle(L *lua.LState) int {
	return handleRoute(L, L.CheckString(1))
}

func apiMethod(method string) lua.LGFunction {
	return func(L *lua.LState) int {
		return handleRoute(L, method)
	}
}

func handleRoute(L *lua.LState, method string) int {
	path := L.CheckString(1)
	handler := L.CheckFunction(2)

	routerInstance.Method(method).HandlerFunc(path, func(w http.ResponseWriter, req *http.Request) {
		// Create new LState per request
		RL := lua.NewState()
		defer RL.Close()

		RL.Push(handler)
		RL.Push(lua.LString(req.Method))
		RL.Push(lua.LString(req.URL.Path))

		params := router.Params(req)
		paramsL := RL.NewTable()
		for k, v := range params {
			RL.SetField(paramsL, k, lua.LString(v))
		}
		RL.Push(paramsL)

		if err := RL.PCall(3, 1, nil); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Process response
		retVal := RL.Get(-1)
		body, headers, status := parseLuaResponse(retVal)

		// Set headers
		for k, v := range headers {
			w.Header().Set(k, v)
		}

		// Send response
		w.WriteHeader(status)
		w.Write([]byte(body))
	})

	return 0
}

func parseLuaResponse(retVal lua.LValue) (string, map[string]string, int) {
	headers := map[string]string{
		"Content-Type": "text/html;charset=UTF-8",
		"Server":       ReleasesName,
	}
	status := http.StatusOK
	var body string

	if respTable, ok := retVal.(*lua.LTable); ok {

		// Extract headers
		if headersLV := respTable.RawGetString("headers"); headersLV != lua.LNil {
			if headersTable, ok := headersLV.(*lua.LTable); ok {
				headersTable.ForEach(func(k, v lua.LValue) {
					headers[k.String()] = v.String()
				})
			}
		}
		// Extract body
		if bodyLV := respTable.RawGetString("body"); bodyLV != lua.LNil {
			body = bodyLV.String()
		}

		// Extract status code
		if statusLV := respTable.RawGetString("status"); statusLV != lua.LNil {
			if statusCode, ok := statusLV.(lua.LNumber); ok {
				status = int(statusCode)
				if status < 100 || status >= 600 {
					status = http.StatusOK
				}
			}
		}

	} else {
		body = retVal.String()
	}

	return body, headers, status
}
