package parser

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"sync"
	"time"

	"parserTest/internal/common/awsS3"
	"parserTest/internal/lamoda/config"

	"github.com/tebeka/selenium"
)

const (
	bucketName      = "wb-spider-lamoda-test"
	lastSkuFilename = "last_sku.json"
	skuLen          = 12
	lamodaApiURL    = "https://www.lamoda.ru/api/v1/product/get?sku="
)

type (
	Parser struct {
		Cfg        *config.Config
		AwsClients []*awsS3.S3

		Drivers []selenium.WebDriver
	}
)

func New(cfg *config.Config, awsClients []*awsS3.S3, drivers []selenium.WebDriver) *Parser {
	b := Parser{
		Cfg:        cfg,
		AwsClients: awsClients,
		Drivers:    drivers,
	}

	return &b
}

var (
	skuChan     chan string
	goodMapChan chan map[string]any
	goodWg      *sync.WaitGroup
	workersWg   *sync.WaitGroup
	lastSku     map[string]string
	skuPatterns []string
	countParsed int64
	mu          *sync.Mutex
)

func ReadLastSku() (err error) {
	file, err := os.Open(lastSkuFilename)
	if err == nil {
		err = json.NewDecoder(file).Decode(&lastSku)
	}
	return
}

func SaveLastSku() (err error) {
	file, err := os.Create(lastSkuFilename)
	if err != nil {
		return err
	}
	defer file.Close()
	data, _ := json.MarshalIndent(lastSku, "", " ")
	_, err = file.Write(data)
	if err != nil {
		return err
	}
	return nil
}

func ValidateLastSku() {
	for _, pattern := range skuPatterns {
		if _, ok := lastSku[pattern]; !ok {
			log.Printf("Последнее обработанное значение для паттерна артикула %v не найдено \n", pattern)
			newSku := ""
			for i := len(pattern); i < skuLen; i++ {
				newSku += "A"
			}
			lastSku[pattern] = newSku
		}
	}
}

func (p *Parser) ParseLamodaBySku(ctx context.Context) error {
	skuChan = make(chan string, 10)
	goodMapChan = make(chan map[string]any)
	goodWg = &sync.WaitGroup{}
	workersWg = &sync.WaitGroup{}
	mu = &sync.Mutex{}
	lastSku = make(map[string]string)

	skuPatterns = []string{
		"MP002XW0", "MP002XW1", "MP002XM0", "MP002XM1", "MP002XM2",
		"MP002XU", "MP002XG", "MP002XB", "RTLA",
		"SA032AU", "AD093FU",
	}

	userAgents := []string{
		"Mozilla/5.0 (Android 4.3; Mobile; rv:54.0) Gecko/54.0 Firefox/54.0",
		"Mozilla/5.0 (Linux; Android 4.3; GT-I9300 Build/JSS15J) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/55.0.2883.91 Mobile Safari/537.36 OPR/42.9.2246.119956",
		"Opera/9.80 (Android; Opera Mini/28.0.2254/66.318; U; en) Presto/2.12.423 Version/12.16",
	}

	if err := ReadLastSku(); err != nil {
		log.Printf("[READ-LAST-SKU] Ошибка при чтении файла последних sku: %q\n", err)
	}
	ValidateLastSku()

	quit := make(chan struct{})
	for _, pattern := range skuPatterns {
		go getSku(pattern, quit)
	}

	for _, userAgent := range userAgents {
		workersWg.Add(1)
		go startWorker(userAgent)
		time.Sleep(4 * time.Second) // fix
	}

	go func() {
		for goodMapFull := range goodMapChan {
			sku, ok := goodMapFull["sku"].(string)
			if !ok {
				log.Printf("[GOOD-MAP]: не удалось получить sku\n")
				goodWg.Done()
				continue
			}
			goodMap, ok := goodMapFull["payload"].(map[string]any)
			if !ok {
				log.Printf("[GOOD-MAP]: не удалось получить payload\n")
				goodWg.Done()
				continue
			}
			if err := p.AwsClients[0].PushJSON(bucketName, sku, goodMap); err != nil {
				log.Printf("[GOOD-MAP][PUSH-JSON]: %q\n", err)
				goodWg.Done()
				continue
			}

			mu.Lock()
			countParsed++
			log.Printf("\n[GOOD-MAP]: %v) %s\n\n", countParsed, sku)
			mu.Unlock()
			goodWg.Done()
		}
	}()

	// for {
	// 	select {
	// 	case <-ctx.Done():
	// 		workersWg.Wait()
	// 	default:
	// 		//goodsWg.Add(1)
	// 		goodMapChan <- GetGoodAPI()
	// 	}
	// }
	quit <- <-ctx.Done()
	close(skuChan)
	if err := SaveLastSku(); err != nil {
		log.Printf("[SAVE-LAST-SKU]: %q\n", err)
	}
	workersWg.Wait()
	close(goodMapChan)
	goodWg.Wait()

	return nil
}

func getSku(pattern string, quit chan struct{}) {
	sku := lastSku[pattern]
	count := 0
	for {
		select {
		case <-quit:
			lastSku[pattern] = sku
			return
		default:
			sku = generateSku(sku)
			skuChan <- pattern + sku
			runtime.Gosched()
			count++
			if count%10 == 0 {
				mu.Lock()
				lastSku[pattern] = sku
				if err := SaveLastSku(); err != nil {
					log.Printf("[SAVE-LAST-SKU]: %q\n", err)
				}
				mu.Unlock()
			}
		}
	}
}

