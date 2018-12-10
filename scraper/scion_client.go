package main

import (
	"encoding/gob"
	"errors"
	"github.com/juagargi/temp_squic" // TODO: remove and import from scionproto
	"github.com/lucas-clemente/quic-go"
	"github.com/lucas-clemente/quic-go/qerr"
	"github.com/netsec-ethz/2SMS/common"
	"github.com/netsec-ethz/2SMS/common/types"
	"github.com/scionproto/scion/go/lib/snet"
	"github.com/scionproto/scion/go/lib/spath"
	"log"
	"net"
	"net/http"
	"strings"
	"time"
)

type SCIONClient struct {
	localAddr snet.Addr
}

func (sc *SCIONClient) TunnelRequest(req *http.Request) (*http.Response, error) {
	// Get target IA from first element of target path
	ia := strings.Split(req.URL.Path, "/")[1]
	// Remove it from the target path
	req.URL.Path = "/" + strings.SplitN(req.URL.Path, "/", 3)[2]
	ip := strings.Split(req.URL.Host, ":")[0]
	port := strings.Split(req.URL.Host, ":")[1]
	remoteAddr, err := snet.AddrFromString(ia + ",[" + ip + "]:" + port)
	if err != nil {
		log.Println("Failed parsing snet address from string:", err)
	}
	if remoteAddr.IA.Eq(sc.localAddr.IA) {
		return &http.Response{}, errors.New("Target in same AS as scraper")
	} else {
		pathEntry := common.ChoosePath(local, *remoteAddr)
		if pathEntry == nil {
			return &http.Response{}, errors.New("No paths available to remote destination")
		} else {
			remoteAddr.Path = spath.New(pathEntry.Path.FwdPath)
			remoteAddr.Path.InitOffsets()
			remoteAddr.NextHopHost = pathEntry.HostInfo.Host()
			remoteAddr.NextHopPort = pathEntry.HostInfo.Port
			// Attempt SQUIC connection
			qsess, err := squic.DialSCION(nil, &local, remoteAddr)
			if err == nil {
				defer qsess.Close()
				qstream, err := qsess.OpenStreamSync()
				if err == nil {
					defer qstream.Close()
					log.Println("Quic stream opened", "local", &local, "remote", &remoteAddr)
					before := time.Now()
					log.Println("Sending request at:", before)
					Send(qstream, req)
					resp := Read(qstream)
					after := time.Now()
					log.Println("Reading response", resp, "at:", after, ". Took:", after.Sub(before))
					return resp, nil
				}
				return &http.Response{}, err
			}
			return &http.Response{}, err
		}
	}
}

func Send(qstream quic.Stream, req *http.Request) {
	encoder := gob.NewEncoder(qstream)
	err := encoder.Encode(common.CopyRequestToQUIC(*req))
	if err != nil {
		log.Fatal("Gob encode error:", err)
	}
}

func Read(qstream quic.Stream) *http.Response {
	decoder := gob.NewDecoder(qstream)
	var resp types.Response
	err := decoder.Decode(&resp)
	if err != nil {
		qer := qerr.ToQuicError(err)
		if qer.ErrorCode == qerr.PeerGoingAway {
			log.Println("Quic peer disconnected")
		}
		if nerr, ok := err.(net.Error); ok && nerr.Timeout() {
			log.Println("ReadDeadline missed:", err)
		} else {
			log.Println("Gob decode error:", err)
		}
	}
	return common.NewResponseFromQUIC(resp)
}
