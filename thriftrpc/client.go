// Copyright 2021 CloudWeGo Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package thriftrpc

import (
	"context"
	"math"
	"math/rand"
	"runtime"
	"time"

	"github.com/apache/thrift/lib/go/thrift"
	"github.com/bytedance/gopkg/cloud/metainfo"
	"github.com/cloudwego/kitex-tests/kitex_gen/thrift/instparam"
	"github.com/cloudwego/kitex-tests/kitex_gen/thrift/stability"
	"github.com/cloudwego/kitex-tests/kitex_gen/thrift/stability/stservice"
	"github.com/cloudwego/kitex-tests/pkg/utils"
	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/pkg/connpool"
	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/cloudwego/kitex/transport"
	"github.com/cloudwego/netpoll"
)

func init() {
	runtime.GOMAXPROCS(4)
	netpoll.SetNumLoops(4 * 2)
	klog.SetLevel(klog.LevelFatal)
}

// ConnectionMode .
type ConnectionMode int

// Modes .
const (
	ShortConnection ConnectionMode = iota
	LongConnection
	ConnectionMultiplexed
)

// ClientInitParam .
type ClientInitParam struct {
	TargetServiceName string
	HostPorts         []string
	Protocol          transport.Protocol
	ConnMode          ConnectionMode
}

// CreateKitexClient .
func CreateKitexClient(param *ClientInitParam, opts ...client.Option) stservice.Client {
	if len(param.HostPorts) > 0 {
		opt := []client.Option{client.WithHostPorts(param.HostPorts...)}
		// the priority of param host port is lower
		opts = append(opt, opts...)
	}

	if param.Protocol != transport.PurePayload {
		opts = append(opts, client.WithTransportProtocol(param.Protocol))
	}

	switch param.ConnMode {
	case LongConnection:
		opts = append(opts, client.WithLongConnection(
			connpool.IdleConfig{
				MaxIdlePerAddress: 1000,
				MaxIdleGlobal:     1000 * 10,
				MaxIdleTimeout:    30 * time.Second,
			}))
	case ConnectionMultiplexed:
		opts = append(opts, client.WithMuxConnection(4))
	default:
	}

	return stservice.MustNewClient(param.TargetServiceName, opts...)
}

// CreateSTRequest .
func CreateSTRequest(ctx context.Context) (context.Context, *stability.STRequest) {
	req := stability.NewSTRequest()
	req.Name = "byted"
	req.On = thrift.BoolPtr(true)
	req.B = 10
	req.Int16 = 10
	req.Int32 = math.MaxInt32
	req.Int64 = math.MaxInt64
	req.D = 0.0
	req.Str = utils.RandomString(100)
	req.Bin = []byte{1, 'a', '*'}
	req.StringMap = map[string]string{
		"key1": utils.RandomString(100),
		"key2": utils.RandomString(10),
	}
	req.StringList = []string{
		utils.RandomString(10),
		utils.RandomString(20),
		utils.RandomString(30),
	}
	req.StringSet = []string{
		utils.RandomString(10),
		utils.RandomString(100),
	}
	req.E = stability.TestEnum_FIRST

	ctx = metainfo.WithValue(ctx, "TK", "TV")
	ctx = metainfo.WithPersistentValue(ctx, "PK", "PV")
	return ctx, req
}

// CreateObjReq .
func CreateObjReq(ctx context.Context) (context.Context, *instparam.ObjReq) {
	id := thrift.Int64Ptr(int64(rand.Intn(100)))
	subMsg1 := &instparam.SubMessage{
		Id:    id,
		Value: thrift.StringPtr(utils.RandomString(100)),
	}
	subMsg2 := &instparam.SubMessage{
		Id:    thrift.Int64Ptr(math.MaxInt64),
		Value: thrift.StringPtr(utils.RandomString(10)),
	}
	subMsgList := []*instparam.SubMessage{subMsg1, subMsg2}

	msg := instparam.NewMessage()
	msg.Id = id
	msg.Value = thrift.StringPtr(utils.RandomString(100))
	msg.SubMessages = subMsgList

	req := instparam.NewObjReq()
	req.Msg = msg
	req.MsgMap = map[*instparam.Message]*instparam.SubMessage{
		msg: subMsg1,
	}

	req.MsgSet = []*instparam.Message{msg}
	req.SubMsgs = subMsgList

	ctx = metainfo.WithValue(ctx, "TK", "TV")
	ctx = metainfo.WithPersistentValue(ctx, "PK", "PV")
	return ctx, req
}
