package main

import (
	"chandyLamportV2/protobuf/pb"
	"chandyLamportV2/utils"
	"context"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ServiceRegistryServer struct {
	pb.UnimplementedServiceRegistryServer
}

var peerList []*pb.Peer

// RegisterPeer Service to register a new process in the process of available processes
func (s ServiceRegistryServer) RegisterPeer(_ context.Context, pToRec *pb.Peer) (*pb.RegisterPeerResponse, error) {
	// Create new peer to add to list with address of peer service
	newPeer := pb.Peer{
		Role: pToRec.Role,
		Addr: pToRec.Addr,
	}

	// emptyPeerList is used to manage contacted peer failures
	var alivePeerList []*pb.Peer
	// Update the list of each peer with newPeer
	for i := range peerList {
		// Informs a peer in PeerList of the entry of a new peer
		err := callServiceNewPeerAdded(peerList[i], &newPeer)
		if err != nil {
			continue
		}

		alivePeerList = append(alivePeerList, peerList[i])
	}
	peerList = alivePeerList

	// Create response with list without newPeer
	registerPeerResponse := pb.RegisterPeerResponse{
		Peer:     pToRec,
		PeerList: peerList,
	}

	// Append newPeer to PeerList
	peerList = append(peerList, &newPeer)

	// Print some log to show ServiceRegistry workload
	log.Print("peerList:", peerList)

	return &registerPeerResponse, nil
}

func callServiceNewPeerAdded(peerToCall *pb.Peer, newPeer *pb.Peer) error {
	conn, err := grpc.Dial(peerToCall.Addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Did not connect: %v", err)
	}
	defer func(conn *grpc.ClientConn) {
		_ = conn.Close()
	}(conn)

	peerService := pb.NewServiceRegistryClient(conn)

	// Register process to Service Registry
	_, err = peerService.NewPeerAdded(context.Background(), newPeer)
	if err != nil {
		return err
	}

	return nil
}

// getServiceRegistryServer return Service Registry server that handle peers registration
func getServiceRegistryServer(addr string, port string) (*grpc.Server, net.Listener) {
	/* Initialize Service Registry service */
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%s", addr, port))
	if err != nil {
		log.Fatalf("Failed to start the Service Registry: %s", err)
	}

	// Create a gRPC server with no service registered
	serverRegister := grpc.NewServer()

	// Register ServiceRegistry as a service
	serviceRegistry := ServiceRegistryServer{}
	pb.RegisterServiceRegistryServer(serverRegister, serviceRegistry)

	return serverRegister, lis
}

func main() {
	srAddr, srPort, _, _ := utils.FetchArgs()
	server, lis := getServiceRegistryServer(srAddr, srPort)

	log.Printf("Service Registry started")

	// Listen for Remote Procedure Call
	if err := server.Serve(lis); err != nil {
		log.Fatalf("Failed to serve process: %s", err)
	}
}
