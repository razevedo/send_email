package main

import (
	"bufio"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"net/mail"
	"net/smtp"
	"os"
	"time"
	"utils"
)

type EmailType struct {
	From_addr string
	Subject   string
	Body      string
}

type ServerType struct {
	Addr string
	Port string
}

type UserType struct {
	Username string
	Password string
}

type ConfStructure struct {
	Time_int int 
	Email  EmailType
	Server ServerType
	User   UserType
}

var conf ConfStructure //to hold the Configuration

func loadConf(filename string) (ConfStructure, error) {
	var _conf ConfStructure

	file, e := ioutil.ReadFile(filename)
	if e != nil {
		log.Printf("File error: %v\n", e)
		return _conf, e
	}
	e = json.Unmarshal(file, &_conf)
	if e != nil {
		log.Printf("Error unmarshaling... %v\n", e)
		return _conf, e
	}
	log.Printf("conf file --- ", _conf)
	return _conf, nil
}

func readEmailAddrs(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	log.Printf("Emails list loaded...")
	return lines, scanner.Err()
}

// SSL/TLS Email Example
func sendEmail(_to string) {

	from := mail.Address{"", conf.Email.From_addr}
	to := mail.Address{"", _to}
	subj := conf.Email.Subject
	body := conf.Email.Body

	// Setup headers
	headers := make(map[string]string)
	headers["From"] = from.String()
	headers["To"] = to.String()
	headers["Subject"] = subj

	// Setup message
	message := ""
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + body

	// Connect to the SMTP Server
	servername := utils.ConcateStrings(conf.Server.Addr, ":", conf.Server.Port)

	host, _, _ := net.SplitHostPort(servername)

	auth := smtp.PlainAuth("", conf.User.Username, conf.User.Password, host)

	// TLS config
	tlsconfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         host,
	}

	conn, err := tls.Dial("tcp", servername, tlsconfig)
	if err != nil {
		log.Panic(err)
	}

	c, err := smtp.NewClient(conn, host)
	if err != nil {
		log.Panic(err)
	}

	// Auth
	if err = c.Auth(auth); err != nil {
		log.Panic(err)
	}

	if err = c.Mail(from.Address); err != nil {
		log.Panic(err)
	}

	if err = c.Rcpt(to.Address); err != nil {
		log.Panic(err)
	}

	// Data
	w, err := c.Data()
	if err != nil {
		log.Panic(err)
	}

	_, err = w.Write([]byte(message))
	if err != nil {
		log.Panic(err)
	}

	err = w.Close()
	if err != nil {
		log.Panic(err)
	}

	c.Quit()

}

var (
	configFile = flag.String("cfg", "./config.json", "Specifies the config file.")
	emailsFile = flag.String("email", "./emails.txt", "Specifies the file with the list of emails.")
)

func main() {
	var cont int

	flag.Parse()
	_conf, err := loadConf(*configFile)
	if err != nil {
		log.Fatal("Impossible to load configuration file at - ", *configFile)
	} else {
		log.Printf("Configuration File loaded")
	}
	conf = _conf
	lines, err := readEmailAddrs(*emailsFile)
	if err != nil {
		fmt.Println("Error reading file")
	}
	now := time.Now()
	fmt.Println(now, " - Starting to send emails")

	//send emails
	for i := range lines {
		cont++
		sendEmail(lines[i])
		fmt.Println("Mail sent to ", lines[i])
                delay := (time.Duration(rand.Intn(conf.Time_int)) * time.Minute)
		fmt.Println("Waiting for ", delay, " seconds.")
		time.Sleep(delay)
	}
	fmt.Println(cont, " Emails sent in ", time.Since(now))
}
