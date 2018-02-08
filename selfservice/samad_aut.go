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
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/aryahadii/sarioself/model"
	"github.com/otiai10/gosseract/v1/gosseract"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
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
		logrus.WithError(err).Fatalln("can't init gosseract client")
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
func (s *SamadAUTClient) GetAvailableFoods() (map[time.Time][]*model.Food, error) {
	availableFoods := make(map[time.Time][]*model.Food)

	// Get page 1
	bodyString, err := s.getSamadReservePage()
	if err != nil {
		return nil, errors.Wrap(err, "can't get first page of Samad")
	}
	foods, err := findSamadFoods(bodyString)
	if err != nil {
		return nil, errors.Wrap(err, "can't find Samad foods")
	}
	for _, food := range foods {
		if food.Status != model.FoodStatusUnavailable {
			if _, ok := availableFoods[*food.Date]; ok {
				availableFoods[*food.Date] = []*model.Food{food}
			} else {
				availableFoods[*food.Date] = append(availableFoods[*food.Date], food)
			}
		}
	}

	// Get page 2
	nextBodyString, err := s.getNextSamadReservePage(bodyString)
	if err != nil {
		return nil, errors.Wrap(err, "can't get second page of Samad")
	}
	foods, err = findSamadFoods(nextBodyString)
	if err != nil {
		return nil, errors.Wrap(err, "can't find Samad foods from second page")
	}
	for _, food := range foods {
		if food.Status != model.FoodStatusUnavailable {
			if _, ok := availableFoods[*food.Date]; ok {
				availableFoods[*food.Date] = []*model.Food{food}
			} else {
				availableFoods[*food.Date] = append(availableFoods[*food.Date], food)
			}
		}
	}

	return availableFoods, nil
}

func (s *SamadAUTClient) ToggleFoodReservation(date *time.Time, foodID string) (bool, error) {
	// Page 1
	bodyString, err := s.getSamadReservePage()
	if err != nil {
		return false, errors.Wrap(err, "can't get first page of Samad")
	}
	toggled, err := s.toggleFoodReservation(bodyString, date, foodID)
	if err != nil {
		if samadError, ok := err.(SamadError); ok {
			return false, samadError
		}
		return false, errors.Wrap(err, "can't toggle food reservation")
	}

	if !toggled {
		// Page 2
		bodyString, err = s.getNextSamadReservePage(bodyString)
		if err != nil {
			return false, errors.Wrap(err, "can't get second page of Samad")
		}
		toggled, err = s.toggleFoodReservation(bodyString, date, foodID)
		if err != nil {
			if samadError, ok := err.(SamadError); ok {
				return false, samadError
			}
			return false, errors.Wrap(err, "can't toggle food reservation")
		}
	}

	return toggled, nil
}

func (s *SamadAUTClient) GetCredit() (int, error) {
	var credit int

	bodyString, err := s.getSamadReservePage()
	if err != nil {
		return credit, errors.Wrap(err, "can't get first page of Samad")
	}

	document, err := goquery.NewDocumentFromReader(strings.NewReader(bodyString))
	if err != nil {
		return credit, errors.Wrap(err, "can't init goquery on document")
	}

	document.Find("#creditId").Each(func(i int, s *goquery.Selection) {
		credit, _ = strconv.Atoi(s.Text())
	})
	return credit, nil
}
