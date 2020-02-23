package main

import (
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"net/smtp"
	"os"
	"time"
)

type Request struct {
	Token string `json:"token"`
	From string `json:"from"`
	To []string `json:"to"`
	Subject string `json:"subject"`
	Body string `json:"body"`
	Files map[string]byte `json:"files"`
	Time int64 `json:"time"`
	IP string
}
func (r *Request) Authorize() bool{
	if r.Token == "0x00-0xff"{
		return true
	}else{
		return false
	}
}
func Parse (c *gin.Context) *Request{
	var req Request
	err := c.ShouldBindJSON(&req)
	if err == nil{
		if req.Time == 0{
			req.Time = time.Now().Unix()
		}
		req.IP = c.Request.RemoteAddr
		return &req
	}
	c.AbortWithStatus(http.StatusBadRequest)
	return &Request{}
}

type Login struct {
	Login string `json:"login"`
	Password string `json:"password"`
	Server string `json:"server"`
	Port string `json:"port"`
}
type Logins struct {
	Content map[string]Login `json:"logins"`
}

type Mailer struct {
	//server must contain port! Ex: smtp.gmail.com:465
	logins Logins

	conn tls.Conn
	client smtp.Client
	auth smtp.Auth
}
func (m *Mailer) Init (config string) error{
	dat, err := ioutil.ReadFile(config)
	if err != nil {
		return err
	}
	err = json.Unmarshal(dat, &m.logins)
	if err != nil {
		return err
	}
	return nil
}
func (m *Mailer) Send(r *Request) error{
	var login Login
	for _,v := range m.logins.Content{
		if v.Login == r.From{
			login = v
			break
		}		
	}
	m.auth = smtp.PlainAuth("", login.Login, login.Password, login.Server+":"+login.Port)
	tlsconf := &tls.Config{
		InsecureSkipVerify: true,
		ServerName: login.Server}
	conn, err := tls.Dial("tcp", login.Server+":"+login.Port, tlsconf)
	if err != nil {return err}
	m.conn = *conn
	c, err := smtp.NewClient(&m.conn, login.Server+":"+login.Port)
	if err != nil {return err}
	m.client = *c
	
	//Construct letter
	msg := "From: "+r.From+"\r\nTo: "
	for _,v := range r.To{
		msg += v+", "
	}
	msg = msg[:len(msg)-3]+"\r\n" //Delete last ", "
	msg += "Subject: "+r.Subject+"\r\n"+
		"Content-Type: text/plain; charset=\"utf-8\"\r\n"+
		"Content-Transfer-Encoding: base64\r\n\r\n" +
		base64.StdEncoding.EncodeToString([]byte(r.Body))

	//Send letter
	err = m.client.Auth(m.auth)
	if err != nil {return err}
	m.client.Mail(r.From)
	for _,v := range r.To{
		m.client.Rcpt(v)
	}
	writer, cw_err := m.client.Data()
	defer writer.Close()
	if cw_err != nil {return cw_err}
	_, w_err := writer.Write([]byte(msg))
	if err != nil {return w_err}

	fmt.Printf("%s: subj:%s -> Sent letter to %s\n", r.IP, r.Subject, r.To)
	return nil
}

func main() {
	config := flag.String("c", "", "Config path")
	flag.Parse()
	if *config==""{
		fmt.Print("Config path not defined. Use \"-c\" flag.")
		os.Exit(-1)
	}

	r := gin.New()
	mailer := &Mailer{}
	err := mailer.Init(*config)
	if err != nil {
		fmt.Print(err)
		os.Exit(-1)
	}

	r.POST("/", func(c *gin.Context) {
		req := Parse(c)
		if req.Authorize(){
			err := mailer.Send(req)
			if err != nil {
				fmt.Print(err)
				c.AbortWithStatus(http.StatusInternalServerError)
				return
			}
			c.AbortWithStatus(http.StatusOK)
			return
		}else {
			fmt.Printf("Subj:%s -> Wrong token from %s\n", req.Subject, req.IP)
			c.JSON(http.StatusBadRequest, map[string]string{
				"Error": "Access denied!",
			})
			return
		}
	})

	r.Run(":5000")
}

