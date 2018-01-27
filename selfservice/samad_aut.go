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
	"time"

	"github.com/aryahadii/sarioself/model"
	"github.com/otiai10/gosseract/v1/gosseract"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/publicsuffix"
)

const (
	samadLoginPageURL         = "http://samad.aut.ac.ir/loginpage.rose"
	samadLoginBackendURL      = "http://samad.aut.ac.ir/j_security_check"
	samadCaptchaURL           = "http://samad.aut.ac.ir/captcha.jpg"
	samadReservationURL       = "http://samad.aut.ac.ir/nurture/user/multi/reserve/reserve.rose"
	samadReservationActionURL = "http://samad.aut.ac.ir/nurture/user/multi/reserve/reserve.rose"
)

var (
	csrfRegex = regexp.MustCompile(`'X-CSRF-TOKEN' : '(.*)'`)
	ocrClient *gosseract.Client
)

// SamadAUTClient is client of Amirkabir Univerity of Technology's restaurant
type SamadAUTClient struct {
	sessionData *userSessionData
	httpClient  *http.Client
}

func init() {
	var err error
	ocrClient, err = gosseract.NewClient()
	if err != nil {
		log.WithError(err).Fatalln("can't init gosseract client")
	}
}

// NewSamadAUTClient creates new instance of SamadAUTClient
func NewSamadAUTClient(username, password string) (*SamadAUTClient, error) {
	samad := &SamadAUTClient{}

	// Cookie Jar
	jarOption := &cookiejar.Options{
		PublicSuffixList: publicsuffix.List,
	}
	cookieJar, _ := cookiejar.New(jarOption)
	samad.httpClient = &http.Client{Jar: cookieJar}

	// Session data
	err := samad.createConnection()
	if err != nil {
		return nil, errors.Wrap(err, "can't create connection")
	}
	if len(samad.sessionData.csrf) == 0 {
		return nil, fmt.Errorf("CSRF isn't valid")
	}
	samad.sessionData.username = username
	samad.sessionData.password = password
	samad.sessionData.jar = cookieJar

	// Captcha
	captcha, err := samad.readCaptcha()
	if err != nil {
		return nil, errors.Wrap(err, "can't read captcha")
	}

	// Login
	err = samad.login(captcha)
	if err != nil {
		return nil, errors.Wrap(err, "can't login to Samad")
	}

	return samad, nil
}

// createConnection creates new connection to Samad and returns
// CSRF token of session
func (s *SamadAUTClient) createConnection() error {
	response, err := s.httpClient.Get(samadLoginPageURL)
	if err != nil {
		return errors.Wrap(err, "can't connect to Samad")
	}
	defer func() {
		io.Copy(ioutil.Discard, response.Body)
		response.Body.Close()
	}()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("Samad returned %v status code", response.StatusCode)
	}

	body, _ := ioutil.ReadAll(response.Body)
	bodyString := string(body)
	s.sessionData = &userSessionData{
		csrf: csrfRegex.FindStringSubmatch(bodyString)[1],
	}
	return nil
}

// readCaptcha gets captcha from Samad and uses OCR to extract text
func (s *SamadAUTClient) readCaptcha() (string, error) {
	response, err := s.httpClient.Get(samadCaptchaURL)
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
func (s *SamadAUTClient) login(captcha string) error {
	form := url.Values{}
	form.Set("_csrf", s.sessionData.csrf)
	form.Set("username", s.sessionData.username)
	form.Set("password", s.sessionData.password)
	form.Set("captcha_input", captcha)
	request, err := http.NewRequest("POST", samadLoginBackendURL, strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("X-Csrf-Token", s.sessionData.csrf)

	response, err := s.httpClient.Do(request)
	if err != nil {
		return errors.Wrap(err, "can't login to Samad")
	}
	defer func() {
		io.Copy(ioutil.Discard, response.Body)
		response.Body.Close()
	}()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("Samad return %v status code when try to generate captcha",
			response.StatusCode)
	}

	body, _ := ioutil.ReadAll(response.Body)
	bodyString := string(body)
	s.sessionData.csrf = csrfRegex.FindStringSubmatch(bodyString)[1]
	return nil
}

// GetAvailableFoods returns a list of all foods that can be reserved
// It checks this week and the next one
func (s *SamadAUTClient) GetAvailableFoods() (map[time.Time]*model.Food, error) {
	availableFoods := make(map[time.Time]*model.Food)

	// Get page 1
	response, err := s.httpClient.Get(samadReservationURL)
	if err != nil {
		return availableFoods, err
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, fmt.Errorf("Samad returned %v status code when tried to serve reservation page ",
			response.StatusCode)
	}
	body, _ := ioutil.ReadAll(response.Body)
	bodyString := string(body)
	s.sessionData.csrf = csrfRegex.FindStringSubmatch(bodyString)[1]
	io.Copy(ioutil.Discard, response.Body)
	response.Body.Close()
	foods, err := findSamadFoods(bodyString)
	if err != nil {
		return nil, errors.Wrap(err, "can't find Samad foods")
	}
	for _, food := range foods {
		if food.Status != model.FoodStatusUnavailable {
			availableFoods[*food.Date] = food
		}
	}

	// Get page 2
	formValues, err := extractFormInputValues(bodyString)
	if err != nil {
		return nil, errors.Wrap(err, "can't extract form input values")
	}
	formValues.Set("method:showNextWeek", "Submit")
	request, err := http.NewRequest("POST", samadReservationActionURL, strings.NewReader(formValues.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("X-Csrf-Token", s.sessionData.csrf)
	response, err = s.httpClient.Do(request)
	defer func() {
		io.Copy(ioutil.Discard, response.Body)
		response.Body.Close()
	}()
	if err != nil {
		return nil, errors.Wrap(err, "can't read second page of Samad")
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, fmt.Errorf("Samad returned %v status code when tried to serve second reservation page ",
			response.StatusCode)
	}
	body, _ = ioutil.ReadAll(response.Body)
	bodyString = string(body)
	s.sessionData.csrf = csrfRegex.FindStringSubmatch(bodyString)[1]

	foods, err = findSamadFoods(bodyString)
	if err != nil {
		return nil, errors.Wrap(err, "can't find Samad foods from second page")
	}
	for _, food := range foods {
		if food.Status != model.FoodStatusUnavailable {
			availableFoods[*food.Date] = food
		}
	}

	return availableFoods, nil
}

func (s *SamadAUTClient) ReserveFood(food *model.Food) error {
	return nil
}
