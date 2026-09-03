package k8smodule

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/gofiber/websocket/v2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"
)

// TerminalResizeMessage describes terminal dimension resize events from the frontend.
type TerminalResizeMessage struct {
	Type string `json:"type"`
	Cols uint16 `json:"cols"`
	Rows uint16 `json:"rows"`
}

// wsTerminalSizeQueue implements remotecommand.TerminalSizeQueue.
type wsTerminalSizeQueue struct {
	resizeChan chan remotecommand.TerminalSize
}

func (q *wsTerminalSizeQueue) Next() *remotecommand.TerminalSize {
	size, ok := <-q.resizeChan
	if !ok {
		return nil
	}
	return &size
}

// wsStreamHandler bridges a fiber WebSocket connection with remotecommand stdin/stdout.
type wsStreamHandler struct {
	conn       *websocket.Conn
	resizeChan chan remotecommand.TerminalSize
	readBuf    bytes.Buffer
	readMutex  sync.Mutex
	writeMutex sync.Mutex
	stopChan   chan struct{}
}

func newWSStreamHandler(conn *websocket.Conn) *wsStreamHandler {
	return &wsStreamHandler{
		conn:       conn,
		resizeChan: make(chan remotecommand.TerminalSize, 10),
		stopChan:   make(chan struct{}),
	}
}

// Read implements io.Reader for container stdin.
func (h *wsStreamHandler) Read(p []byte) (int, error) {
	for {
		h.readMutex.Lock()
		if h.readBuf.Len() > 0 {
			n, err := h.readBuf.Read(p)
			h.readMutex.Unlock()
			return n, err
		}
		h.readMutex.Unlock()

		msgType, msg, err := h.conn.ReadMessage()
		if err != nil {
			return 0, err
		}

		if msgType == websocket.TextMessage || msgType == websocket.BinaryMessage {
			// Check if this is a resize JSON event
			var resizeMsg TerminalResizeMessage
			if err := json.Unmarshal(msg, &resizeMsg); err == nil && resizeMsg.Type == "resize" && resizeMsg.Cols > 0 && resizeMsg.Rows > 0 {
				select {
				case h.resizeChan <- remotecommand.TerminalSize{Width: resizeMsg.Cols, Height: resizeMsg.Rows}:
				default:
				}
				continue
			}

			// Otherwise, it's keystroke / stdin data
			h.readMutex.Lock()
			h.readBuf.Write(msg)
			n, readErr := h.readBuf.Read(p)
			h.readMutex.Unlock()
			return n, readErr
		}
	}
}

// Write implements io.Writer for container stdout/stderr.
func (h *wsStreamHandler) Write(p []byte) (int, error) {
	h.writeMutex.Lock()
	defer h.writeMutex.Unlock()
	err := h.conn.WriteMessage(websocket.TextMessage, p)
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

// ExecPodTerminal establishes an interactive shell connection inside a pod container.
func (s *K8sService) ExecPodTerminal(ctx context.Context, conn *websocket.Conn, namespace, podName, container, shell string) error {
	if namespace == "" {
		namespace = "default"
	}
	if shell == "" {
		shell = "sh"
	}

	streamHandler := newWSStreamHandler(conn)
	defer close(streamHandler.resizeChan)

	// Send welcoming prompt
	welcome := fmt.Sprintf("\r\n\x1b[38;2;16;185;129m[KubeEnv Web Terminal]\x1b[0m Connected to pod \x1b[1m%s\x1b[0m (container: %s, ns: %s)\r\n\r\n", podName, container, namespace)
	_ = conn.WriteMessage(websocket.TextMessage, []byte(welcome))

	if !s.clientMgr.Connected || s.clientMgr.Clientset == nil || s.clientMgr.Config == nil {
		// Run interactive simulated shell for offline / demo mode
		return s.runSimulatedShell(conn, streamHandler, namespace, podName, container)
	}

	// Build pod exec request
	req := s.clientMgr.Clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(podName).
		Namespace(namespace).
		SubResource("exec")

	execOpts := &corev1.PodExecOptions{
		Container: container,
		Command:   []string{shell},
		Stdin:     true,
		Stdout:    true,
		Stderr:    true,
		TTY:       true,
	}

	req.VersionedParams(execOpts, scheme.ParameterCodec)

	executor, err := remotecommand.NewSPDYExecutor(s.clientMgr.Config, "POST", req.URL())
	if err != nil {
		errMsg := fmt.Sprintf("\r\n\x1b[31mFailed to initialize SPDY executor for %s: %v\x1b[0m\r\n", shell, err)
		_ = conn.WriteMessage(websocket.TextMessage, []byte(errMsg))
		return err
	}

	sizeQueue := &wsTerminalSizeQueue{resizeChan: streamHandler.resizeChan}
	select {
	case streamHandler.resizeChan <- remotecommand.TerminalSize{Width: 120, Height: 35}:
	default:
	}

	streamErr := executor.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdin:             streamHandler,
		Stdout:            streamHandler,
		Stderr:            streamHandler,
		Tty:               true,
		TerminalSizeQueue: sizeQueue,
	})

	if streamErr != nil {
		_ = conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("\r\n\x1b[33mSession ended: %v\x1b[0m\r\n", streamErr)))
	} else {
		_ = conn.WriteMessage(websocket.TextMessage, []byte("\r\n\x1b[32mSession closed normally.\x1b[0m\r\n"))
	}

	return streamErr
}

