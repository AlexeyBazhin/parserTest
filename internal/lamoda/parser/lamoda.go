package parser

import (
	"fmt"
	"strings"

	"github.com/tebeka/selenium"
)

const (
	lamodaSearchURL = "https://www.lamoda.ru/catalogsearch/result/?q="
)

func GetGoodsByRequest(driver selenium.WebDriver, request string) []string {
	normalizedRequest := strings.TrimSpace(request)
	normalizedRequest = strings.ReplaceAll(normalizedRequest, " ", "+")
	url := strings.Join([]string{lamodaSearchURL, normalizedRequest}, "")

	fmt.Println(url)

	err := driver.Get(url)
	if err != nil {
		panic(err)
	}
	// time.Sleep(100 * time.Millisecond)

	// "window.scrollTo(5,4000)"
	// "window.scrollBy(0,350)"
	// прокручиваем страницу, чтоб динамический контент подгрузился
	if _, err := driver.ExecuteScript("window.scrollTo(0, document.body.scrollHeight)", nil); err != nil {
		panic(err)
	}

	// ищи что-нибудь в поиске на ламоде и смотри все указанные классы у тэгов
	elements, err := driver.FindElements(selenium.ByXPATH, "//div[starts-with(@class, 'x-product-card__card')]")
	if err != nil {
		panic(err)
	}
	goodsId := make([]string, 0)
	for _, v := range elements {
		id, err := v.GetAttribute("id") // я изначально собирал артикулы товаров, а потом на их основе переходил на саму страницу товара (будет дальше по коду), но надо переделать и собирать ссылки сразу
		if err != nil {
			panic(err)
		}
		goodsId = append(goodsId, id)
		fmt.Println(v.GetAttribute("id"))
	}

	return goodsId
}
