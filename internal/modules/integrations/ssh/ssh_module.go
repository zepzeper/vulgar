package ssh

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"os"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"

	lua "github.com/yuin/gopher-lua"
	"github.com/zepzeper/vulgar/internal/modules"
	"github.com/zepzeper/vulgar/internal/modules/util"
)

const (
	ModuleName        = "integrations.ssh"
	luaClientTypeName = "ssh_client"
	luaTunnelTypeName = "ssh_tunnel"
)

type sshClient struct {
	client *ssh.Client
	host   string
	user   string
	mu     sync.Mutex
	closed bool
}

type sshTunnel struct {
	listener   net.Listener
	client     *sshClient
	remoteHost string
	remotePort int
	done       chan struct{}
	mu         sync.Mutex
	closed     bool
}

var clientMethods = map[string]lua.LGFunction{
	"exec":     luaClientExec,
	"run":      luaClientRun,
	"upload":   luaClientUpload,
	"download": luaClientDownload,
	"close":    luaClientClose,
}

var tunnelMethods = map[string]lua.LGFunction{
	"close": luaTunnelClose,
	"port":  luaTunnelPort,
}

func registerTypes(L *lua.LState) {
	// Register client type
	mtClient := L.NewTypeMetatable(luaClientTypeName)
	L.SetField(mtClient, "__index", L.SetFuncs(L.NewTable(), clientMethods))
	L.SetField(mtClient, "__gc", L.NewFunction(clientGC))

	// Register tunnel type
	mtTunnel := L.NewTypeMetatable(luaTunnelTypeName)
	L.SetField(mtTunnel, "__index", L.SetFuncs(L.NewTable(), tunnelMethods))
	L.SetField(mtTunnel, "__gc", L.NewFunction(tunnelGC))
}

func checkClient(L *lua.LState, idx int) *sshClient {
	ud := L.CheckUserData(idx)
	if c, ok := ud.Value.(*sshClient); ok {
		return c
	}
	L.ArgError(idx, "ssh_client expected")
	return nil
}

func checkTunnel(L *lua.LState) *sshTunnel {
	ud := L.CheckUserData(1)
	if t, ok := ud.Value.(*sshTunnel); ok {
		return t
	}
	L.ArgError(1, "ssh_tunnel expected")
	return nil
}

func (c *sshClient) close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return
	}
	c.closed = true

	if c.client != nil {
		c.client.Close()
	}
}

func (t *sshTunnel) close() {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed {
		return
	}
	t.closed = true

	close(t.done)
	if t.listener != nil {
		t.listener.Close()
	}
}

func clientGC(L *lua.LState) int {
	ud := L.CheckUserData(1)
	if c, ok := ud.Value.(*sshClient); ok {
		c.close()
	}
	return 0
}

func tunnelGC(L *lua.LState) int {
	ud := L.CheckUserData(1)
	if t, ok := ud.Value.(*sshTunnel); ok {
		t.close()
	}
	return 0
}

// Usage: local client, err = ssh.connect({host = "example.com", port = 22, user = "user", password = "pass"})
// Usage: local client, err = ssh.connect({host = "example.com", port = 22, user = "user", key_file = "/path/to/key"})
func luaConnect(L *lua.LState) int {
	configTbl := L.CheckTable(1)

	host := getStringField(configTbl, "host", "")
	port := int(getNumberField(configTbl, "port", 22))
	user := getStringField(configTbl, "user", "")
	password := getStringField(configTbl, "password", "")
	keyFile := getStringField(configTbl, "key_file", "")
	timeout := int(getNumberField(configTbl, "timeout", 30))

	if host == "" {
		return util.PushError(L, "host is required")
	}
	if user == "" {
		return util.PushError(L, "user is required")
	}

	var authMethods []ssh.AuthMethod

	// Password auth
	if password != "" {
		authMethods = append(authMethods, ssh.Password(password))
	}

	// Key file auth
	if keyFile != "" {
		key, err := os.ReadFile(keyFile)
		if err != nil {
			return util.PushError(L, "failed to read key file: %v", err)
		}

		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			return util.PushError(L, "failed to parse private key: %v", err)
		}

		authMethods = append(authMethods, ssh.PublicKeys(signer))
	}

	if len(authMethods) == 0 {
		return util.PushError(L, "password or key_file is required")
	}

	config := &ssh.ClientConfig{
		User:            user,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         time.Duration(timeout) * time.Second,
	}

	addr := fmt.Sprintf("%s:%d", host, port)
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return util.PushError(L, "failed to connect: %v", err)
	}

	c := &sshClient{
		client: client,
		host:   host,
		user:   user,
	}

	ud := util.NewUserData(L, c, luaClientTypeName)
	return util.PushSuccess(L, ud)
}

