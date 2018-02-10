package selfservice

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/aryahadii/sarioself/model"
	"github.com/pkg/errors"
	"github.com/yaa110/go-persian-calendar/ptime"
)

func (s *SamadAUTClient) getSamadReservePage() (string, error) {
	var bodyString string

	response, err := s.httpClient.Get(samadReservationURL)
	if err != nil {
		return bodyString, err
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return bodyString, fmt.Errorf("Samad returned %v status code when tried to serve reservation page ",
			response.StatusCode)
	}
	body, _ := ioutil.ReadAll(response.Body)
	bodyString = string(body)
	s.sessionData.csrf = csrfRegex.FindStringSubmatch(bodyString)[1]
	io.Copy(ioutil.Discard, response.Body)
	response.Body.Close()

	return bodyString, nil
}

func (s *SamadAUTClient) getNextSamadReservePage(bodyString string) (string, error) {
	var nextBodyString string

	formValues, err := extractFormInputValues(bodyString)
	if err != nil {
		return nextBodyString, errors.Wrap(err, "can't extract form input values")
	}
	formValues.Set("method:showNextWeek", "Submit")
	request, err := http.NewRequest("POST", samadReservationActionURL, strings.NewReader(formValues.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("X-Csrf-Token", s.sessionData.csrf)
	response, err := s.httpClient.Do(request)
	defer func() {
		io.Copy(ioutil.Discard, response.Body)
		response.Body.Close()
	}()
	if err != nil {
		return nextBodyString, errors.Wrap(err, "can't read second page of Samad")
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nextBodyString, fmt.Errorf("Samad returned %v status code when tried to serve second reservation page ",
			response.StatusCode)
	}
	body, _ := ioutil.ReadAll(response.Body)
	nextBodyString = string(body)
	s.sessionData.csrf = csrfRegex.FindStringSubmatch(bodyString)[1]

	return nextBodyString, nil
}

func (s *SamadAUTClient) toggleFoodReservation(samadPage string, date *time.Time, foodID string) (bool, error) {
	var toggled bool

	document, err := goquery.NewDocumentFromReader(strings.NewReader(samadPage))
	if err != nil {
		return toggled, errors.Wrap(err, "can't init goquery on document")
	}

	document.Find(":input[type=checkbox]").Each(func(i int, s *goquery.Selection) {
		food := makeFoodObject(s)

		if food.Date.Equal(*date) && food.ID == foodID {
			if _, ok := s.Attr("checked"); ok {
				s.RemoveAttr("checked")

				document.Find(":input[name=remainCredit]").Each(func(i int, s *goquery.Selection) {
					currentCredit, _ := s.Attr("value")
					intCredit, _ := strconv.Atoi(currentCredit)
					s.SetAttr("value", strconv.Itoa(intCredit+food.PriceTooman))
					toggled = true
				})

				s.Parent().Siblings().Children().First().SetAttr("value", "0")
				s.Parent().Siblings().Children().Find(":select").SetAttr("disabled", "true")
			} else {
				if _, ok := s.Attr("disabled"); !ok {
					s.SetAttr("checked", "true")

					document.Find(":input[name=remainCredit]").Each(func(i int, s *goquery.Selection) {
						currentCredit, _ := s.Attr("value")
						intCredit, _ := strconv.Atoi(currentCredit)
						s.SetAttr("value", strconv.Itoa(intCredit-food.PriceTooman))
						toggled = true
					})

					s.Parent().Siblings().Children().First().SetAttr("value", "1")
					s.Parent().Siblings().Children().Find(":select").RemoveAttr("disabled")
					s.Parent().Siblings().Children().Find(":option").SetAttr("selected", "true")
				}
			}
		}
	})

	if toggled {
		htmlString, _ := document.Html()
		formValues, err := extractFormInputValues(htmlString)
		if err != nil {
			return toggled, errors.Wrap(err, "can't extract form input values")
		}
		formValues.Set("method:doReserve", "Submit")
		request, err := http.NewRequest("POST", samadReservationActionURL, strings.NewReader(formValues.Encode()))
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		request.Header.Set("X-Csrf-Token", s.sessionData.csrf)
		response, err := s.httpClient.Do(request)
		defer func() {
			io.Copy(ioutil.Discard, response.Body)
			response.Body.Close()
		}()

		if err = getErrorOnPage(response.Body); err != nil {
			if samadError, ok := err.(SamadError); ok {
				return toggled, samadError
			}
			return toggled, errors.Wrap(err, "can't check for error after reservation")
		}
	}
	return toggled, nil
}

func getErrorOnPage(page io.Reader) error {
	document, err := goquery.NewDocumentFromReader(page)
	if err != nil {
		return errors.Wrap(err, "can't init goquery on document")
	}

	var samadError error
	document.Find("#errorMessages").Each(func(i int, s *goquery.Selection) {
		if len(s.Text()) > 0 {
			samadError = SamadError{
				What: s.Text(),
				When: time.Now(),
			}
		}
	})
	return samadError
}

func getMealTimeLunch(year, month, day int) *time.Time {
	jalaliDate := ptime.Date(year, ptime.Month(month), day, 10, 0, 0, 0, ptime.Iran())
	gregorianDate := jalaliDate.Time()
	return &gregorianDate
}

func getMealDate(year, month, day int, mealTime model.MealTime) *time.Time {
	var jalaliDate ptime.Time
	if mealTime == model.MealTimeLunch {
		jalaliDate = ptime.Date(year, ptime.Month(month), day, 11, 30, 0, 0, ptime.Iran())
	} else {
		jalaliDate = ptime.Date(year, ptime.Month(month), day, 19, 0, 0, 0, ptime.Iran())
	}
	gregorianDate := jalaliDate.Time()
	return &gregorianDate
}

func makeFoodObject(s *goquery.Selection) *model.Food {
	food := &model.Food{}

	// Extract meal time
	food.MealTime = model.MealTime(s.ParentsFiltered("table[align=\"center\"]").Parent().Index())

	// Extract date
	foodDate := s.ParentsFiltered("table[align=\"center\"]").Parent().
		SiblingsFiltered("td[valign=\"middle\"]").ChildrenFiltered("div").Text()
	year, _ := strconv.Atoi(foodDate[:4])
	month, _ := strconv.Atoi(foodDate[5:7])
	day, _ := strconv.Atoi(foodDate[8:10])
	food.Date = getMealDate(year, month, day, food.MealTime)

	// Extract descriptions
	foodDesc := strings.Split(strings.TrimSpace(s.SiblingsFiltered("span").Text()), " | ")
	food.Name = foodDesc[1]
	if len(foodDesc) > 2 {
		food.SideDish = foodDesc[2]
	}

	// Extract price
	stringPrice := strings.Split(strings.TrimSpace(s.SiblingsFiltered("div").Text()), " ")[0]
	intPrice, _ := strconv.Atoi(stringPrice)
	food.PriceTooman = intPrice

	// Extract status
	if _, ok := s.Attr("disabled"); ok {
		food.Status = model.FoodStatusUnavailable
	} else if _, ok := s.Attr("checked"); ok {
		food.Status = model.FoodStatusReserved
	} else {
		food.Status = model.FoodStatusReservable
	}

	// FoodID
	food.ID, _ = s.Attr("id")

	return food
}
