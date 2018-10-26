package common

import (
	"net/http"
	"bytes"
	"log"

	sd "github.com/scionproto/scion/go/lib/sciond"
	"github.com/scionproto/scion/go/lib/snet"
	"os"

	"fmt"
	"path"
	"github.com/juagargi/temp_squic"
	"io/ioutil"
	"github.com/scionproto/scion/go/lib/addr"
	"github.com/baehless/2SMS/common/types"
	"github.com/scionproto/scion/go/cert_srv/conf"
)

func LoadConfig(ia addr.IA) (*conf.Conf, error) {
	scionDir := os.Getenv("SC")
	eName := "cs" + ia.FileFmt(false) + "-1"
	confDir := scionDir + "/gen/ISD" + fmt.Sprint(ia.I) + "/AS" + ia.A.FileFmt() + "/" + eName
	cacheDir := scionDir + "/gen-cache"
	stateDir := confDir
	return conf.Load(eName, confDir, cacheDir, stateDir)
}

func InitNetwork(local snet.Addr, sciond, dispatcher *string) {
	// Initialize default SCION networking context
	if err := snet.Init(local.IA, *sciond, *dispatcher); err != nil {
		log.Fatal("Unable to initialize SCION network", "err", err)
	}
	log.Println("SCION network successfully initialized")
	SC := os.Getenv("SC")
	keypath := path.Join(SC, "gen-certs", "tls.key")
	pempath := path.Join(SC, "gen-certs", "tls.pem")
	if err := squic.Init(keypath, pempath); err != nil {
		log.Fatal("Unable to initialize QUIC/SCION", "err", err)
	}
	log.Println("QUIC/SCION successfully initialized")
}

func CopyRequestToQUIC(r http.Request) types.Request {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println("Failed reading request body:", err)
	}
	return types.Request{
		Method: r.Method,
		URL: r.URL,
		Proto: r.Proto,
		ProtoMajor: r.ProtoMajor,
		ProtoMinor: r.ProtoMinor,
		Header: r.Header,
		Body: body,
		GetBody: r.GetBody,
		ContentLength: r.ContentLength,
		TransferEncoding: r.TransferEncoding,
		Close: r.Close,
		Host: r.Host,
		Form: r.Form,
		PostForm: r.PostForm,
		MultipartForm: r.MultipartForm,
		Trailer: r.Trailer,
		RemoteAddr: r.RemoteAddr,
		RequestURI: r.RequestURI,
		TLS: r.TLS,
		Cancel: r.Cancel,
		Response: r.Response,
	}
}

func NewRequestFromQUIC(sr types.Request) http.Request {
	body := ioutil.NopCloser(bytes.NewReader(sr.Body))
	return http.Request{
		Method: sr.Method,
		URL: sr.URL,
		Proto: sr.Proto,
		ProtoMajor: sr.ProtoMajor,
		ProtoMinor: sr.ProtoMinor,
		Header: sr.Header,
		Body: body,
		GetBody: sr.GetBody,
		ContentLength: sr.ContentLength,
		TransferEncoding: sr.TransferEncoding,
		Close: sr.Close,
		Host: sr.Host,
		Form: sr.Form,
		PostForm: sr.PostForm,
		MultipartForm: sr.MultipartForm,
		Trailer: sr.Trailer,
		RemoteAddr: sr.RemoteAddr,
		RequestURI: sr.RequestURI,
		TLS: sr.TLS,
		Cancel: sr.Cancel,
		Response: sr.Response,
	}
}

func CopyResponseToQUIC(r http.Response) types.Response {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
	log.Println("Failed reading request body:", err)
	}
	return types.Response {
		Status: r.Status,
		StatusCode: r.StatusCode,
		Proto: r.Proto,
		ProtoMajor: r.ProtoMajor,
		ProtoMinor: r.ProtoMinor,
		Header: r.Header,
		Body: body,
		ContentLength: r.ContentLength,
		TransferEncoding: r.TransferEncoding,
		Close: r.Close,
		Uncompressed: r.Uncompressed,
		Trailer: r.Trailer,
		Request: r.Request,
		TLS: r.TLS,
	}
}

func NewResponseFromQUIC(sr types.Response) *http.Response {
	body := ioutil.NopCloser(bytes.NewReader(sr.Body))
	return &http.Response {
		Status: sr.Status,
		StatusCode: sr.StatusCode,
		Proto: sr.Proto,
		ProtoMajor: sr.ProtoMajor,
		ProtoMinor: sr.ProtoMinor,
		Header: sr.Header,
		Body: body,
		ContentLength: sr.ContentLength,
		TransferEncoding: sr.TransferEncoding,
		Close: sr.Close,
		Uncompressed: sr.Uncompressed,
		Trailer: sr.Trailer,
		Request: sr.Request,
		TLS: sr.TLS,
	}
}

func ChoosePath(local, remote snet.Addr) *sd.PathReplyEntry {
	var paths []*sd.PathReplyEntry
	var pathIndex uint64

	pathMgr := snet.DefNetwork.PathResolver()
	pathSet := pathMgr.Query(local.IA, remote.IA)

	if len(pathSet) == 0 {
		return nil
	}
	for _, p := range pathSet {
		paths = append(paths, p.Entry)
	}
	fmt.Printf("Using path:\n  %s\n", paths[pathIndex].Path.String())
	return paths[pathIndex]
}

func GetNeighboringASes(localIA addr.IA) []*addr.IA {
	c, err := LoadConfig(localIA)
	if err != nil {
		return []*addr.IA{}
	}
	neighMap := make(map[addr.IA]bool)
	for _, info := range c.Topo.BR {
		for _, id := range info.IFIDs {
			neighMap[c.Topo.IFInfoMap[id].ISD_AS] = true
		}
	}
	neighList := make([]*addr.IA, len(neighMap))
	i := 0
	for ia := range neighMap {
		neighList[i] = &ia
	}
	return neighList
}

func GetCoreASes(localIA addr.IA) []*addr.IA {
	c, err := LoadConfig(localIA)
	if err != nil {
		return []*addr.IA{}
	}
	maxTRC := c.Store.GetNewestTRC(localIA.I)
	coreIAs := make([]*addr.IA, len(maxTRC.CoreASes))
	i := 0
	for coreIA := range maxTRC.CoreASes {
		coreIAs[i] = &coreIA
		i++
	}
	return coreIAs
}

func DRKeyAuthenticate(*snet.Addr) error {
	// TODO: Authenticate source with DRKeys
	return nil
}