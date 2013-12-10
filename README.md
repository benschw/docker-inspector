## Docker Inspector

### build

	go build insepctor.go
	
	docker build -t benschw/inspector


### run

	docker run -d -v /var/run/:/docker benschw/inspector


### Usage

When you run your container, link in an instance of `benschw/inspector`:

	docker run -i -p 1234 -link crazy_name:foo -t ubuntu /bin/bash

You may then issue an inspect request like:

	GET http://$FOO_PORT_8080_TCP_ADDR:8080/$HOSTNAME

to get back a json object like:

	{
	    "Id": "30b193cacf05eb8561769857ec798c49f88acb51b4d6129d04767c5e3f49e7d2",
	    "IpAddress": "172.17.0.19",
	    "Gateway": "172.17.42.1",
	    "Ports": {
	        "1234/tcp": "49158"
	    }
	}

which tells you how to access your app from outside of your container or the local Docker host (`172.17.42.1:49158`is the public route to `172.17.0.19:1234`)
