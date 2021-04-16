package main

import (
	"bufio"
	"io"
	"os/exec"
	"syscall"
)

type command struct {
	args   []string
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout io.ReadCloser
	stderr io.ReadCloser
}

func newCommand(cmd string, args []string) (*command, error) {
	r := &command{}
	r.args = args
	r.cmd = exec.Command(cmd, args...)
	var err error

	if r.stdin, err = r.cmd.StdinPipe(); err != nil {
		return nil, err
	}

	if r.stdout, err = r.cmd.StdoutPipe(); err != nil {
		return nil, err
	}

	if r.stderr, err = r.cmd.StderrPipe(); err != nil {
		return nil, err
	}
	return r, nil
}

func (r *command) writeToStdin(data []byte) (err error) {
	index := 0
	for len(data[index:]) > 0 {
		var n = 0
		if n, err = r.stdin.Write(data[index:]); err != nil {
			logger.Errorln(err)
			return err
		}
		index += n
	}
	return nil
}

func (r *command) run() (err error) {
	if err = r.cmd.Start(); err != nil {
		return err
	}

	if err = r.cmd.Wait(); err != nil {
		return err
	}
	return nil
}

type ffmpegCommand struct {
	command
}

func newFfmpegCommand(cmd string, args []string) (*ffmpegCommand, error) {
	c, err := newCommand(cmd, args)
	if err != nil {
		return nil, err
	}

	return &ffmpegCommand{
		command: *c,
	}, nil
}

func (c *ffmpegCommand) run() (err error) {
	go func() {
		var line string
		var err error
		r := bufio.NewReader(c.stderr)
		for {
			if line, err = r.ReadString('\n'); err != nil {
				break
			}
			logger.Info(line)
		}
		if err != io.EOF {
			logger.Errorf("failed to read from ffmpeg stderr, %v\n", err)
		}
	}()
	return c.command.run()
}

func (c *ffmpegCommand) stop() {
	if err := c.cmd.Process.Signal(syscall.SIGINT); err != nil {
		logger.Errorf("failed to stop ffmpeg, %v", err)
	}
}
