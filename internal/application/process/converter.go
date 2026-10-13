package process

import (
	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
	"github.com/eirikur-ari/pidshooter/internal/core/process"
)

func toInfos(infos []outbound.ProcessInfo) []process.Info {
	out := make([]process.Info, len(infos))
	for i, info := range infos {
		out[i] = process.NewInfo(info.PID, info.Name, info.Rss, info.UID)
	}
	return out
}

func toFindResults(infos []process.Info) []FindResult {
	out := make([]FindResult, len(infos))
	for i, info := range infos {
		out[i] = FindResult{PID: info.PID, Name: info.Name, Rss: info.Rss, UID: info.UID}
	}
	return out
}

func toProcessInfos(infos []process.Info) []outbound.ProcessInfo {
	out := make([]outbound.ProcessInfo, len(infos))
	for i, info := range infos {
		out[i] = outbound.ProcessInfo{PID: info.PID, Name: info.Name, Rss: info.Rss, UID: info.UID}
	}
	return out
}