// Usage: local output, err = ssh.exec(client, "ls -la")
func luaExec(L *lua.LState) int {
	if L.Get(1) == lua.LNil {
		return util.PushError(L, "client is nil")
	}

	c := checkClient(L, 1)
	cmd := L.CheckString(2)

	c.mu.Lock()
	if c.closed || c.client == nil {
		c.mu.Unlock()
		return util.PushError(L, "client is closed")
	}
	client := c.client
	c.mu.Unlock()

	session, err := client.NewSession()
	if err != nil {
		return util.PushError(L, "failed to create session: %v", err)
	}
	defer session.Close()

	output, err := session.CombinedOutput(cmd)
	if err != nil {
		if len(output) > 0 {
			L.Push(lua.LString(string(output)))
			L.Push(lua.LString(err.Error()))
			return 2
		}
		return util.PushError(L, "exec failed: %v", err)
	}

	L.Push(lua.LString(string(output)))
	L.Push(lua.LNil)
	return 2
}

// Usage: local code, stdout, stderr = ssh.run(client, "make build")
func luaRun(L *lua.LState) int {
	if L.Get(1) == lua.LNil {
		return util.PushError(L, "client is nil")
	}

	c := checkClient(L, 1)
	cmd := L.CheckString(2)

	c.mu.Lock()
	if c.closed || c.client == nil {
		c.mu.Unlock()
		L.Push(lua.LNumber(-1))
		L.Push(lua.LNil)
		L.Push(lua.LString("client is closed"))
		return 3
	}
	client := c.client
	c.mu.Unlock()

	session, err := client.NewSession()
	if err != nil {
		L.Push(lua.LNumber(-1))
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 3
	}
	defer session.Close()

	var stdout, stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr

	exitCode := 0
	if err := session.Run(cmd); err != nil {
		if exitErr, ok := err.(*ssh.ExitError); ok {
			exitCode = exitErr.ExitStatus()
		} else {
			L.Push(lua.LNumber(-1))
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 3
		}
	}

	L.Push(lua.LNumber(exitCode))
	L.Push(lua.LString(stdout.String()))
	L.Push(lua.LString(stderr.String()))
	return 3
}

// Usage: local session, err = ssh.shell(client)
func luaShell(L *lua.LState) int {
	if L.Get(1) == lua.LNil {
		return util.PushError(L, "client is nil")
	}

	// Shell sessions are complex - for now return error
	// A full implementation would need terminal handling
	return util.PushError(L, "interactive shell not yet implemented; use ssh.exec() instead")
}

// Usage: local tunnel, err = ssh.tunnel(client, local_port, remote_host, remote_port)
func luaTunnel(L *lua.LState) int {
	if L.Get(1) == lua.LNil {
		return util.PushError(L, "client is nil")
	}

	c := checkClient(L, 1)
	localPort := int(L.CheckNumber(2))
	remoteHost := L.CheckString(3)
	remotePort := int(L.CheckNumber(4))

	c.mu.Lock()
	if c.closed || c.client == nil {
		c.mu.Unlock()
		return util.PushError(L, "client is closed")
	}
	c.mu.Unlock()

	// Start local listener
	listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", localPort))
	if err != nil {
		return util.PushError(L, "failed to start local listener: %v", err)
	}

	t := &sshTunnel{
		listener:   listener,
		client:     c,
		remoteHost: remoteHost,
		remotePort: remotePort,
		done:       make(chan struct{}),
	}

	// Handle connections in background
	go func() {
		for {
			select {
			case <-t.done:
				return
			default:
			}

			conn, err := listener.Accept()
			if err != nil {
				select {
				case <-t.done:
					return
				default:
					continue
				}
			}

			go t.handleConnection(conn)
		}
	}()

	ud := util.NewUserData(L, t, luaTunnelTypeName)
	return util.PushSuccess(L, ud)
}

