package selfservice

import (
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/aryahadii/sarioself/model"
	"github.com/pkg/errors"
)

// findSamadFoods creates a list of all foods in Samad's HTML file
func findSamadFoods(samadPage string) ([]*model.Food, error) {
	var foods []*model.Food
	document, err := goquery.NewDocumentFromReader(strings.NewReader(samadPage))
	if err != nil {
		return foods, errors.Wrap(err, "can't init goquery on document")
	}

	document.Find(":input[type=checkbox]").Each(func(i int, s *goquery.Selection) {
		foods = append(foods, makeFoodObject(s))
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
	document.Find(":input[type=checkbox]").Each(func(i int, s *goquery.Selection) {
		if name, got := s.Attr("name"); got {
			if _, got := s.Attr("checked"); got {
				values.Set(name, "true")
			}
		}
	})
	document.Find("select").Each(func(i int, s *goquery.Selection) {
		if name, got := s.Attr("name"); got {
			if _, got := s.Parent().Siblings().Children().First().Attr("checked"); got {
				values.Set(name, "1")
			}
		}
	})

	return values, nil
}
