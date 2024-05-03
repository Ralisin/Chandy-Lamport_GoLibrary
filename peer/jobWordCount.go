package main

import (
	chLam "chandyLamportV2/chLamLib"
	"chandyLamportV2/protobuf/pb"
	"chandyLamportV2/utils"
	"errors"
	"fmt"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"regexp"
	"strings"
	"sync"
	"time"
)

type wordCountData struct {
	Addr     string
	FileName string
	Line     string
}

var (
	lineList      = make([]wordCountData, 0)
	lineListMutex sync.Mutex

	wordCountMap      = make(map[string]int32)
	wordCountMapMutex sync.Mutex
)

func jobWordCount(storedLineList []wordCountData, storedWordCountMap map[string]int32) {
	wordCountMapMutex.Lock()
	lineList = storedLineList
	wordCountMapMutex.Unlock()

	wordCountMapMutex.Lock()
	wordCountMap = storedWordCountMap
	wordCountMapMutex.Unlock()

	for {
		lineListMutex.Lock()
		lineListLength := len(lineList)
		lineListMutex.Unlock()

		if lineListLength == 0 {
			time.Sleep(time.Second)

			continue
		}

		peer, err := utils.GetPeerAddrWithRole(peerList, pb.CountingRole_SAVER)
		if err != nil {
			time.Sleep(time.Second)

			continue
		}

		if len(wordCountMap) != 0 {
			wordCountMapMutex.Lock()
			err = sendWordCountToSaver1(peer.Addr, lineList[0].FileName, wordCountMap, lineList[0].Addr)
			wordCountMapMutex.Unlock()
			if err != nil {
				continue
			}

			wordCountMapMutex.Lock()
			lineList = lineList[1:]
			wordCountMapMutex.Unlock()

			wordCountMapMutex.Lock()
			wordCountMap = make(map[string]int32)
			wordCountMapMutex.Unlock()

			continue
		}

		// Define a regular expression pattern to match non-letter and non-number characters (excluding spaces)
		reg := regexp.MustCompile("[^a-zA-Z0-9\\s]+")
		lineListMutex.Lock()
		cleanStr := reg.ReplaceAllString(lineList[0].Line, "")
		lineListMutex.Unlock()

		// Split
		words := strings.Fields(cleanStr)

		for len(words) > 0 {
			wordCountMapMutex.Lock()
			wordCountMap[words[0]]++
			wordCountMapMutex.Unlock()

			words = words[1:]

			lineListMutex.Lock()
			lineList[0].Line = strings.Join(words, " ")
			lineListMutex.Unlock()
		}

		wordCountMapMutex.Lock()
		err = sendWordCountToSaver1(peer.Addr, lineList[0].FileName, wordCountMap, lineList[0].Addr)
		wordCountMapMutex.Unlock()
		if err != nil {
			continue
		}

		wordCountMapMutex.Lock()
		lineList = lineList[1:]
		wordCountMapMutex.Unlock()

		wordCountMapMutex.Lock()
		wordCountMap = make(map[string]int32)
		wordCountMapMutex.Unlock()
	}
}

func sendWordCountToSaver1(addr string, fileName string, lineCount map[string]int32, originalSenderAddr string) error {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return errors.New(fmt.Sprintf("did not connect: %v", err))
	}
	defer func(conn *grpc.ClientConn) {
		_ = conn.Close()
	}(conn)

	saverClient := pb.NewCounterSaverClient(conn)

	message := &pb.LineCount{
		FileName:  fileName,
		LineCount: lineCount,
	}

	// Register process to Service Registry
	ctx := chLam.SetContextChLam(context.Background(), peerServiceAddr)
	ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs(ctxKey, originalSenderAddr))

	_, err = saverClient.SaveCount(ctx, message)
	if err != nil {
		return errors.New(fmt.Sprintf("error when calling SaveCount: %s", err))
	}

	return nil
}
