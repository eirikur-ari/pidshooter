package process

import (
	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
	"github.com/eirikur-ari/pidshooter/internal/core/process"
)

func toInfo(info outbound.ProcessInfo) process.Info {
	return process.NewInfo(info.PID, info.Name, info.Rss, info.UID)
}

func toInfos(infos []outbound.ProcessInfo) []process.Info {
	out := make([]process.Info, len(infos))
	for i, info := range infos {
		out[i] = toInfo(info)
	}
	return out
}

func toProcessInfo(info process.Info) outbound.ProcessInfo {
	return outbound.ProcessInfo{PID: info.PID, Name: info.Name, Rss: info.Rss, UID: info.UID}
}

func toProcessInfos(infos []process.Info) []outbound.ProcessInfo {
	out := make([]outbound.ProcessInfo, len(infos))
	for i, info := range infos {
		out[i] = toProcessInfo(info)
	}
	return out
}
