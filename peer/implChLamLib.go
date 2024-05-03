package main

import (
	chLam "chandyLamportV2/chLamLib"
	"chandyLamportV2/protobuf/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"math/rand"
	"time"
)

type CountingSnapshot struct {
	ReadFileFileName string
	ReadFileCurrLine int

	WordCountWordCount map[string]int32
	WordCountLineList  []wordCountData

	SaverTotalCount    map[string]map[string]map[string]int32
	SaverLineCountList []saverData
}

func registerCountingType() {
	// Proto structs
	chLam.RegisterType((*pb.Peer)(nil))
	chLam.RegisterType((*pb.CountingRole)(nil))
	chLam.RegisterType((*pb.EmptyResponse)(nil))
	chLam.RegisterType((*pb.Line)(nil))
	chLam.RegisterType((*pb.LineCount)(nil))

	// Internal structs
	chLam.RegisterType((*wordCountData)(nil))
	chLam.RegisterType((*saverData)(nil))
	chLam.RegisterType((*CountingSnapshot)(nil))
}

func retrieveChLamClient(methodFullName string) (interface{}, error) {
	conn, err := grpc.Dial(":", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	defer func(conn *grpc.ClientConn) {
		_ = conn.Close()
	}(conn)

	var clientConn interface{}

	switch methodFullName {
	case "/ReaderCounter/CountWord":
		clientConn = pb.NewReaderCounterClient(conn)
		break
	case "/CounterSaver/SaveCount":
		clientConn = pb.NewCounterSaverClient(conn)
		break
	case "/ServiceRegistry/RegisterPeer", "/ServiceRegistry/NewPeerAdded":
		clientConn = pb.NewServiceRegistryClient(conn)
		break
	}

	return &clientConn, nil
}

func startRandomlyChLamSnapshot(peerServiceAddr string) {
	source := rand.NewSource(time.Now().UnixNano())
	random := rand.New(source)

	// Generate random int
	randWaiting := random.Intn(16) + 15

	for {
		time.Sleep(time.Duration(randWaiting) * time.Second)

		log.Printf("Starting snapshot from peer %s", peerServiceAddr)
		err := chLam.StartSnapshot()
		if err != nil {
			log.Printf("Error starting snapshot: %s", err)
		}
	}
}

func peerSnapshot() (interface{}, error) {
	var countingSnap = &CountingSnapshot{}

	switch role {
	case pb.CountingRole_FILE_READER:
		globFileNameMutex.Lock()
		countingSnap.ReadFileFileName = globFileName
		globFileNameMutex.Unlock()

		globCurrLineMutex.Lock()
		countingSnap.ReadFileCurrLine = globCurrLine
		globCurrLineMutex.Unlock()

		break
	case pb.CountingRole_WORD_COUNTER:
		lineListMutex.Lock()
		// Copy list
		copiedList := make([]wordCountData, len(lineList))
		copy(copiedList, lineList)
		lineListMutex.Unlock()
		countingSnap.WordCountLineList = copiedList

		copiedMap := make(map[string]int32)
		// Copy map
		wordCountMapMutex.Lock()
		for k, v := range wordCountMap {
			copiedMap[k] = v
		}
		wordCountMapMutex.Unlock()
		countingSnap.WordCountWordCount = copiedMap

		break
	case pb.CountingRole_SAVER:
		lineCountListMutex.Lock()
		// Copy list
		copiedList := make([]saverData, len(lineCountList))
		copy(copiedList, lineCountList)
		lineCountListMutex.Unlock()

		copiedMap := make(map[string]map[string]map[string]int32)
		globTotalCountMutex.Lock()
		// Copy map
		for k, v := range globTotalCount {
			copiedMap[k] = make(map[string]map[string]int32)
			for k1, v1 := range v {
				copiedMap[k][k1] = make(map[string]int32)
				for k2, v2 := range v1 {
					copiedMap[k][k1][k2] = v2
				}
			}
		}
		globTotalCountMutex.Unlock()
		countingSnap.SaverTotalCount = copiedMap

		break
	}

	return countingSnap, nil

}
