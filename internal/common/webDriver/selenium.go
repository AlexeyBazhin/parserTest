package webDriver

import (
	"time"

	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/chrome"
)

type (
	Env     string
	Browser string
)

const (
	DevEnd Env = "dev"

	Firefox Browser = "firefox"
	Chrome  Browser = "chrome"

	chromeDriverPath = "/usr/local/bin/chromedriver" // TODO это все дело потом в конфиг
	geckoDriverPath  = "/usr/local/bin/geckodriver"
)

func NewChromeDriver() (selenium.WebDriver, *selenium.Service, error) {
	// Run Chrome browser
	service, err := selenium.NewChromeDriverService(chromeDriverPath, 4444)
	if err != nil {
		panic(err)
	}

	caps := selenium.Capabilities{}
	caps.AddChrome(chrome.Capabilities{Args: []string{
		//"-proxy http://<iparchitect_27434_13_07_23>:<nQ5aNAAreT9G4zEbha>@188.143.169.27:30049",
		"window-size=1920x1080",
		//"--headless",
		//"--blink-settings=imagesEnabled=false",
		//"disable-2d-canvas-image-chromium",
		//"disable-webgl-image-chromium",
		"--disable-blink-features=AutomationControlled",
		// "-proxy socks5://<iparchitect_27434_13_07_23>:<nQ5aNAAreT9G4zEbha>@188.143.169.27:40049",
		// "--proxy-server=socks5://188.143.169.27:40049",
		//"--host-resolver-rules=MAP * ~NOTFOUND , EXCLUDE 188.143.169.27",

		//"--no-sandbox",

		// это пока хз - копировал со stackoverflow)
		// "--disable-dev-shm-usage",
		// "disable-gpu",
	}})

	// caps.AddProxy(selenium.Proxy{
	// 	Type: selenium.Manual,
	// 	// HTTP: "http://<iparchitect_27434_13_07_23>:<nQ5aNAAreT9G4zEbha>@188.143.169.27:30049",
	// 	// HTTPPort:      8000,
	// 	SOCKS:        "185.54.178.193:1080",
	// 	SOCKSVersion: 5,
	// 	// SOCKSUsername: "iparchitect_27434_13_07_23",
	// 	// SOCKSPassword: "nQ5aNAAreT9G4zEbha",
	// })

	driver, err := selenium.NewRemote(caps, "")
	if err != nil {
		panic(err)
	}
	driver.SetImplicitWaitTimeout(1000 * time.Millisecond)

	time.Sleep(1 * time.Second)

	return driver, service, nil
}

// драйвер для firefox, но у меня так и не запустился
// service, err := selenium.NewGeckoDriverService(geckoDriverPath, 4444)

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
