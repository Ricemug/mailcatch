package smtp

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"fakesmtp/internal/models"
)

type Server struct {
	port     string
	listener net.Listener
	onEmail  func(*models.Email)
}

type session struct {
	conn   net.Conn
	reader *bufio.Reader
	writer *bufio.Writer
	from   string
	to     []string
	data   string
}

func NewServer(port string, onEmail func(*models.Email)) *Server {
	return &Server{
		port:    port,
		onEmail: onEmail,
	}
}

func (s *Server) Start() error {
	var err error
	s.listener, err = net.Listen("tcp", ":"+s.port)
	if err != nil {
		return fmt.Errorf("failed to start SMTP server: %w", err)
	}

	log.Printf("SMTP server listening on port %s", s.port)

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %v", err)
			continue
		}

		go s.handleConnection(conn)
	}
}

func (s *Server) Stop() error {
	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()
	
	sess := &session{
		conn:   conn,
		reader: bufio.NewReader(conn),
		writer: bufio.NewWriter(conn),
		to:     make([]string, 0),
	}

	// Send greeting
	sess.writeLine("220 fakesmtp ready")

	for {
		line, err := sess.readLine()
		if err != nil {
			log.Printf("Error reading from client: %v", err)
			break
		}

		cmd := strings.ToUpper(strings.TrimSpace(line))
		parts := strings.Fields(cmd)
		
		if len(parts) == 0 {
			continue
		}

		switch parts[0] {
		case "HELO", "EHLO":
			sess.handleHelo()
		case "MAIL":
			sess.handleMail(line)
		case "RCPT":
			sess.handleRcpt(line)
		case "DATA":
			sess.handleData(s.onEmail)
		case "QUIT":
			sess.writeLine("221 Bye")
			return
		case "RSET":
			sess.reset()
			sess.writeLine("250 OK")
		default:
			sess.writeLine("502 Command not implemented")
		}
	}
}

func (sess *session) readLine() (string, error) {
	return sess.reader.ReadString('\n')
}

func (sess *session) writeLine(line string) error {
	_, err := sess.writer.WriteString(line + "\r\n")
	if err != nil {
		return err
	}
	return sess.writer.Flush()
}

func (sess *session) handleHelo() {
	sess.writeLine("250 Hello")
}

func (sess *session) handleMail(line string) {
	// Extract email from "MAIL FROM:<email@example.com>"
	start := strings.Index(line, "<")
	end := strings.Index(line, ">")
	
	if start != -1 && end != -1 && end > start {
		sess.from = line[start+1 : end]
		sess.writeLine("250 OK")
	} else {
		sess.writeLine("501 Syntax error")
	}
}

func (sess *session) handleRcpt(line string) {
	// Extract email from "RCPT TO:<email@example.com>"
	start := strings.Index(line, "<")
	end := strings.Index(line, ">")
	
	if start != -1 && end != -1 && end > start {
		to := line[start+1 : end]
		sess.to = append(sess.to, to)
		sess.writeLine("250 OK")
	} else {
		sess.writeLine("501 Syntax error")
	}
}

func (sess *session) handleData(onEmail func(*models.Email)) {
	sess.writeLine("354 Start mail input; end with <CRLF>.<CRLF>")
	
	var data strings.Builder
	for {
		line, err := sess.readLine()
		if err != nil {
			log.Printf("Error reading data: %v", err)
			return
		}
		
		// Check for end of data
		if strings.TrimSpace(line) == "." {
			break
		}
		
		data.WriteString(line)
	}
	
	sess.data = data.String()
	
	// Parse email and create model
	email := sess.parseEmail()
	if onEmail != nil {
		onEmail(email)
	}
	
	sess.writeLine("250 OK: Message accepted")
	sess.reset()
}

func (sess *session) parseEmail() *models.Email {
	lines := strings.Split(sess.data, "\n")
	
	email := &models.Email{
		From:      sess.from,
		To:        strings.Join(sess.to, ", "),
		Raw:       sess.data,
		CreatedAt: time.Now(),
	}
	
	// Simple header parsing
	inHeaders := true
	var bodyLines []string
	
	for _, line := range lines {
		line = strings.TrimRight(line, "\r")
		
		if inHeaders {
			if line == "" {
				inHeaders = false
				continue
			}
			
			if strings.HasPrefix(strings.ToLower(line), "subject:") {
				email.Subject = strings.TrimSpace(line[8:])
			}
		} else {
			bodyLines = append(bodyLines, line)
		}
	}
	
	email.Body = strings.Join(bodyLines, "\n")
	
	// Simple HTML detection
	if strings.Contains(strings.ToLower(email.Body), "<html") ||
		strings.Contains(strings.ToLower(email.Body), "<!doctype") {
		email.HTML = email.Body
	}
	
	return email
}

func (sess *session) reset() {
	sess.from = ""
	sess.to = make([]string, 0)
	sess.data = ""
}