func (t *sshTunnel) handleConnection(localConn net.Conn) {
	defer localConn.Close()

	t.client.mu.Lock()
	if t.client.closed || t.client.client == nil {
		t.client.mu.Unlock()
		return
	}
	client := t.client.client
	t.client.mu.Unlock()

	remoteAddr := fmt.Sprintf("%s:%d", t.remoteHost, t.remotePort)
	remoteConn, err := client.Dial("tcp", remoteAddr)
	if err != nil {
		return
	}
	defer remoteConn.Close()

	// Bidirectional copy
	done := make(chan struct{}, 2)

	go func() {
		io.Copy(remoteConn, localConn)
		done <- struct{}{}
	}()

	go func() {
		io.Copy(localConn, remoteConn)
		done <- struct{}{}
	}()

	select {
	case <-done:
	case <-t.done:
	}
}

// Usage: local err = ssh.upload(client, local_path, remote_path)
func luaUpload(L *lua.LState) int {
	if L.Get(1) == lua.LNil {
		L.Push(lua.LString("client is nil"))
		return 1
	}

	c := checkClient(L, 1)
	localPath := L.CheckString(2)
	remotePath := L.CheckString(3)

	c.mu.Lock()
	if c.closed || c.client == nil {
		c.mu.Unlock()
		L.Push(lua.LString("client is closed"))
		return 1
	}
	client := c.client
	c.mu.Unlock()

	// Read local file
	data, err := os.ReadFile(localPath)
	if err != nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}

	// Create session and use cat to write file
	session, err := client.NewSession()
	if err != nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}
	defer session.Close()

	stdin, err := session.StdinPipe()
	if err != nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}

	go func() {
		defer stdin.Close()
		stdin.Write(data)
	}()

	if err := session.Run(fmt.Sprintf("cat > %s", remotePath)); err != nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}

	L.Push(lua.LNil)
	return 1
}

// Usage: local err = ssh.download(client, remote_path, local_path)
func luaDownload(L *lua.LState) int {
	if L.Get(1) == lua.LNil {
		L.Push(lua.LString("client is nil"))
		return 1
	}

	c := checkClient(L, 1)
	remotePath := L.CheckString(2)
	localPath := L.CheckString(3)

	c.mu.Lock()
	if c.closed || c.client == nil {
		c.mu.Unlock()
		L.Push(lua.LString("client is closed"))
		return 1
	}
	client := c.client
	c.mu.Unlock()

	// Create session and use cat to read file
	session, err := client.NewSession()
	if err != nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}
	defer session.Close()

	output, err := session.Output(fmt.Sprintf("cat %s", remotePath))
	if err != nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}

	if err := os.WriteFile(localPath, output, 0644); err != nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}

	L.Push(lua.LNil)
	return 1
}

// Usage: local err = ssh.close(client)
func luaClose(L *lua.LState) int {
	if L.Get(1) == lua.LNil {
		L.Push(lua.LNil)
		return 1
	}

	c := checkClient(L, 1)
	c.close()

	L.Push(lua.LNil)
	return 1
}

// Usage: local output, err = client:exec("ls -la")
func luaClientExec(L *lua.LState) int {
	c := checkClient(L, 1)
	cmd := L.CheckString(2)

	c.mu.Lock()
	if c.closed || c.client == nil {
		c.mu.Unlock()
		return util.PushError(L, "client is closed")
	}
	client := c.client
	c.mu.Unlock()

	session, err := client.NewSession()
	if err != nil {
		return util.PushError(L, "failed to create session: %v", err)
	}
	defer session.Close()

	output, err := session.CombinedOutput(cmd)
	if err != nil {
		if len(output) > 0 {
			L.Push(lua.LString(string(output)))
			L.Push(lua.LString(err.Error()))
			return 2
		}
		return util.PushError(L, "exec failed: %v", err)
	}

	L.Push(lua.LString(string(output)))
	L.Push(lua.LNil)
	return 2
}

