package main

import (
	"log"
	"os/exec"
	"sync"
	"os"
	"strings"
	"bufio"
	"fmt"
)

const OPENVPN_OK_STRING="Initialization Sequence Completed"

// VPNClient defines the interface for managing VPN connections
type VPNClient interface {
	Start(configPath string) error
	Stop() error
}

// OpenVPNClient implements VPNClient for managing OpenVPN connections
type OpenVPNClient struct {
	cmd    *exec.Cmd
	mutex  sync.RWMutex // To safely access the cmd
	authPath string     
}

// NewOpenVPNClient creates a new OpenVPNClient instance
func NewOpenVPNClient(authPath string) *OpenVPNClient {
	return &OpenVPNClient{
        authPath: authPath,
    }
}

// Start starts the VPN connection using the provided configuration
func (c *OpenVPNClient) Start(configPath string) error {
	fmt.Println("setting VPN...")
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Stop any existing connection
	if c.cmd != nil && c.cmd.Process != nil {
		if err := c.cmd.Process.Kill(); err != nil {
			log.Printf("Failed to kill existing OpenVPN process: %v", err)
			// Consider how you want to handle this error
		}
	}

	// Start a new OpenVPN process
	c.cmd = exec.Command("openvpn", "--config", configPath,"--auth-user-pass",c.authPath)
	//redirect outputs
	//c.cmd.Stdout = os.Stdout //moved to scanner
    c.cmd.Stderr = os.Stderr

    cliPipe, err := c.cmd.StdoutPipe()

    if err != nil {
        return err
    }

    if err := c.cmd.Start(); err != nil {
        return err
    }

	//wait for the OK
	scanner := bufio.NewScanner(cliPipe)
    for scanner.Scan() {
        //fmt.Println(scanner.Text()) // Debug: Print every line to verify output
        if strings.Contains(scanner.Text(), OPENVPN_OK_STRING) {
            fmt.Println("VPN active")
            break
        }
    }

    if err := scanner.Err(); err != nil {
	    return fmt.Errorf("error reading openvpn output: %w", err)
	}

	return nil
}

// Stop terminates the VPN connection
func (c *OpenVPNClient) Stop() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.cmd != nil && c.cmd.Process != nil {
		if err := c.cmd.Process.Kill(); err != nil {
			return err
		}
		c.cmd = nil // Clear the cmd after stopping
	}

	return nil
}
