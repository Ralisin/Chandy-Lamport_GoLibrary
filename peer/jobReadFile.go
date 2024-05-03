package main

import (
	"bufio"
	chLam "chandyLamportV2/chLamLib"
	"chandyLamportV2/protobuf/pb"
	"chandyLamportV2/utils"
	"context"
	"errors"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
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
	globFileName      = ""
	globFileNameMutex sync.Mutex

	globCurrLine      = 0
	globCurrLineMutex sync.Mutex
)

func jobReadFile(fileName string, currLine int) {
	globFileNameMutex.Lock()
	globFileName = fileName
	globFileNameMutex.Unlock()

	globCurrLineMutex.Lock()
	globCurrLine = currLine
	globCurrLineMutex.Unlock()

	for {
		var err error

		// Get fileName randomly
		globFileNameMutex.Lock()
		if globFileName == "" {
			globFileName, err = utils.GetRandomTxtFilePath(dir)
			if err != nil {
				globFileNameMutex.Unlock()
				log.Printf("error utils.GetRandomTxtFilePath: %v", err)

				continue
			}
		}
		globFileNameMutex.Unlock()

		// Obtain the peer to which to send the lines of the file
		var peer *pb.Peer
		peer, err = utils.GetPeerAddrWithRole(peerList, pb.CountingRole_WORD_COUNTER)
		if err != nil {
			time.Sleep(time.Second)

			continue
		}

		log.Printf("File: %s, peer: %v, role: %v\n", globFileName, peer.Addr, peer.Role)

		err = sendFileLines(peer)
		if err != nil {
			peerList = utils.RemovePeer(peerList, peer)

			continue
		}

		// Reset struct
		globFileNameMutex.Lock()
		globFileName = ""
		globFileNameMutex.Unlock()

		globCurrLineMutex.Lock()
		globCurrLine = 0
		globCurrLineMutex.Unlock()
	}
}

func sendFileLines(peerAddr *pb.Peer) error {
	// Get connection to the chosen peer
	conn, err := grpc.Dial(peerAddr.Addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return errors.New(fmt.Sprintf("did not connect: %v", err))
	}
	defer func(conn *grpc.ClientConn) {
		_ = conn.Close()
	}(conn)

	// Open the file
	var file *os.File
	globFileNameMutex.Lock()
	file, err = os.Open(globFileName)
	globFileNameMutex.Unlock()
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("failed to open file: %v", err))
	}
	defer func(file *os.File) {
		if err := file.Close(); err != nil {
			return
		}
	}(file)

	// Set current file seek position to begin
	if _, err = file.Seek(0, io.SeekStart); err != nil {
		return errors.New(fmt.Sprintf("error seek: %v", err))
	}

	scanner := bufio.NewScanner(file)
	currLine := 0
	for scanner.Scan() {
		// Read file line
		lineText := scanner.Text()

		// Check if curr line index is right
		globCurrLineMutex.Lock()
		if currLine < globCurrLine {
			globCurrLineMutex.Unlock()

			currLine++

			continue
		}
		globCurrLineMutex.Unlock()

		log.Printf("lineText: %v\n", lineText)

		globFileNameMutex.Lock()
		err = sendLineToCounter(conn, globFileName, lineText)
		globFileNameMutex.Unlock()
		if err != nil {
			return errors.New(fmt.Sprintf("error sendLineToCounter: %v", err))
		}

		log.Printf("SENT lineText: %v\n", lineText)

		currLine++

		globCurrLineMutex.Lock()
		globCurrLine = currLine
		globCurrLineMutex.Unlock()

		time.Sleep(time.Second)
	}

	return nil
}

func sendLineToCounter(conn *grpc.ClientConn, fileName string, line string) error {
	wordCounter := pb.NewReaderCounterClient(conn)

	message := &pb.Line{
		FileName: fileName,
		Line:     line,
	}

	ctx := chLam.SetContextChLam(context.Background(), peerServiceAddr)
	ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs(ctxKey, peerServiceAddr))

	// Register process to Service Registry
	_, err := wordCounter.CountWord(ctx, message)
	if err != nil {
		return errors.New(fmt.Sprintf("error when calling CountWord: %s", err))
	}

	return nil
}
