package loadbalance

import (
	"fmt"
	"strings"
	"sync"
	"sync/atomic"

	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
)

var _ base.PickerBuilder = (*Picker)(nil)
var _ balancer.Picker = (*Picker)(nil)

type Picker struct {
	mu        sync.RWMutex
	leader    balancer.SubConn
	followers []balancer.SubConn
	current   uint64
}

func init() {
	balancer.Register(base.NewBalancerBuilder(ShopResolverName, &Picker{}, base.Config{}))
}

func (p *Picker) Build(buildInfo base.PickerBuildInfo) balancer.Picker {
	p.mu.Lock()
	defer p.mu.Unlock()

	var followers []balancer.SubConn
	for sc, scInfo := range buildInfo.ReadySCs {
		isLeader := scInfo.Address.Attributes.Value("is_leader").(bool)
		if isLeader {
			p.leader = sc
			continue
		}
		followers = append(followers, sc)
	}
	p.followers = followers
	return p
}

func (p *Picker) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	fmt.Printf("picking connection for: %s, curr followers: %d\n", info.FullMethodName, len(p.followers))

	var result balancer.PickResult

	if len(p.followers) == 0 {
		fmt.Println("no followers, picking leader")
		result.SubConn = p.leader
	} else if p.isWriteRequest(info) {
		fmt.Println("is write request, picking leader")
		result.SubConn = p.leader
	} else if p.isReadRequest(info) {
		fmt.Println("is read request, picking next follower")
		result.SubConn = p.nextFollower()
	} else if p.isCommonRequest(info) {
		fmt.Println("is common request, picking current")
		result.SubConn = p.getCurrent()
	}
	if result.SubConn == nil {
		return result, balancer.ErrNoSubConnAvailable
	}
	return result, nil
}

func (p *Picker) nextFollower() balancer.SubConn {
	cur := atomic.AddUint64(&p.current, uint64(1))
	len := uint64(len(p.followers))
	idx := int(cur % len)
	return p.followers[idx]
}

func (p *Picker) getCurrent() balancer.SubConn {
	curr := p.followers[p.current]
	if curr == nil {
		return p.leader
	}
	return curr
}

func (p *Picker) isWriteRequest(info balancer.PickInfo) bool {
	return strings.Contains(info.FullMethodName, "AddShop") ||
		strings.Contains(info.FullMethodName, "UpdateShop") ||
		strings.Contains(info.FullMethodName, "DeleteShop")
}

func (p *Picker) isReadRequest(info balancer.PickInfo) bool {
	return strings.Contains(info.FullMethodName, "GetShop") ||
		strings.Contains(info.FullMethodName, "SearchShops") ||
		strings.Contains(info.FullMethodName, "GetAllShops")
}

func (p *Picker) isCommonRequest(info balancer.PickInfo) bool {
	return strings.Contains(info.FullMethodName, "GetServers")
}
