package chLam

import (
	"chandyLamportV2/chLamLib/utils"
	"context"
)

func RegisterServerAddr(addr string) {
	peerServerAddr.Addr = addr
}

func RegisterSnapFileName(filename string) {
	snapFileName = filename
}

// RegisterType records the type of the pointer passed to the function in an internal library structure.
//   - usage: chLam.RegisterType((typeCast)(pointer))
//
// example: chLam.RegisterType((typeCast)(nil))
func RegisterType(ptr interface{}) {
	utils.RegisterType(ptr)
}

func RegisterNewPeer(serverAddr string) {
	peerAddrList = append(peerAddrList, serverAddr)
}

func RegisterFunctionForSnapshot(takeSnap func() (interface{}, error)) {
	takePeerSnapshot = takeSnap
}

func RegisterFunctionRetrieveChLamClient(retrieveFunc func(string) (interface{}, error)) {
	retrievePeerClient = retrieveFunc
}

// SetContextChLam is used to set the context with the client's server address so that it can be traced by the interceptor
func SetContextChLam(ctx context.Context, peerAddr string) context.Context {
	return context.WithValue(ctx, peerSrvAddrKey, peerAddr)
}
