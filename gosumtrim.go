// Binary gosumtrim trims the go.sum file based on the given go.mod file.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/golang/glog"
)

var (
	modFile = flag.String("mod", "", "input go.mod file")
	sumFile = flag.String("sum", "", "input go.sum file")
	outFile = flag.String("out", "", "output go.sum file")
)

func main() {
	flag.Parse()

	modIn, err := os.Open(*modFile)
	if err != nil {
		glog.Exitf("Open(%q) failed: %v", *modFile, err)
	}
	sumIn, err := os.Open(*sumFile)
	if err != nil {
		glog.Exitf("Open(%q) failed: %v", *sumFile, err)
	}
	sumOut, err := os.Create(*outFile)
	if err != nil {
		glog.Exitf("Create(%q) failed: %v", *outFile, err)
	}

	if err := trim(modIn, sumIn, sumOut); err != nil {
		glog.Exitf("Trim() failed: %v", err)
	}
	if err := sumOut.Close(); err != nil {
		glog.Exitf("os.Close() fialed: %v", err)
	}
}

func trim(mod io.Reader, sum io.Reader, o io.Writer) error {
	m := bufio.NewReader(mod)
	s := bufio.NewReader(sum)

	// Process require lines and add dependencies into used map.
	used := make(map[string]bool)

readLoop:
	for {
		// Read till we get to the "require (\n".
		for {
			line, err := m.ReadString('\n')
			if err == io.EOF {
				break readLoop
			}
			if err != nil {
				return fmt.Errorf("reading mod file failed: %v", err)
			}

			glog.Infof("line: %v", line)

			if line == "require (\n" {
				break
			}
		}

		// Read the lines and process them until we reach ")\n".
		for {
			line, err := m.ReadString('\n')
			if err == io.EOF {
				break readLoop
			}
			if err != nil {
				return fmt.Errorf("reading input go.mod failed: %v", err)
			}

			glog.Infof("line: %v", line)

			if line == ")\n" {
				break
			}

			line = strings.Trim(line, "\n \t")
			line = strings.Split(line, " // ")[0]
			used[line] = true
		}
	}

	glog.Infof("used: %v", used)

	// Filter sum->out, dropping what is not used in mod.
	for {
		line, err := s.ReadString('\n')
		if err == io.EOF {
			break
		}
		glog.Infof("line: %v", line)
		if err != nil {
			return fmt.Errorf("reading input go.sum failed: %v", err)
		}
		parts := strings.Split(strings.TrimSpace(line), " ")
		if len(parts) != 3 {
			return fmt.Errorf("expecting 3 parts per go.sum line, got %v", parts)
		}
		pkg := parts[0]
		ver := strings.Split(parts[1], "/")[0]
		if !used[strings.Join([]string{pkg, ver}, " ")] {
			continue
		}
		if _, err := o.Write([]byte(line)); err != nil {
			return fmt.Errorf("writing output go.sum failed: %v", err)
		}
	}
	return nil
}
