package selfservice

import (
	"net/url"
	"strings"

	"strconv"

	"github.com/PuerkitoBio/goquery"
	"github.com/aryahadii/sarioself/model"
	"github.com/pkg/errors"
	"github.com/yaa110/go-persian-calendar/ptime"
)

// findSamadFoods creates a list of all foods in Samad's HTML file
func findSamadFoods(samadPage string) ([]*model.Food, error) {
	var foods []*model.Food
	document, err := goquery.NewDocumentFromReader(strings.NewReader(samadPage))
	if err != nil {
		return foods, errors.Wrap(err, "can't init goquery on document")
	}

	document.Find(":input[type=checkbox]").Each(func(i int, s *goquery.Selection) {
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
		food.PriceTooman = strings.Split(strings.TrimSpace(s.SiblingsFiltered("div").Text()), " ")[0]

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

		foods = append(foods, food)
	})

	return foods, nil
}

// extractFormInputValues returns form values which are needed to get next page
// in Samad reservation page
func extractFormInputValues(samadPage string) (*url.Values, error) {
	values := &url.Values{}

	document, err := goquery.NewDocumentFromReader(strings.NewReader(samadPage))
	if err != nil {
		return nil, errors.Wrap(err, "can't init goquery on document")
	}
	document.Find(":input[type=hidden]").Each(func(i int, s *goquery.Selection) {
		if name, got := s.Attr("name"); got {
			if val, got := s.Attr("value"); got {
				values.Set(name, val)
			}
		}
	})

	return values, nil
}
