package selfservice

import (
	"net/http"
	"net/http/cookiejar"
	"testing"

	"github.com/aryahadii/sarioself/configuration"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/publicsuffix"
)

func init() {
	configuration.SarioselfConfigPath = "../test/config.yaml"
	err := configuration.LoadConfig()
	if err != nil {
		log.Fatalln("can't load config file")
	}
}

func TestLogin(t *testing.T) {
	samad := NewSamadAUTClient()
	jarOption := &cookiejar.Options{
		PublicSuffixList: publicsuffix.List,
	}
	cookieJar, _ := cookiejar.New(jarOption)
	client := &http.Client{Jar: cookieJar}

	sessionData, err := samad.createConnection(client)
	if err != nil {
		t.Fatal(err)
	}
	if len(sessionData.csrf) == 0 {
		t.Fatalf("CSRF is not valid, %v", sessionData.csrf)
	}
	sessionData.username = configuration.SarioselfConfig.GetString("test-user.username")
	sessionData.password = configuration.SarioselfConfig.GetString("test-user.password")
	sessionData.jar = cookieJar

	captcha, err := samad.readCaptcha(client)
	if err != nil {
		t.Fatal(err)
	}

	err = samad.login(captcha, sessionData, client)
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetAvailableFoods(t *testing.T) {
	samad := NewSamadAUTClient()
	jarOption := &cookiejar.Options{
		PublicSuffixList: publicsuffix.List,
	}
	cookieJar, _ := cookiejar.New(jarOption)
	client := &http.Client{Jar: cookieJar}

	sessionData, err := samad.createConnection(client)
	if err != nil {
		t.Fatal(err)
	}
	if len(sessionData.csrf) == 0 {
		t.Fatalf("CSRF is not valid, %v", sessionData.csrf)
	}
	sessionData.username = configuration.SarioselfConfig.GetString("test-user.username")
	sessionData.password = configuration.SarioselfConfig.GetString("test-user.password")
	sessionData.jar = cookieJar

	captcha, err := samad.readCaptcha(client)
	if err != nil {
		t.Fatal(err)
	}

	err = samad.login(captcha, sessionData, client)
	if err != nil {
		t.Fatal(err)
	}

	foods, err := samad.GetAvailableFoods(sessionData, client)
	if err != nil {
		t.Fatal(err)
	}
	if len(foods) == 0 {
		t.FailNow()
	}
}
