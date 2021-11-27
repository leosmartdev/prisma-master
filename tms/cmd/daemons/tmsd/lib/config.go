// Package lib has tmsd work process functions.
package lib

import (
	"prisma/tms/log"
	_ "prisma/tms/tmsg"

	"bufio"
	"bytes"
	"errors"
	"flag"
	"io"
	"os"
	"path"
	"strings"

	"github.com/kardianos/osext"
)

var (
	Config TmsdConfig
)

type TmsdConfig struct {
	ConfigFile      string
	PidFile         string
	ControlSocket   string
	StatusInterval  int
	SecondIntTime   int
	KillTime        int
	SLaunchInterval int
	SLaunchTimes    int
	LLaunchInterval int
	LLaunchTimes    int
	ReportInterval  int
	Kill            bool
	DefaultUsers    string

	Processes     []ProcessConfig
	GlobalOptions []string
}

type ProcessConfig struct {
	Prog string
	Args []string
}

func init() {
	flag.StringVar(&Config.ConfigFile, "config", "/etc/trident/tmsd.conf",
		"Location of configuration file")
	flag.StringVar(&Config.ControlSocket, "control", "/tmp/tmsd.sock",
		"Location of TMSD control socket")
	flag.StringVar(&Config.PidFile, "pid", "/tmp/tmsd.pid",
		"Location of file for tmsd to place subprocess PIDs")
	flag.IntVar(&Config.StatusInterval, "si", 5,
		"Time interval to print status of subprocesses when tmsd -stop")
	flag.IntVar(&Config.SecondIntTime, "se", 5,
		"Wait time to try the second time to make subprocesses exit gracefully when tmsd -stop")
	flag.IntVar(&Config.KillTime, "kt", 60,
		"Wait time to kill the subprocesses when tmsd -stop")
	flag.IntVar(&Config.SLaunchInterval, "sli", 1,
		"Short time interval to try launching not started or crashed sub-processes in the first phase")
	flag.IntVar(&Config.SLaunchTimes, "slt", 30,
		"Times to try launching not started or crashed sub-processes in the first phase")
	flag.IntVar(&Config.LLaunchInterval, "lli", 10,
		"Long time interval to try launching not started or crashed sub-processes in the second phase")
	flag.IntVar(&Config.LLaunchTimes, "llt", 30,
		"Times to try launching not started or crashed sub-processes in the second phase")
	flag.IntVar(&Config.ReportInterval, "ri", 60*5,
		"Interval to send tms info to front end")
	flag.BoolVar(&Config.Kill, "kill", false,
		"Not try making subprocesses exit gracefully by sending sigint the second time when tmsd -stop, kill them instead")
	flag.StringVar(&Config.DefaultUsers, "users", "prisma",
		"A list of default users to switch to launch tmsd instead if a root try to launch it, will be tried in order")
}

func (pc ProcessConfig) findProg() string {
	prog := pc.Prog
	if !path.IsAbs(prog) {
		// An absolute path was NOT specified, so resolve relative to us.
		exedir, err := osext.ExecutableFolder()
		if err != nil {
			log.Error("Error getting executable folder: %v", err)
		} else {
			proposed := path.Join(path.Clean(exedir), prog)
			stat, err := os.Stat(proposed)
			// If the file exists, use it
			if err == nil && !stat.IsDir() {
				prog = proposed
			}
		}
	}
	return prog
}

func ReadConfig() (*TmsdConfig, error) {
	Config.Processes = make([]ProcessConfig, 0, 16)

	comments, err := readConfigEliminateComments()

	if err != nil {
		return &Config, err
	}

	buf := bytes.NewBuffer(comments)

	var line string = ""

	for {
		line, err = buf.ReadString('\n')
		if err != nil {
			break
		}

		line = strings.TrimSpace(line)
		if line == "" {
			// Skip blank lines
			continue
		}

		if line[0] == '{' {
			// Begin reading process def
			err = readDef(buf, line[1:])
			if err != nil {
				break
			}
		} else if strings.Contains(line, "=") {
			eqIdx := strings.Index(line, "=")
			key := line[0:eqIdx]
			value := line[eqIdx+1:]
			err = flag.Set(key, value)
			if err != nil {
				log.Error("Error setting config setting: %v", err)
				break
			}
		} else {
			log.Error("Could not parse config file line: %s", line)
			err = errors.New("Config comment format error with line: %s")
			break
		}
	}

	if err != io.EOF {
		log.Error("Got error reading config file: %v", err)
		return &Config, err
	}

	for index, _ := range Config.Processes {
		Config.Processes[index].Args = append(Config.GlobalOptions, Config.Processes[index].Args...)
	}

	return &Config, nil
}

func readDef(buf *bytes.Buffer, line string) error {
	if line[len(line)-1] != '}' {
		rest, err := buf.ReadString('}')
		if err != nil {
			log.Error("Could not read process or global options definition: %s", line)
			return errors.New("invalid process definition")
		}

		line = line + rest[:len(rest)-1]
	} else {
		line = line[:len(line)-1]
	}

	line = strings.Replace(line, "\n", "", -1)
	line = strings.TrimSpace(line)

	if strings.Index(line, "global") == 0 {
		Config.GlobalOptions = strings.Fields(line)[1:]
	} else {
		fields := strings.Fields(line)
		Config.Processes = append(Config.Processes,
			ProcessConfig{
				Prog: fields[0],
				Args: fields[1:],
			})
	}
	return nil
}

func readConfigEliminateComments() ([]byte, error) {
	cf, err := os.Open(Config.ConfigFile)
	if err != nil {
		log.Error("Could not open configuation file: %v", err)
		return []byte{}, err
	}

	buf := bufio.NewReader(cf)
	outbuf := bytes.NewBuffer([]byte{})

	for {
		line, err := buf.ReadString('\n')
		if err != nil {
			break
		}

		idx := strings.Index(line, "#")
		if idx >= 0 {
			line = line[0:idx]
		}
		outbuf.WriteString(line + "\n")
	}

	return outbuf.Bytes(), err
}
