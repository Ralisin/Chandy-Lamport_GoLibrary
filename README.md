# Chandy-Lamport
## SDCC project - take a snapshot of a distributed system

Chandy-Lamport algorithm is the first algorithm proposed for taking a distributed snapshot.

We can highlight its assumptions:
- The distributed system consist of a finite set of processes and a finite set of channels
- Channels are assumed to have infinite buffers, be error-free and deliver messages in FIFO (First In First Out) order
- The delay experienced by a message in a channel is arbitrary but finite
- The sequence of messages received along a channel is an initial subsequence of the sequence of messages sent along the channel.
- The state of a channel is the sequence of messages sent along the channel, excluding the messages received along the channel.

### Algorithm

The algorithm can be split in two distinct parts.

The initiator process(es) (one or more) that start the algorithm
```
Initiator process:
    Records its status
    Send marker message to all his outgoing channels
    Start recording messages from all channels
```

The other processes P<sub>i</sub>
```
The process receive a marker message on channel C_{k,i}:
    If is the first time process P_i sees a marker message (sent or receive)
        P_i records its status
        P_i marks the C_{k,i} channel as empty
        P_i send marker message to all its known outgoing channels
        P_i start recording messages from all channels
    Otherwise
        P_i stor recording messages coming from channel C_{k,i}
```

## Chandy Lamport library
The algorithm was implemented within the chLamLib folder.

Here you can go and analyse the code I have written to be able to use the snapshot algorithm in library mode. by a programmer   
who is developing a distributed system in Go using gRPC technology.

The library is designed to be used by a programmer who is developing a distributed application using the Go language and the gRPC tool.

<div style="display: flex;">
<img src="https://grpc.io/img/logos/grpc-logo.png" width="125" style="margin-right: 50px;" alt="gRPC">
<img src="https://go.dev/images/go-logo-white.svg" width="125"  alt="GoLang">
</div>

### Library utilization

Things the library user must do:
- To set up his server, instead of `grpc.NewServer`, he must go and use the library function `chLam.NewServer()`
  - it is used exactly like `grpc.NewServer`, only it integrates an internal library interceptor to do internal snapshot
    utilities
- It must register all its data types inside the library via `chLam.RegisterType`
  - the library function `chLam.RegisterType` merely calls a more internal library function, namely `utils.RegisterType` 
    which is fundamental to the correct operation of the library: what you need to pass to it is the pointer to the
    struct you have defined, and not a pointer to pointer. Take a look in the code in the folder `peer/implChLamLib.go`,
    the function `registerCountingType`
- It must provide a function with the following interface: func() (interface{}, error)
    - this function must snapshot the current state of the peer and return an interface{}. For a reference look at
      `peer/implChLamLib.go` function `peerSnapshot()`
    - The interface can either be the struct directly or also pointer to the struct returned with snapshot inside.
      The library can decently handle both cases
- Whenever it becomes aware of a new peer server, it must invoke the `chLam.RegisterNewPeer` method and pass it the
  address on which the service is made available
  - The library allows the snapshot to be taken on registered peers only. In this way, it is possible to take even a
    partial snapshot of the system
  - the registered peer must implement this library and must be able to receive remote procedure calls
- Every time it makes a call to a remote procedure, it must set the context with its server address as well via
  `chLam.SetContextChLam`. An example of this is in file `peer/jobReadFile.go` in the function `sendLineToCounter` it has
  been used this way: `ctx := chLam.SetContextChLam(context.Background(), peerServiceAddr)`
- It must register the address of its service via the function `chLam.RegisterServerAddr(addr string)`
- It must provide a function with the following interface: `func(string) (interface{}, error)`
  - this function must return the correct client depending on the string passed to it. In particular this function will be passed a string with the name of the full gRPC method
  and this must return the correct client to execute the method invocation. An example is in file `peer/implChLamLib.go`,
  the function `func retrieveChLamClient(methodFullName string) (interface{}, error)`

## AWS EC2 setup
### PowerShell SSH
Connect to EC2 instance via ssh from PowerShell:
```sh
ssh -i <path_to_PEM> ec2-user@<ip-EC2-instance>
```

### EC2 instance
Install docker
```sh
sudo yum update -y
sudo yum install -y docker
```

Install docker-compose
```sh
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose
```

Run docker deamon
```sh
sudo service docker start
```

Install git and clone repository
```sh
sudo yum install git -y
git clone https://github.com/Ralisin/chandyLamportLibrary
```

Run docker compose:
```sh
sudo docker-compose -f compose.yml up
```

### Docker: some useful commands to test the program
Stop one container:
```sh
sudo docker kill <Container ID>
```
Restart one container:
```sh
sudo docker restart <Container ID>
```
Stop all containers:
```sh
sudo docker stop $(sudo docker ps -a -q)
```