// runSimulatedShell provides an interactive simulated shell session when running in demo/offline mode.
func (s *K8sService) runSimulatedShell(conn *websocket.Conn, handler *wsStreamHandler, namespace, pod, container string) error {
	prompt := fmt.Sprintf("\x1b[32mroot@%s\x1b[0m:\x1b[34m/app\x1b[0m# ", pod)
	_ = conn.WriteMessage(websocket.TextMessage, []byte(prompt))

	inputBuf := make([]byte, 256)
	cmdLine := ""

	for {
		n, err := handler.Read(inputBuf)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		raw := string(inputBuf[:n])

		for _, ch := range raw {
			if ch == '\r' || ch == '\n' {
				// Execute command
				_ = conn.WriteMessage(websocket.TextMessage, []byte("\r\n"))
				s.executeSimulatedCommand(conn, cmdLine, namespace, pod, container)
				cmdLine = ""
				_ = conn.WriteMessage(websocket.TextMessage, []byte(prompt))
			} else if ch == 127 || ch == '\b' {
				// Backspace
				if len(cmdLine) > 0 {
					cmdLine = cmdLine[:len(cmdLine)-1]
					_ = conn.WriteMessage(websocket.TextMessage, []byte("\b \b"))
				}
			} else if ch == 3 {
				// Ctrl+C
				_ = conn.WriteMessage(websocket.TextMessage, []byte("^C\r\n"))
				cmdLine = ""
				_ = conn.WriteMessage(websocket.TextMessage, []byte(prompt))
			} else if ch >= 32 && ch <= 126 {
				cmdLine += string(ch)
				_ = conn.WriteMessage(websocket.TextMessage, []byte(string(ch)))
			}
		}
	}
}

func (s *K8sService) executeSimulatedCommand(conn *websocket.Conn, cmd, namespace, pod, container string) {
	cmd = bytes.NewBufferString(cmd).String()
	switch cmd {
	case "help":
		_ = conn.WriteMessage(websocket.TextMessage, []byte("Available commands: help, ls, pwd, uname, env, ps, date, whoami, exit\r\n"))
	case "pwd":
		_ = conn.WriteMessage(websocket.TextMessage, []byte("/app\r\n"))
	case "whoami":
		_ = conn.WriteMessage(websocket.TextMessage, []byte("root\r\n"))
	case "date":
		_ = conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("%s\r\n", time.Now().Format(time.RFC1123))))
	case "uname", "uname -a":
		_ = conn.WriteMessage(websocket.TextMessage, []byte("Linux k8s-container 6.6.0-k8s #1 SMP PREEMPT x86_64 GNU/Linux\r\n"))
	case "ls", "ls -la":
		_ = conn.WriteMessage(websocket.TextMessage, []byte("total 48\r\ndrwxr-xr-x 1 root root 4096 Jan 01 00:00 .\r\ndrwxr-xr-x 1 root root 4096 Jan 01 00:00 ..\r\n-rwxr-xr-x 1 root root 8192 Jan 01 00:00 server\r\n-rw-r--r-- 1 root root  240 Jan 01 00:00 package.json\r\n-rw-r--r-- 1 root root 1024 Jan 01 00:00 config.yaml\r\n"))
	case "ps", "ps aux":
		_ = conn.WriteMessage(websocket.TextMessage, []byte("PID   USER     TIME  COMMAND\r\n    1 root      0:05 /app/server --port=8080\r\n   14 root      0:00 sh\r\n"))
	case "env":
		_ = conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("HOSTNAME=%s\r\nKUBERNETES_PORT=tcp://10.96.0.1:443\r\nKUBERNETES_SERVICE_HOST=10.96.0.1\r\nNODE_ENV=production\r\nPORT=8080\r\n", pod)))
	case "exit":
		_ = conn.WriteMessage(websocket.TextMessage, []byte("exit\r\n"))
		_ = conn.Close()
	default:
		if cmd != "" {
			_ = conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("sh: %s: command not found (demo shell)\r\n", cmd)))
		}
	}
}
