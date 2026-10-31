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

	"github.com/cygy/ginamite/common/broker"
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

// newTestMailer returns a mailer sending its mails to the SMTP server, with the templates.
func newTestMailer(t *testing.T, server *smtpServer, htmlTemplate, textTemplate string, htmlVars, textVars map[string]map[string]any) *Mailer {
	t.Helper()

	host, port, _ := net.SplitHostPort(server.addr)
	var portNumber int
	_, _ = fmt.Sscan(port, &portNumber)

	return New(Settings{
		Username: "user", Password: "pass", Host: host, Port: portNumber, Sender: testSender,
		HTMLVars: htmlVars, TextVars: textVars,
		HTMLTemplates: template.Must(template.New("notification.html").Parse(htmlTemplate)),
		TextTemplates: ttemplate.Must(ttemplate.New("notification.txt").Parse(textTemplate)),
	})
}

// testSender is the sender of the mails of the tests.
var testSender = Sender{Name: "Sender", EmailAddress: "sender@example.com", NoReplyEmailAddress: "noreply@example.com"}

func TestMailer_Send(t *testing.T) {
	message := Message{
		Subject: "Hello", Recipient: address, Template: TemplateNotification, Locale: "en",
		HTMLVars: map[string]any{"Title": "html-title"},
		TextVars: map[string]any{"Title": "text-title"},
	}

	t.Run("sent", func(t *testing.T) {
		hook.Reset()
		server := startSMTP(t)
		mailer := newTestMailer(t, server, "<p>{{.Title}} {{.Host}}</p>", "{{.Title}} {{.Host}}",
			map[string]map[string]any{"Host": {"en": "https://html"}},
			map[string]map[string]any{"Host": {"en": "https://text"}})

		if err := mailer.Send(message); err != nil {
			t.Fatalf("Send() error = %v", err)
		}

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
		mailer := New(Settings{
			Username: "user", Password: "pass", Host: "127.0.0.1", Port: 1, Sender: testSender,
			HTMLTemplates: template.Must(template.New("notification.html").Parse("{{.Title}}")),
			TextTemplates: ttemplate.Must(ttemplate.New("notification.txt").Parse("{{.Title}}")),
		})

		err := mailer.Send(message)

		if err == nil || broker.IsPermanent(err) {
			t.Errorf("Send() error = %v, want an error a retry can fix", err)
		}
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
				server := startSMTP(t)
				mailer := newTestMailer(t, server, tt.html, tt.text, nil, nil)

				if err := mailer.Send(message); !broker.IsPermanent(err) {
					t.Errorf("Send() error = %v, want a permanent error", err)
				}

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

func TestNew(t *testing.T) {
	settings := Settings{Identity: "id", Username: "user", Password: "pass", Host: "smtp.example.com", Port: 587, Sender: testSender}

	mailer := New(settings)

	if mailer.auth == nil || mailer.smtpHost != "smtp.example.com:587" {
		t.Errorf("auth = %v, smtpHost = %q", mailer.auth, mailer.smtpHost)
	}
	if mailer.settings.Sender != testSender {
		t.Errorf("sender = %+v, want %+v", mailer.settings.Sender, testSender)
	}
	if mailer.translate == nil || mailer.routeURL == nil || mailer.deliver == nil {
		t.Error("the mailer must have its functions")
	}
}