func generateSku(prevSku string) string {
	newSku := prevSku
	for i := 0; i < len(newSku); i++ {
		left := ""
		right := ""
		if i != 0 {
			left = newSku[:i]
		}
		if i != len(newSku)-1 {
			right = newSku[i+1:]
		}

		switch string(newSku[i]) {
		case "Z":
			newSku = left + "0" + right
			return newSku
		case "9":
			newSku = left + "A" + right
			continue
		default:
			newSku = left + string(newSku[i]+1) + right
			return newSku
		}
	}
	return newSku
}

func startWorker(userAgent string) {
	log.Printf("Воркер под UserAgent %s начал работу\n", userAgent[:20])
	for sku := range skuChan {

		if goodMap, err := getGoodAPI(userAgent, sku); err != nil {
			log.Printf("[WORKER]: %q\n", err) // TODO кастомные ошибки
		} else {
			fullMap := make(map[string]any)
			fullMap["sku"] = sku
			fullMap["payload"] = goodMap
			goodWg.Add(1)
			goodMapChan <- fullMap
		}
		log.Printf("Воркер %s обработал sku: %v\n", userAgent[:20], sku)
		runtime.Gosched()
		time.Sleep(3 * time.Second)
	}
	workersWg.Done()
	log.Printf("Воркер под номером %v завершил работу\n", userAgent[:12])
}

func getGoodAPI(userAgent string, sku string) (map[string]any, error) {
	req, err := http.NewRequest(http.MethodGet, lamodaApiURL+sku, nil)
	if err != nil {
		return nil, fmt.Errorf("[GetRequest]: client: не удалось создать запрос: %q\n", err)
	}
	req.Header.Set("User-Agent", userAgent)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == 400 {
		return nil, fmt.Errorf("[GetRequest]: Status code 400\n")
	}
	defer resp.Body.Close()
	good := make(map[string]any)
	if err := json.NewDecoder(resp.Body).Decode(&good); err != nil {
		return nil, fmt.Errorf("[GetRequest]: не удалось спарсить в JSON: %q\n", err)
	}

	return good, nil
}

func (p *Parser) ParseLamodaByRequest() error {
	// if err := p.AwsClients[0].CreateBucket(bucketName); err != nil {
	// 	return err
	// }

	if err := p.AwsClients[0].ListBucketsWithLocation(); err != nil {
		return err
	}
	p.AwsClients[0].ListObjects(bucketName)

	//рюкзак желтый", "платье женское", "лоферы женские", "куртка женская", "лоферы мужские", "куртка мужская",
	//"шорты мужские", "шорты женские", "джинсы женские", "джинсы мужские", "сапоги женские", "сапоги мужские",
	//"кроссовки женские", "кроссовки мужские", "футболки мужские", "футболки женские", "леггинсы", "майки", "топики",
	// "нижнее белье", "носки мужские", "рубашки мужские", "блузки женские", "сумка женская", "шоппер"
	// "брюки женские", "брюки мужские", "кеды летние мужские", "кеды летние женские", "пуховик мужской", "пуховик женский",
	// 	"косуха мужская", "косуха женская", "сандалии женские", "сандалии мужские", "шапка женская", "шапка мужская", "часы мужские", "часы женские",
	// 	"юбка женская", "свитшот женский", "свитшот мужской", "шуба женская", "колготки женские", "солнечные очки женские",
	// 	"солнечные очки мужские", "ветровка мужская", "ветровка женская", "пиджак мужской", "пиджак женский",
	// 	"худи женские", "худи мужские", "туфли женские", "туфли мужские", "спортивные костюмы женские", "спортивные костюмы мужские",
	// 	"свитер женский", "свитер мужской", "поло мужская", "поло женская",
	// 	"купальник женский", "пальто женское", "пальто мужское", "тренч женский", "тренч мужской", "тушь для ресниц", "спортивная сумка мужская", "рюкзак мужской",
	// 	"шампунь", "кофта мальчик", "кофта девочка", "кроссовки мальчик", "кроссовки девочка", "плед", "постельное белье",

	requests := []string{}

	for i, request := range requests {
		goodsId := GetGoodsByRequest(p.Drivers[0], request)
		for j, goodId := range goodsId {

			goodMap := ParseGoodByHtml(p.Drivers[0], goodId)
			if err := p.AwsClients[0].PushJSON(bucketName, goodId, goodMap); err != nil {
				log.Fatal(err)
			}
			fmt.Println(i, ".", j)
			// file, err := p.AwsClients[0].DownloadFile(bucketName, goodId)
			// if err != nil {
			// 	log.Fatal(err)
			// }
			// fmt.Println("bucket content: ")
			// fmt.Println(file)
			// result := make(map[string]any)
			// if err := json.Unmarshal(file, &result); err != nil {
			// 	log.Fatal(err)
			// }
			// fmt.Println(result)
			// if err := WriteFileLamodaHtml("data3.json", []map[string]any{result}); err != nil {
			// 	panic(err)
			// }
			time.Sleep(3 * time.Second)
		}
	}

	return nil
}
