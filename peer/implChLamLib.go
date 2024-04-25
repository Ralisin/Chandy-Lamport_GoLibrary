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

// CountingSnapshot TODO da creare la struct che memorizzi lo snapshot dell'applicazione
type CountingSnapshot struct {
	ReadFileGlobalFileName string
	ReadFileGlobalCurrLine int

	WordCountGlobalWordCount map[string]map[string]int32
	WordCountGlobalLine      []*pb.Line

	SaverGlobalTotalCount map[string]map[string]int32
	SaverGlobalLineCount  *pb.LineCount
}

func registerCountingType() {
	chLam.RegisterType((*pb.Peer)(nil))
	chLam.RegisterType((*pb.CountingRole)(nil))
	chLam.RegisterType((*pb.EmptyResponse)(nil))
	chLam.RegisterType((*pb.Line)(nil))
	chLam.RegisterType((*pb.LineCount)(nil))

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

// TODO popolare la funzione per fare lo snapshot
func peerSnapshot() (interface{}, error) {
	var countingSnap = &CountingSnapshot{}

	// *** Store readFile peer ***
	globalFileNameMutex.Lock()
	countingSnap.ReadFileGlobalFileName = globalFileName
	globalFileNameMutex.Unlock()

	globalCurrLineMutex.Lock()
	countingSnap.ReadFileGlobalCurrLine = globalCurrLine
	globalCurrLineMutex.Unlock()

	// *** Store globalWordCount peer ***
	copiedWordCount := make(map[string]map[string]int32)
	wordCountMutex.Lock()
	for key, innerMap := range globalWordCount {
		copiedWordCount[key] = make(map[string]int32)
		for innerKey, innerValue := range innerMap {
			copiedWordCount[key][innerKey] = innerValue
		}
	}
	wordCountMutex.Unlock()
	countingSnap.WordCountGlobalWordCount = copiedWordCount

	globalLineListMutex.Lock()
	copiedGlobalLineList := make([]*pb.Line, len(globalLineList))
	copy(copiedGlobalLineList, globalLineList)
	globalLineListMutex.Unlock()
	countingSnap.WordCountGlobalLine = copiedGlobalLineList

	// *** Store saver peer ***
	copiedGlobalTotalCount := make(map[string]map[string]int32)
	globalTotalCountMutex.Lock()
	for key, innerMap := range globalTotalCount {
		copiedGlobalTotalCount[key] = make(map[string]int32)
		for innerKey, innerValue := range innerMap {
			copiedGlobalTotalCount[key][innerKey] = innerValue
		}
	}
	globalTotalCountMutex.Unlock()
	countingSnap.SaverGlobalTotalCount = copiedGlobalTotalCount

	globalLineCountMutex.Lock()
	countingSnap.SaverGlobalLineCount = globalLineCount
	globalLineCountMutex.Unlock()

	return countingSnap, nil
}
