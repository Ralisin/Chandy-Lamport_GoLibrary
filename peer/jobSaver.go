package main

import (
	"chandyLamportV2/protobuf/pb"
	"fmt"
	"log"
	"sync"
	"time"
)

var (
	// Key1: fileName, Key2: word, Value: count
	globalTotalCount      = make(map[string]map[string]int32)
	globalTotalCountMutex sync.Mutex

	globalLineCount      = &pb.LineCount{}
	globalLineCountMutex sync.Mutex
)

func jobSaver(storedTotalCount map[string]map[string]int32, lineCount *pb.LineCount) {
	if storedTotalCount == nil {
		globalTotalCountMutex.Lock()
		globalTotalCount = make(map[string]map[string]int32)
		globalTotalCountMutex.Unlock()
	} else {
		globalTotalCountMutex.Lock()
		globalTotalCount = storedTotalCount
		globalTotalCountMutex.Unlock()
	}

	if lineCount != nil {
		err := storeTotalCount(lineCount)
		if err != nil {
			log.Printf("[jobSaver] error restoring snapshot: %v", err)
		}
	}

	globalLineCountMutex.Lock()
	globalLineCount = &pb.LineCount{}
	globalLineCountMutex.Unlock()
}

func storeTotalCount(lineCount *pb.LineCount) error {
	globalTotalCountMutex.Lock()
	if globalTotalCount == nil {
		globalTotalCount = make(map[string]map[string]int32)
	}
	globalTotalCountMutex.Unlock()

	if lineCount == nil {
		return fmt.Errorf("nil lineCount")
	}

	globalLineCountMutex.Lock()
	globalLineCount = lineCount
	globalLineCountMutex.Unlock()

	globalTotalCountMutex.Lock()
	if globalTotalCount[lineCount.FileName] == nil {
		globalTotalCount[lineCount.FileName] = make(map[string]int32)
	}
	globalTotalCountMutex.Unlock()

	for word, value := range lineCount.LineCount {
		globalTotalCountMutex.Lock()
		globalTotalCount[lineCount.FileName][word] += value
		globalTotalCountMutex.Unlock()

		delete(lineCount.LineCount, word)

		globalLineCountMutex.Lock()
		globalLineCount = lineCount
		globalLineCountMutex.Unlock()

		time.Sleep(250 * time.Millisecond)
	}

	globalTotalCountMutex.Lock()
	for key, subMap := range globalTotalCount {
		fmt.Println(key, ":")
		for subKey, value := range subMap {
			fmt.Printf("\t%s: %d\n", subKey, value)
		}
	}
	globalTotalCountMutex.Unlock()

	return nil
}
