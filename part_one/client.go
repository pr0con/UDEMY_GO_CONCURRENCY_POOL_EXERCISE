package main

import(
	"fmt"
	"math/rand"
	"time"
)

const (
	requestsPerClient = 100
	maxBatchSize = (requestsPerClient / 10 ) * 2
	
	ReqAdd        = iota
	ReqAvg        = iota
	ReqRandom     = iota
	ReqSpellCheck = iota
	ReqSearch     = iota		
)

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
)

func main() {
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
			for y := 0; y < req.Size; y++ {
				req.Data[y] = byte(y + 1)
			}
			fmt.Println(req) // send to server			
		}
	}
}

