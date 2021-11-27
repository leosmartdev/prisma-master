package lib

import (
	. "prisma/tms"
	"prisma/tms/libmain"
	"prisma/tms/log"
	. "prisma/tms/routing"
	. "prisma/tms/tmsg"

	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"os/user"
	"prisma/gogroup"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"text/tabwriter"
	"time"

	pb "github.com/golang/protobuf/proto"

	"golang.org/x/net/context"
)

type fileHandler struct {
	file *os.File
	mu   *sync.Mutex
}

type Tmsd struct {
	config TmsdConfig
	ctxt   context.Context
	cancel context.CancelFunc
	conn   net.Conn
	fh     *fileHandler

	monitors []processMonitor
}

type Status uint

const (
	New Status = iota
	Running
	Died
	Stopped
)

type processMonitor struct {
	tmsd      *Tmsd
	status    Status
	startTime time.Time
	starts    uint64
	prog      string
	args      []string
	ctxt      context.Context
	cancel    context.CancelFunc
	pid       int
}

func NewTmsd(config *TmsdConfig) *Tmsd {
	usr, err := user.Current()
	if err != nil {
		log.Fatal("Could not recognize the user to launch tmsd: %v", err)
	} else if (usr.Username == "root") && (usr.Gid == "0") && (usr.Uid == "0") {
		_, err = checkUser()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: running tmsd as root but no available user in configure file to switch\n")
			os.Exit(1)
		}
	}

	err = os.MkdirAll(config.ControlSocket[:strings.Index(config.ControlSocket, "tmsd.sock")], 0777)
	if err != nil {
		log.Fatal("Could not create directory %v:, %v", config.ControlSocket[:strings.Index(config.ControlSocket, "tmsd.sock")], err)
	}
	ret := &Tmsd{
		config: *config,
	}
	return ret
}

func (t *Tmsd) Start(ctxt context.Context) {
	t.Cleanup()
	t.ctxt, t.cancel = context.WithCancel(ctxt)
	err := t.serve()
	if err != nil {
		log.Fatal("Could not start TMSD: %v", err)
	}
}

func (t *Tmsd) Info() {

	client, err := t.connect()
	if err != nil {
		fmt.Println("There is no running tmsd now")
		log.Fatal("Could not connect to currently running TMSD: %v", err)
	}
	defer client.Close()

	fmt.Fprintln(client, "info")
	io.Copy(os.Stdout, client)

}

func (t *Tmsd) Status() {

	client, err := t.connect()
	if err != nil {
		fmt.Println("There is no running tmsd now")
		log.Fatal("Could not connect to currently running TMSD: %v", err)
	}
	defer client.Close()

	fmt.Fprintln(client, "status")
	io.Copy(os.Stdout, client)

}

func (t *Tmsd) Stop() {
	client, err := t.connect()
	if err != nil {
		fmt.Println("There is no running tmsd now")
		log.Fatal("Could not connect to currently running TMSD: %v", err)
	}

	defer client.Close()

	if t.config.Kill {
		fmt.Fprintln(client, "kill")
	}

	fmt.Fprintln(client, "stop")

	io.Copy(os.Stdout, client)
}

func (t *Tmsd) Restart(ctxt context.Context) {
	t.Stop()
	time.Sleep(time.Duration(250) * time.Millisecond)
	t.Start(ctxt)
}

func (t *Tmsd) Cleanup() {
	client, err := t.connect()
	if err == nil {
		client.Close()
		fmt.Println("There is a running tmsd now")
		log.Fatal("There is a running tmsd now")
	} else {
		t.doClean()
	}

}

func (t *Tmsd) connect() (net.Conn, error) {
	conn, err := net.Dial("unix", t.config.ControlSocket)
	return conn, err
}

func (t *Tmsd) handle(conn net.Conn) {
	r := bufio.NewReader(conn)
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			break
		}
		line = strings.TrimSpace(line)
		switch line {
		case "stop":
			log.Info("Received STOP request", line, conn)
			t.cancel()
		case "kill":
			t.config.Kill = true
		case "status":
			t.echoStatus(conn)
		case "info":
			t.echoInfo(conn)
		case "cleanup":
			fmt.Fprintln(conn, "There is already a tmsd --start running now")
		}
	}
}

