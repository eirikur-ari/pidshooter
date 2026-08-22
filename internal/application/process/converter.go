package process

import (
	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
	"github.com/eirikur-ari/pidshooter/internal/core/process"
)

func toProcessInfo(info outbound.ProcessInfo) process.Info {
	return process.NewInfo(info.Pid, info.Name, info.Rss)
}

func toProcessInfos(infos []outbound.ProcessInfo) []process.Info {
	out := make([]process.Info, len(infos))
	for i, info := range infos {
		out[i] = toProcessInfo(info)
	}
	return out
}
