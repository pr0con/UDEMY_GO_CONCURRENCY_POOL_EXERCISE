package main

import(
	"fmt"
	"time"
	"net/http"
	"math/rand"
	"encoding/json"
	
	"sync/atomic"
	log "github.com/mgutz/logxi/v1"
)


var (
	s = rand.NewSource(time.Now().Unix())
	rn = rand.New(s)
)

const ReqDataSize = 1 * 1024 // 1kb
type ClientReq struct {
	ID      uint
	ReqType int               // one of ReqX defined above
	Data    [ReqDataSize]byte // request specific encoded data
	Size    int               // how many byte in Data
}

type Server interface {
	Start() error
	Stop()
}

type TCPServer struct {
	numReqs uint64
	port    string
	log     log.Logger
	s       *http.Server
}

//POOL ADDITIONS
var pool chan *ClientReq
var total, allocated, reused uint

func init() {
	pool = make(chan *ClientReq, 1000)
}

func Alloc() *ClientReq {
	total++
	select {
		case r := <-pool:
			reused++
			return r
		default:
			allocated++
			r := &ClientReq{}
			return r
	}
}

func Release(r *ClientReq ) {
	select{
		case pool <- r:
		default:
	}
}

func poolStats() {
	fmt.Printf("Total: %v, Allocated: %v, Reused %v \n", total, allocated, reused)
}
//End Pool Additions...

func newTCPServer(port string) Server {
	srv := &TCPServer{port: port}
	srv.log = log.New("server")
	
	//srv.log.SetLevel(log.LevelDebug)
	srv.log.SetLevel(log.LevelInfo) 

	// TODO - create and configure http.Server
	s := &http.Server{}
	// configure http server
	s.WriteTimeout = 500 * time.Millisecond
	s.ReadTimeout = 1000 * time.Millisecond
	s.Addr = fmt.Sprintf(":%v", srv.port) // listen on all interfaces
	s.Handler = srv
	srv.s = s // store a reference to the http.Server
	return srv	
}


func (srv *TCPServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	//Safely increment a number from multiple go routines... better than mutex locks etc...
	atomic.AddUint64(&srv.numReqs, 1)
	
	//srv.log.Debug("TCPServer - message from", r.RemoteAddr)

	go func() {
		dec := json.NewDecoder(r.Body)
		defer r.Body.Close()
	
		//msg := &ClientReq{}
		msg := Alloc()
		
		dec.Decode(msg)
		// INFO - pretent we do some work on with the msg
		time.Sleep(time.Duration(rn.Intn(5)) * time.Millisecond)
		
		Release(msg)
	}()
}


func (srv *TCPServer) Start() error {
	if nil == srv {
		return fmt.Errorf("Start() called on nil TCPServer object")
	}

	srv.log.Info("Starting HTTP server")
	
	// Used with stop server thing 
	var err error
	go func() {
		err = srv.s.ListenAndServe()
	}()
	time.Sleep(200 * time.Millisecond)
	return err
}

// Stop listening and close all client connections
func (srv *TCPServer) Stop() {
	if nil == srv {
		return
	}

	srv.log.Info("Stopping HTTP server")
	srv.s.Close()
	
	//for server close thing
	srv.log.Info("Messages processed:", srv.numReqs)
	poolStats()
}

func main() {
	srv := newTCPServer("7000")
	//srv.Start()	
	// Start server with stop 
	err := srv.Start()
	if err != nil {
		log.Error("Failed to start TCPServer", err)
		return
	}
	d := 20 * time.Second
	fmt.Printf("Sleeping for %v\n", d)
	time.Sleep(d)
	srv.Stop()
}