func (t *Tmsd) echoStatus(conn net.Conn) {
	w := new(tabwriter.Writer)
	w.Init(conn, 0, 0, 1, ' ', 0)
	for _, monitor := range t.monitors {
		index := strings.LastIndex(monitor.prog, "/")
		fmt.Fprint(w, monitor.prog[index+1:], "\t")
		switch monitor.status {
		case New:
			fmt.Fprintln(w, "not launch")
		case Running:
			fmt.Fprintln(w, "running")
		case Died:
			fmt.Fprintln(w, "crashed")
		case Stopped:
			fmt.Fprintln(w, "stopped")
		default:
			fmt.Fprintln(w, "unknown")
		}
	}
	w.Flush()
	conn.Close()
}

func (t *Tmsd) echoInfo(conn net.Conn) {
	w := new(tabwriter.Writer)
	w.Init(conn, 0, 0, 1, ' ', 0)
	fmt.Fprintln(w, "pid\tname\tstatus\t started\t the last successful start\t command line args")
	for _, monitor := range t.monitors {
		index := strings.LastIndex(monitor.prog, "/")
		fmt.Fprint(w, monitor.pid, "\t", monitor.prog[index+1:], "\t")
		var argline string
		for _, arg := range monitor.args {
			argline = argline + " " + arg
		}
		switch monitor.status {
		case New:
			fmt.Fprintln(w, "not launch\t", monitor.starts, "times")
		case Running:
			fmt.Fprintln(w, "running\t", monitor.starts, "times\t", monitor.startTime.Format("2006-01-02 15:04:05")+" UTC\t", argline)
		case Died:
			fmt.Fprintln(w, "crashed\t", monitor.starts, "times")
		case Stopped:
			fmt.Fprintln(w, "stopped\t", monitor.starts, "times\t", monitor.startTime.Format("2006-01-02 15:04:05")+" UTC\t", argline)
		default:
			fmt.Fprintln(w, "unknown")
		}
	}
	w.Flush()
	conn.Close()
}

func (t *Tmsd) doClean() {

	sockFile, err_check := os.Stat(t.config.ControlSocket)
	if err_check == nil && !sockFile.IsDir() {
		err_remove := os.Remove(t.config.ControlSocket)
		if err_remove != nil {
			log.Fatal("Error removing tmsd.sock: %v", err_remove)
		}
	}

	t.fh = &fileHandler{
		mu: &sync.Mutex{},
	}

	if _, err_file := os.Stat(t.config.PidFile); err_file == nil {
		t.processPidFile()
	}
	t.emptyPidFile()
}

func (t *Tmsd) waitForAllStopped() {
	for {
		stopped := true
		for _, m := range t.monitors {
			if m.status == Running {
				stopped = false
			}
		}

		if stopped {
			break
		}
		time.Sleep(time.Duration(100) * time.Millisecond)
	}

	err_file := syscall.Flock(int(t.fh.file.Fd()), syscall.LOCK_UN)
	if err_file != nil {
		log.Error("Error unlocking tmsd.pid: %v", err_file)
	}
	err_file = t.fh.file.Close()
	if err_file != nil {
		log.Error("Error closing tmsd.pid: %v", err_file)
	}

	log.Info("All processes stopped.", t)
}

