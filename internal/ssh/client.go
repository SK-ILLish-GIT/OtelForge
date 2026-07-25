package sshclient

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"net"
	"time"

	"golang.org/x/crypto/ssh"
)

type Config struct {
	Host                       string
	Port                       int
	User                       string
	Password                   string
	PrivateKeyPEM              string
	ExpectedHostKeyFingerprint string
	PinHostKey                 func(fingerprint string) error
}

type Client struct {
	cfg Config
}

func New(cfg Config) *Client {
	return &Client{cfg: cfg}
}

func (c *Client) connect() (*ssh.Client, error) {
	auth, err := c.authMethods()
	if err != nil {
		return nil, err
	}
	config := &ssh.ClientConfig{
		User:            c.cfg.User,
		Auth:            auth,
		HostKeyCallback: verifyHostKey(c.cfg.ExpectedHostKeyFingerprint, c.cfg.PinHostKey),
		Timeout:         15 * time.Second,
	}
	addr := net.JoinHostPort(c.cfg.Host, fmt.Sprintf("%d", c.cfg.Port))
	return ssh.Dial("tcp", addr, config)
}

func (c *Client) authMethods() ([]ssh.AuthMethod, error) {
	var methods []ssh.AuthMethod
	if c.cfg.PrivateKeyPEM != "" {
		signer, err := ssh.ParsePrivateKey([]byte(c.cfg.PrivateKeyPEM))
		if err != nil {
			return nil, fmt.Errorf("parse private key: %w", err)
		}
		methods = append(methods, ssh.PublicKeys(signer))
	}
	if c.cfg.Password != "" {
		methods = append(methods, ssh.Password(c.cfg.Password))
	}
	if len(methods) == 0 {
		return nil, fmt.Errorf("no ssh credentials configured")
	}
	return methods, nil
}

func (c *Client) RunRemote(command string) (stdout, stderr string, exitCode int, err error) {
	client, err := c.connect()
	if err != nil {
		return "", err.Error(), 1, err
	}
	defer client.Close()
	return runSession(client, command)
}

func runSession(client *ssh.Client, command string) (stdout, stderr string, exitCode int, err error) {
	session, err := client.NewSession()
	if err != nil {
		return "", err.Error(), 1, err
	}
	defer session.Close()

	var outBuf, errBuf bytes.Buffer
	session.Stdout = &outBuf
	session.Stderr = &errBuf
	runErr := session.Run(command)
	exitCode = 0
	if runErr != nil {
		if ee, ok := runErr.(*ssh.ExitError); ok {
			exitCode = ee.ExitStatus()
		} else {
			return outBuf.String(), errBuf.String(), 1, runErr
		}
	}
	return outBuf.String(), errBuf.String(), exitCode, nil
}

func (c *Client) RunScript(script []byte, configContent *string) (stdout, stderr string, exitCode int, err error) {
	client, err := c.connect()
	if err != nil {
		return "", err.Error(), 1, err
	}
	defer client.Close()

	remoteDir := fmt.Sprintf("/tmp/otelforge-%d", time.Now().UnixNano())
	scriptB64 := base64.StdEncoding.EncodeToString(script)
	setup := fmt.Sprintf(`set -e
mkdir -p %s
echo '%s' | base64 -d > %s/task.sh
chmod +x %s/task.sh`, remoteDir, scriptB64, remoteDir, remoteDir)

	cmd := setup
	if configContent != nil {
		cfgB64 := base64.StdEncoding.EncodeToString([]byte(*configContent))
		cmd += fmt.Sprintf(`
echo '%s' | base64 -d > %s/config.yaml`, cfgB64, remoteDir)
		cmd += fmt.Sprintf("\nbash %s/task.sh %s/config.yaml", remoteDir, remoteDir)
	} else {
		cmd += fmt.Sprintf("\nbash %s/task.sh", remoteDir)
	}

	return runSession(client, cmd)
}

func (c *Client) TestConnectivity() error {
	_, _, code, err := c.RunRemote("echo ok")
	if err != nil {
		return err
	}
	if code != 0 {
		return fmt.Errorf("ssh test failed exit %d", code)
	}
	return nil
}
