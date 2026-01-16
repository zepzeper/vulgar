package smtp

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"strings"
	"sync"

	lua "github.com/yuin/gopher-lua"
	"github.com/zepzeper/vulgar/internal/modules"
	"github.com/zepzeper/vulgar/internal/modules/util"
)

const ModuleName = "integrations.smtp"
const luaSMTPClientTypeName = "smtp_client"

type smtpClient struct {
	host     string
	port     int
	user     string
	password string
	useTLS   bool
	auth     smtp.Auth
	mu       sync.Mutex
	closed   bool
}

var clientMethods = map[string]lua.LGFunction{
	"send":     luaClientSend,
	"send_raw": luaClientSendRaw,
	"close":    luaClientClose,
}

func registerSMTPClientType(L *lua.LState) {
	mt := L.NewTypeMetatable(luaSMTPClientTypeName)
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), clientMethods))
	L.SetField(mt, "__gc", L.NewFunction(smtpClientGC))
}

func checkSMTPClient(L *lua.LState, idx int) *smtpClient {
	ud := L.CheckUserData(idx)
	if v, ok := ud.Value.(*smtpClient); ok {
		return v
	}
	L.ArgError(idx, "smtp_client expected")
	return nil
}

func smtpClientGC(L *lua.LState) int {
	ud := L.CheckUserData(1)
	if client, ok := ud.Value.(*smtpClient); ok {
		client.close()
	}
	return 0
}

func (c *smtpClient) close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return
	}
	c.closed = true
	// SMTP connections are stateless per-send, no persistent connection to close
}

func (c *smtpClient) addr() string {
	return fmt.Sprintf("%s:%d", c.host, c.port)
}

func (c *smtpClient) sendMail(from string, to []string, msg []byte) error {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return fmt.Errorf("client is closed")
	}
	c.mu.Unlock()

	addr := c.addr()

	if c.useTLS {
		return c.sendMailTLS(from, to, msg)
	}

	return smtp.SendMail(addr, c.auth, from, to, msg)
}

func (c *smtpClient) sendMailTLS(from string, to []string, msg []byte) error {
	addr := c.addr()

	// Connect to the server
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	// Create SMTP client
	client, err := smtp.NewClient(conn, c.host)
	if err != nil {
		conn.Close()
		return fmt.Errorf("failed to create client: %w", err)
	}
	defer client.Close()

	// Start TLS
	tlsConfig := &tls.Config{
		ServerName: c.host,
	}
	if err := client.StartTLS(tlsConfig); err != nil {
		return fmt.Errorf("failed to start TLS: %w", err)
	}

	// Authenticate if credentials provided
	if c.auth != nil {
		if err := client.Auth(c.auth); err != nil {
			return fmt.Errorf("authentication failed: %w", err)
		}
	}

	// Set sender
	if err := client.Mail(from); err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}

	// Set recipients
	for _, rcpt := range to {
		if err := client.Rcpt(rcpt); err != nil {
			return fmt.Errorf("failed to set recipient %s: %w", rcpt, err)
		}
	}

	// Send message body
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to open data writer: %w", err)
	}

	_, err = w.Write(msg)
	if err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	err = w.Close()
	if err != nil {
		return fmt.Errorf("failed to close data writer: %w", err)
	}

	return client.Quit()
}

func buildMessage(from, subject, body string, to, cc, bcc []string, html bool) []byte {
	var msg strings.Builder

	msg.WriteString(fmt.Sprintf("From: %s\r\n", from))
	msg.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(to, ", ")))

	if len(cc) > 0 {
		msg.WriteString(fmt.Sprintf("Cc: %s\r\n", strings.Join(cc, ", ")))
	}

	msg.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))

	if html {
		msg.WriteString("MIME-Version: 1.0\r\n")
		msg.WriteString("Content-Type: text/html; charset=\"UTF-8\"\r\n")
	} else {
		msg.WriteString("Content-Type: text/plain; charset=\"UTF-8\"\r\n")
	}

	msg.WriteString("\r\n")
	msg.WriteString(body)

	return []byte(msg.String())
}

func parseStringArray(L *lua.LState, val lua.LValue) []string {
	var result []string
	if tbl, ok := val.(*lua.LTable); ok {
		tbl.ForEach(func(_, v lua.LValue) {
			if s, ok := v.(lua.LString); ok {
				result = append(result, string(s))
			}
		})
	}
	return result
}

