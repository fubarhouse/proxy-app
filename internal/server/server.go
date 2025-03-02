package server

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

// RunParams is passed to the Run() function.
type RunParams struct {
	// Address which the server will respond to requests.
	Addr string
	// Endpoint which this server will proxy to.
	Endpoint string
	// Username which will be used to authenticate with the proxy endpoint with basic authentication.
	Username string
	// Password which will be used to authenticate with the proxy endpoint with basic authentication.
	Password string
	// TrimPathPrefix from backend requests.
	TrimPathPrefix string
}

// Validate the server parameters.
func (p RunParams) Validate() error {
	if p.Addr == "" {
		return fmt.Errorf("not provided: addr")
	}

	if p.Endpoint == "" {
		return fmt.Errorf("not provided: endpoint")
	}

	return nil
}

// Run the server.
func Run(params RunParams) error {
	if err := params.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	endpoint, err := url.Parse(params.Endpoint)
	if err != nil {
		return fmt.Errorf("failed to parse endpoint: %w", err)
	}

	proxy := httputil.NewSingleHostReverseProxy(endpoint)

	d := proxy.Director

	// Don't recompute this for every request...
	basicAuthHeader := ""
	if params.Username != "" {
		basicAuthHeader = fmt.Sprintf("Basic %s", basicAuth(params.Username, params.Password))
	}

	proxy.Director = func(r *http.Request) {
		d(r) // call default director

		if basicAuthHeader != "" {
			r.Header.Add("Authorization", basicAuthHeader)
		}

		if params.TrimPathPrefix != "" {
			r.URL.Path = strings.TrimPrefix(r.URL.Path, params.TrimPathPrefix)
		}
	}

	// Debug messaging - also add target to LB.
	fmt.Printf("Starting proxy on addr: %s for endpoint: %s\n", params.Addr, endpoint)

	http.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Ready!"))
	})

	http.Handle("/", proxy)

	return http.ListenAndServe(params.Addr, nil)
}

// Helper function to generate basic auth value.
func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}
