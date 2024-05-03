package main

import (
	chLam "chandyLamportV2/chLamLib"
	"chandyLamportV2/protobuf/pb"
	"golang.org/x/net/context"
	"google.golang.org/grpc/metadata"
	"log"
)

type ServiceRegistryServer struct {
	pb.UnimplementedServiceRegistryServer
}

type ReaderCounterServer struct {
	pb.UnimplementedReaderCounterServer
}

type CounterSaverServer struct {
	pb.UnimplementedCounterSaverServer
}

func (s ServiceRegistryServer) NewPeerAdded(_ context.Context, peerToRegister *pb.Peer) (*pb.EmptyResponse, error) {
	peerList = append(peerList, peerToRegister)
	log.Printf("[NewPeerAdded] peerList updated: %v", peerList)

	// Register peer address into chLam library
	chLam.RegisterNewPeer(peerToRegister.Addr)

	return &pb.EmptyResponse{}, nil
}

//func (s ReaderCounterServer) CountWord(_ context.Context, line *pb.Line) (*pb.EmptyResponse, error) {
//	log.Printf("[CountWord]: %v", line)
//
//	globalLineListMutex.Lock()
//	globalLineList = append(globalLineList, line)
//	globalLineListMutex.Unlock()
//
//	count := 0
//	for {
//		if err := countLine(line); err != nil {
//			if count < 10 {
//				time.Sleep(time.Second)
//
//				count++
//
//				continue
//			}
//
//			return nil, err
//		}
//
//		break
//	}
//
//	globalTotalCountMutex.Lock()
//	delete(globalTotalCount, line.FileName)
//	globalTotalCountMutex.Unlock()
//
//	return &pb.EmptyResponse{}, nil
//}

func (s ReaderCounterServer) CountWord(ctx context.Context, line *pb.Line) (*pb.EmptyResponse, error) {
	log.Printf("[CountWord]: %v", line)

	md, _ := metadata.FromIncomingContext(ctx)

	// Add to list of lines to precess
	lineListMutex.Lock()
	lineList = append(lineList, wordCountData{
		Addr:     md.Get(ctxKey)[0],
		FileName: line.FileName,
		Line:     line.Line,
	})
	lineListMutex.Unlock()

	return &pb.EmptyResponse{}, nil
}

//func (s CounterSaverServer) SaveCount(_ context.Context, lineCount *pb.LineCount) (*pb.EmptyResponse, error) {
//	log.Printf("[SaveCount]: %v", lineCount)
//
//	if err := storeTotalCount(lineCount); err != nil {
//		return nil, err
//	}
//
//	return &pb.EmptyResponse{}, nil
//}

func (s CounterSaverServer) SaveCount(ctx context.Context, lineCount *pb.LineCount) (*pb.EmptyResponse, error) {
	md, _ := metadata.FromIncomingContext(ctx)
	log.Printf("[SaveCount]: %v", lineCount)

	lineCountListMutex.Lock()
	lineCountList = append(lineCountList, saverData{
		Addr:      md.Get(ctxKey)[0],
		FileName:  lineCount.FileName,
		LineCount: lineCount.LineCount,
	})
	lineCountListMutex.Unlock()

	return &pb.EmptyResponse{}, nil
}
