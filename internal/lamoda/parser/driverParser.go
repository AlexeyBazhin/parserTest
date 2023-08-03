package parser

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/tebeka/selenium"
)

type (
	LamodaGoodInfo struct {
		Name         string            `json:"name"`
		Brand        string            `json:"brand"`
		CurrentPrice string            `json:"current_price"`
		OldPrice     string            `json:"old_price"`
		Description  string            `json:"description"`
		Attributes   map[string]string `json:"attributes"`
	}
)

func WriteFileLamodaDriver(filename string, goods []LamodaGoodInfo) error {
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

func ParseGoodsByDriver(driver selenium.WebDriver, goodsIds []string) {
	generalStart := time.Now()

	baseUrl := "https://www.lamoda.ru/p/"

	goodsInfo := make([]LamodaGoodInfo, len(goodsIds))
	for i, v := range goodsIds {
		goodStart := time.Now()
		getStart := time.Now()
		driver.Get(strings.Join([]string{baseUrl, v}, "")) // перейдет по артиклу и сам редиректнет куда надо
		// time.Sleep(300 * time.Millisecond)
		// if _, err := driver.ExecuteScript("window.scrollTo(0, document.body.scrollHeight)", nil); err != nil {
		// 	panic(err)
		// }
		fmt.Println(i, goodsIds[i])
		fmt.Println("Get url: ", time.Since(getStart))

		// брэнд товара
		brandStart := time.Now()
		if element, err := driver.FindElement(selenium.ByXPATH, "//span[contains(@class, 'product-title')]"); err != nil {
			panic(err)
		} else {
			goodsInfo[i].Brand, err = element.Text()
			if err != nil {
				panic(err)
			}
		}
		fmt.Printf("Brand: %v\n", time.Since(brandStart))

		// название товара
		nameStart := time.Now()
		if element, err := driver.FindElement(selenium.ByXPATH, "//div[contains(@class, '_modelName')]"); err != nil {
			panic(err) // пока панику выкидываю
		} else {
			goodsInfo[i].Name, err = element.Text()
			if err != nil {
				panic(err)
			}
		}
		fmt.Printf("Name: %v\n", time.Since(nameStart))

		// цена товара
		priceStart := time.Now()
		if elements, err := driver.FindElements(selenium.ByXPATH, "//div[contains(@class, '_prices')]//span[contains(@class, '_price')]"); err != nil {
			panic(err)
		} else {
			for _, v := range elements {
				// не важно есть скидка на товар или нет - всегда будет данный атрибут - это наша текущая цена
				if attributeText, err := v.GetAttribute("aria-label"); err == nil && attributeText == "Итоговая цена" {
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
		fmt.Printf("Price: %v\n", time.Since(priceStart))

		// описание товара
		descriptionStart := time.Now()
		descriptionElement, err := driver.FindElement(selenium.ByXPATH, "//div[contains(@class, '_content')]//div[contains(@class, '_root')]//div[contains(@class, '_description')]")
		if err != nil {
			fmt.Println("Нет описания" + goodsIds[i]) // для дебага (не паникую потому что не у всех товаров есть описание)
		} else {
			if moreButton, err := descriptionElement.FindElement(selenium.ByTagName, "a"); err != nil { // ищем кнопку "Подробнее" у описания товара
				fmt.Println("Нет кнопки подробнее у описания" + goodsIds[i])
			} else {
				if err = moreButton.Click(); err != nil {
					panic(err)
				}
			}
			// time.Sleep(100 * time.Millisecond)

			if descriptionSpan, err := descriptionElement.FindElement(selenium.ByTagName, "span"); err != nil {
				panic(err)
			} else {
				goodsInfo[i].Description, err = descriptionSpan.Text()
				fmt.Println(goodsInfo[i].Description)
				if err != nil {
					panic(err)
				}
			}
		}
		fmt.Printf("Description: %v\n", time.Since(descriptionStart))

		// атрибуты товара

		// rootElement, err := driver.FindElement(selenium.ByClassName, "_root_1ga1h_2")
		// if err != nil {
		// 	panic(err)
		// }
		// if moreButton, err := rootElement.FindElement(selenium.ByClassName, "_root_f9xmk_2"); err != nil {
		// 	fmt.Println("Нет кнопки подробнее у атрибутов" + goodsIds[i])
		// } else {
		// 	if err = moreButton.Click(); err != nil {
		// 		panic(err) // иногда сюда прилетает потому что DOM-дерево не перестроилось и он берет первую кнопку подробнее (а не которую нам надо) - а там уже текст
		// 	}
		// }

		time.Sleep(200 * time.Millisecond)
		if moreButton, err := driver.FindElement(selenium.ByXPATH, "//div[contains(@class, '_content')]//div[contains(@class, '_root')]//a[contains(@role, 'button')]"); err != nil {
			panic(err)
		} else {
			if err = moreButton.Click(); err != nil {
				panic(err)
			}
		}
		// time.Sleep(100 * time.Millisecond)
		attrStart := time.Now()
		goodsInfo[i].Attributes = make(map[string]string)
		// if attributes, err := driver.FindElements(selenium.ByXPATH, "//p[starts-with(@class, '_item')]"); err != nil {
		// 	panic(err)
		// } else {
		// 	for _, attr := range attributes {
		// 		var attrName, attrValue string

		// 		if attrNameElement, err := attr.FindElement(selenium.ByXPATH, "_attributeName_ajirn_14"); err != nil {
		// 			panic(err)
		// 		} else {
		// 			if attrName, err = attrNameElement.Text(); err != nil {
		// 				panic(err)
		// 			}
		// 		}

		// 		if attrValueElement, err := attr.FindElement(selenium.ByClassName, "_value_ajirn_27"); err != nil {
		// 			panic(err)
		// 		} else {
		// 			if attrValue, err = attrValueElement.Text(); err != nil {
		// 				panic(err)
		// 			}
		// 		}
		// 		fmt.Println(attrName, attrValue)
		// 		goodsInfo[i].Attributes[attrName] = attrValue
		// 	}
		// }
		attributeNames, err := driver.FindElements(selenium.ByXPATH, "//p[contains(@class, '_item')]//span[contains(@class, '_attributeName')]")
		if err != nil {
			panic(err)
		}
		attributeValues, err := driver.FindElements(selenium.ByXPATH, "//p[contains(@class, '_item')]//span[contains(@class, '_value')]")
		if err != nil {
			panic(err)
		}
		for j, attrName := range attributeNames {
			name, err := attrName.Text()
			if err != nil {
				panic(err)
			}
			value, err := attributeValues[j].Text()
			if err != nil {
				panic(err)
			}
			fmt.Println(name, value)
			goodsInfo[i].Attributes[name] = value
		}
		fmt.Printf("Attrs: %v\n", time.Since(attrStart))

		fmt.Printf("Товар(%v): %v\n\n", i, time.Since(goodStart))
		// fmt.Println(goodsInfo[i])
	}
	fmt.Printf("Итоговое время: %v\n", time.Since(generalStart))

	err := WriteFileLamodaDriver("data.json", goodsInfo)
	if err != nil {
		panic(err)
	}
}
