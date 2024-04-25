package utils

import (
	"chandyLamportV2/protobuf/pb"
	"fmt"
	"math/rand"
	"time"
)

func GetPeerAddrWithRole(peerList []*pb.Peer, role pb.CountingRole) (*pb.Peer, error) {
	peerWithRoleList := make([]*pb.Peer, 0)
	for _, peer := range peerList {
		if peer.Role == role {
			peerWithRoleList = append(peerWithRoleList, peer)
		}
	}

	if len(peerWithRoleList) == 0 {
		return nil, fmt.Errorf("peer with role %v not found", role)
	}

	source := rand.NewSource(time.Now().UnixNano())
	random := rand.New(source)

	// Generate random int
	randIndex := random.Intn(len(peerWithRoleList))

	return peerWithRoleList[randIndex], nil

}

func RemovePeer(peerList []*pb.Peer, peerToRemove *pb.Peer) []*pb.Peer {
	var updatedPeerList []*pb.Peer

	for _, peer := range peerList {
		if peer.Addr != peerToRemove.Addr {
			updatedPeerList = append(updatedPeerList, peer)
		}
	}

	return updatedPeerList
}
