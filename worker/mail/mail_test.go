package mail

import (
	"bufio"
	"context"
	"fmt"
	"html/template"
	"net"
	"strings"
	"sync"
	"testing"
	ttemplate "text/template"

	"github.com/cygy/ginamite/common/config"
	"github.com/cygy/ginamite/common/template/html"
	"github.com/cygy/ginamite/common/template/text"
)

func TestIsEmailAddressValidFunc_check(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name  string
		f     IsEmailAddressValidFunc
		want  bool
		wantL string // The expected log, if any.
	}{
		{"valid", valid, true, ""},
		{"invalid", invalid, false, ""},
		{"failing", failing, false, "unable to check the email address"},
		{"undefined", nil, false, "the function is undefined"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hook.Reset()

			if got := tt.f.check(ctx, address); got != tt.want {
				t.Errorf("check = %v, want %v", got, tt.want)
			}

			logged := ""
			if entry := hook.LastEntry(); entry != nil {
				logged = entry.Message
			}
			if !strings.Contains(logged, tt.wantL) || (tt.wantL != "" && logged == "") {
				t.Errorf("logged %q, want it to contain %q", logged, tt.wantL)
			}
		})
	}
}

func TestMergeVars(t *testing.T) {
	common := map[string]map[string]any{
		"Host": {"en": "https://en", "fr": "https://fr"},
		"Home": {"en": "home-en"},
	}

	got := mergeVars(common, "en", map[string]any{"Home": "custom", "Title": "t"})
	want := map[string]any{"Host": "https://en", "Home": "custom", "Title": "t"}
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Errorf("mergeVars = %v, want %v", got, want)
	}

	if got := mergeVars(common, "de", nil); got["Host"] != nil {
		t.Errorf("mergeVars for an unknown locale = %v", got)
	}
}

// smtpServer is a fake SMTP server which records the data of the mails.
type smtpServer struct {
	addr string

	mu   sync.Mutex
	data []string
}

// startSMTP starts a fake SMTP server, stopped at the end of the test.
func startSMTP(t *testing.T) *smtpServer {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = listener.Close() })

	s := &smtpServer{addr: listener.Addr().String()}
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go s.serve(conn)
		}
	}()

	return s
}

// serve talks SMTP to the connection until it is closed.
func (s *smtpServer) serve(conn net.Conn) {
	defer func() { _ = conn.Close() }()

	reply := func(format string, args ...any) { _, _ = fmt.Fprintf(conn, format+"\r\n", args...) }
	reader := bufio.NewReader(conn)
	reply("220 localhost ready")

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return
		}

		switch command := strings.ToUpper(strings.TrimSpace(line)); {
		case strings.HasPrefix(command, "EHLO"), strings.HasPrefix(command, "HELO"):
			reply("250-localhost")
			reply("250 AUTH PLAIN")
		case strings.HasPrefix(command, "AUTH"):
			reply("235 ok")
		case strings.HasPrefix(command, "DATA"):
			reply("354 go ahead")

			var data strings.Builder
			for {
				l, err := reader.ReadString('\n')
				if err != nil || l == ".\r\n" {
					break
				}
				data.WriteString(l)
			}

			s.mu.Lock()
			s.data = append(s.data, data.String())
			s.mu.Unlock()
			reply("250 queued")
		case strings.HasPrefix(command, "QUIT"):
			reply("221 bye")
			return
		default:
			reply("250 ok")
		}
	}
}

// received returns the data of the mails received.
func (s *smtpServer) received() []string {
	s.mu.Lock()
	defer s.mu.Unlock()

	return append([]string(nil), s.data...)
}

// useTemplates sets the templates and the configuration used to send the mails, for the duration of the test.
func useTemplates(t *testing.T, htmlTemplate, textTemplate string) {
	t.Helper()

	oldHTML, oldText, oldConfig := html.Main, text.Main, config.Main
	t.Cleanup(func() { html.Main, text.Main, config.Main = oldHTML, oldText, oldConfig })

	html.Main = template.Must(template.New("notification.html").Parse(htmlTemplate))
	text.Main = ttemplate.Must(ttemplate.New("notification.txt").Parse(textTemplate))
	config.Main = &config.Configuration{
		SenderName: "Sender", SenderEmailAddress: "sender@example.com", NoReplyEmailAddress: "noreply@example.com",
	}
}

