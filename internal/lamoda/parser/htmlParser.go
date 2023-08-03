package parser

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/tebeka/selenium"
	"golang.org/x/net/html"
)

func WriteFileLamodaHtml(filename string, goods []map[string]any) error {
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

func htmlParse(pageSource string) map[string]any {
	startIndex := strings.Index(pageSource, "payload: ")
	endIndex := strings.Index(pageSource, "LMDA.pageState")

	payload := pageSource[startIndex+len("payload: ") : endIndex]

	payloadStart := strings.Index(payload, "payload: ")
	payloadEnd := strings.Index(payload, "settings:")
	payloadStr := payload[payloadStart+len("payload: ") : payloadEnd]
	payloadStr = payloadStr[:strings.LastIndex(payloadStr, "}")+1]
	// fmt.Println(payloadStr)
	result := make(map[string]any)
	if err := json.Unmarshal([]byte(payloadStr), &result); err != nil {
		panic(err)
	}

	return result
}

func ParseGoodByHtml(driver selenium.WebDriver, goodId string) map[string]any {
	baseUrl := "https://www.lamoda.ru/p/"

	goodStart := time.Now()
	getStart := time.Now()
	driver.Get(strings.Join([]string{baseUrl, goodId}, "")) // перейдет по артиклу и сам редиректнет куда надо

	fmt.Println("+", goodId)
	fmt.Println("Get url: ", time.Since(getStart))

	pageSource, err := driver.PageSource()
	if err != nil {
		panic(err)
	}

	// err = WriteFilePageSource(goodsIds[i], pageSource)
	// if err != nil {
	// 	panic(err)
	// }

	goodMap := htmlParse(pageSource)

	fmt.Printf("Товар(%v): %v\n\n", goodId, time.Since(goodStart))

	return goodMap
}

func TestingParseGoodsByHtml(driver selenium.WebDriver, goodsIds []string) {
	generalStart := time.Now()

	baseUrl := "https://www.lamoda.ru/p/"

	goodsMap := make([]map[string]any, len(goodsIds))
	for i, v := range goodsIds {
		goodStart := time.Now()
		getStart := time.Now()
		driver.Get(strings.Join([]string{baseUrl, v}, "")) // перейдет по артиклу и сам редиректнет куда надо

		fmt.Println(i, goodsIds[i])
		fmt.Println("Get url: ", time.Since(getStart))

		pageSource, err := driver.PageSource()
		if err != nil {
			panic(err)
		}

		// err = WriteFilePageSource(goodsIds[i], pageSource)
		// if err != nil {
		// 	panic(err)
		// }

		goodMap := htmlParse(pageSource)
		goodsMap[i] = goodMap

		fmt.Printf("Товар(%v): %v\n\n", i, time.Since(goodStart))
	}
	fmt.Printf("Итоговое время: %v\n", time.Since(generalStart))

	err := WriteFileLamodaHtml("data2.json", goodsMap)
	if err != nil {
		panic(err)
	}
}

func WriteFilePageSource(filename, pageSource string) error {
	file, err := os.Create(filename + ".html")
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = html.Parse(strings.NewReader(pageSource))
	if err != nil {
		return err
	}
	_, err = file.Write([]byte(pageSource))
	if err != nil {
		return err
	}
	return nil
}
