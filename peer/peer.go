package main

import (
	"chandyLamportV2/chLamLib"
	"chandyLamportV2/protobuf/pb"
	"chandyLamportV2/utils"
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"net"
)

const ctxKey = "countingPeerAddr"

var (
	peerServiceAddr string
	role            pb.CountingRole

	peerList []*pb.Peer // Pointer so the list will be updated with same one of service registry
)

// initPeerServiceServer serves to initialize peer server
func initPeerServiceServer(peerAddr string, peerRole pb.CountingRole) (net.Listener, *grpc.Server) {
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:", peerAddr))
	if err != nil {
		log.Fatalf("Failed to start the Peer service: %s", err)
	}

	// Get a gRPC server from chLam library
	server := chLam.NewServer()

	srSrv := ServiceRegistryServer{}
	pb.RegisterServiceRegistryServer(server, srSrv)

	switch peerRole {
	case pb.CountingRole_WORD_COUNTER:
		readCounterSrv := ReaderCounterServer{}
		pb.RegisterReaderCounterServer(server, readCounterSrv)

		break
	case pb.CountingRole_SAVER:
		counterSaverSrv := CounterSaverServer{}
		pb.RegisterCounterSaverServer(server, counterSaverSrv)

		break
	}

	return lis, server
}

// RegisterPeerServiceOnServiceRegistry register peer service on Service Registry
func regSvcAddrOnSrvReg(srAddr string, srPort string, peerServerAddr string, role pb.CountingRole) (*pb.Peer, []*pb.Peer) {
	conn, err := grpc.Dial(fmt.Sprintf("%s:%s", srAddr, srPort), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Did not connect: %v", err)
	}
	defer func(conn *grpc.ClientConn) {
		_ = conn.Close()
	}(conn)

	srClient := pb.NewServiceRegistryClient(conn)

	ctx := chLam.SetContextChLam(context.Background(), peerServerAddr)

	// Register process to Service Registry
	registerResponse, err := srClient.RegisterPeer(ctx, &pb.Peer{Role: role, Addr: peerServerAddr})
	if err != nil {
		log.Fatalf("Error when calling RegisterPeer: %s", err)
	}

	return registerResponse.Peer, registerResponse.PeerList
}

func main() {
	var srAddr, srPort, peerHostname string

	srAddr, srPort, peerHostname, role = utils.FetchArgs()
	lis, server := initPeerServiceServer(peerHostname, role)

	peerServiceAddr = lis.Addr().String()

	fmt.Printf("New peer service on address: %s\n", peerServiceAddr)

	var snapFileName = fmt.Sprintf("gobSnap_%v.gob", role.String())

	// Register peer service address into the chLam library
	chLam.RegisterServerAddr(peerServiceAddr)

	chLam.RegisterSnapFileName(snapFileName)

	chLam.RegisterFunctionForSnapshot(peerSnapshot)

	chLam.RegisterFunctionRetrieveChLamClient(retrieveChLamClient)

	// Used to register all reflect.Type used in gRPC calls
	registerCountingType()

	// restore snapshot struct
	var err error
	var data interface{}
	if data, err = chLam.RetrieveDataSnapshot(snapFileName); err != nil {
		log.Printf("[chLam.RetrieveDataSnapshot(snapFileName)] no file to retrieve snapshot, filename: %s\n", snapFileName)
	}

	var snapStruct *CountingSnapshot = nil
	if data != nil {
		snapStruct = data.(*CountingSnapshot)
		log.Print(snapStruct)
	}

	if err = chLam.RestoreMethodsSnapshot(snapFileName); err != nil {
		log.Printf("[retrieveSnap] failed to restore methods from snap: %v", err)
	}

	_, peerList = regSvcAddrOnSrvReg(srAddr, srPort, peerServiceAddr, role)
	for _, peer := range peerList {
		// Register peer address into chLam library
		chLam.RegisterNewPeer(peer.Addr)
	}

	switch role {
	case pb.CountingRole_FILE_READER:
		if snapStruct == nil {
			go jobReadFile("", 0)
		} else {
			go jobReadFile(snapStruct.ReadFileFileName, snapStruct.ReadFileCurrLine)
		}

		break
	case pb.CountingRole_WORD_COUNTER:
		if snapStruct == nil {
			go jobWordCount(make([]wordCountData, 0), make(map[string]int32))
		} else {
			go jobWordCount(snapStruct.WordCountLineList, snapStruct.WordCountWordCount)
		}

		break
	case pb.CountingRole_SAVER:
		if snapStruct == nil {
			go jobSaver(make([]saverData, 0), make(map[string]map[string]map[string]int32))
		} else {
			go jobSaver(snapStruct.SaverLineCountList, snapStruct.SaverTotalCount)
		}
		break
	}

	// Start chandy lamport snapshot randomly
	go startRandomlyChLamSnapshot(peerServiceAddr)

	// Listen for RPC calls
	if err := server.Serve(lis); err != nil {
		log.Fatalf("Failed to serve process over port []: %s", err)
	}
}
