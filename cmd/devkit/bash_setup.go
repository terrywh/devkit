package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/terrywh/devkit/util"
)

type BashSetup struct {
	server BashBackend
}

func (s *BashSetup) Serve(ctx context.Context, arch string) (err error) {
	log.Println("<bash-setup> preparing ...")
	time.Sleep(1 * time.Second)
	log.Println("<bash-setup> installing ...")
	time.Sleep(1 * time.Second)
	// name := "trzsz_1.1.7_linux_x86_64"
	// name := "trzsz_1.1.7_linux_aarch64"
	s.install(fmt.Sprintf("trzsz_1.1.7_linux_%s", arch))
	log.Println("<bash-setup> done.")
	return
}

func (s *BashSetup) install(name string) {
	path := fmt.Sprintf("/Users/terryhaowu/data/htdocs/github.com/terrywh/devkit/var/%s.tar.gz", name)
	file, _ := os.Open(path)
	defer file.Close()
	io.WriteString(s.server, "rm -rf /usr/local/trzsz\r")
	io.WriteString(s.server, fmt.Sprintf("base64 -di -w 128 > /tmp/%s.tar.gz\r", name))
	time.Sleep(100 * time.Millisecond)
	rfc2045 := util.NewRfc2045(128, s.server)
	encoder := base64.NewEncoder(base64.StdEncoding, rfc2045)
	io.Copy(encoder, file)
	encoder.Close()
	// rfc2045.Flush()
	time.Sleep(100 * time.Millisecond)
	// Ctrl+D x2
	io.WriteString(s.server, "\x04")
	time.Sleep(100 * time.Millisecond)
	io.WriteString(s.server, "\x04")
	time.Sleep(100 * time.Millisecond)
	io.WriteString(s.server, fmt.Sprintf("tar x -C /tmp -f /tmp/%s.tar.gz\r", name))
	io.WriteString(s.server, fmt.Sprintf("tar x -C /tmp -f /tmp/%s.tar.gz\r", name))
	io.WriteString(s.server, fmt.Sprintf("mv /tmp/%s /usr/local/trzsz\r", name))
	io.WriteString(s.server, fmt.Sprintf("rm -rf /tmp/%s.tar.gz\r", name))
	io.WriteString(s.server, "ln -s /usr/local/trzsz/trz /usr/bin/trz\r")
	io.WriteString(s.server, "ln -s /usr/local/trzsz/tsz /usr/bin/tsz\r")
}
