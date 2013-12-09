package main

import (
	"flag"
	"fmt"
	"errors"
	// "io"
	"log"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httputil"
	"encoding/json"
)

var (
	containerId string
	socketPath  string
)

func getJsonBytes(socketPath string, path string) ([]byte, error) {
	req, err := http.NewRequest("GET", path, nil)
	if err != nil {
        return nil, err
	}
	dial, err := net.Dial("unix", socketPath)
	if err != nil {
        return nil, err
	}

	var resp *http.Response
	clientconn := httputil.NewClientConn(dial, nil)
	resp, err = clientconn.Do(req)
	defer clientconn.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		return nil, errors.New("bad status code")
	}


	if resp.Header.Get("Content-Type") != "application/json" {
        return nil, errors.New("expected application/json")
	}

	defer resp.Body.Close()
    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return nil, err
    }

	return body, nil
}


type address struct {
	HostIp string
	HostPort string
}


type networkSettings struct {
	IPAddress string
	Gateway string
	Ports map[string][]address
}

type container struct {
    ID string
    Created string
    Path string
    NetworkSettings networkSettings
}


type myContainer struct {
	Id string
	IpAddress string
	Gateway string
	Ports map[string]string
}


func getJsonBytesResp(socketPath string, containerId string) ([]byte, error) {
	b, err := getJsonBytes(socketPath, "/containers/"+containerId+"/json")
    if err != nil {
        return nil, err
    }

	var t container
	err = json.Unmarshal(b, &t)
    if err != nil {
        return nil, err
    }

    out := new(myContainer)
    out.Id = t.ID
    out.IpAddress = t.NetworkSettings.IPAddress
    out.Gateway = t.NetworkSettings.Gateway
    out.Ports = make(map[string]string)


	for k, v := range t.NetworkSettings.Ports {
		for _, add := range v {
			out.Ports[k] = add.HostPort
		}
	}

	bOut, err := json.Marshal(out)
    if err != nil {
        return nil, err
    }

    return bOut, nil
}

func main() {

	flag.StringVar(&socketPath, "s", "/var/run/docker.sock", "unix socket to connect to")
	flag.Parse()



	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
	    cid := r.URL.Path[len("/"):]
	 	log.Println("handling "+cid)

		b, err := getJsonBytesResp(socketPath, cid)
	    if err != nil {
	        return
	    }

		s := string(b[:])

		fmt.Fprintf(w, "%s", s)
	})

	log.Fatal(http.ListenAndServe(":8080", nil))




}
