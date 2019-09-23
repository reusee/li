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

func (_ Provide) LSP(
	on On,
	j AppendJournal,
) Init2 {

	endpoints := make(map[Language]*LSPEndpoint)

	// start lsp process
	on(EvBufferLanguageChanged, func(
		buffer *Buffer,
		langs [2]Language,
		configDir ConfigDir,
	) {
		j("%s changed language from %v to %v", buffer.Path, langs[0], langs[1])

		lang := langs[1]
		if _, ok := endpoints[lang]; ok {
			return
		}

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
						delete(endpoints, endpoint.Language)
					},
				)
				endpoints[lang] = endpoint

				var ret any
				ce(endpoint.Req("initialize", M{
					"processId": syscall.Getpid(),
					"rootUri":   "li://buffers/",
				}).Wait(&ret))

				j("language server for %s started: %v", lang, ret)
			}

		}
	})

	return nil
}

type LSPEndpoint struct {
	*sync.Mutex
	*sync.Cond

	Language Language
	RW       io.ReadWriter
	OnErr    func(error)

	calls     []*LSPCall
	nextReqID int64
}

type LSPCall struct {
	endpoint *LSPEndpoint
	id       int64
	bs       []byte
	err      error
}

func NewLSPEndpoint(
	rw io.ReadWriter,
	lang Language,
	onErr func(error),
) *LSPEndpoint {
	l := new(sync.Mutex)
	cond := sync.NewCond(l)
	endpoint := &LSPEndpoint{
		Mutex:    l,
		Cond:     cond,
		Language: lang,
		RW:       rw,
		OnErr:    onErr,
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
				ID *int64
			}
			if err = json.Unmarshal(bs, &data); err != nil {
				break
			}
			if data.ID != nil {
				l.Lock()
				for i := 0; i < len(l.calls); i++ {
					call := l.calls[i]
					if call.id == *data.ID {
						call.bs = bs
						l.calls = append(l.calls[:i], l.calls[i+1:]...)
					}
				}
				l.Unlock()
				l.Broadcast()
			}
		}

	}
	if err != nil {
		if l.OnErr != nil {
			l.OnErr(err)
		}
	}
}

func (c *LSPCall) Wait(target any) error {
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