// Usage: local code, stdout, stderr = client:run("make build")
func luaClientRun(L *lua.LState) int {
	c := checkClient(L, 1)
	cmd := L.CheckString(2)

	c.mu.Lock()
	if c.closed || c.client == nil {
		c.mu.Unlock()
		L.Push(lua.LNumber(-1))
		L.Push(lua.LNil)
		L.Push(lua.LString("client is closed"))
		return 3
	}
	client := c.client
	c.mu.Unlock()

	session, err := client.NewSession()
	if err != nil {
		L.Push(lua.LNumber(-1))
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 3
	}
	defer session.Close()

	var stdout, stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr

	exitCode := 0
	if err := session.Run(cmd); err != nil {
		if exitErr, ok := err.(*ssh.ExitError); ok {
			exitCode = exitErr.ExitStatus()
		} else {
			L.Push(lua.LNumber(-1))
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 3
		}
	}

	L.Push(lua.LNumber(exitCode))
	L.Push(lua.LString(stdout.String()))
	L.Push(lua.LString(stderr.String()))
	return 3
}

// Usage: local err = client:upload(local_path, remote_path)
func luaClientUpload(L *lua.LState) int {
	c := checkClient(L, 1)
	localPath := L.CheckString(2)
	remotePath := L.CheckString(3)

	c.mu.Lock()
	if c.closed || c.client == nil {
		c.mu.Unlock()
		L.Push(lua.LString("client is closed"))
		return 1
	}
	client := c.client
	c.mu.Unlock()

	data, err := os.ReadFile(localPath)
	if err != nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}

	session, err := client.NewSession()
	if err != nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}
	defer session.Close()

	stdin, err := session.StdinPipe()
	if err != nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}

	go func() {
		defer stdin.Close()
		stdin.Write(data)
	}()

	if err := session.Run(fmt.Sprintf("cat > %s", remotePath)); err != nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}

	L.Push(lua.LNil)
	return 1
}

// Usage: local err = client:download(remote_path, local_path)
func luaClientDownload(L *lua.LState) int {
	c := checkClient(L, 1)
	remotePath := L.CheckString(2)
	localPath := L.CheckString(3)

	c.mu.Lock()
	if c.closed || c.client == nil {
		c.mu.Unlock()
		L.Push(lua.LString("client is closed"))
		return 1
	}
	client := c.client
	c.mu.Unlock()

	session, err := client.NewSession()
	if err != nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}
	defer session.Close()

	output, err := session.Output(fmt.Sprintf("cat %s", remotePath))
	if err != nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}

	if err := os.WriteFile(localPath, output, 0644); err != nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}

	L.Push(lua.LNil)
	return 1
}

// Usage: client:close()
func luaClientClose(L *lua.LState) int {
	c := checkClient(L, 1)
	c.close()
	L.Push(lua.LNil)
	return 1
}

// Usage: tunnel:close()
func luaTunnelClose(L *lua.LState) int {
	t := checkTunnel(L)
	t.close()
	L.Push(lua.LNil)
	return 1
}

// Usage: local port = tunnel:port()
func luaTunnelPort(L *lua.LState) int {
	t := checkTunnel(L)
	if t.listener != nil {
		addr := t.listener.Addr().(*net.TCPAddr)
		L.Push(lua.LNumber(addr.Port))
		return 1
	}
	L.Push(lua.LNumber(0))
	return 1
}

// Helper functions
func getStringField(tbl *lua.LTable, key, defaultVal string) string {
	val := tbl.RawGetString(key)
	if str, ok := val.(lua.LString); ok {
		return string(str)
	}
	return defaultVal
}

func getNumberField(tbl *lua.LTable, key string, defaultVal float64) float64 {
	val := tbl.RawGetString(key)
	if num, ok := val.(lua.LNumber); ok {
		return float64(num)
	}
	return defaultVal
}

var exports = map[string]lua.LGFunction{
	"connect":  luaConnect,
	"exec":     luaExec,
	"run":      luaRun,
	"shell":    luaShell,
	"tunnel":   luaTunnel,
	"upload":   luaUpload,
	"download": luaDownload,
	"close":    luaClose,
}

func Loader(L *lua.LState) int {
	registerTypes(L)
	mod := L.SetFuncs(L.NewTable(), exports)
	L.Push(mod)
	return 1
}

func init() {
	modules.Register(ModuleName, Loader)
}
