package main

import (
	chLam "chandyLamportV2/chLamLib"
	"chandyLamportV2/protobuf/pb"
	"chandyLamportV2/utils"
	"context"
	"errors"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"regexp"
	"strings"
	"sync"
	"time"
)

var (
	// Key1: fileName, Key2: word, Value: count
	globalWordCount = make(map[string]map[string]int32)
	wordCountMutex  sync.Mutex

	globalLineList      []*pb.Line
	globalLineListMutex sync.Mutex
)

func jobWordCount(storedWordCount map[string]map[string]int32, storedLine []*pb.Line) {
	wordCountMutex.Lock()
	globalWordCount = storedWordCount
	wordCountMutex.Unlock()

	globalLineListMutex.Lock()
	globalLineList = storedLine
	globalLineListMutex.Unlock()

	if storedWordCount == nil {
		wordCountMutex.Lock()
		globalWordCount = make(map[string]map[string]int32)
		wordCountMutex.Unlock()
	}

	if storedLine == nil {
		storedLine = make([]*pb.Line, 0)
	}

	var newStoredLine []*pb.Line
	for _, line := range storedLine {
		if line != nil {
			count := 0

			// Try more than once to count word in a line
			for {
				err := countLine(line)
				if err != nil {
					if count > 10 {
						break
					}

					// If no peer with role SAVER has been found
					if err.Error() == "peer with role SAVER not found" {
						count++

						time.Sleep(time.Second)

						continue
					}

					log.Printf("[jobWordCount] error restoring snapshot: %v", err)

					break
				}
			}

			newStoredLine = append(newStoredLine, line)
		}
	}

	storedLine = newStoredLine
}

func countLine(line *pb.Line) error {
	var err error

	var peer *pb.Peer
	peer, err = utils.GetPeerAddrWithRole(peerList, pb.CountingRole_SAVER)
	if err != nil {
		return err
	}

	// Define a regular expression pattern to match non-letter and non-number characters (excluding spaces)
	reg := regexp.MustCompile("[^a-zA-Z0-9\\s]+")
	cleanedStr := reg.ReplaceAllString(line.Line, "")

	// Split
	words := strings.Fields(cleanedStr)

	// Theoretically, this scenario should never occur
	wordCountMutex.Lock()
	if globalWordCount == nil {
		globalWordCount = make(map[string]map[string]int32)
	}
	wordCountMutex.Unlock()

	wordCountMutex.Lock()
	if globalWordCount[line.FileName] == nil {
		globalWordCount[line.FileName] = make(map[string]int32)
	}
	wordCountMutex.Unlock()

	// Count words in line sent
	for len(words) > 0 {
		wordCountMutex.Lock()
		globalWordCount[line.FileName][words[0]]++
		wordCountMutex.Unlock()

		words = words[1:]
		line.Line = strings.Join(words, " ")
	}

	wordCountMutex.Lock()
	if err = sendWordCountToSaver(peer.Addr, line.FileName, globalWordCount[line.FileName]); err != nil {
		wordCountMutex.Unlock()
		return fmt.Errorf("[countLine] error sendWordCountToSaver: %v", err)
	}
	wordCountMutex.Unlock()

	wordCountMutex.Lock()
	globalWordCount[line.FileName] = make(map[string]int32)
	wordCountMutex.Unlock()

	globalLineListMutex.Lock()
	index := 0
	for i, item := range globalLineList {
		if item == line {
			index = i
		}
	}
	globalLineList = append(globalLineList[:index], globalLineList[index+1:]...)
	globalLineListMutex.Unlock()

	return nil
}

func sendWordCountToSaver(addr string, fileName string, lineCount map[string]int32) error {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return errors.New(fmt.Sprintf("did not connect: %v", err))
	}
	defer func(conn *grpc.ClientConn) {
		_ = conn.Close()
	}(conn)

	saverClient := pb.NewCounterSaverClient(conn)

	message := &pb.LineCount{
		// TODO capire quale file name lasciare
		//FileName:  fmt.Sprintf("%s:%s", peerServiceAddr, fileName),
		FileName:  fileName,
		LineCount: lineCount,
	}

	// Register process to Service Registry
	ctx := chLam.SetContextChLam(context.Background(), peerServiceAddr)
	_, err = saverClient.SaveCount(ctx, message)
	if err != nil {
		return errors.New(fmt.Sprintf("error when calling SaveCount: %s", err))
	}

	return nil
}
