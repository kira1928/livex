package build

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

func BuildGoBinary(isDev bool) {
	goHostOS := runtime.GOOS
	goHostArch := runtime.GOARCH
	goVersion := runtime.Version()
	goTags := "release"
	gcflags := ""
	debug_build_flags := " -s -w "
	if isDev {
		goTags = "dev"
		gcflags = "all=-N -l"
		debug_build_flags = ""
	}
	fmt.Printf("building xlive (Platform: %s, Arch: %s, GoVersion: %s, Tags: %s)\n", goHostOS, goHostArch, goVersion, goTags)

	constsPath := "github.com/kira1928/xlive/src/consts"
	now := fmt.Sprintf("%d", time.Now().Unix())
	t := template.Must(template.New("ldFlags").Parse(
		`{{.DebugBuildFlags}} \
		-X {{.ConstsPath}}.BuildTime={{.Now}} \
		-X {{.ConstsPath}}.AppVersion={{.AppVersion}} \
		-X {{.ConstsPath}}.GitHash={{.GitHash}}`))

	var buf bytes.Buffer
	t.Execute(&buf, map[string]string{
		"DebugBuildFlags": debug_build_flags,
		"ConstsPath":      constsPath,
		"Now":             now,
		"AppVersion":      getGitTagString(),
		"GitHash":         getGitHash(),
	})
	ldflags := buf.String()

	cmd := exec.Command(
		"go", "build",
		"-tags", goTags,
		`-gcflags=`+gcflags,
		"-o", "bin/"+generateBinaryName(),
		"-ldflags="+ldflags,
		"./src/cmd/xlive",
	)
	cmd.Env = append(
		os.Environ(),
		"GOOS="+goHostOS,
		"GOARCH="+goHostArch,
		"CGO_ENABLED=0",
		"UPX_ENABLE=0",
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	log.Print(cmd.String())
	err := cmd.Run()
	if err != nil {
		fmt.Printf("Command finished with error: %v", err)
	}
}

func generateBinaryName() string {
	binaryName := "xlive-" + runtime.GOOS + "-" + runtime.GOARCH
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}
	return binaryName
}

func getGitHash() string {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(out))
}

func getGitTagString() string {
	cmd := exec.Command("git", "describe", "--tags", "--always")
	out, err := cmd.Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(out))
}
