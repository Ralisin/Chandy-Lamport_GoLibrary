package main

import (
	"log"
	"sync"
	"time"
)

type saverData struct {
	Addr      string
	FileName  string
	LineCount map[string]int32
}

var (
	lineCountList      = make([]saverData, 0)
	lineCountListMutex sync.Mutex

	// peerAddrSender, fileName, word, count
	globTotalCount      = make(map[string]map[string]map[string]int32)
	globTotalCountMutex sync.Mutex
)

func jobSaver(storedLineCountList []saverData, storedGlobTotalCount map[string]map[string]map[string]int32) {
	globTotalCountMutex.Lock()
	globTotalCount = storedGlobTotalCount
	globTotalCountMutex.Unlock()

	lineCountListMutex.Lock()
	lineCountList = storedLineCountList
	lineCountListMutex.Unlock()

	for {
		lineCountListMutex.Lock()
		lineCountListLength := len(lineCountList)
		lineCountListMutex.Unlock()

		if lineCountListLength == 0 {
			time.Sleep(time.Second)

			continue
		}

		copiedMap := make(map[string]int32)
		lineCountListMutex.Lock()
		for key, value := range lineCountList[0].LineCount {
			copiedMap[key] = value
		}
		lineCountListMutex.Unlock()

		lineCountListMutex.Lock()
		addr, fileName := lineCountList[0].Addr, lineCountList[0].FileName
		lineCountListMutex.Unlock()

		for word, count := range copiedMap {
			globTotalCountMutex.Lock()
			if globTotalCount[addr] == nil {
				globTotalCount[addr] = make(map[string]map[string]int32)
			}
			if globTotalCount[addr][fileName] == nil {
				globTotalCount[addr][fileName] = make(map[string]int32)
			}
			globTotalCount[addr][fileName][word] += count
			globTotalCountMutex.Unlock()

			lineCountListMutex.Lock()
			delete(lineCountList[0].LineCount, word)
			lineCountListMutex.Unlock()
		}

		lineCountListMutex.Lock()
		lineCountList = lineCountList[1:]
		lineCountListMutex.Unlock()

		globTotalCountMutex.Lock()
		log.Printf("globTotalCount:\n")
		for peerAddrSender, fileMap := range globTotalCount {
			log.Printf("\tpeerAddrSender: %v", peerAddrSender)
			for fName, wordMap := range fileMap {
				log.Printf("\t\tfileMap: %v", fName)
				for word, count := range wordMap {
					log.Printf("\t\t\tword: %v - %v", word, count)
				}
			}
		}
		globTotalCountMutex.Unlock()
	}
}
