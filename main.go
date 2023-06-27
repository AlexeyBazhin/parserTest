package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/chrome"
)

type (
	Env     string
	Browser string

	LamodaGoodInfo struct {
		Name         string            `json:"name"`
		Brand        string            `json:"brand"`
		CurrentPrice string            `json:"current_price"`
		OldPrice     string            `json:"old_price"`
		Description  string            `json:"description"`
		Attributes   map[string]string `json:"attributes"`
	}
)

const (
	DevEnd = "dev"

	Firefox Browser = "firefox"
	Chrome  Browser = "chrome"

	chromeDriverPath = "/usr/local/bin/chromedriver" // это все дело потом в конфиг
	geckoDriverPath  = "/usr/local/bin/geckodriver"

	lamodaSearchURL = "https://www.lamoda.ru/catalogsearch/result/?q="
)

func WriteFileLamoda(filename string, goods []LamodaGoodInfo) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	data, _ := json.MarshalIndent(goods, "", " ")
	_, err = file.Write(data)
	if err != nil {
		return err
	}
	return nil
}

func main() {
	// Run Chrome browser
	service, err := selenium.NewChromeDriverService(chromeDriverPath, 4444)
	// драйвер для firefox, но у меня так и не запустился
	// service, err := selenium.NewGeckoDriverService(geckoDriverPath, 4444)
	if err != nil {
		panic(err)
	}
	defer service.Stop()

	caps := selenium.Capabilities{}
	caps.AddChrome(chrome.Capabilities{Args: []string{
		"window-size=1920x1080",
		// "--headless",  // режим без запуска отдельного окна браузера (GUI)
		//"--no-sandbox",

		//это пока хз - копировал со stackoverflow)
		//"--disable-dev-shm-usage",
		//"disable-gpu",
	}})

	// для гугла и мозилы как-то по разному нужно capabilities указывать, нужно лучше посмотреть
	// caps := selenium.Capabilities{
	// 	"browserName": "firefox",
	// }
	// caps.AddFirefox(firefox.Capabilities{Args: []string{
	// 	fmt.Sprintf("--width=%d", 1920),
	// 	fmt.Sprintf("--height=%d", 1080),
	// 	// "--headless",  // comment out this line to see the browser
	// }})

	// driver, err := selenium.NewRemote(caps, fmt.Sprintf("http://localhost:%d", 4444))
	driver, err := selenium.NewRemote(caps, "")
	if err != nil {
		panic(err)
	}

	GetLamoda(driver, "рюкзак желтый")
}

func GetLamoda(driver selenium.WebDriver, request string) {
	normalizedRequest := strings.TrimSpace(request)
	normalizedRequest = strings.ReplaceAll(normalizedRequest, " ", "+")
	url := strings.Join([]string{lamodaSearchURL, normalizedRequest}, "")

	err := driver.Get(url)
	if err != nil {
		panic(err)
	}
	time.Sleep(1 * time.Second)

	// "window.scrollTo(5,4000)"
	// "window.scrollBy(0,350)"
	// прокручиваем страницу, чтоб динамический контент подгрузился
	if _, err := driver.ExecuteScript("window.scrollTo(0, document.body.scrollHeight)", nil); err != nil {
		panic(err)
	}

	// ищи что-нибудь в поиске на ламоде и смотри все указанные классы у тэгов
	elements, err := driver.FindElements(selenium.ByClassName, "x-product-card__card")
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

	parsedGoods := ParseGoodsLamoda(driver, goodsId) // чтоб долго не ждать - goodsId[:n]
	err = WriteFileLamoda("data.json", parsedGoods)

	// driver.PageSource()
}

