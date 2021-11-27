package main

import (
	"flag"
	"io"
	"io/ioutil"
	"os"
	"prisma/gogroup"
	"prisma/tms/libmain"
	"prisma/tms/log"
	"prisma/tms/tmsg"
	"regexp"
	"sort"
	"strconv"
	"time"
)

var (
	srcDir string
	dstDir string
)

type FileAttributes struct {
	filename  string
	timestamp int64
}

func init() {
	flag.StringVar(&srcDir, "src-dir", "~/capture", "location of the mcc data to replayed")
	flag.StringVar(&dstDir, "dst-dir", "/srv/ftp", "location of the mcc ftp server")
}

func main() {
	flag.Parse()
	libmain.Main(tmsg.APP_ID_TMCCRD, realmain)
}

func realmain(ctxt gogroup.GoGroup) {

	for {
		files, err := ioutil.ReadDir(srcDir)
		if err != nil {
			log.Fatal("unable to scan directory %s: %v", srcDir, err)
		}

		var list = make([]FileAttributes, len(files))

		log.Debug("scanning files in %s ...", srcDir)

		for key, file := range files {
			value, err := strconv.ParseInt(file.Name(), 10, 64)
			if err != nil {
				continue
			}
			list[key].timestamp = value
			list[key].filename = file.Name()
		}

		log.Debug("sorting the list of files in %s ...", srcDir)
		sort.Slice(list, func(i, j int) bool { return list[i].timestamp < list[j].timestamp })

		log.Debug("files in %s are sorted", srcDir)

		for key, file := range list {

			log.Debug("trying to replay %s...", file.filename)

			if file.filename == "" && file.timestamp == 0 {
				continue
			}

			mccmsg, err := ioutil.ReadFile(srcDir + "/" + file.filename)
			if err != nil {
				log.Error("unable to read file %s: %v", file.filename, err)
			}

			dateregexp, err := regexp.Compile("date=\"(.+)\"")
			if err != nil {
				log.Error("unable to compile date search reg exp: %v", err)
			}

			raw := dateregexp.ReplaceAllString(string(mccmsg), "date=\""+time.Now().Format("2006-01-02T15:04:05Z")+"\" ")

			tcaregexp, err := regexp.Compile("<tca>(.+)</tca>")
			if err != nil {
				log.Error("unable to compile tca search reg exp: %v", err)
			}

			raw = tcaregexp.ReplaceAllString(raw, "<tca>"+time.Now().Format("2006-01-02T15:04:05.000Z")+"</tca>")

			//Open file to start writting xml raw data
			f, err := os.OpenFile(srcDir+"/"+file.filename, os.O_RDWR|os.O_CREATE, 0755)
			if err != nil {
				log.Error("unable to open file %s: %v", file.filename, err)
				continue
			}

			_, err = f.Write([]byte(raw))
			if err != nil {
				log.Error("unable to write to file %s: %v", file.filename, err)
				continue
			}

			if err := f.Close(); err != nil {
				log.Warn("unable to close file %s: %v", file.filename, err)
			}

			err = copyfile(srcDir+"/"+file.filename, dstDir+"/"+file.filename+".tmp")
			if err != nil {
				log.Error("unable to copy file %s from %s to %s : %v", file.filename, srcDir, dstDir, err)
				continue
			}

			err = os.Rename(dstDir+"/"+file.filename+".tmp", dstDir+"/"+file.filename+".txt")
			if err != nil {
				log.Error("unable to rename %s from %s.tmp to %s.txt: err", file.filename, file.filename, file.filename, err)
			}

			if key != len(list)-1 {
				log.Debug("will replay next message in %d seconds", list[key+1].timestamp-file.timestamp)
				time.Sleep(time.Second * time.Duration(list[key+1].timestamp-file.timestamp))
			}
		}
	}
}

func copyfile(src, dst string) error {

	input, err := os.Open(src)
	if err != nil {
		return err
	}
	defer input.Close()

	output, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer output.Close()

	_, err = io.Copy(output, input)
	if err != nil {
		return err
	}
	return nil
}