func (t *Tmsd) serve() error {
	l, err := net.Listen("unix", t.config.ControlSocket)
	if err != nil {
		return err
	}
	defer l.Close()

	err = os.Chmod(t.config.ControlSocket, 0777)
	if err != nil {
		fmt.Println(err)
	}

	log.Info("started")

	// Request to be notified of SIGINT rather than just dying!
	sigch := make(chan os.Signal)
	signal.Notify(sigch, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	sigch_hup := make(chan os.Signal)
	signal.Notify(sigch_hup, syscall.SIGHUP)
	go t.report()
	t.startServices()

	go func() {
		for {
			fd, err := l.Accept()
			if err != nil {
				log.Error("Accept error: %v", err)
				l, err = net.Listen("unix", t.config.ControlSocket)
			} else {
				t.conn = fd
				go t.handle(fd)
			}
		}
	}()

Loop:
	for {
		select {
		// Wait until context closes
		case sig := <-sigch_hup:
			log.Info("Got SIGHUP, exiting gracefully and restarting a new tmsd process", sig)
			for _, m := range t.monitors {
				m.cancel()
			}
			t.waitForAllStopped()
			time.Sleep(time.Duration(250) * time.Millisecond)
			l.Close()
			ReadConfig()
			t.config = Config
			t.doClean()
			t.startServices()
			continue Loop
		case <-t.ctxt.Done():
			// Do nothing -- just die
			break Loop
		case sig := <-sigch:
			// Someone signalled us to die. We should kill everything gracefully
			log.Info("Got SIGINT, exiting gracefully", sig)
			t.cancel()
			break Loop
		}
	}
	// Wait for everyone to stop
	t.waitForAllStopped()

	// Give everyone time to clean up. We don't really care if they don't quite
	// finish, though
	time.Sleep(time.Duration(250) * time.Millisecond)

	return nil
}

func (t *Tmsd) startServices() {
	numProcs := len(t.config.Processes)
	t.monitors = make([]processMonitor, numProcs)

	for i, proc := range t.config.Processes {
		t.monitors[i] = processMonitor{
			tmsd:   t,
			status: New,
			starts: 0,
			prog:   proc.findProg(),
			args:   proc.Args,
			//ctxt:   t.ctxt,
		}
		t.monitors[i].ctxt, t.monitors[i].cancel = context.WithCancel(t.ctxt)
		go (&t.monitors[i]).monitor()
	}

}

func (t *Tmsd) processPidFile() {
	var err_file error
	t.fh.mu.Lock()
	t.fh.file, err_file = os.Open(t.config.PidFile)
	if err_file != nil {
		log.Error("Error openning tmsd.pid: %v", err_file)
	}
	err_file = syscall.Flock(int(t.fh.file.Fd()), syscall.LOCK_EX)
	if err_file != nil {
		log.Error("Error locking tmsd.pid: %v", err_file)
	}

	var wg sync.WaitGroup
	scanner := bufio.NewScanner(t.fh.file)
	for scanner.Scan() {
		line := scanner.Text()
		index := strings.Index(line, "/")
		if (index >= 0) && (index < len(line)) {
			processId, err_parse := strconv.Atoi(line[:index])
			processName := line[index+1:]
			if err_parse != nil {
				log.Error("Error parsing string to int for line %v: %v", line, err_parse)
			}
			_, err_find := os.Stat("/proc/" + line[:index])
			if err_find != nil {
				log.Info("Could not find process for pid %v %v: %v", processId, processName, err_find)
			} else {
				process, _ := os.FindProcess(processId)
				wg.Add(1)
				go func() {
					defer wg.Done()
					shutDown(process, processId, processName, t, nil)
				}()
			}
		}
	}

	wg.Wait()

	err_file = syscall.Flock(int(t.fh.file.Fd()), syscall.LOCK_UN)
	if err_file != nil {
		log.Error("Error unlocking tmsd.pid: %v", err_file)
	}
	err_file = t.fh.file.Close()
	if err_file != nil {
		log.Error("Error closing tmsd.pid: %v", err_file)
	}
	t.fh.mu.Unlock()
}

func (t *Tmsd) emptyPidFile() {
	var err_file error
	t.fh.mu.Lock()
	t.fh.file, err_file = os.OpenFile(t.config.PidFile, os.O_RDWR|os.O_CREATE, 0666)
	if err_file != nil {
		log.Error("Error opening tmsd.pid: %v", err_file)
	}

	err_file = syscall.Flock(int(t.fh.file.Fd()), syscall.LOCK_EX)
	if err_file != nil {
		log.Error("Error locking tmsd.pid: %v", err_file)
	}

	_, err_file = t.fh.file.Seek(0, 0)
	if err_file != nil {
		fmt.Printf("Error seeking tmsd.pid: %v \n", err_file)
	}

	err_file = t.fh.file.Truncate(0)
	if err_file != nil {
		fmt.Printf("Error deleting content from tmsd.pid: %v \n", err_file)
	}

	t.fh.mu.Unlock()
}

func (m *processMonitor) monitor() {
	for (m.status != Stopped) && (m.status != Running) {
		counter_1 := 0
		for (m.status != Stopped) && (counter_1 < m.tmsd.config.SLaunchTimes) {
			select {
			case <-time.After(time.Duration(m.tmsd.config.SLaunchInterval) * time.Second):
				// Wait 1 second before starting
				m.starts += 1
				counter_1 += 1
				go func() {
					time.Sleep(time.Duration(m.tmsd.config.SLaunchInterval) * time.Second)
					if m.status == Running {
						counter_1 = 0
					}
				}()
				m.runOnce()

			}
		}

		counter_2 := 0
		for (m.status != Stopped) && (counter_2 < m.tmsd.config.LLaunchTimes) && (counter_1 == m.tmsd.config.SLaunchTimes) {
			select {
			case <-time.After(time.Duration(m.tmsd.config.LLaunchInterval) * time.Second):
				// Wait 1 second before starting
				m.starts += 1
				counter_2 += 1
				go func() {
					time.Sleep(time.Duration(m.tmsd.config.SLaunchInterval) * time.Second)
					if m.status == Running {
						counter_1 = 0
						counter_2 = 0
					}
				}()
				m.runOnce()

			}
		}
		if (counter_1 == m.tmsd.config.SLaunchTimes) && (counter_2 == m.tmsd.config.LLaunchTimes) {
			break
		}
	}
}

func (m *processMonitor) runOnce() error {

	proc := exec.Command(m.prog, m.args...)

	proc.Stdin = strings.NewReader("")
	stderr, stderrw := io.Pipe()
	proc.Stderr = stderrw
	stdout, stdoutw := io.Pipe()
	proc.Stdout = stdoutw

	usr, err := user.Current()
	if err != nil {
		log.Error("Could not recognize the user to launch tmsd: %v", err)
		return err
	} else if (usr.Username == "root") && (usr.Gid == "0") && (usr.Uid == "0") {
		err = switchUser(proc)
		if err != nil {
			log.Error("Could not find available user to switch: %v", err)
			return err
		}
	}

	err = proc.Start()
	if err != nil {
		log.Error("Error starting process %v %v: %v", m.prog, m.args, err)
		return err
	}

	m.status = Running
	log.Info("Started %v %v", m.prog, m.args, m, proc)

	m.startTime = time.Now().UTC()

	m.pid = proc.Process.Pid

	index := strings.LastIndex(m.prog, "/")
	prog_short := m.prog[index+1:]

	m.addPidLine()

	logwriter := func(name string, s io.Reader) {
		b := bufio.NewReaderSize(s, 10*1024)
		for {
			line, _, err := b.ReadLine()
			if err == io.EOF {
				return
			}
			if err != nil {
				log.Error("Error reading subprocess io stream %v: %v", name, err)
			} else {
				var pid int = -1
				if proc.Process != nil {
					pid = proc.Process.Pid
				}
				m.pid = pid
				log.Debug("%v(%v) %v: %v", m.prog, pid, name, string(line))
			}
		}
	}
	go logwriter("stdout", stdout)
	go logwriter("stderr", stderr)

	exitCh := make(chan error)

	go func() {
		exitCh <- proc.Wait()
	}()
	done := m.ctxt.Done()

	for {
		select {
		case <-done:
			err_shutdown := shutDown(proc.Process, m.pid, prog_short, m.tmsd, exitCh)
			m.status = Stopped
			m.removePidLine()
			return err_shutdown

		case err := <-exitCh:
			if m.status != Stopped {
				m.status = Died
				m.removePidLine()
				log.Warn("Process %v %v died with error: %v", m.pid, prog_short, err)
			} else {
				m.status = Stopped
				m.removePidLine()
				log.Info("Process %v %v died with error: %v", m.pid, prog_short, err)
			}
			return err
		}
	}
}

func (m *processMonitor) addPidLine() {

	m.tmsd.fh.mu.Lock()
	_, err_file := m.tmsd.fh.file.Seek(0, 2)
	if err_file != nil {
		log.Error("Error seeking tmsd.pid: %v", err_file)
	}

	index := strings.LastIndex(m.prog, "/")
	prog_short := m.prog[index+1:]

	newPid := strconv.Itoa(m.pid) + "/" + prog_short + "\n"

	_, err_file = m.tmsd.fh.file.WriteString(newPid)
	if err_file != nil {
		log.Error("Error writing the line %v into tmsd.pid: %v", newPid, err_file)
	}

	err_file = m.tmsd.fh.file.Sync()
	if err_file != nil {
		log.Error("Error committing change to tmsd.pid: %v", err_file)
	}
	m.tmsd.fh.mu.Unlock()
}

func (m *processMonitor) removePidLine() {

	m.tmsd.fh.mu.Lock()

	_, err_file := m.tmsd.fh.file.Seek(0, 0)
	if err_file != nil {
		fmt.Printf("Error seeking tmsd.pid: %v \n", err_file)
	}

	var buf []string
	scanner := bufio.NewScanner(m.tmsd.fh.file)
	for scanner.Scan() {
		line := scanner.Text()
		buf = append(buf, line)
	}

	_, err_file = m.tmsd.fh.file.Seek(0, 0)
	if err_file != nil {
		fmt.Printf("Error seeking tmsd.pid: %v \n", err_file)
	}

	err_file = m.tmsd.fh.file.Truncate(0)
	if err_file != nil {
		fmt.Printf("Error deleting content from tmsd.pid: %v \n", err_file)
	}

	index := strings.LastIndex(m.prog, "/")
	prog_short := m.prog[index+1:]

	for s := 0; s < len(buf); s++ {
		if strings.Contains(buf[s], strconv.Itoa(m.pid)) && strings.Contains(buf[s], prog_short) {
			continue
		} else {
			_, err_file = m.tmsd.fh.file.WriteString(buf[s] + "\n")
			if err_file != nil {
				fmt.Printf("Error writing back the line %v: %v", buf[s], err_file)
			}
		}
	}

	err_file = m.tmsd.fh.file.Sync()
	if err_file != nil {
		fmt.Printf("Error committing to tmsd.pid: %v \n", err_file)
	}
	m.tmsd.fh.mu.Unlock()
}

func shutDown(process *os.Process, processId int, processName string, t *Tmsd, exitCh chan error) error {
	err := process.Signal(syscall.SIGINT)
	if err != nil {
		log.Error("Error stopping process %v %v: %v", processId, processName, err)
	}

	ticker := time.NewTicker(time.Second * time.Duration(t.config.StatusInterval))
	timer_kill := time.NewTimer(time.Second * time.Duration(t.config.KillTime))
	timer_int := time.NewTimer(time.Second * time.Duration(t.config.SecondIntTime))

	for {
		select {
		case err := <-exitCh:
			if err == nil {
				if t.conn != nil {
					fmt.Fprintln(t.conn, "Process", processId, processName, "is stopped successfully")
				}
				log.Info("Process %v %v is stopped successfully", processId, processName)
			} else {
				if t.conn != nil {
					fmt.Fprintln(t.conn, "Process", processId, processName, "is stopped with error:", err)
				}
				log.Info("Process %v %v is stopped with error: %v", processId, processName, err)
			}
			ticker.Stop()
			timer_kill.Stop()
			timer_int.Stop()
			return err

		case <-timer_kill.C:
			ticker.Stop()
			err := process.Kill()
			if err == nil {
				if t.conn != nil {
					fmt.Fprintln(t.conn, "Process", processId, processName, "is killed")
				}
				log.Info("Process %v %v is killed", processId, processName)
			} else {
				log.Error("Error killing process %v %v: %v", processId, processName, err)
			}
			return err

		case <-timer_int.C:
			if !t.config.Kill {
				err := process.Signal(syscall.SIGINT)
				if err != nil {
					log.Error("Error stopping process %v %v: %v", processId, processName, err)
				}
			} else {
				ticker.Stop()
				err := process.Kill()
				if err == nil {
					if t.conn != nil {
						fmt.Fprintln(t.conn, "Process", processId, processName, "is killed")
					}
					log.Info("Process %v %v is killed", processId, processName)
				} else {
					log.Error("Error killing process %v %v: %v", processId, processName, err)
				}
				return err
			}

		case <-ticker.C:
			_, err := os.Stat("/proc/" + strconv.Itoa(processId))
			if err != nil {
				log.Info("Process %v %v is stopped successfully", processId, processName)
				ticker.Stop()
				timer_kill.Stop()
				timer_int.Stop()
				return err
			} else {
				if t.conn != nil {
					fmt.Fprintln(t.conn, "Process", processId, processName, "is still running")
				}
				log.Info("Process %v %v is still running", processId, processName)
			}

		}
	}
}

func (t *Tmsd) createMsgBody() pb.Message {
	loadAvg1, loadAvg10, loadAvg15, kernalRate, err := getLoadAvg()
	if err != nil {
		log.Error("Error reading load average of the system: %v", err)
	}
	cpuUsageRate, err := getCpuUsage()
	if err != nil {
		log.Error("Error reading cpu usage of the system: %v", err)
	}
	size, used, avail, err := getDiskUsage()
	if err != nil {
		log.Error("Error reading disk usage of the system: %v", err)
	}
	total, memUsed, free, err := getMemUsage()
	if err != nil {
		log.Error("Error reading memory usage of the system: %v", err)
	}
	cpuUsage := &CpuUsage{
		LoadAvg_1:     loadAvg1,
		LoadAvg_10:    loadAvg10,
		LoadAvg_15:    loadAvg15,
		ExecExistRate: kernalRate,
		CpuUsageRate:  cpuUsageRate + "%",
	}
	memUsage := &MemUsage{
		Total: total + "M",
		Used:  memUsed + "M",
		Free:  free + "M",
	}
	diskUsage := &DiskUsage{
		Size:  size,
		Used:  used,
		Avail: avail,
	}
	tmsInfo := &TmsInfo{
		Processes: []*ProcessInfo{},
		CpuUsage:  cpuUsage,
		MemUsage:  memUsage,
		DiskUsage: diskUsage,
	}

	for _, monitor := range t.monitors {
		index := strings.LastIndex(monitor.prog, "/")
		processInfo := &ProcessInfo{
			Pid:          uint32(monitor.pid),
			Name:         monitor.prog[index+1:],
			StartedTimes: uint32(monitor.starts),
			LastStart:    ToTimestamp(monitor.startTime),
		}
		switch monitor.status {
		case New:
			processInfo.Status = ProcessInfo_NOT_LAUNCH
		case Running:
			processInfo.Status = ProcessInfo_RUNNING
		case Died:
			processInfo.Status = ProcessInfo_CRASHED
		case Stopped:
			processInfo.Status = ProcessInfo_STOPPED
		default:
			processInfo.Status = ProcessInfo_UNKNOWN
		}
		tmsInfo.Processes = append(tmsInfo.Processes, processInfo)
	}
	return tmsInfo
}

func (t *Tmsd) report() {

	libmain.Main(APP_ID_TMSD, func(ctxt gogroup.GoGroup) {
		go func() {
			msgChan := GClient.Listen(ctxt, Listener{
				Destination: &EndPoint{
					Site: TMSG_LOCAL_SITE,
					Aid:  APP_ID_TMSD,
				},
				MessageType: "prisma.tms.TmsInfo",
			})
			for {
				select {
				case <-ctxt.Done():
					return
				case tmsg := <-msgChan:
					infoMsg, err := t.createReportMsg(GClient.Local(), tmsg.TsiMessage.Source.Site)
					if err != nil {
						log.Error("Error packing pb.message: %v", err)
						break
					}
					GClient.Send(ctxt, &infoMsg)
				}
			}
		}()

		ticker := time.NewTicker(time.Duration(t.config.ReportInterval) * time.Second)
		for {
			select {
			case <-ctxt.Done():
				ticker.Stop()
				return
			case <-ticker.C:
				infoMsg, err := t.createReportMsg(GClient.Local(), TMSG_HQ_SITE)
				if err != nil {
					log.Error("Error packing pb.message: %v", err)
					break
				}
				GClient.Send(ctxt, &infoMsg)
			}
		}
	})
}

func (t *Tmsd) createReportMsg(source *EndPoint, site uint32) (TsiMessage, error) {
	body, err := PackFrom(t.createMsgBody())
	if err != nil {
		return TsiMessage{}, err
	}
	infoMsg := TsiMessage{
		Source: source,
		Destination: []*EndPoint{
			&EndPoint{
				Site: site,
			},
		},
		WriteTime: Now(),
		SendTime:  Now(),
		Status:    TsiMessage_BROADCAST,
		Body:      body,
	}
	return infoMsg, nil
}

func getLoadAvg() (string, string, string, string, error) {
	loadAvg1 := ""
	loadAvg10 := ""
	loadAvg15 := ""
	kernalRate := ""
	file, err := os.Open("/proc/loadavg")
	if err != nil {
		return loadAvg1, loadAvg10, loadAvg15, kernalRate, err
	}
	err = syscall.Flock(int(file.Fd()), syscall.LOCK_EX)
	if err != nil {
		return loadAvg1, loadAvg10, loadAvg15, kernalRate, err
	}
	err = errors.New("Load average is not found in /proc/loadavg")
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		if line == "" {
			// Skip blank lines
			continue
		}
		fields := strings.Fields(line)
		if len(fields) >= 5 {
			loadAvg1 = fields[0]
			loadAvg10 = fields[1]
			loadAvg15 = fields[2]
			kernalRate = fields[3]
			err = nil
		}

	}
	err = syscall.Flock(int(file.Fd()), syscall.LOCK_UN)
	if err != nil {
		return loadAvg1, loadAvg10, loadAvg15, kernalRate, err
	}
	err = file.Close()
	if err != nil {
		return loadAvg1, loadAvg10, loadAvg15, kernalRate, err
	}
	return loadAvg1, loadAvg10, loadAvg15, kernalRate, err
}

