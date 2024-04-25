package chLam

import (
	"chandyLamportV2/chLamLib/chLamProto"
	"chandyLamportV2/chLamLib/utils"
	"fmt"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"reflect"
)

func StartSnapshot() error {
	// * Store snapshot
	if err := storePeerSnapshot(); err != nil {
		return err
	}

	isSnapMutex.Lock()
	isSnap = !isSnap
	isSnapMutex.Unlock()

	// * Send snapshot request to all other peers
	go sendChLamRequestToAllPeers()

	return nil
}

func storePeerSnapshot() error {
	// 	Get peer's system snapshot
	var peerSnap, err = takePeerSnapshot() // Invoke function to get peerSnapshot struct
	if err != nil {
		return fmt.Errorf("error StartSnapshot calling takePeerSnapshot: %v", err)
	}

	peerSnapBytes, err := utils.ConvertInterfaceToBytes(&peerSnap)
	if err != nil {
		return fmt.Errorf("error StartSnapshot [utils.ConvertInterfaceToBytes]: %v", err)
	}

	// Take Req type to save it in methodSnap
	peerSnapType := reflect.TypeOf(peerSnap)
	// Check if peerSnapType is a pointer, to get data type without pointer
	if peerSnapType.Kind() == reflect.Ptr {
		peerSnapType = peerSnapType.Elem()
	}

	// Get Req type string
	peerSnapTypeStr := peerSnapType.String()

	// Register snapshot into the snapshotWrap
	snapshotWrap.Data = interfaceSnap{
		Bytes: peerSnapBytes,
		Type:  peerSnapTypeStr,
	}

	return nil
}

func sendChLamRequestToAllPeers() {
	var emptyPeerAddrList = make([]string, 0)
	for _, peerAddr := range peerAddrList {
		if err := sendChLamRequest(peerAddr); err != nil {
			log.Printf("[sendChLamRequest] error sending chLamRequest to peer: %v", peerAddr)
			if _, exist := peerMap[peerAddr]; !exist {
				delete(peerMap, peerAddr)
			}

			continue
		}

		emptyPeerAddrList = append(emptyPeerAddrList, peerAddr)
	}
	peerAddrList = emptyPeerAddrList
}

// sendMessageToPeer send a string message to the peer "dest" using RPC
func sendChLamRequest(dest string) error {
	conn, err := grpc.Dial(dest, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("error did not connect: %s", err)
	}
	defer func(conn *grpc.ClientConn) {
		_ = conn.Close()
	}(conn)

	client := chLamProto.NewChandyLamportSnapshotClient(conn)

	// Send message to peers
	_, err = client.ChLamSnapshot(context.Background(), &peerServerAddr)
	if err != nil {
		return fmt.Errorf("error when calling ChLamSnapshot on peer: %s, err: %s", dest, err)
	}

	return nil
}
