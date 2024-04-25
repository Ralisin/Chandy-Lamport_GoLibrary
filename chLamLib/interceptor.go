package chLam

import (
	"chandyLamportV2/chLamLib/utils"
	"context"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"reflect"
	"strings"
)

func interceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	var resp interface{}
	var err error

	// Invoke gRPC method
	resp, err = handler(ctx, req)
	if err != nil {
		return nil, err
	}

	if info.FullMethod == "/ChandyLamportSnapshot/ChLamSnapshot" {
		// Check if snapshot is ended
		if len(peerAddrList) != 0 && len(peerMap) == len(peerAddrList) {
			// Nel file sto salvando solamente la struct sopra definita
			if err := utils.SaveToFile(snapFileName, snapshotWrap); err != nil {
				log.Printf("[ChLamSnapshot] SaveToFile err: %v", err)
				return nil, err
			}

			log.Printf("[ChLamSnapshot] Snapshot taken")

			// Reset snapshot structs
			isSnapMutex.Lock()
			isSnap = false
			isSnapMutex.Unlock()
			snapshotWrap = &snapWrap{}
			peerMap = make(map[string]bool)
		}
	} else {
		// * Store incoming remote procedures requests only if doing snapshot
		if err := storeRemoteProcedureRequestIfSnapshotting(ctx, req, info); err != nil {
			return nil, err
		}
	}

	return resp, nil
}

// Funzione per estrarre il nome del servizio e del metodo dal nome completo del metodo
func parseMethodName(fullMethodName string) (serviceName string, methodName string) {
	parts := strings.SplitN(fullMethodName, "/", 3)
	return parts[1], parts[2]
}

func storeRemoteProcedureRequestIfSnapshotting(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo) error {
	// Must cast .(string) to get result as string cause return of .Value is "any"
	var peerSrvAddr string
	if ctx.Value(peerSrvAddrKey) != nil {
		peerSrvAddr = ctx.Value(peerSrvAddrKey).(string)
	}

	isSnapMutex.Lock()
	defer isSnapMutex.Unlock()

	// If snapshot in ongoing and peer doesn't still send a marker message
	if isSnap && !peerMap[peerSrvAddr] {
		// Convert Req interface to []byte to save it in methodSnap
		reqBytes, err := utils.ConvertInterfaceToBytes(req)
		if err != nil {
			return err
		}

		// Take Req type to save it in methodSnap
		reqType := reflect.TypeOf(req)
		// Check if reqType is a pointer, to get data type without pointer
		if reqType.Kind() == reflect.Ptr {
			reqType = reqType.Elem()
		}

		// Get Req type string
		reqTypeStr := reqType.String()

		// Get method's full name from UnaryServerInfo struct
		serviceName, methodName := parseMethodName(info.FullMethod)

		// Populate methodSnap struct
		rpcSnap := methodSnap{
			Req: interfaceSnap{
				Bytes: reqBytes,
				Type:  reqTypeStr,
			},
			ServiceName: serviceName,
			MethodName:  methodName,
		}

		// Insert request into snapshotWrap.MethodList
		snapshotWrap.MethodList = append(snapshotWrap.MethodList, rpcSnap)
	}

	return nil
}

/* Funzione che chiama una procedura gRPC dal suo nome */
func callGRPCMethod(ctx context.Context, methodName string, req interface{}, client *interface{}) (interface{}, error) {
	// Trova il metodo gRPC utilizzando la reflection
	svc := reflect.ValueOf(*client)
	log.Printf("[callGRPCMethod] svc: %v", svc)

	method := svc.MethodByName(methodName)
	if !method.IsValid() {
		return nil, fmt.Errorf("metodo gRPC non trovato: %s", methodName)
	}
	log.Printf("[callGRPCMethod] method: %v", method)

	// Prepara gli argomenti per il metodo gRPC
	args := make([]reflect.Value, 2)
	args[0] = reflect.ValueOf(ctx)
	args[1] = reflect.ValueOf(req)

	// Chiama il metodo gRPC
	result := method.Call(args)
	if len(result) != 2 {
		return nil, fmt.Errorf("risposta non valida dal metodo gRPC: %v", result)
	}

	// Estrai il risultato dal metodo chiamato
	resp := result[0].Interface()

	return resp, nil
}