// Usage: local client, err = smtp.connect({host = "smtp.example.com", port = 587, user = "user", password = "pass", tls = true})
func luaConnect(L *lua.LState) int {
	opts := L.OptTable(1, nil)
	if opts == nil {
		return util.PushError(L, "config table required")
	}

	host := ""
	port := 587
	user := ""
	password := ""
	useTLS := true

	if v := L.GetField(opts, "host"); v != lua.LNil {
		host = lua.LVAsString(v)
	}
	if v := L.GetField(opts, "port"); v != lua.LNil {
		if n, ok := v.(lua.LNumber); ok {
			port = int(n)
		}
	}
	if v := L.GetField(opts, "user"); v != lua.LNil {
		user = lua.LVAsString(v)
	}
	if v := L.GetField(opts, "password"); v != lua.LNil {
		password = lua.LVAsString(v)
	}
	if v := L.GetField(opts, "tls"); v != lua.LNil {
		useTLS = lua.LVAsBool(v)
	}

	if host == "" {
		return util.PushError(L, "host is required")
	}

	// Create auth if credentials provided
	var auth smtp.Auth
	if user != "" && password != "" {
		auth = smtp.PlainAuth("", user, password, host)
	}

	// Test connection by dialing
	addr := net.JoinHostPort(host, fmt.Sprintf("%d", port))
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return util.PushError(L, "failed to connect to %s: %v", addr, err)
	}
	conn.Close()

	client := &smtpClient{
		host:     host,
		port:     port,
		user:     user,
		password: password,
		useTLS:   useTLS,
		auth:     auth,
	}

	ud := L.NewUserData()
	ud.Value = client
	L.SetMetatable(ud, L.GetTypeMetatable(luaSMTPClientTypeName))

	return util.PushSuccess(L, ud)
}

// Usage: local err = smtp.send(client, {from = "me@example.com", to = {"user@example.com"}, subject = "Hello", body = "World"})
func luaSend(L *lua.LState) int {
	if L.Get(1) == lua.LNil {
		L.Push(lua.LString("client is required"))
		return 1
	}

	client := checkSMTPClient(L, 1)
	if client == nil {
		L.Push(lua.LString("invalid client"))
		return 1
	}

	opts := L.OptTable(2, nil)
	if opts == nil {
		L.Push(lua.LString("email options required"))
		return 1
	}

	from := ""
	subject := ""
	body := ""
	html := false
	var to, cc, bcc []string

	if v := L.GetField(opts, "from"); v != lua.LNil {
		from = lua.LVAsString(v)
	}
	if v := L.GetField(opts, "to"); v != lua.LNil {
		to = parseStringArray(L, v)
	}
	if v := L.GetField(opts, "cc"); v != lua.LNil {
		cc = parseStringArray(L, v)
	}
	if v := L.GetField(opts, "bcc"); v != lua.LNil {
		bcc = parseStringArray(L, v)
	}
	if v := L.GetField(opts, "subject"); v != lua.LNil {
		subject = lua.LVAsString(v)
	}
	if v := L.GetField(opts, "body"); v != lua.LNil {
		body = lua.LVAsString(v)
	}
	if v := L.GetField(opts, "html"); v != lua.LNil {
		html = lua.LVAsBool(v)
	}

	if from == "" {
		L.Push(lua.LString("from is required"))
		return 1
	}
	if len(to) == 0 {
		L.Push(lua.LString("to is required"))
		return 1
	}

	// Build recipient list (to + cc + bcc)
	allRecipients := append([]string{}, to...)
	allRecipients = append(allRecipients, cc...)
	allRecipients = append(allRecipients, bcc...)

	msg := buildMessage(from, subject, body, to, cc, bcc, html)

	if err := client.sendMail(from, allRecipients, msg); err != nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}

	L.Push(lua.LNil)
	return 1
}

