package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

var update = flag.Bool("update", false, "update golden files")

func runSimulator(path string, cycles int) ([]byte, error) {
	cmd := exec.Command("./eniacsim", "-g", "-t", strconv.Itoa(cycles), path)
	return cmd.Output()
}

func TestGolden(t *testing.T) {
	files, err := ioutil.ReadDir("testdata")
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range files {
		path := filepath.Join("testdata", file.Name())
		if strings.HasSuffix(file.Name(), ".e") {
			fmt.Printf("%s...", path)
			out, err := runSimulator(path, 50000)
			if err != nil {
				t.Errorf("%s", err)
			}

			golden, err := ioutil.ReadFile(path + ".out")
			if err != nil {
				t.Fatalf("missing golden file for %s", path)
			}
			if bytes.Compare(out, golden) != 0 {
				if *update {
					fmt.Printf("update %s.out\n", path)
					err = ioutil.WriteFile(path+".out", out, 0644)
				} else {
					fmt.Printf("fail %s.bad\n", path)
					err = ioutil.WriteFile(path+".bad", out, 0644)
				}
				if err != nil {
					t.Fatalf("error saving output %s", err)
				}
				t.Errorf("failed %s", file.Name())
			} else {
				fmt.Printf("ok\n")
			}
		}
	}
}
