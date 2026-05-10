package process

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"syscall"
)

// Info holds the details of a single process.
type Info struct {
	PID     int    `json:"pid"`
	PPID    int    `json:"ppid"`
	Name    string `json:"name"`
	State   string `json:"state"`
	User    string `json:"user"`
	MemRSS  int64  `json:"mem_rss_kb"`
	CPU     float64 `json:"cpu_pct"` // average since process start (not instantaneous)
	Command string `json:"command"`
	Nice    int    `json:"nice"`
}

// List returns all processes visible in /proc.
func List() ([]Info, error) {
	uptime, hertz, numCPU := systemParams()

	entries, err := os.ReadDir("/proc")
	if err != nil {
		return nil, err
	}

	uidMap := buildUIDMap()
	var procs []Info
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		pid, err := strconv.Atoi(e.Name())
		if err != nil {
			continue // not a numeric dir
		}
		p, err := readProc(pid, uptime, hertz, numCPU, uidMap)
		if err != nil {
			continue // process may have exited
		}
		procs = append(procs, p)
	}
	sort.Slice(procs, func(i, j int) bool { return procs[i].PID < procs[j].PID })
	return procs, nil
}

// Kill sends signal sig to process pid.
func Kill(pid int, sig syscall.Signal) error {
	if pid <= 1 {
		return fmt.Errorf("refusing to signal PID %d", pid)
	}
	proc, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	return proc.Signal(sig)
}

// Renice sets the scheduling priority of pid to priority (−20..19).
func Renice(pid, priority int) error {
	if priority < -20 || priority > 19 {
		return fmt.Errorf("priority must be in range -20..19")
	}
	return syscall.Setpriority(syscall.PRIO_PROCESS, pid, priority)
}

// ─── internal helpers ────────────────────────────────────────────────────────

func systemParams() (uptimeSec float64, hertz float64, numCPU int) {
	data, err := os.ReadFile("/proc/uptime")
	if err == nil {
		fmt.Sscanf(strings.Fields(string(data))[0], "%f", &uptimeSec)
	}
	hertz = 100 // CLK_TCK; getconf CLK_TCK on Linux is almost always 100
	numCPU = 1
	data, err = os.ReadFile("/proc/cpuinfo")
	if err == nil {
		for _, line := range strings.Split(string(data), "\n") {
			if strings.HasPrefix(line, "processor") {
				numCPU++
			}
		}
	}
	return
}

func readProc(pid int, uptimeSec, hertz float64, numCPU int, uidMap map[string]string) (Info, error) {
	base := fmt.Sprintf("/proc/%d", pid)

	// --- /proc/{pid}/stat --------------------------------------------------
	statData, err := os.ReadFile(filepath.Join(base, "stat"))
	if err != nil {
		return Info{}, err
	}
	statStr := string(statData)
	// comm is wrapped in parens and may contain spaces; find the last ')'
	rp := strings.LastIndex(statStr, ")")
	if rp < 0 {
		return Info{}, fmt.Errorf("bad stat")
	}
	comm := statStr[strings.Index(statStr, "(")+1 : rp]
	rest := strings.Fields(statStr[rp+2:])
	if len(rest) < 20 {
		return Info{}, fmt.Errorf("short stat")
	}
	state := rest[0]
	ppid, _ := strconv.Atoi(rest[1])
	nice, _ := strconv.Atoi(rest[16])
	utime, _ := strconv.ParseFloat(rest[11], 64)
	stime, _ := strconv.ParseFloat(rest[12], 64)
	starttime, _ := strconv.ParseFloat(rest[19], 64)

	// CPU% averaged since start
	totalTime := (utime + stime) / hertz
	elapsed := uptimeSec - (starttime / hertz)
	var cpuPct float64
	if elapsed > 0 {
		cpuPct = (totalTime / elapsed) * 100 / float64(numCPU)
	}

	// --- /proc/{pid}/status ------------------------------------------------
	var memRSS int64
	var uid string
	if statusData, err := os.ReadFile(filepath.Join(base, "status")); err == nil {
		for _, line := range strings.Split(string(statusData), "\n") {
			if strings.HasPrefix(line, "VmRSS:") {
				fields := strings.Fields(line)
				if len(fields) >= 2 {
					memRSS, _ = strconv.ParseInt(fields[1], 10, 64)
				}
			}
			if strings.HasPrefix(line, "Uid:") {
				fields := strings.Fields(line)
				if len(fields) >= 2 {
					uid = fields[1]
				}
			}
		}
	}
	username := uidMap[uid]
	if username == "" {
		username = uid
	}

	// --- /proc/{pid}/cmdline -----------------------------------------------
	var cmd string
	if cmdData, err := os.ReadFile(filepath.Join(base, "cmdline")); err == nil {
		cmd = strings.ReplaceAll(string(cmdData), "\x00", " ")
		cmd = strings.TrimSpace(cmd)
		if len(cmd) > 256 {
			cmd = cmd[:256] + "…"
		}
	}
	if cmd == "" {
		cmd = "[" + comm + "]"
	}

	return Info{
		PID: pid, PPID: ppid, Name: comm, State: state,
		User: username, MemRSS: memRSS, CPU: cpuPct,
		Command: cmd, Nice: nice,
	}, nil
}

func buildUIDMap() map[string]string {
	m := make(map[string]string)
	f, err := os.Open("/etc/passwd")
	if err != nil {
		return m
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		parts := strings.SplitN(sc.Text(), ":", 7)
		if len(parts) >= 4 {
			m[parts[2]] = parts[0]
		}
	}
	return m
}
