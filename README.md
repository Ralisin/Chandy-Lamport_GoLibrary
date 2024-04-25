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

## My implementation
The algorithm was implemented within the chLamLib folder.

Here you can go and analyse the code I have written to be able to use the snapshot algorithm in library mode. by a programmer   
who is developing a distributed system in Go using gRPC technology.

The library is designed to be used by a programmer who is developing a distributed application using the Go language and the gRPC tool.

<img src="https://grpc.io/img/logos/grpc-logo.png" width="150" style="margin-right: 50px;" alt="gRPC">
<img src="https://go.dev/images/go-logo-white.svg" width="150"  alt="GoLang">


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
git clone https://github.com/Ralisin/chandy-lamport
```

Run docker compose:
```sh
sudo docker-compose -f compose.yml up
```

### Tech

Language choose for development is [GoLang], and for process communication is used [gRPC]

[//]: # (Reference links)

[GoLang]: <https://go.dev/>
[gRPC]: <https://grpc.io/>