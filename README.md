# Mailer
Kind of microservice for sends emails

## 1. Config

Must contain credentials as the next:

```
{
	"logins":
	{		
		"basic" : {
		"login": "noreply@prismcorp.com",
		"password": "SitelenEKamaSoweliSuli",
		"server": "smtp.prismcorp.com",
		"port": "465"
		}
	},
	"additional":{
		...
	}
	...
}

```
## 2. Build

I don't know about you, but i building in docker

```
FROM golang:latest
USER 0
RUN go get -v -u "github.com/gin-gonic/gin"


WORKDIR /opt
```
And then

```
docker run --rm -it -v D:\Temp\:/mnt/tmp:rw -p 5000:5000 my_container /bin/bash

#git clone https://github.com/EvgenyOvsov/Mailer.git .
#export GOPATH=/opt/:/go
#go build -o /mnt/tmp/mailer main
#./mailer -h
#mv ./mailer /mnt/tmp/ 
```
## 3. Run
Log's don't have enough information, but better than nothing...

$mailer -c /path/to/config.json >> ./log.txt

## 4. Use

```
curl -X POST -H "Content-Type: application/json" \
-d "@/opt/data.json"
```
...when data.json is...

```
{
	"Token":"0x00-1xff",
	"From": "noreply@prismcorp.com",
	"To": ["evgeny.ovsov@prismcorp.com"],
	"Subject": "Test",
	"Body": "This is the test!"
}
```