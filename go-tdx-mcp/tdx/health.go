package tdx

import (
	"math"
	"sort"
	"sync"
	"time"
)

// ---------------------------------------------------------------------------
// 健康分引擎（移植自 easy_tdx _health.py）
// ---------------------------------------------------------------------------
// 设计要点：
//   1. 进程级单例，多 client 共享同一份健康记录——一台服务器的好坏不分协议。
//   2. score ∈ (0, 1.0]，初始 1.0。recordFailure 乘性衰减（×0.5），
//      连续失败 ≥ cooldownFailThreshold 进入 cooldownSec 秒冷却期；
//      recordSuccess 加性恢复（+0.2，上限 1.0）并重置计数。
//   3. rankByHealth 把 ping 结果按 latency / score 重排（score 越低惩罚越大），
//      冷却中的主机直接剔除。
//   4. 全健康时 rankByHealth 近似恒等映射，对既有行为零影响。

const (
	healthFailureDecay         = 0.5
	healthSuccessRecover       = 0.2
	healthCooldownFailThreshold = 3
	healthCooldownSec          = 120.0
	healthScoreFloor           = 1e-3
)

type hostHealth struct {
	score              float64
	consecutiveFailures int
	cooldownUntil      time.Time
}

type healthBook struct {
	mu    sync.Mutex
	hosts map[string]*hostHealth
}

var globalHealthBook = &healthBook{hosts: make(map[string]*hostHealth)}

func (b *healthBook) get(host string) *hostHealth {
	hh := b.hosts[host]
	if hh == nil {
		hh = &hostHealth{score: 1.0}
		b.hosts[host] = hh
	}
	return hh
}

// RecordFailure 记录一次失败：score 乘性衰减，连续失败达阈值则进入冷却。
func RecordFailure(host string) float64 {
	now := time.Now()
	b := globalHealthBook
	b.mu.Lock()
	hh := b.get(host)
	hh.score = math.Max(healthScoreFloor, hh.score*healthFailureDecay)
	hh.consecutiveFailures++
	if hh.consecutiveFailures >= healthCooldownFailThreshold {
		hh.cooldownUntil = now.Add(time.Duration(healthCooldownSec) * time.Second)
	}
	score := hh.score
	b.mu.Unlock()
	return score
}

// RecordSuccess 记录一次成功：score 加性恢复（上限 1.0），重置连续失败与冷却。
func RecordSuccess(host string) {
	b := globalHealthBook
	b.mu.Lock()
	hh := b.get(host)
	hh.score = math.Min(1.0, hh.score+healthSuccessRecover)
	hh.consecutiveFailures = 0
	hh.cooldownUntil = time.Time{}
	b.mu.Unlock()
}

// IsInCooldown 该主机是否处于冷却期。
func IsInCooldown(host string) bool {
	b := globalHealthBook
	b.mu.Lock()
	hh := b.hosts[host]
	if hh == nil {
		b.mu.Unlock()
		return false
	}
	cd := hh.cooldownUntil.After(time.Now())
	b.mu.Unlock()
	return cd
}

// GetScore 返回该主机当前 score（无记录则 1.0）。
func GetScore(host string) float64 {
	b := globalHealthBook
	b.mu.Lock()
	hh := b.hosts[host]
	if hh == nil {
		b.mu.Unlock()
		return 1.0
	}
	s := hh.score
	b.mu.Unlock()
	return s
}

// ResetHealth 清空全部健康记录。主要供测试使用。
func ResetHealth() {
	b := globalHealthBook
	b.mu.Lock()
	b.hosts = make(map[string]*hostHealth)
	b.mu.Unlock()
}

// RankEntry 表示一个主机及其延迟。
type RankEntry struct {
	Host    string
	Latency time.Duration
}

// RankByHealth 按健康分重排主机列表。
// 输入为已按延迟升序的列表，输出：
//  1. 剔除处于冷却期的主机；
//  2. 按 latency / score 升序排序——score 越低惩罚越大。
func RankByHealth(entries []RankEntry) []RankEntry {
	now := time.Now()
	b := globalHealthBook
	b.mu.Lock()
	kept := make([]struct {
		host    string
		latency time.Duration
		score   float64
	}, 0, len(entries))
	for _, e := range entries {
		hh := b.hosts[e.Host]
		if hh != nil && hh.cooldownUntil.After(now) {
			continue
		}
		s := 1.0
		if hh != nil {
			s = hh.score
		}
		kept = append(kept, struct {
			host    string
			latency time.Duration
			score   float64
		}{e.Host, e.Latency, s})
	}
	b.mu.Unlock()

	sort.Slice(kept, func(i, j int) bool {
		return float64(kept[i].latency)/kept[i].score < float64(kept[j].latency)/kept[j].score
	})

	result := make([]RankEntry, len(kept))
	for i, k := range kept {
		result[i] = RankEntry{Host: k.host, Latency: k.latency}
	}
	return result
}

// SelectBestHost 重新测速并选出优于当前主机的最佳主机。
// 返回选中的新主机；若无更优选择则返回空字符串。
func SelectBestHost(
	hosts []string,
	pingFn func(host string) (time.Duration, error),
	currentHost string,
) string {
	type pingResult struct {
		host    string
		latency time.Duration
		ok      bool
	}
	results := make([]pingResult, len(hosts))
	var wg sync.WaitGroup
	for i, h := range hosts {
		wg.Add(1)
		go func(idx int, host string) {
			defer wg.Done()
			lat, err := pingFn(host)
			if err == nil {
				results[idx] = pingResult{host: host, latency: lat, ok: true}
			}
		}(i, h)
	}
	wg.Wait()

	var entries []RankEntry
	for _, r := range results {
		if r.ok {
			entries = append(entries, RankEntry{Host: r.host, Latency: r.latency})
		}
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Latency < entries[j].Latency
	})

	ranked := RankByHealth(entries)
	for _, e := range ranked {
		if e.Host != currentHost {
			return e.Host
		}
	}
	return ""
}

// FindWorkingHost 按延迟顺序逐台测试候选主机，返回第一台可用的。
// tryFn 返回 true 表示该主机可用（如返回非空数据）。
// 最多尝试 maxAttempts 台。
func FindWorkingHost(
	entries []RankEntry,
	tryFn func(host string) bool,
	currentHost string,
	maxAttempts int,
) string {
	ranked := RankByHealth(entries)
	tried := 0
	for _, e := range ranked {
		if e.Host == currentHost {
			continue
		}
		if tried >= maxAttempts {
			break
		}
		tried++
		if tryFn(e.Host) {
			RecordSuccess(e.Host)
			return e.Host
		}
		RecordFailure(e.Host)
	}
	return ""
}

// DefaultMaxAttempts 是空数据故障转移时最多尝试的候选主机数。
const DefaultMaxAttempts = 5