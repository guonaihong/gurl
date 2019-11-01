package main

import (
	"fmt"
	"os/exec"
)

func main() {
	os := []string{"openbsd", "windows", "linux", "freebsd", "netbsd", "aix", "darwin", "solaris"}
	arch := []string{"arm", "arm64", "386", "amd64", "ppc64", "ppc64le", "mips", "mipsle", "mips64", "mips64le"}

	for _, o := range os {
		for _, a := range arch {
			testCmd := fmt.Sprintf("env GOPATH=`pwd` CGO_ENABLED=0 GOOS=%s GOARCH=%s go build -o gurl github.com/guonaihong/gurl/", o, a)
			cmd := exec.Command("bash", "-c", testCmd)
			_, err := cmd.Output()
			if err != nil {
				fmt.Printf("err :%s:%s\n", err, testCmd)
				return
			}

			cmd2 := exec.Command("bash", "-c", fmt.Sprintf("tar zcvf %s_%s.tar.gz gurl", o, a))
			if err = cmd2.Run(); err != nil {
				fmt.Printf("err: %s\n", err)
				return
			}

		}
	}
}
