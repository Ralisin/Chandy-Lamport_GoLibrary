package chLam

import (
	"chandyLamportV2/chLamLib/chLamProto"
	"chandyLamportV2/chLamLib/utils"
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"sync"
)

type ChandyLamportSnapshotServer struct {
	chLamProto.UnimplementedChandyLamportSnapshotServer
}

type interfaceSnap struct {
	Bytes []byte
	Type  string
}

type methodSnap struct {
	Ctx         context.Context
	Req         interfaceSnap
	ServiceName string
	MethodName  string
}

type snapWrap struct {
	// Data store peer snapshot
	Data interfaceSnap

	// MethodList list of all remote procedure requests received during the snapshot
	MethodList []methodSnap
}

// Context key to retrieve peerServerAddress
const peerSrvAddrKey = "chLamPeerSrvAddr"

// Function passed by the programmer via a library method
// Must return struct with the snapshot and nil if its call was successful, error != nil otherwise
var takePeerSnapshot func() (interface{}, error)

var retrievePeerClient func(string) (interface{}, error)

var (
	peerServerAddr = chLamProto.ChLamPeer{Addr: ""}

	peerAddrList []string

	peerMap = make(map[string]bool)

	snapFileName string

	isSnap       = false
	isSnapMutex  sync.Mutex
	snapshotWrap = &snapWrap{}
)

/*
 ! Cose che l'utilizzatore della libreria deve fare:
	* Per impostare il suo server, invece di grpc.NewServer, deve andare a utilizzare la funzione di libreria chLam.NewServer()
		- il funzionamento è lo stesso di grpc.NewServer, soltanto che integra un interceptor interno alla libreria per riuscire e fare lo snapshot

	* Deve registrare tutti i suoi tipi di dato all'interno della libreria tramite chLam.RegisterType
		- la funzione di libreria chLam.RegisterType non fa altro che richiamare una funzione più interna della libreria, ossia utils.RegisterType
		- fondamentale per il corretto funzionamento della libreria: quello che tocca passargli è il puntatore alla struct che si è definita, e non
			un puntatore a puntatore

	* Deve fornire una funzione con la seguente interfaccia: func() (interface{}, error)
		- questa funzione deve fare lo snapshot dello stato corrente del peer e ritornare una interface{}.
			? La interface può essere sia direttamente la struct che anche il suo indirizzo di memoria. La libreria riesce a gestire decentemente
			? entrambi i casi.

	* Ogni volta che viene a conoscenza di un nuovo peer server, deve invocare il metodo RegisterNewPeer e passargli l'indirizzo su cui il servizio è messo a disposizione
		- la libreria permette quindi di effettuare lo snapshot solamente sui peer registrati. In questo modo si riesce a fare anche solo uno snapshot parziale del sistema
		- il peer registrato deve naturalmente implementare questa libreria e deve poter ricevere richieste a procedura remota.

	* Ogni volta che esegue una chiamata a una procedura remota, deve impostare il context con anche l'indirizzo del suo server

	* Deve registrare l'indirizzo del proprio servizio tramite la funzione chLam.RegisterServerAddr(addr string)

	* Deve fornire una funzione con la seguente interfaccia: func(string) (interface{}, error)
		- questa funzione deve restituire il client corretto a seconde della stringa passata. In particolare a questa funzione gli si passerà una stringa con il nome del metodo gRPC completo
			e questa dovrà ritornare il client corretto per eseguire l'invocazione
*/

func (s ChandyLamportSnapshotServer) ChLamSnapshot(_ context.Context, senderPeer *chLamProto.ChLamPeer) (*chLamProto.ChLamPeer, error) {
	if peerServerAddr.Addr == "" {
		return nil, fmt.Errorf("error library initialization: address of the peer server must be registered via chlam.RegisterServerAddr before serving the request")
	}

	if senderPeer.Addr == "" {
		return nil, fmt.Errorf("error parameters: senderPeer.Addr cannot be empty string")
	}

	peerMap[senderPeer.Addr] = true

	// Take a snapshot of a peer if it is the first time it receives the request and store it in snapshotWrap.Data
	isSnapMutex.Lock()
	if !isSnap {
		isSnap = !isSnap
		isSnapMutex.Unlock()

		if err := storePeerSnapshot(); err != nil {
			return nil, err
		}

		go sendChLamRequestToAllPeers()
	} else {
		isSnapMutex.Unlock()
	}

	return &chLamProto.ChLamPeer{}, nil
}

// NewServer abstracts the library function grpc.NewServer via implementation of an internal interceptor
func NewServer(opts ...grpc.ServerOption) *grpc.Server {
	opts = append(opts, grpc.UnaryInterceptor(interceptor))

	grpcServer := grpc.NewServer(opts...)

	chLamService := ChandyLamportSnapshotServer{}
	chLamProto.RegisterChandyLamportSnapshotServer(grpcServer, chLamService)

	reflection.Register(grpcServer)

	return grpcServer
}

func RetrieveDataSnapshot(fileName string) (data interface{}, err error) {
	var retrieveSnapWrap snapWrap
	if err = utils.RetrieveFromFile(fileName, &retrieveSnapWrap); err != nil {
		return nil, err
	}

	log.Printf("[RetrieveSnapshot] retrieveSnapWrap: %v\n", retrieveSnapWrap)

	dataInterface, err := utils.ConvertDataBytesToInterface(retrieveSnapWrap.Data.Bytes, retrieveSnapWrap.Data.Type)
	if err != nil {
		return nil, err
	}

	return dataInterface, nil
}

func RestoreMethodsSnapshot(fileName string) error {
	var retrieveSnapWrap snapWrap
	if err := utils.RetrieveFromFile(fileName, &retrieveSnapWrap); err != nil {
		return err
	}

	var err error = nil
	for _, method := range retrieveSnapWrap.MethodList {
		var client interface{}
		client, err = retrievePeerClient(method.MethodName)

		// I don't care about the answer, they're being only dummy calls
		_, err = callGRPCMethod(method.Ctx, method.MethodName, method.Req, &client)
	}

	return err
}
