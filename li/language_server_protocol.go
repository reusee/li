package li

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
)

/*
TODO gopls
code action
completion
definition
document format
document highlight
document link
document symbol
hover
signature help
type definition
*/

type LanguageServerProtocolConfig struct {
	Enable bool
}

func (_ Provide) LSP(
	on On,
	j AppendJournal,
	getConfig GetConfig,
) Init2 {

	var c struct {
		LanguageServerProtocol LanguageServerProtocolConfig
	}
	ce(getConfig(&c))
	config := c.LanguageServerProtocol

	if !config.Enable {
		return nil
	}

	endpoints := make(map[string]*LSPEndpoint)

	// start lsp process
	on(EvBufferLanguageChanged, func(
		buffer *Buffer,
		langs [2]Language,
		configDir ConfigDir,
		linkedOne LinkedOne,
	) {
		j("%s changed language from %v to %v", buffer.Path, langs[0], langs[1])

		if _, ok := endpoints[buffer.AbsDir]; ok {
			return
		}

		lang := langs[1]
		switch lang {

		case LanguageGo:
			exePath, err := exec.LookPath("gopls")
			if err != nil {
				j("gopls executable not found in PATH")
			} else {
				cmd := exec.Command(
					exePath,
					"-logfile", filepath.Join(string(configDir), "gopls.log"),
					"-rpc.trace",
					"-v",
				)
				w, err := cmd.StdinPipe()
				ce(err)
				r, err := cmd.StdoutPipe()
				ce(err)
				ce(cmd.Start())

				var endpoint *LSPEndpoint
				endpoint = NewLSPEndpoint(
					struct {
						io.Writer
						io.Reader
					}{w, r},
					lang,
					func(err error) {
						j("language server for %s error: %v", endpoint.Language, err)
						delete(endpoints, buffer.AbsDir)
					},
					func(format string, args ...any) {
						j(format, args...)
					},
				)
				endpoints[buffer.AbsDir] = endpoint

				var ret any
				ce(endpoint.Req("initialize", M{
					"processId": syscall.Getpid(),
					"rootUri":   buffer.AbsDir,
				}).Unmarshal(&ret))
				endpoint.Notify("initialized", M{})

				var moment *Moment
				linkedOne(buffer, &moment)
				if moment != nil {
					endpoint.Notify("textDocument/didOpen", M{
						"textDocument": M{
							"uri":        buffer.AbsPath,
							"languageId": "go",
							"version":    moment.ID,
							"text":       moment.GetContent(),
						},
					})
				}

				j("language server for %s started:\n%s", lang, toJSON(ret))
			}

		}
	})

	// format
	on(EvMomentSwitched, func(
		buffer *Buffer,
	) {
		endpoint, ok := endpoints[buffer.AbsDir]
		if !ok {
			return
		}
		endpoint.Req("textDocument/formatting", M{
			"textDocument": M{
				"uri": buffer.AbsPath,
			},
			"options": M{
				"foo": "bar",
			},
		}).Then(func(c *LSPCall) {
			var ret any
			ce(c.Unmarshal(&ret))
			j("format %s", toJSON(ret))
			//TODO apply change
		})
	})

	// sync change
	on(EvMomentSwitched, func(
		buffer *Buffer,
		moments [2]*Moment,
	) {
		//TODO
	})

	return nil
}

type LSPEndpoint struct {
	*sync.Mutex
	*sync.Cond

	Language Language
	RW       io.ReadWriter
	OnErr    func(error)
	OnLog    func(format string, args ...any)

	calls     []*LSPCall
	nextReqID int64
}

type LSPCall struct {
	endpoint *LSPEndpoint
	id       int64
	bs       []byte
	err      error
	then     func(*LSPCall)
}

func NewLSPEndpoint(
	rw io.ReadWriter,
	lang Language,
	onErr func(error),
	onLog func(format string, args ...any),
) *LSPEndpoint {
	l := new(sync.Mutex)
	cond := sync.NewCond(l)
	endpoint := &LSPEndpoint{
		Mutex:    l,
		Cond:     cond,
		Language: lang,
		RW:       rw,
		OnErr:    onErr,
		OnLog:    onLog,
	}
	go endpoint.startHandler()
	return endpoint
}

