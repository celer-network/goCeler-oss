// Copyright 2018-2020 Celer Network

// for now only have clientStreamInterceptor to drop recv/send msg

package cnode

import (
	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc"
)

// DO NOT enable dropRecv/dropSend in production code!
// the values are not protected by concurrent access
type clientStreamMsgDropper struct {
	grpc.ClientStream
	dropRecv bool
	dropSend bool
}

// RecvMsg blocks until receive msg into m. so if dropRecv is true, it doesn't return unless error
func (s *clientStreamMsgDropper) RecvMsg(m interface{}) error {
	var err error
	pb := m.(proto.Message) // m must be a proto.Message

	if !s.dropRecv {
		err = s.ClientStream.RecvMsg(m)
		if err != nil {
			return err
		}
		// now a msg is in m, but we need to check dropRecv again and only return if it's still false
		if !s.dropRecv {
			return nil
		}
	}
	// if we're here, means dropRecv is true. and as long as dropRecv is true, loop recv and not return until dropRecv becomes false
	for s.dropRecv {
		pb.Reset() // clear m data
		err = s.ClientStream.RecvMsg(m)
		if err != nil {
			return err
		}
	}
	// dropRecv becomes false, return nil (last err must be nil otherwise we return early)
	return nil
}

func (s *clientStreamMsgDropper) SendMsg(m interface{}) error {
	if s.dropSend {
		return nil
	}
	return s.ClientStream.SendMsg(m)
}
