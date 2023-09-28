package main

import (
	"io"
	"net"
	"strings"
	"testing"

	"github.com/emersion/go-smtp"
)

var (
	cases = []testCase{
		{[]string{"drone-email"}, "failed sanity check: no finish time for build"},
		{[]string{"drone-email", "--recipients=foo@example.com", "--no.starttls", "--build.finished=1"}, ""},
	}
)

type testCase struct {
	args      []string
	errorText string
}

type testServerBackend struct {
}

func (t *testServerBackend) NewSession(c *smtp.Conn) (smtp.Session, error) {
	return &testServerSession{}, nil
}

type testServerSession struct {
}

func (t *testServerSession) AuthPlain(username, password string) error {
	return nil
}

func (t *testServerSession) Mail(from string, opts *smtp.MailOptions) error {
	return nil
}

func (t *testServerSession) Rcpt(to string, opts *smtp.RcptOptions) error {
	return nil
}

func (t *testServerSession) Data(r io.Reader) error {
	_, err := io.Copy(io.Discard, r)
	return err
}

func (t *testServerSession) Reset() {
}

func (t *testServerSession) Logout() error {
	return nil
}

func testServer() (string, error) {
	s := smtp.NewServer(new(testServerBackend))
	l, err := net.Listen("tcp", "127.0.0.1:0") // XXX: not closed
	if err != nil {
		return "", err
	}
	go s.Serve(l)
	return l.Addr().String(), nil
}

func TestPlugin(t *testing.T) {

	addr, err := testServer()
	if err != nil {
		t.Fatalf("Failed to set up test server: %v", err)
	}
	splitAddr := strings.Split(addr, ":")

	for i, tcase := range cases {
		args := append(tcase.args, []string{"--host=" + splitAddr[0], "--port=" + splitAddr[1]}...)
		err := app.Run(args)
		if tcase.errorText == "" {
			if err != nil {
				t.Fatalf("Test %d failed: unexpected error: (%s)", i, err.Error())
			}
		} else {
			if err == nil {
				t.Fatalf("Test %d failed: expected error: (%s) but got nothing", i, tcase.errorText)
			}
			if err.Error() != tcase.errorText {
				t.Fatalf("Test %d failed: expected error: (%s) did not match actual: (%s)", i, tcase.errorText, err.Error())
			}
		}
	}
}
