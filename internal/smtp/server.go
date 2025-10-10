package smtp

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
	"net"
	"net/mail"
	"strings"
	"time"

	"mailcatch/internal/models"
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
	sess.writeLine("220 mailcatch ready")

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
	email := &models.Email{
		From:      sess.from,
		To:        strings.Join(sess.to, ", "),
		Raw:       sess.data,
		CreatedAt: time.Now(),
	}

	// Parse using net/mail
	msg, err := mail.ReadMessage(strings.NewReader(sess.data))
	if err != nil {
		log.Printf("Error parsing email: %v", err)
		// Fallback to simple parsing
		email.Body = sess.data
		return email
	}

	// Extract headers
	email.Subject = msg.Header.Get("Subject")

	// Decode subject if needed
	dec := new(mime.WordDecoder)
	if decodedSubject, err := dec.DecodeHeader(email.Subject); err == nil {
		email.Subject = decodedSubject
	}

	// Parse body
	contentType := msg.Header.Get("Content-Type")
	if contentType == "" {
		// Plain text email
		body, _ := io.ReadAll(msg.Body)
		email.Body = string(body)
		return email
	}

	mediaType, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		body, _ := io.ReadAll(msg.Body)
		email.Body = string(body)
		return email
	}

	if strings.HasPrefix(mediaType, "multipart/") {
		sess.parseMultipart(msg.Body, params["boundary"], email)
	} else {
		// Single part message
		body, _ := io.ReadAll(msg.Body)
		decoded := sess.decodeContent(string(body), msg.Header.Get("Content-Transfer-Encoding"))

		if strings.HasPrefix(mediaType, "text/html") {
			email.HTML = decoded
			email.Body = decoded
		} else {
			email.Body = decoded
		}
	}

	return email
}

func (sess *session) parseMultipart(body io.Reader, boundary string, email *models.Email) {
	mr := multipart.NewReader(body, boundary)

	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("Error reading multipart: %v", err)
			break
		}

		contentType := part.Header.Get("Content-Type")
		mediaType, _, _ := mime.ParseMediaType(contentType)
		transferEncoding := part.Header.Get("Content-Transfer-Encoding")

		partBody, _ := io.ReadAll(part)
		decoded := sess.decodeContent(string(partBody), transferEncoding)

		if strings.HasPrefix(mediaType, "text/plain") {
			if email.Body == "" {
				email.Body = decoded
			}
		} else if strings.HasPrefix(mediaType, "text/html") {
			email.HTML = decoded
		} else if strings.HasPrefix(mediaType, "multipart/") {
			// Nested multipart
			_, params, _ := mime.ParseMediaType(contentType)
			sess.parseMultipart(bytes.NewReader(partBody), params["boundary"], email)
		}
	}
}

func (sess *session) decodeContent(content string, encoding string) string {
	encoding = strings.ToLower(strings.TrimSpace(encoding))

	switch encoding {
	case "base64":
		decoded, err := base64.StdEncoding.DecodeString(content)
		if err != nil {
			log.Printf("Error decoding base64: %v", err)
			return content
		}
		return string(decoded)
	case "quoted-printable":
		reader := quotedprintable.NewReader(strings.NewReader(content))
		decoded, err := io.ReadAll(reader)
		if err != nil {
			log.Printf("Error decoding quoted-printable: %v", err)
			return content
		}
		return string(decoded)
	default:
		return content
	}
}

func (sess *session) reset() {
	sess.from = ""
	sess.to = make([]string, 0)
	sess.data = ""
}