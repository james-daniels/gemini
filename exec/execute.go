package exec

import (
	"cgem/conf"
	"cgem/order"
	"fmt"
	"log"
	"os"
	"time"

	"gopkg.in/ini.v1"
)

func Execute(symbol, side string, amount, offset int) {

	loadConfig()

	switch repeat {
	case true:
		if freq <= 0 {
			logger(logFile).Fatalln("enter frequency value greater than 0")
		} else {
			multiInst(symbol, side, baseURL, amount, offset)
		}
	default:
		oneInst(symbol, side, baseURL, amount, offset)
	}
}

func oneInst(symbol, side, baseURL string, amount, offset int) {

	p, err := order.PriceFeed(symbol, baseURL)
	errHandler(err)

	if iOffset != 0 {
		offset = iOffset
	}
	price, err := order.PriceOffset(p.Price, offset)
	errHandler(err)

	payload, err := order.PayloadBuilder(symbol, price, side, amount)
	errHandler(err)

	signature := order.SigBuilder(payload, apiSecret)

	response, err := order.NewOrder(baseURL, payload, apiKey, signature)
	errHandler(err)

	if pretty {
		order.MakePretty(response)
		logger(logFile).Printf("%+v\n", response)
	} else {
		fmt.Printf("%+v\n", response)
		logger(logFile).Printf("%+v\n", response)
	}
}

func multiInst(symbol, side, baseURL string, amount, offset int) {

	logger(logFile).Println("app started")

	for {
		p, err := order.PriceFeed(symbol, baseURL)
		errHandler(err)

		if iOffset != 0 {
			offset = iOffset
		}
		price, err := order.PriceOffset(p.Price, offset)
		errHandler(err)

		payload, err := order.PayloadBuilder(symbol, price, side, amount)
		errHandler(err)

		signature := order.SigBuilder(payload, apiSecret)

		response, err := order.NewOrder(baseURL, payload, apiKey, signature)
		errHandler(err)

		logger(logFile).Printf("%+v\n", response)

		time.Sleep(time.Hour * time.Duration(freq))
	}
}

func GetPrice(symbol string) {

	p, err := order.PriceFeed(symbol, conf.GetEnv(env))
	errHandler(err)

	fmt.Printf("\n%v: %v\n", p.Pair, p.Price)
}

func logger(logfile string) *log.Logger {
	if logfile == "" {
		logfile = "cgem.log"
	}
	file, err := os.OpenFile(logfile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalln(err)
	}

	return log.New(file, "cgem: ", log.LstdFlags|log.Lshortfile)
}

func loadConfig() {

	_, err := os.Stat(conf.ConfigFile)
	if err != nil {
		if os.IsNotExist(err) {
			logger(logFile).Fatalln(conf.ConfigFile, "missing: run 'cgem init' to get started")
		}
	}
	cfg, err := ini.Load(conf.ConfigFile)
	errHandler(err)

	apiKey = cfg.Section("credentials").Key("apikey").String()
	apiSecret = cfg.Section("credentials").Key("apisecret").String()
	env = cfg.Section("").Key("environment").String()
	logFile = cfg.Section("logging").Key("logfile").String()
	pretty, _ = cfg.Section("").Key("pretty").Bool()
	iOffset, _ = cfg.Section("orders").Key("offset").Int()
	repeat, _ = cfg.Section("recurrence").Key("repeat").Bool()
	freq, _ = cfg.Section("recurrence").Key("frequency").Int()

	baseURL = conf.GetEnv(env)
}

func errHandler(err error) {
	if err != nil {
		logger(logFile).Fatalln(err)
	}
}