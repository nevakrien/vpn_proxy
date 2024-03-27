package main

import (
	"log"
	"os/exec"
	"sync"
)

// VPNClient defines the interface for managing VPN connections
type VPNClient interface {
	Start(configPath string) error
	Stop() error
}

// OpenVPNClient implements VPNClient for managing OpenVPN connections
type OpenVPNClient struct {
	cmd    *exec.Cmd
	mutex  sync.Mutex // To safely access the cmd
	//config string     // Store the config path if needed
}

// NewOpenVPNClient creates a new OpenVPNClient instance
func NewOpenVPNClient() *OpenVPNClient {
	return &OpenVPNClient{}
}

// Start starts the VPN connection using the provided configuration
func (c *OpenVPNClient) Start(configPath string) error {
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
	c.cmd = exec.Command("openvpn", "--config", configPath, "--daemon")
	//c.config = configPath // Store the config path if needed

	if err := c.cmd.Start(); err != nil {
		return err
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
