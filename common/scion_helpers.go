package common

import (
	"log"

	"github.com/scionproto/scion/go/lib/snet"
	"os"

	"fmt"
	"github.com/netsec-ethz/2SMS/common/types"
	"github.com/scionproto/scion/go/lib/addr"
)

func LoadConfig(ia addr.IA) (*types.State, error) {
	scionDir := os.Getenv("SC")
	eName := "endhost"
	confDir := scionDir + "/gen/ISD" + fmt.Sprint(ia.I) + "/AS" + ia.A.FileFmt() + "/" + eName
	return types.LoadState(confDir, false)
}

func InitNetwork(local snet.Addr, sciond, dispatcher *string) {
	// Initialize default SCION networking context
	if err := snet.Init(local.IA, *sciond, *dispatcher); err != nil {
		log.Fatal("Unable to initialize SCION network", "err", err)
	}
	log.Println("SCION network successfully initialized")
}

//func GetNeighboringASes(localIA addr.IA) []*addr.IA {
//	c, err := LoadConfig(localIA)
//	if err != nil {
//		return []*addr.IA{}
//	}
//	neighMap := make(map[addr.IA]bool)
//	for _, info := range c.Topo.BR {
//		for _, id := range info.IFIDs {
//			neighMap[c.Topo.IFInfoMap[id].ISD_AS] = true
//		}
//	}
//	neighList := make([]*addr.IA, len(neighMap))
//	i := 0
//	for ia := range neighMap {
//		neighList[i] = &ia
//	}
//	return neighList
//}

//func GetCoreASes(localIA addr.IA) []*addr.IA {
//	c, err := LoadConfig(localIA)
//	if err != nil {
//		return []*addr.IA{}
//	}
//	maxTRC := c.Store.GetNewestTRC(localIA.I)
//	coreIAs := make([]*addr.IA, len(maxTRC.CoreASes))
//	i := 0
//	for coreIA := range maxTRC.CoreASes {
//		coreIAs[i] = &coreIA
//		i++
//	}
//	return coreIAs
//}

func DRKeyAuthenticate(*snet.Addr) error {
	// TODO: Authenticate source with DRKeys
	return nil
}