func ParseGoodsLamoda(driver selenium.WebDriver, goodsIds []string) []LamodaGoodInfo {
	baseUrl := "https://www.lamoda.ru/p/"

	goodsInfo := make([]LamodaGoodInfo, len(goodsIds))
	for i, v := range goodsIds {
		driver.Get(strings.Join([]string{baseUrl, v}, "")) // перейдет по артиклу и сам редиректнет куда надо
		time.Sleep(500 * time.Millisecond)

		// название товара
		if element, err := driver.FindElement(selenium.ByClassName, "_modelName_rumwo_22"); err != nil {
			panic(err) // пока панику выкидываю
		} else {
			goodsInfo[i].Name, err = element.Text()
			if err != nil {
				panic(err)
			}
		}

		// брэнд товара
		if element, err := driver.FindElement(selenium.ByClassName, "product-title__brand-name"); err != nil {
			panic(err)
		} else {
			goodsInfo[i].Brand, err = element.Text()
			if err != nil {
				panic(err)
			}
		}

		// цена товара
		if elements, err := driver.FindElements(selenium.ByClassName, "_price_1gga1_7"); err != nil {
			panic(err)
		} else {
			for _, v := range elements {
				attributeText, err := v.GetAttribute("aria-label") // не важно есть скидка на товар или нет - всегда будет данный атрибут - это наша текущая цена
				if err == nil && attributeText == "Итоговая цена" {
					goodsInfo[i].CurrentPrice, err = v.Text()
					if err != nil {
						panic(err)
					}
					continue
				}
				// атрибута нет, но элемент есть -> это старая цена
				goodsInfo[i].OldPrice, err = v.Text()
				if err != nil {
					panic(err)
				}
			}
		}

		// описание товара
		descriptionElement, err := driver.FindElement(selenium.ByClassName, "_description_1ga1h_20")
		if err != nil {
			fmt.Println("Нет описания" + goodsIds[i]) // для дебага (не паникую потому что не у всех товаров есть описание)
		} else {
			if moreButton, err := descriptionElement.FindElement(selenium.ByClassName, "_root_f9xmk_2"); err != nil { // ищем кнопку "Подробнее" у описания товара
				fmt.Println("Нет кнопки подробнее у описания" + goodsIds[i])
			} else {
				if err = moreButton.Click(); err != nil {
					panic(err)
				}
			}

			time.Sleep(500 * time.Millisecond) // подгружается полное описание

			if descriptionSpan, err := descriptionElement.FindElement(selenium.ByTagName, "span"); err != nil {
				panic(err)
			} else {
				goodsInfo[i].Description, err = descriptionSpan.Text()
				if err != nil {
					panic(err)
				}
			}
		}

		// атрибуты товара
		rootElement, err := driver.FindElement(selenium.ByClassName, "_root_1ga1h_2")
		if err != nil {
			panic(err)
		}
		if moreButton, err := rootElement.FindElement(selenium.ByClassName, "_root_f9xmk_2"); err != nil {
			fmt.Println("Нет кнопки подробнее у атрибутов" + goodsIds[i])
		} else {
			if err = moreButton.Click(); err != nil {
				panic(err)
			}
		}

		goodsInfo[i].Attributes = make(map[string]string)
		if attributes, err := driver.FindElements(selenium.ByClassName, "_item_ajirn_2"); err != nil {
			panic(err)
		} else {
			for _, attr := range attributes {
				var attrName, attrValue string

				if attrNameElement, err := attr.FindElement(selenium.ByClassName, "_attributeName_ajirn_14"); err != nil {
					panic(err)
				} else {
					if attrName, err = attrNameElement.Text(); err != nil {
						panic(err)
					}
				}

				if attrValueElement, err := attr.FindElement(selenium.ByClassName, "_value_ajirn_27"); err != nil {
					panic(err)
				} else {
					if attrValue, err = attrValueElement.Text(); err != nil {
						panic(err)
					}
				}
				// fmt.Println(j, attrName, attrValue)
				goodsInfo[i].Attributes[attrName] = attrValue
			}
		}

		// fmt.Println(goodsInfo[i])
	}
	return goodsInfo
}