func TestMessage_Send(t *testing.T) {
	oldAuth, oldHost, oldHTMLVars, oldTextVars := mailAuth, smtpHost, commonHTMLVars, commonTextVars
	t.Cleanup(func() {
		mailAuth, smtpHost, commonHTMLVars, commonTextVars = oldAuth, oldHost, oldHTMLVars, oldTextVars
	})

	message := Message{
		Subject: "Hello", Recipient: address, Template: TemplateNotification, Locale: "en",
		HTMLVars: map[string]any{"Title": "html-title"},
		TextVars: map[string]any{"Title": "text-title"},
	}

	t.Run("sent", func(t *testing.T) {
		hook.Reset()
		useTemplates(t, "<p>{{.Title}} {{.Host}}</p>", "{{.Title}} {{.Host}}")
		server := startSMTP(t)

		host, port, _ := net.SplitHostPort(server.addr)
		var portNumber int
		_, _ = fmt.Sscan(port, &portNumber)
		Initialize("", "user", "pass", host, portNumber,
			map[string]map[string]any{"Host": {"en": "https://html"}},
			map[string]map[string]any{"Host": {"en": "https://text"}})

		message.Send()

		data := server.received()
		if len(data) != 1 {
			t.Fatalf("received %d mails, want 1 (logs: %v)", len(data), hook.AllEntries())
		}
		for _, want := range []string{"Subject: Hello", address, "<p>html-title https://html</p>", "text-title https://text", "sender@example.com", "noreply@example.com"} {
			if !strings.Contains(data[0], want) {
				t.Errorf("the mail does not contain %q:\n%s", want, data[0])
			}
		}
	})

	t.Run("SMTP failure", func(t *testing.T) {
		hook.Reset()
		useTemplates(t, "{{.Title}}", "{{.Title}}")
		Initialize("", "user", "pass", "127.0.0.1", 1, nil, nil)

		message.Send()

		if entry := hook.LastEntry(); entry == nil || entry.Message != "unable to send mail" {
			t.Errorf("last log = %+v", entry)
		}
	})

	t.Run("template failures send nothing", func(t *testing.T) {
		tests := []struct{ name, html, text string }{
			{"HTML", "{{template \"missing\"}}", "ok"},
			{"text", "ok", "{{template \"missing\"}}"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				hook.Reset()
				useTemplates(t, tt.html, tt.text)
				server := startSMTP(t)
				smtpHost = server.addr

				message.Send()

				if entry := hook.LastEntry(); entry == nil || entry.Message != "unable to execute the template" {
					t.Errorf("last log = %+v", entry)
				}
				if got := server.received(); len(got) != 0 {
					t.Errorf("received %d mails, want none", len(got))
				}
			})
		}
	})
}

func TestInitialize(t *testing.T) {
	oldAuth, oldHost, oldHTMLVars, oldTextVars := mailAuth, smtpHost, commonHTMLVars, commonTextVars
	t.Cleanup(func() {
		mailAuth, smtpHost, commonHTMLVars, commonTextVars = oldAuth, oldHost, oldHTMLVars, oldTextVars
	})

	htmlVars := map[string]map[string]any{"a": {"en": 1}}
	textVars := map[string]map[string]any{"b": {"en": 2}}
	Initialize("id", "user", "pass", "smtp.example.com", 587, htmlVars, textVars)

	if mailAuth == nil || smtpHost != "smtp.example.com:587" {
		t.Errorf("mailAuth = %v, smtpHost = %q", mailAuth, smtpHost)
	}
	if fmt.Sprint(commonHTMLVars) != fmt.Sprint(htmlVars) || fmt.Sprint(commonTextVars) != fmt.Sprint(textVars) {
		t.Errorf("vars = %v, %v", commonHTMLVars, commonTextVars)
	}
}
