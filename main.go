package main

import (
	"bufio"
	"fmt"
	"gopkg.in/yaml.v3"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

// Process defines the structure for each process to run
type Process struct {
	Name       string   `yaml:"name"`
	Command    string   `yaml:"command"`
	Args       []string `yaml:"args"`
	WorkingDir string   `yaml:"working_dir"`
}

// Config holds all processes to run
type Config struct {
	Processes []Process `yaml:"processes"`
}

// loadConfig loads the configuration from a YAML file
func loadConfig(configFile string) (*Config, error) {
	config := &Config{}
	file, err := os.ReadFile(configFile)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(file, config)
	if err != nil {
		return nil, err
	}
	return config, nil
}

// runProcess starts a process and restarts it if it crashes
func runProcess(proc Process) {
	for {
		cmd := exec.Command(proc.Command, proc.Args...)
		cmd.Dir, _ = filepath.Abs(proc.WorkingDir)
		cmd.Stdin = os.Stdin

		stdoutPipe, err := cmd.StdoutPipe()
		if err != nil {
			log.Fatalf("Failed to get stdout for process %s: %v", proc.Name, err)
		}
		stderrPipe, err := cmd.StderrPipe()
		if err != nil {
			log.Fatalf("Failed to get stderr for process %s: %v", proc.Name, err)
		}

		if err := cmd.Start(); err != nil {
			log.Fatalf("Failed to start process %s: %v", proc.Name, err)
		}

		go pipeOutput(proc.Name, stdoutPipe)
		go pipeOutput(proc.Name, stderrPipe)

		fmt.Printf("Started process: %s\n", proc.Name)

		// Wait for the process to exit and log if it exits unexpectedly
		err = cmd.Wait()
		if err != nil {
			fmt.Printf("Process %s exited unexpectedly: %v\n", proc.Name, err)
		} else {
			fmt.Printf("Process %s exited normally.\n", proc.Name)
		}

		// Wait a bit before restarting
		time.Sleep(5 * time.Second)
	}
}

// pipeOutput reads the output from a reader and prints each line with a prefix
func pipeOutput(prefix string, reader io.ReadCloser) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		fmt.Printf("[%s] %s\n", prefix, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading output from %s: %v\n", prefix, err)
	} else {
		fmt.Printf("[%s] Process completed with no output.\n", prefix)
	}
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: runner <config-file>")
	}
	configFile := os.Args[1]

	config, err := loadConfig(configFile)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Trap interrupt signal to gracefully shutdown processes
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	for _, proc := range config.Processes {
		go runProcess(proc)
	}

	<-sigChan
	fmt.Println("Shutting down runner...")
}
