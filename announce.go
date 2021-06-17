package main

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type GeneralInfo struct {
	Organization string
	Address      string
}

type OrganizatorInfo struct {
	Name  string
	Email string
	Phone string
}

type Lot struct {
	Name         string
	ExtendedInfo string
	PricePerUnit string
	Amount       string
}

type Announce struct {
	URL             string
	Date            time.Time
	GeneralInfo     GeneralInfo
	OrganizatorInfo OrganizatorInfo
	Lots            []Lot
}

const (
	base    = "https://www.goszakup.gov.kz"
	search  = base + "/ru/search/lots?filter[name]=ноутбук"
	lotsTab = "?tab=lots"
)

var (
	client = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
)

func parsePage(url string) (*goquery.Document, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status)
	}

	// Load the HTML document
	return goquery.NewDocumentFromReader(res.Body)
}

func newAnnounce(link string) (Announce, error) {
	url := base + link
	announce := Announce{
		URL: url,
	}

	doc, err := parsePage(url)
	if err != nil {
		return announce, err
	}

	doc.Find("th").Each(func(_ int, s *goquery.Selection) {
		key := s.Text()
		val := getSiblingText(s)
		announce.Set(key, val)
	})

	return announce, nil
}

func getLots(link string) ([]Lot, error) {
	doc, err := parsePage(base + link + lotsTab)
	if err != nil {
		return nil, err
	}

	lots := make([]Lot, 0)
	doc.Find("tbody").Children().Each(func(i int, s *goquery.Selection) {
		lot := Lot{}
		s.Children().Each(func(j int, s *goquery.Selection) {
			val := s.Text()
			lot.Set(j, val)
		})
		if lot.Name != "" && lot.Name != "Наименование" {
			lots = append(lots, lot)
		}
	})

	return lots, nil
}

func getSiblingText(s *goquery.Selection) (str string) {
	s.Siblings().Each(func(i int, s *goquery.Selection) {
		str = s.Text()
	})

	return
}

func (announce *Announce) Set(key, val string) {
	switch key {
	case "Организатор":
		announce.GeneralInfo.Organization = val
	case "Юр. адрес организатора":
		announce.GeneralInfo.Address = val
	case "ФИО представителя":
		announce.OrganizatorInfo.Name = val
	case "Контактный телефон":
		announce.OrganizatorInfo.Phone = val
	case "E-Mail":
		announce.OrganizatorInfo.Email = val
	}
}

func (lot *Lot) Set(key int, val string) {
	switch key {
	case 3:
		lot.Name = val
	case 4:
		lot.ExtendedInfo = val
	case 5:
		lot.PricePerUnit = val
	case 6:
		lot.Amount = val
	}
}

func getLatestAnnouncesLinks() ([]string, error) {
	links := make([]string, 0)

	doc, err := parsePage(search)
	if err != nil {
		return nil, err
	}

	doc.Find("a[href]").Each(func(_ int, s *goquery.Selection) {
		link, _ := s.Attr("href")
		if strings.Contains(link, "announce/index") {
			links = append(links, link)
		}
	})

	return links, nil
}

func extractNew(links []string) []string {
	newLinks := make([]string, 0)
	for _, link := range links {
		if !defaultStorage.Exists(link) {
			newLinks = append(newLinks, link)
			defaultStorage.Store(link)
		}
	}

	return newLinks
}

func getNewAnnounces() ([]Announce, error) {
	announcesLinks, err := getLatestAnnouncesLinks()
	if err != nil {
		return nil, err
	}

	announces := make([]Announce, 0)

	announcesLinks = extractNew(announcesLinks)

	for _, link := range announcesLinks {
		announce, err := newAnnounce(link)
		if err != nil {
			return nil, err
		}

		announce.Lots, err = getLots(link)
		if err != nil {
			return nil, err
		}

		announces = append(announces, announce)
	}

	return announces, nil
}

func (announce *Announce) String() string {
	organizator := fmt.Sprintf("Организатор: \n%v\n", announce.OrganizatorInfo.String())
	generalInfo := fmt.Sprintf("Обзщая информация: \n%v\n", announce.GeneralInfo.String())

	lots := "Лоты: \n"
	for _, lot := range announce.Lots {
		lots += lot.String()
	}

	return organizator + generalInfo + lots
}

func (oi *OrganizatorInfo) String() string {
	return fmt.Sprintf("  %v\n  %v\n  %v\n", oi.Name, oi.Email, oi.Phone)
}

func (gi *GeneralInfo) String() string {
	return fmt.Sprintf("  Организация:%v\n  Адрес: %v\n", gi.Organization, gi.Address)
}

func (l *Lot) String() string {
	res := "\n"
	res += fmt.Sprintf("  Наименование: %v\n", l.Name)
	res += fmt.Sprintf("  Цена: %v\n", l.PricePerUnit)
	res += fmt.Sprintf("  Количество: %v\n", l.Amount)
	res += fmt.Sprintf("  Описание: %v\n", l.ExtendedInfo)
	return res
}
