package service

import (
	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
	"github.com/eirikur-ari/pidshooter/internal/core/process"
)

func toProcessInfo(p outbound.ProcessInfo) process.Info {
	return process.Info{
		Pid:  p.Pid,
		Name: p.Name,
		Rss:  p.Rss,
	}
}

func toProcessInfos(ps []outbound.ProcessInfo) []process.Info {
	out := make([]process.Info, len(ps))
	for i, p := range ps {
		out[i] = toProcessInfo(p)
	}
	return out
}
