package selfservice

import (
	"fmt"
	"image/jpeg"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/aryahadii/sarioself/model"
	"github.com/otiai10/gosseract"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
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
		return fmt.Errorf("Samad return %v status code when try to generate captcha",
			response.StatusCode)
	}

	body, _ := ioutil.ReadAll(response.Body)
	bodyString := string(body)
	sessionData.csrf = csrfRegex.FindStringSubmatch(bodyString)[1]
	err = ioutil.WriteFile("/Users/aryahadi/Desktop/login.html", body, 0644)
	return nil
}

// GetAvailableFoods returns a list of all foods that can be reserved
// It checks this week and the next one
func (s *SamadAUTClient) GetAvailableFoods(sessionData *userSessionData, client *http.Client) (map[time.Time]*model.Food, error) {
	availableFoods := make(map[time.Time]*model.Food)

	// Get page 1
	response, err := client.Get(samadReservationURL)
	if err != nil {
		return availableFoods, err
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, fmt.Errorf("Samad returned %v status code when tried to serve reservation page ",
			response.StatusCode)
	}
	body, _ := ioutil.ReadAll(response.Body)
	bodyString := string(body)
	sessionData.csrf = csrfRegex.FindStringSubmatch(bodyString)[1]
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
	request.Header.Set("X-Csrf-Token", sessionData.csrf)
	response, err = client.Do(request)
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
	sessionData.csrf = csrfRegex.FindStringSubmatch(bodyString)[1]

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
