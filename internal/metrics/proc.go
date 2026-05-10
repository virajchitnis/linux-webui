package metrics

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Snapshot struct {
	Timestamp  time.Time `json:"ts"`
	CPUPercent float64   `json:"cpu_pct"`
	MemTotal   uint64    `json:"mem_total_kb"`
	MemFree    uint64    `json:"mem_free_kb"`
	MemUsed    uint64    `json:"mem_used_kb"`
	SwapTotal  uint64    `json:"swap_total_kb"`
	SwapUsed   uint64    `json:"swap_used_kb"`
	LoadAvg1   float64   `json:"load1"`
	LoadAvg5   float64   `json:"load5"`
	LoadAvg15  float64   `json:"load15"`
	Uptime     float64   `json:"uptime_seconds"`
	NetRxBytes uint64    `json:"net_rx_bytes"`
	NetTxBytes uint64    `json:"net_tx_bytes"`
}

type Collector struct {
	mu        sync.RWMutex
	latest    *Snapshot
	ring      [60]*Snapshot
	ringIdx   int
	prevCPU   cpuStat
	prevNet   netStat
	listeners []chan *Snapshot
	listMu    sync.Mutex
}

type cpuStat struct {
	user, nice, system, idle, iowait, irq, softirq uint64
}

type netStat struct {
	rx, tx uint64
}

func NewCollector() *Collector {
	c := &Collector{}
	c.prevCPU = readCPUStat()
	c.prevNet = readNetStat()
	return c
}

func (c *Collector) Start() {
	go func() {
		for range time.Tick(time.Second) {
			snap := c.collect()
			c.mu.Lock()
			c.latest = snap
			c.ring[c.ringIdx%60] = snap
			c.ringIdx++
			c.mu.Unlock()
			c.broadcast(snap)
		}
	}()
}

func (c *Collector) Latest() *Snapshot {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.latest
}

func (c *Collector) Subscribe() chan *Snapshot {
	ch := make(chan *Snapshot, 2)
	c.listMu.Lock()
	c.listeners = append(c.listeners, ch)
	c.listMu.Unlock()
	return ch
}

func (c *Collector) Unsubscribe(ch chan *Snapshot) {
	c.listMu.Lock()
	defer c.listMu.Unlock()
	for i, l := range c.listeners {
		if l == ch {
			c.listeners = append(c.listeners[:i], c.listeners[i+1:]...)
			close(ch)
			return
		}
	}
}

func (c *Collector) broadcast(snap *Snapshot) {
	c.listMu.Lock()
	defer c.listMu.Unlock()
	for _, ch := range c.listeners {
		select {
		case ch <- snap:
		default:
		}
	}
}

func (c *Collector) collect() *Snapshot {
	cur := readCPUStat()
	prev := c.prevCPU
	c.prevCPU = cur

	totalDiff := (cur.user + cur.nice + cur.system + cur.idle + cur.iowait + cur.irq + cur.softirq) -
		(prev.user + prev.nice + prev.system + prev.idle + prev.iowait + prev.irq + prev.softirq)
	idleDiff := cur.idle - prev.idle
	cpuPct := 0.0
	if totalDiff > 0 {
		cpuPct = float64(totalDiff-idleDiff) / float64(totalDiff) * 100
	}

	curNet := readNetStat()
	prevNet := c.prevNet
	c.prevNet = curNet

	mem := readMemInfo()
	load := readLoadAvg()
	uptime := readUptime()

	return &Snapshot{
		Timestamp:  time.Now(),
		CPUPercent: cpuPct,
		MemTotal:   mem["MemTotal"],
		MemFree:    mem["MemAvailable"],
		MemUsed:    mem["MemTotal"] - mem["MemAvailable"],
		SwapTotal:  mem["SwapTotal"],
		SwapUsed:   mem["SwapTotal"] - mem["SwapFree"],
		LoadAvg1:   load[0],
		LoadAvg5:   load[1],
		LoadAvg15:  load[2],
		Uptime:     uptime,
		NetRxBytes: curNet.rx - prevNet.rx,
		NetTxBytes: curNet.tx - prevNet.tx,
	}
}

func readCPUStat() cpuStat {
	f, err := os.Open("/proc/stat")
	if err != nil {
		return cpuStat{}
	}
	defer f.Close()
	s := bufio.NewScanner(f)
	for s.Scan() {
		line := s.Text()
		if !strings.HasPrefix(line, "cpu ") {
			continue
		}
		var cs cpuStat
		fmt.Sscanf(line, "cpu %d %d %d %d %d %d %d",
			&cs.user, &cs.nice, &cs.system, &cs.idle, &cs.iowait, &cs.irq, &cs.softirq)
		return cs
	}
	return cpuStat{}
}

func readMemInfo() map[string]uint64 {
	f, err := os.Open("/proc/meminfo")
	if err != nil {
		return nil
	}
	defer f.Close()
	m := make(map[string]uint64)
	s := bufio.NewScanner(f)
	for s.Scan() {
		line := s.Text()
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		key := strings.TrimSuffix(parts[0], ":")
		v, _ := strconv.ParseUint(parts[1], 10, 64)
		m[key] = v
	}
	return m
}

func readLoadAvg() [3]float64 {
	data, err := os.ReadFile("/proc/loadavg")
	if err != nil {
		return [3]float64{}
	}
	var la [3]float64
	fmt.Sscanf(string(data), "%f %f %f", &la[0], &la[1], &la[2])
	return la
}

func readUptime() float64 {
	data, err := os.ReadFile("/proc/uptime")
	if err != nil {
		return 0
	}
	var up float64
	fmt.Sscanf(string(data), "%f", &up)
	return up
}

func readNetStat() netStat {
	f, err := os.Open("/proc/net/dev")
	if err != nil {
		return netStat{}
	}
	defer f.Close()
	var rx, tx uint64
	s := bufio.NewScanner(f)
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if strings.HasPrefix(line, "lo:") || strings.HasPrefix(line, "Inter") || strings.HasPrefix(line, "face") {
			continue
		}
		colonIdx := strings.Index(line, ":")
		if colonIdx < 0 {
			continue
		}
		fields := strings.Fields(line[colonIdx+1:])
		if len(fields) < 9 {
			continue
		}
		r, _ := strconv.ParseUint(fields[0], 10, 64)
		t, _ := strconv.ParseUint(fields[8], 10, 64)
		rx += r
		tx += t
	}
	return netStat{rx: rx, tx: tx}
}