func getCpuUsage() (string, error) {
	cpuUsageRate := ""
	file, err := os.Open("/proc/stat")
	if err != nil {
		return cpuUsageRate, err
	}
	err = syscall.Flock(int(file.Fd()), syscall.LOCK_EX)
	if err != nil {
		return cpuUsageRate, err
	}
	err = errors.New("Cpu usage is not found in /proc/stat")
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		if line == "" {
			// Skip blank lines
			continue
		}
		fields := strings.Fields(line)
		if (len(fields) >= 5) && (fields[0] == "cpu") {
			var total int = 0
			var idle int = 0
			for i := 1; i < len(fields); i++ {
				val, err := strconv.Atoi(fields[i])
				if err != nil {
					return cpuUsageRate, err
				}
				total += val
				if i == 4 {
					idle = val
				}
			}
			cpuUsageRate = strconv.Itoa(100 * (total - idle) / total)
			err = nil
		}
	}
	err = syscall.Flock(int(file.Fd()), syscall.LOCK_UN)
	if err != nil {
		return cpuUsageRate, err
	}
	err = file.Close()
	if err != nil {
		return cpuUsageRate, err
	}
	return cpuUsageRate, err
}

func getMemUsage() (string, string, string, error) {
	out, err := exec.Command("free", "-m").Output()
	if err != nil {
		log.Error("Error executing free -m: %v", err)
		return "", "", "", err
	}
	total := ""
	used := ""
	free := ""
	buf := bytes.NewBuffer(out)
	for {
		line, err := buf.ReadString('\n')
		if err != nil {
			break
		}
		line = strings.TrimSpace(line)
		if line == "" {
			// Skip blank lines
			continue
		}
		if strings.Index(line, "Mem") == 0 {
			fields := strings.Fields(line)
			if len(fields) >= 4 {
				total = fields[1]
				used = fields[2]
				free = fields[3]
			}
		}
	}
	if total == "" {
		return "", "", "", errors.New("Mem line is not found from free -m")
	}
	return total, used, free, nil
}

