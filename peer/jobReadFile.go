package main

import (
	"bufio"
	chLam "chandyLamportV2/chLamLib"
	"chandyLamportV2/protobuf/pb"
	"chandyLamportV2/utils"
	"errors"
	"fmt"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"io"
	"log"
	"os"
	"sync"
	"time"
)

// TODO path hardcoded
const dir = "./filesToRead"

// Variabili globali per mantenere lo stato del sistema
var (
	globalFileName      = ""
	globalFileNameMutex sync.Mutex

	globalCurrLine      = 0
	globalCurrLineMutex sync.Mutex
)

func jobReadFile(fileName string, currLine int) {
	globalFileNameMutex.Lock()
	globalFileName = fileName
	globalFileNameMutex.Unlock()

	globalCurrLineMutex.Lock()
	globalCurrLine = currLine
	globalCurrLineMutex.Unlock()

	for {
		var err error
		globalFileNameMutex.Lock()
		if globalFileName == "" {
			globalFileName, err = utils.GetRandomTxtFilePath(dir)
			if err != nil {
				globalFileNameMutex.Unlock()
				log.Printf("error utils.GetRandomTxtFilePath: %v", err)

				continue
			}
		}
		globalFileNameMutex.Unlock()

		var peer *pb.Peer
		peer, err = utils.GetPeerAddrWithRole(peerList, pb.CountingRole_WORD_COUNTER)
		if err != nil {
			// TODO verbose?
			// log.Printf("[utils.GetPeerAddrWithRole] error: %v", err)

			time.Sleep(time.Second)

			continue
		}

		log.Printf("File: %s, peer: %v, role: %v\n", globalFileName, peer.Addr, peer.Role)

		err = sendFileLines(peer)
		if err != nil {
			peerList = utils.RemovePeer(peerList, peer)

			continue
		}

		// Reset struct
		globalFileNameMutex.Lock()
		globalFileName = ""
		globalFileNameMutex.Unlock()

		globalCurrLineMutex.Lock()
		globalCurrLine = 0
		globalCurrLineMutex.Unlock()
	}
}

func sendFileLines(peerAddr *pb.Peer) error {
	var err error

	// Open the file
	var file *os.File
	globalFileNameMutex.Lock()
	if file, err = os.Open(globalFileName); err != nil {
		globalFileNameMutex.Unlock()
		return fmt.Errorf(fmt.Sprintf("failed to open file: %v", err))
	}
	defer func(file *os.File) {
		if err := file.Close(); err != nil {
			return
		}
	}(file)
	globalFileNameMutex.Unlock()

	// Set current file seek position to begin
	if _, err = file.Seek(0, io.SeekStart); err != nil {
		return errors.New(fmt.Sprintf("error seek: %v", err))
	}

	scanner := bufio.NewScanner(file)
	currLine := 0
	for scanner.Scan() {
		lineText := scanner.Text()
		log.Printf("lineText: %v\n", lineText)

		globalCurrLineMutex.Lock()
		if currLine < globalCurrLine {
			globalCurrLineMutex.Unlock()

			currLine++

			continue
		}
		globalCurrLineMutex.Unlock()

		globalFileNameMutex.Lock()
		err = sendLineToCounter(peerAddr.Addr, globalFileName, lineText)
		if err != nil {
			globalFileNameMutex.Unlock()
			return errors.New(fmt.Sprintf("error sendLineToCounter: %v", err))
		}
		globalFileNameMutex.Unlock()

		currLine++

		globalCurrLineMutex.Lock()
		globalCurrLine = currLine
		globalCurrLineMutex.Unlock()

		time.Sleep(time.Second)
	}

	return nil
}

func sendLineToCounter(addr string, fileName string, line string) error {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return errors.New(fmt.Sprintf("did not connect: %v", err))
	}
	defer func(conn *grpc.ClientConn) {
		_ = conn.Close()
	}(conn)

	wordCounter := pb.NewReaderCounterClient(conn)

	message := &pb.Line{
		// TODO capire quale file name lasciare
		//FileName: fmt.Sprintf("%s-%s", peerServiceAddr, fileName),
		FileName: fileName,
		Line:     line,
	}

	ctx := chLam.SetContextChLam(context.Background(), peerServiceAddr)

	// Register process to Service Registry
	_, err = wordCounter.CountWord(ctx, message)
	if err != nil {
		return errors.New(fmt.Sprintf("error when calling CountWord: %s", err))
	}

	return nil
}
