package main

import (
	//Native Packages
	"io"
	//"fmt"
	"time"
	"bytes"
	"net/http"
	"math/rand"
	"encoding/json"
	
	//3rd party 
	log "github.com/mgutz/logxi/v1"
)

const (
	requestsPerClient = 1000
	maxBatchSize      = (requestsPerClient / 10) * 2 // 20% of total request
	
	ReqAdd        = iota
	ReqAvg        = iota
	ReqRandom     = iota
	ReqSpellCheck = iota
	ReqSearch     = iota	
)

// ReqDataSize is the max bytes per ClentReq.Data byte array
const ReqDataSize = 1 * 1024 // 1k=b

type ClientReq struct {
		ID      uint
		ReqType int               // one of ReqX defined above
		Data    [ReqDataSize]byte // request specific encoded data
		Size    int               // how many byte in Data
}


var (
	s = rand.NewSource(time.Now().Unix())
	r = rand.New(s)
	logger = log.New("client")	
)


func encodeReq(req *ClientReq) io.Reader {
	var buf = &bytes.Buffer{}
	jsonEnc := json.NewEncoder(buf)
	jsonEnc.Encode(req)
	return buf
}

func submitRequests(url string) {
	var req *ClientReq
	msgLeft := requestsPerClient
	var reqID uint

	for 0 < msgLeft {
		batch := r.Intn(maxBatchSize)
		if batch > msgLeft {
			batch = msgLeft
		}
		msgLeft -= batch

		for i := 0; i < batch; i++ {
			req = &ClientReq{}
			reqID++
			req.ID = reqID
			req.Size = r.Intn(ReqDataSize)
			
			buf := encodeReq(req)
			resp, err := http.Post(url, "text/json", buf)
			if nil != err {
				logger.Error("Post error:", err)
				break // try again later
			}			
			defer resp.Body.Close()
			//fmt.Println(req) // send to server
		}
		// pause a bit between batches
		//time.Sleep(time.Duration(r.Intn(200)) * time.Millisecond)
	}	
}

func main() {
	submitRequests("http://localhost:7000")
}