/*
  inspector is a rest app which will query a Docker socket
  using the Docker API to get gateway / port information about
  a running container.

  It's design, is to be run in a container with /var/run mounted
  locally so it can query the docker instance running on its
  host. This allows another container to link the inspector container
  in and ask for it's route details (i.e. what is the gateway IP
  and what port am i publically exposed on) 

  e.g. 
  request:
	GET http://$FOO_PORT_8080_TCP_ADDR:8080/$HOSTNAME

  response:
        {
            "Id": "30b193cacf05eb8561769857ec798c49f88acb51b4d6129d04767c5e3f49e7d2",
            "IpAddress": "172.17.0.19",
            "Gateway": "172.17.42.1",
            "Ports": {
                "1234/tcp": "49158"
            }
        }


*/
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


var socketPath string


// Types for parsing docker's json response
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


/* Type for encoding our app's response
 */
type myContainer struct {
	Id string
	IpAddress string
	Gateway string
	Ports map[string]string
}


func inspectDockerSocket(socketPath string, path string) ([]byte, error) {
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

func getInspectionResponse(socketPath string, containerId string) ([]byte, error) {
	b, err := inspectDockerSocket(socketPath, "/containers/"+containerId+"/json")
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
		log.Println("inspecting "+cid)

		b, err := getInspectionResponse(socketPath, cid)
		if err != nil {
			return
		}

		s := string(b[:])

	    w.Header().Set("Content-Type", "application/json")
	
		fmt.Fprintf(w, "%s", s)
	})

	log.Fatal(http.ListenAndServe(":8080", nil))

}
