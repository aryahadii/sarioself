package selfservice

import (
	"fmt"
	"image/jpeg"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strings"

	"github.com/otiai10/gosseract"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

const (
	samadLoginPageURL    = "http://samad.aut.ac.ir/loginpage.rose"
	samadLoginBackendURL = "http://samad.aut.ac.ir/j_security_check"
	samadCaptchaURL      = "http://samad.aut.ac.ir/captcha.jpg"
)

var (
	csrfRegex = regexp.MustCompile("'X-CSRF-TOKEN' : '(.*)'")
	ocrClient *gosseract.Client
)

// SamadAUTClient is client of Amirkabir Univerity of Technology's restaurant
type SamadAUTClient struct {
}

type userSessionData struct {
	username string
	password string
	csrf     string
	jar      *cookiejar.Jar
}

func init() {
	var err error
	ocrClient, err = gosseract.NewClient()
	if err != nil {
		log.WithError(err).Fatalln("can't init gosseract client")
	}
}

// NewSamadAUTClient creates new instance of SamadAUTClient
func NewSamadAUTClient() *SamadAUTClient {
	samad := &SamadAUTClient{}
	return samad
}

// createConnection creates new connection to Samad and returns
// CSRF token of session
func (s *SamadAUTClient) createConnection(client *http.Client) (*userSessionData, error) {
	response, err := client.Get(samadLoginPageURL)
	if err != nil {
		return nil, errors.Wrap(err, "can't connect to Samad")
	}
	defer func() {
		io.Copy(ioutil.Discard, response.Body)
		response.Body.Close()
	}()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, fmt.Errorf("Samad returned %v status code", response.StatusCode)
	}

	body, _ := ioutil.ReadAll(response.Body)
	bodyString := string(body)
	sessionData := &userSessionData{
		csrf: csrfRegex.FindStringSubmatch(bodyString)[1],
	}
	return sessionData, nil
}

// readCaptcha gets captcha from Samad and uses OCR to extract text
func (s *SamadAUTClient) readCaptcha(client *http.Client) (string, error) {
	response, err := client.Get(samadCaptchaURL)
	if err != nil {
		return "", errors.Wrap(err, "can't connect to Samad")
	}
	defer func() {
		io.Copy(ioutil.Discard, response.Body)
		response.Body.Close()
	}()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return "", fmt.Errorf("Samad returned %v status code when try to get captcha",
			response.StatusCode)
	}

	image, _ := jpeg.Decode(response.Body)
	captcha, _ := ocrClient.Image(image).Out()
	return strings.TrimSpace(captcha), nil
}

// login tries to log into Samad website
func (s *SamadAUTClient) login(captcha string, sessionData *userSessionData,
	client *http.Client) error {
	form := url.Values{}
	form.Set("_csrf", sessionData.csrf)
	form.Set("username", sessionData.username)
	form.Set("password", sessionData.password)
	form.Set("captcha_input", captcha)
	request, err := http.NewRequest("POST", samadLoginBackendURL, strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("X-Csrf-Token", sessionData.csrf)

	response, err := client.Do(request)
	if err != nil {
		return errors.Wrap(err, "can't login to Samad")
	}
	defer func() {
		io.Copy(ioutil.Discard, response.Body)
		response.Body.Close()
	}()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("Samad returned %v status code when try to get captcha",
			response.StatusCode)
	}

	body, _ := ioutil.ReadAll(response.Body)
	err = ioutil.WriteFile("/Users/aryahadi/Desktop/login.html", body, 0644)
	return nil
}
