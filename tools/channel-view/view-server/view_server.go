// Copyright 2018-2020 Celer Network

// This binary provides a web interface for channel view tool so we can easily check channel and pay
// status in browser.
package main

import (
	"flag"
	"fmt"
	"net/http"
	"os/exec"
	"strings"
)

var (
	profile       = flag.String("profile", "config/profile.json", "Path to profile json file")
	viewBinary    = flag.String("viewbin", "/usr/local/bin/channel-view", "location of view binary")
	storedir      = flag.String("storedir", "", "sql api entry")
	staticFileDir = flag.String("staticfiledir", "/etc/cv_static/", "location of static file dir")
	port          = flag.String("port", "10080", "port to serve on")
)

func view(argsToAppend ...string) ([]byte, error) {
	commonArgs := []string{
		"-profile", *profile,
		"-storedir", *storedir,
	}
	args := make([]string, len(commonArgs), len(commonArgs)+len(argsToAppend))
	copy(args, commonArgs)
	args = append(args, argsToAppend...)
	cmd := exec.Command(*viewBinary, args...)
	return cmd.CombinedOutput()
}
func main() {
	flag.Parse()
	// Each handler handles one HTML "form" in static/index.html
	http.HandleFunc("/db/channel", func(w http.ResponseWriter, r *http.Request) {
		token := r.URL.Query()["token"]
		peer := r.URL.Query()["peer"]
		if len(token) != 1 || len(peer) != 1 {
			fmt.Fprintf(w, "Input Wrong")
			return
		}
		out, err := view("-dbview", "channel", "-peer", peer[0], "-token", token[0])
		if err != nil {
			fmt.Fprintf(w, "cmd.Run() failed with %s\n", err)
		}
		fmt.Fprintf(w, string(out))
	})
	http.HandleFunc("/db/channel-cid", func(w http.ResponseWriter, r *http.Request) {
		cid := r.URL.Query()["cid"]
		if len(cid) != 1 {
			fmt.Fprintf(w, "Input Wrong")
			return
		}
		out, err := view("-dbview", "channel", "-cid", cid[0])
		if err != nil {
			fmt.Fprintf(w, "cmd.Run() failed with %s\n", err)
		}
		fmt.Fprintf(w, string(out))
	})
	http.HandleFunc("/db/pay", func(w http.ResponseWriter, r *http.Request) {
		payid := r.URL.Query()["payid"]
		if len(payid) != 1 {
			fmt.Fprintf(w, "Input Wrong")
			return
		}
		out, err := view("-dbview", "pay", "-payid", payid[0])
		if err != nil {
			fmt.Fprintf(w, "cmd.Run() failed with %s\n", err)
		}
		fmt.Fprintf(w, string(out))
	})
	http.HandleFunc("/db/deposit-id", func(w http.ResponseWriter, r *http.Request) {
		depositid := r.URL.Query()["depositid"]
		if len(depositid) != 1 {
			fmt.Fprintf(w, "Input Wrong")
			return
		}
		out, err := view("-dbview", "deposit", "-depositid", depositid[0])
		if err != nil {
			fmt.Fprintf(w, "cmd.Run() failed with %s\n", err)
		}
		fmt.Fprintf(w, string(out))
	})
	http.HandleFunc("/db/deposits-cid", func(w http.ResponseWriter, r *http.Request) {
		cid := r.URL.Query()["cid"]
		if len(cid) != 1 {
			fmt.Fprintf(w, "Input Wrong")
			return
		}
		out, err := view("-dbview", "deposit", "-cid", cid[0])
		if err != nil {
			fmt.Fprintf(w, "cmd.Run() failed with %s\n", err)
		}
		fmt.Fprintf(w, string(out))
	})
	http.HandleFunc("/onchain/channel", func(w http.ResponseWriter, r *http.Request) {
		cid := r.URL.Query()["cid"]
		if len(cid) != 1 {
			fmt.Fprintf(w, "Input Wrong")
			return
		}
		out, err := view("-chainview", "channel", "-cid", cid[0])
		if err != nil {
			fmt.Fprintf(w, "cmd.Run() failed with %s\n", err)
		}
		fmt.Fprintf(w, string(out))
	})
	http.HandleFunc("/onchain/pay", func(w http.ResponseWriter, r *http.Request) {
		payid := r.URL.Query()["payid"]
		if len(payid) != 1 {
			fmt.Fprintf(w, "Input Wrong")
			return
		}
		out, err := view("-chainview", "pay", "-payid", payid[0])
		if err != nil {
			fmt.Fprintf(w, "cmd.Run() failed with %s\n", err)
		}
		fmt.Fprintf(w, string(out))
	})
	http.HandleFunc("/onchain/app", func(w http.ResponseWriter, r *http.Request) {
		appAddr := r.URL.Query()["app_addr"]
		argOutcome := r.URL.Query()["arg_outcome"]
		argFinalize := r.URL.Query()["arg_finalize"]
		if len(appAddr) != 1 || len(argOutcome) != 1 {
			fmt.Fprintf(w, "Input Wrong")
			return
		}
		var err error
		var out []byte
		if argFinalize == nil || len(argFinalize) == 0 {
			out, err = view("-chainview", "app", "-appaddr", appAddr[0], "-outcome", argOutcome[0])
		} else {
			out, err = view("-chainview", "app", "-appaddr", appAddr[0], "-outcome", argOutcome[0], "-finalize", argFinalize[0])
		}
		if err != nil {
			fmt.Fprintf(w, "cmd.Run() failed with %s\n", err)
		}
		fmt.Fprintf(w, string(out))
	})
	fs := http.FileServer(FileSystem{http.Dir(*staticFileDir)})
	http.Handle("/", fs)

	fmt.Println("Running view server on port ", *port)
	http.ListenAndServe(":"+*port, nil)
}

// FileSystem custom file system handler
type FileSystem struct {
	fs http.FileSystem
}

// Open opens file
func (fs FileSystem) Open(path string) (http.File, error) {
	f, err := fs.fs.Open(path)
	if err != nil {
		return nil, err
	}

	s, err := f.Stat()
	if s.IsDir() {
		index := strings.TrimSuffix(path, "/") + "/index.html"
		if _, err2 := fs.fs.Open(index); err != nil {
			return nil, err2
		}
	}

	return f, nil
}
