package util

import (
	"bufio"
	"io"
	"log/slog"
	"os/exec"
	"strings"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

func convertGBKToUTF8(r io.Reader) io.Reader {
	return transform.NewReader(r, simplifiedchinese.GBK.NewDecoder())
}

func ExecuteCommand(command string, args []string, printErrorOnly bool) error {
	slog.Info("Executing command: " + command + " " + strings.Join(args, " "))
	cmd := exec.Command(command, args...)
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	go func() {
		scanner := bufio.NewScanner(convertGBKToUTF8(stderr))
		for scanner.Scan() {
			slog.Debug(scanner.Text())
		}
	}()
	go func() {
		scanner := bufio.NewScanner(convertGBKToUTF8(stdout))
		for scanner.Scan() {
			text := scanner.Text()
			if !printErrorOnly || strings.Contains(text, "ERROR") {
				slog.Debug(text)
			}
		}
	}()
	err = cmd.Wait()
	return err
}