// Usage: local err = smtp.send_raw(client, from, to, raw_message)
func luaSendRaw(L *lua.LState) int {
	if L.Get(1) == lua.LNil {
		L.Push(lua.LString("client is required"))
		return 1
	}

	client := checkSMTPClient(L, 1)
	if client == nil {
		L.Push(lua.LString("invalid client"))
		return 1
	}

	from := L.CheckString(2)
	toVal := L.Get(3)
	rawMsg := L.CheckString(4)

	var to []string
	if s, ok := toVal.(lua.LString); ok {
		to = []string{string(s)}
	} else if tbl, ok := toVal.(*lua.LTable); ok {
		to = parseStringArray(L, tbl)
	} else {
		L.Push(lua.LString("to must be string or table"))
		return 1
	}

	if err := client.sendMail(from, to, []byte(rawMsg)); err != nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}

	L.Push(lua.LNil)
	return 1
}

// Usage: local err = smtp.close(client)
func luaClose(L *lua.LState) int {
	if L.Get(1) == lua.LNil {
		L.Push(lua.LNil)
		return 1
	}

	client := checkSMTPClient(L, 1)
	if client != nil {
		client.close()
	}

	L.Push(lua.LNil)
	return 1
}

// Usage: client:send({from = "...", to = {...}, subject = "...", body = "..."})
func luaClientSend(L *lua.LState) int {
	client := checkSMTPClient(L, 1)
	if client == nil {
		L.Push(lua.LString("invalid client"))
		return 1
	}

	opts := L.OptTable(2, nil)
	if opts == nil {
		L.Push(lua.LString("email options required"))
		return 1
	}

	from := ""
	subject := ""
	body := ""
	html := false
	var to, cc, bcc []string

	if v := L.GetField(opts, "from"); v != lua.LNil {
		from = lua.LVAsString(v)
	}
	if v := L.GetField(opts, "to"); v != lua.LNil {
		to = parseStringArray(L, v)
	}
	if v := L.GetField(opts, "cc"); v != lua.LNil {
		cc = parseStringArray(L, v)
	}
	if v := L.GetField(opts, "bcc"); v != lua.LNil {
		bcc = parseStringArray(L, v)
	}
	if v := L.GetField(opts, "subject"); v != lua.LNil {
		subject = lua.LVAsString(v)
	}
	if v := L.GetField(opts, "body"); v != lua.LNil {
		body = lua.LVAsString(v)
	}
	if v := L.GetField(opts, "html"); v != lua.LNil {
		html = lua.LVAsBool(v)
	}

	if from == "" {
		L.Push(lua.LString("from is required"))
		return 1
	}
	if len(to) == 0 {
		L.Push(lua.LString("to is required"))
		return 1
	}

	allRecipients := append([]string{}, to...)
	allRecipients = append(allRecipients, cc...)
	allRecipients = append(allRecipients, bcc...)

	msg := buildMessage(from, subject, body, to, cc, bcc, html)

	if err := client.sendMail(from, allRecipients, msg); err != nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}

	L.Push(lua.LNil)
	return 1
}

// Usage: client:send_raw(from, to, raw_message)
func luaClientSendRaw(L *lua.LState) int {
	client := checkSMTPClient(L, 1)
	if client == nil {
		L.Push(lua.LString("invalid client"))
		return 1
	}

	from := L.CheckString(2)
	toVal := L.Get(3)
	rawMsg := L.CheckString(4)

	var to []string
	if s, ok := toVal.(lua.LString); ok {
		to = []string{string(s)}
	} else if tbl, ok := toVal.(*lua.LTable); ok {
		to = parseStringArray(L, tbl)
	} else {
		L.Push(lua.LString("to must be string or table"))
		return 1
	}

	if err := client.sendMail(from, to, []byte(rawMsg)); err != nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}

	L.Push(lua.LNil)
	return 1
}

// Usage: client:close()
func luaClientClose(L *lua.LState) int {
	client := checkSMTPClient(L, 1)
	if client != nil {
		client.close()
	}
	L.Push(lua.LNil)
	return 1
}

var exports = map[string]lua.LGFunction{
	"connect":  luaConnect,
	"send":     luaSend,
	"send_raw": luaSendRaw,
	"close":    luaClose,
}

func Loader(L *lua.LState) int {
	registerSMTPClientType(L)
	mod := L.SetFuncs(L.NewTable(), exports)
	L.Push(mod)
	return 1
}

func init() {
	modules.Register(ModuleName, Loader)
}
