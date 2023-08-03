package parser

// import (
// 	"context"
// 	"fmt"
// 	"runtime"
// 	"sync"
// )

// type (
// 	ApiParser struct {
// 		GoodIdsChan chan string
// 		GoodMapChan chan map[string]any
// 		GoodsWg     *sync.WaitGroup
// 		WorkersWg   *sync.WaitGroup
// 	}
// )

// func (apiParser *ApiParser) GetLamodaAPI(skuChan chan string) map[string]any {
// 	// for sku := range skuChan {

// 	// }

// 	return nil
// }

// func (apiParser *ApiParser) StartWorker(ctx context.Context) {
// 	fmt.Printf("Worker with UserAgent %v starts\n", userAgent)
// 	for data := range goodIdsChan { // каждый воркер слушает общий канал с данными

// 		// fmt.Printf("Worker %v is getting : %v\n", workerNum, data)
// 		// TODO запрос к АПИ
// 		dataWg.Done()
// 		runtime.Gosched()
// 	}
// 	fmt.Printf("Воркер под номером %v завершил работу\n", workerNum)
// 	workerWg.Done()
// }

// func (apiParser *ApiParser) GenerateId(ctx context.Context, pattern string) {
// }
