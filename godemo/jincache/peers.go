package jincache

import "jincache/pb/pb"

type PeerPicker interface {
	PickPeer(key string) (peer PeerGetter, ok bool)
}

//type PeerGetter interface {
//	Get(group string, key string)([]byte, error)
//}

type PeerGetter interface {
	Get(in *pb.Request, out *pb.Response) error
}