func (l *LSPEndpoint) Req(method string, params M) *LSPCall {
	l.Lock()
	defer l.Unlock()
	id := l.nextReqID
	l.nextReqID++
	data := M{
		"jsonrpc": "2.0",
		"id":      id,
		"method":  method,
		"params":  params,
	}
	call := &LSPCall{
		endpoint: l,
		id:       id,
	}
	buf := new(bytes.Buffer)
	err := json.NewEncoder(buf).Encode(data)
	ce(err)
	bs := buf.Bytes()
	if _, err := io.WriteString(l.RW, fmt.Sprintf("Content-Length: %d\r\n\r\n", len(bs))); err != nil {
		if l.OnErr != nil {
			l.OnErr(err)
		}
		call.err = err
		return call
	}
	if _, err := l.RW.Write(bs); err != nil {
		if l.OnErr != nil {
			l.OnErr(err)
		}
		call.err = err
		return call
	}
	l.calls = append(l.calls, call)
	return call
}

func (l *LSPEndpoint) Notify(method string, params M) {
	l.Lock()
	defer l.Unlock()
	data := M{
		"jsonrpc": "2.0",
		"method":  method,
		"params":  params,
	}
	buf := new(bytes.Buffer)
	err := json.NewEncoder(buf).Encode(data)
	ce(err)
	bs := buf.Bytes()
	if _, err := io.WriteString(l.RW, fmt.Sprintf("Content-Length: %d\r\n\r\n", len(bs))); err != nil {
		if l.OnErr != nil {
			l.OnErr(err)
		}
		return
	}
	if _, err := l.RW.Write(bs); err != nil {
		if l.OnErr != nil {
			l.OnErr(err)
		}
		return
	}
}

func (l *LSPEndpoint) startHandler() {
	r := bufio.NewReader(l.RW)
	var err error
	var contentLen int
	for {
		header, err := r.ReadString('\n')
		if err != nil {
			break
		}
		header = strings.TrimSpace(header)

		if strings.HasPrefix(header, "Content-Length:") {
			contentLen, err = strconv.Atoi(
				strings.TrimSpace(header[len("Content-Length:"):]),
			)
			if err != nil {
				pt("%v\n", err)
				break
			}

		} else if len(header) > 0 {
			continue

		} else if len(header) == 0 {
			bs := make([]byte, contentLen)
			if _, err = io.ReadFull(r, bs); err != nil {
				break
			}
			var data struct {
				ID     *int64
				Method string
				Params struct {
					Type    LSPMessageType
					Message string
				}
			}
			if err = json.Unmarshal(bs, &data); err != nil {
				break
			}

			if data.ID != nil {
				// response
				l.Lock()
				for i := 0; i < len(l.calls); i++ {
					call := l.calls[i]
					if call.id == *data.ID {
						call.bs = bs
						l.calls = append(l.calls[:i], l.calls[i+1:]...)
						if call.then != nil {
							go call.then(call)
						}
					}
				}
				l.Unlock()
				l.Broadcast()

			} else if data.Method == "window/logMessage" && data.Params.Type <= LSPWarning {
				if l.OnLog != nil {
					l.OnLog("%s - %s: %s", l.Language, data.Params.Type, data.Params.Message)
				}

			}

		}

	}
	if err != nil {
		if l.OnErr != nil {
			l.OnErr(err)
		}
	}
}

type LSPMessageType uint8

const (
	LSPError LSPMessageType = iota + 1
	LSPWarning
	LSPInfo
	LSPLog
)

func (c *LSPCall) Unmarshal(target any) error {
	if c.err != nil {
		return c.err
	}
	c.endpoint.Lock()
	for c.bs == nil {
		c.endpoint.Wait()
	}
	c.endpoint.Unlock()
	if err := json.Unmarshal(c.bs, target); err != nil {
		return err
	}
	return nil
}

func (c *LSPCall) Then(fn func(*LSPCall)) {
	c.endpoint.Lock()
	defer c.endpoint.Unlock()
	if c.bs != nil {
		go fn(c)
	} else {
		c.then = fn
	}
}