func getDiskUsage() (string, string, string, error) {
	out, err := exec.Command("df", "-h").Output()
	if err != nil {
		log.Error("Error executing df -h: %v", err)
		return "", "", "", err
	}
	buf := bytes.NewBuffer(out)
	for {
		line, err := buf.ReadString('\n')
		if err != nil {
			break
		}
		line = strings.TrimSpace(line)
		if line == "" {
			// Skip blank lines
			continue
		}
		if strings.Index(line, "/dev/") == 0 {
			fields := strings.Fields(line)
			if len(fields) >= 6 {
				return fields[1], fields[2], fields[3], nil
			}
		}
	}
	err = errors.New("No /dev/ line is found from df -h")
	return "", "", "", err
}

func switchUser(proc *exec.Cmd) error {

	newUsr, err := checkUser()

	if err == nil {

		newGid, err := strconv.Atoi(newUsr.Gid)
		if err != nil {
			log.Error("Could not change found gid (%s) of %s from string to int: %v", newUsr.Gid, newUsr.Username, err)
			return err
		}

		newUid, err := strconv.Atoi(newUsr.Uid)
		if err != nil {
			log.Error("Could not change found uid (%s) of %s from string to int: %v", newUsr.Uid, newUsr.Username, err)
			return err
		}

		proc.SysProcAttr = &syscall.SysProcAttr{}
		proc.SysProcAttr.Credential = &syscall.Credential{Uid: uint32(newUid), Gid: uint32(newGid)}
	}

	return err

}

func checkUser() (*user.User, error) {

	users := strings.Split(Config.DefaultUsers, ",")

	var newUsr *user.User
	var err error
	foundUser := false

	for _, userOption := range users {
		userOption = strings.TrimSpace(userOption)
		if userOption != "" {
			newUsr, err = user.Lookup(userOption)
			if err == nil {
				foundUser = true
				break
			}
		}
	}

	if foundUser {
		return newUsr, nil
	}
	err = errors.New("no available user from configure file")
	return nil, err
}
