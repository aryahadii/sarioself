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

func (s *SamadAUTClient) toggleFoodReservation(samadPage string, date *time.Time, foodID string) error {
	document, err := goquery.NewDocumentFromReader(strings.NewReader(samadPage))
	if err != nil {
		return errors.Wrap(err, "can't init goquery on document")
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
					})

					s.Parent().Siblings().Children().First().SetAttr("value", "1")
					s.Parent().Siblings().Children().Find(":select").RemoveAttr("disabled")
					s.Parent().Siblings().Children().Find(":option").SetAttr("selected", "true")
				}
			}
		}
	})

	htmlString, _ := document.Html()
	formValues, err := extractFormInputValues(htmlString)
	if err != nil {
		return errors.Wrap(err, "can't extract form input values")
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

	return nil
}

func makeFoodObject(s *goquery.Selection) *model.Food {
	food := &model.Food{}

	// Extract date
	foodDate := s.ParentsFiltered("table[align=\"center\"]").Parent().
		SiblingsFiltered("td[valign=\"middle\"]").ChildrenFiltered("div").Text()
	year, _ := strconv.Atoi(foodDate[:4])
	month, _ := strconv.Atoi(foodDate[5:7])
	day, _ := strconv.Atoi(foodDate[8:10])
	jalaliDate := ptime.Date(year, ptime.Month(month), day, 10, 0, 0, 0, ptime.Iran())
	gregorianDate := jalaliDate.Time()
	food.Date = &gregorianDate

	// Extract descriptions
	foodDesc := strings.Split(strings.TrimSpace(s.SiblingsFiltered("span").Text()), " | ")
	food.Name = foodDesc[1]
	food.SideDish = foodDesc[2]

	// Extract price
	stringPrice := strings.Split(strings.TrimSpace(s.SiblingsFiltered("div").Text()), " ")[0]
	intPrice, _ := strconv.Atoi(stringPrice)
	food.PriceTooman = intPrice

	// Extract meal time
	food.MealTime = model.MealTime(s.ParentsFiltered("table[align=\"center\"]").Parent().Index() - 1)

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
