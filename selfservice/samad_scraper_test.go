package selfservice

import (
	"io/ioutil"
	"testing"
)

func TestFindSamadFoods(t *testing.T) {
	fileBytes, err := ioutil.ReadFile("../test/samad/reserve_notavailable.html")
	if err != nil {
		t.Fatalf("can't open test html, %v", err)
	}

	foods, err := findSamadFoods(string(fileBytes))
	if err != nil {
		t.Fatalf("can't find Samad foods, %v", err)
	}
	if len(foods) == 0 {
		t.Error("Foods list is empty!")
	}
}

func TestExtractFormInputValues(t *testing.T) {
	notavailableReserve, err := ioutil.ReadFile("../test/samad/reserve_notavailable.html")
	if err != nil {
		t.Fatalf("can't open test html, %v", err)
	}

	expectedValues := map[string]string{
		"userMessageId":                        "",
		"weekStartDateTime":                    "1508614772535",
		"remainCredit":                         "-24549",
		"selfChangeReserveId":                  "",
		"weekStartDateTimeAjx":                 "1508614772520",
		"selectedSelfDefId":                    "1",
		"userWeekReserves[0].selected":         "true",
		"userWeekReserves[0].selectedCount":    "1",
		"userWeekReserves[0].id":               "8903111",
		"userWeekReserves[0].programId":        "986113",
		"userWeekReserves[0].mealTypeId":       "2",
		"userWeekReserves[0].programDateTime":  "1508531400000",
		"userWeekReserves[0].selfId":           "1",
		"userWeekReserves[0].foodTypeId":       "517",
		"userWeekReserves[0].freeFoodSelected": "false",
		"userWeekReserves[1].selected":         "false",
		"userWeekReserves[1].selectedCount":    "0",
		"userWeekReserves[1].id":               "",
		"userWeekReserves[1].programId":        "986076",
		"userWeekReserves[1].mealTypeId":       "2",
		"userWeekReserves[1].programDateTime":  "1508531400000",
		"userWeekReserves[1].selfId":           "1",
		"userWeekReserves[1].foodTypeId":       "154",
		"userWeekReserves[1].freeFoodSelected": "false",
		"userWeekReserves[2].selected":         "false",
		"userWeekReserves[2].selectedCount":    "0",
		"userWeekReserves[2].id":               "",
		"userWeekReserves[2].programId":        "986216",
		"userWeekReserves[2].mealTypeId":       "2",
		"userWeekReserves[2].programDateTime":  "1508617800000",
		"userWeekReserves[2].selfId":           "1",
		"userWeekReserves[2].foodTypeId":       "152",
		"userWeekReserves[2].freeFoodSelected": "false",
		"userWeekReserves[3].selected":         "true",
		"userWeekReserves[3].selectedCount":    "1",
		"userWeekReserves[3].id":               "8903112",
		"userWeekReserves[3].programId":        "986253",
		"userWeekReserves[3].mealTypeId":       "2",
		"userWeekReserves[3].programDateTime":  "1508617800000",
		"userWeekReserves[3].selfId":           "1",
		"userWeekReserves[3].foodTypeId":       "153",
		"userWeekReserves[3].freeFoodSelected": "false",
		"userWeekReserves[4].selected":         "true",
		"userWeekReserves[4].selectedCount":    "1",
		"userWeekReserves[4].id":               "8903110",
		"userWeekReserves[4].programId":        "986356",
		"userWeekReserves[4].mealTypeId":       "2",
		"userWeekReserves[4].programDateTime":  "1508704200000",
		"userWeekReserves[4].selfId":           "1",
		"userWeekReserves[4].foodTypeId":       "152",
		"userWeekReserves[4].freeFoodSelected": "false",
		"userWeekReserves[5].selected":         "false",
		"userWeekReserves[5].selectedCount":    "0",
		"userWeekReserves[5].id":               "",
		"userWeekReserves[5].programId":        "986393",
		"userWeekReserves[5].mealTypeId":       "2",
		"userWeekReserves[5].programDateTime":  "1508704200000",
		"userWeekReserves[5].selfId":           "1",
		"userWeekReserves[5].foodTypeId":       "153",
		"userWeekReserves[5].freeFoodSelected": "false",
		"userWeekReserves[6].selected":         "true",
		"userWeekReserves[6].selectedCount":    "1",
		"userWeekReserves[6].id":               "8903108",
		"userWeekReserves[6].programId":        "986496",
		"userWeekReserves[6].mealTypeId":       "2",
		"userWeekReserves[6].programDateTime":  "1508790600000",
		"userWeekReserves[6].selfId":           "1",
		"userWeekReserves[6].foodTypeId":       "153",
		"userWeekReserves[6].freeFoodSelected": "false",
		"userWeekReserves[7].selected":         "false",
		"userWeekReserves[7].selectedCount":    "0",
		"userWeekReserves[7].id":               "",
		"userWeekReserves[7].programId":        "986533",
		"userWeekReserves[7].mealTypeId":       "2",
		"userWeekReserves[7].programDateTime":  "1508790600000",
		"userWeekReserves[7].selfId":           "1",
		"userWeekReserves[7].foodTypeId":       "154",
		"userWeekReserves[7].freeFoodSelected": "false",
		"userWeekReserves[8].selected":         "false",
		"userWeekReserves[8].selectedCount":    "0",
		"userWeekReserves[8].id":               "",
		"userWeekReserves[8].programId":        "986673",
		"userWeekReserves[8].mealTypeId":       "2",
		"userWeekReserves[8].programDateTime":  "1508877000000",
		"userWeekReserves[8].selfId":           "1",
		"userWeekReserves[8].foodTypeId":       "153",
		"userWeekReserves[8].freeFoodSelected": "false",
		"userWeekReserves[9].selected":         "true",
		"userWeekReserves[9].selectedCount":    "1",
		"userWeekReserves[9].id":               "8903109",
		"userWeekReserves[9].programId":        "986636",
		"userWeekReserves[9].mealTypeId":       "2",
		"userWeekReserves[9].programDateTime":  "1508877000000",
		"userWeekReserves[9].selfId":           "1",
		"userWeekReserves[9].foodTypeId":       "154",
		"userWeekReserves[9].freeFoodSelected": "false",
		"_csrf": "d6a2102a-afcd-42dd-af7a-8e623973ae32",
	}

	values, err := extractFormInputValues(string(notavailableReserve))
	if err != nil {
		t.Fatal(err)
	}
	for key, val := range expectedValues {
		if values.Get(key) != val {
			t.Errorf("extractFormInputValues() made a mistake. (%v, %v) instead of (%v, %v)",
				key, values.Get(key), key, val)
		}
	}
	if len(map[string][]string(*values)) != len(expectedValues) {
		t.Errorf("extractFormInputValues(): len aren't the same: %v, %v", len(expectedValues),
			len(map[string][]string(*values)))
	}
}
