package main

import (
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"strings"
	"time"

	"github.com/JackSmithThu/bs_common/conf"
	model "github.com/JackSmithThu/bs_common/model/data_platform_model"
	"github.com/JackSmithThu/common-ares/frame/logs"
)

func HandleMessage() {
	logs.Info("[HandleMessage] ==== start ====")
	conf.InitPlatformDBConnect()
	date := time.Now().Format("2006-01-02")
	hour := time.Now().Format("2006-01-02-15")
	filename := fmt.Sprintf("%v.csv", hour)
	pwd, _ := os.Getwd()
	path := os.Getenv("W2B_UPDATE_PATH")
	if path == "" {
		path = fmt.Sprintf("/data/www/static/w2b/%v/", date)
	}
	logs.Info("[HandleMessage] path env: %v, pwd: %v", path, pwd)

	fileInfoList, err := ioutil.ReadDir(path)
	if err != nil {
		logs.Error("[HandleMessage] read dir fail! add new dir: %v", path)
		err = os.MkdirAll(path, os.ModePerm)
		if err != nil {
			logs.Error("[HandleMessage] new dir fail! dir: %v", path)
			panic("get and new dir fail!")
		}
	}

	for _, fileItem := range fileInfoList {
		logs.Info("[HandleMessage] current file list: %v", fileItem.Name()) //打印当前文件或目录下的文件或目录名
		if strings.Contains(fileItem.Name(), "america-"+filename) {
			logs.Info("[HandleMessage] current filename is exist, dirFilename: %v, filename %v", fileItem.Name(), filename) //打印当前文件或目录下的文件或目录名
			time.Sleep(1800 * time.Second)
			return
		}
	}

	param := GenerateFileParam{
		IsCanada: false,
		FilePath: path,
		FileName: filename,
	}
	// GeneratePriceFile(param)
	GenerateStockFile(param)

	// param.IsCanada = true
	GeneratePriceFile(param)
}

type GenerateFileParam struct {
	IsCanada bool
	FileName string
	FilePath string
}

func GenerateStockFile(param GenerateFileParam) {

	offset := 0
	limit := 2000
	condition := "on_sale = 1"

	totalW2bProducts := []model.W2bProducts{}
	for {
		w2bProducts := []model.W2bProducts{}
		conf.DataPlatformDB.Where(condition).Offset(offset).Limit(limit).Find(&w2bProducts)
		if len(w2bProducts) == 0 {
			break
		}
		offset += limit
		totalW2bProducts = append(totalW2bProducts, w2bProducts...)
	}

	for i := 1; i < 20; i++ {
		GenerateStoreStockFile(int32(i), totalW2bProducts, param)
	}
}

func GenerateStoreStockFile(storeNum int32, totalW2bProducts []model.W2bProducts, param GenerateFileParam) {

	logs.Info("[GenerateStoreStockFile] the %v store stock begin!", storeNum)

	offset := 0
	limit := 2000

	amazonStocks := []model.AmazonStock{}
	for {
		stocks := []model.AmazonStock{}
		conf.DataPlatformDB.Where(fmt.Sprintf(" store=%v ", storeNum)).Offset(offset).Limit(limit).Find(&stocks)
		if len(stocks) == 0 {
			break
		}
		offset += limit
		amazonStocks = append(amazonStocks, stocks...)
	}

	if len(amazonStocks) == 0 {
		logs.Info("[GenerateStoreStockFile] store %v has no stock!", storeNum)
		return
	}

	w2bIndex := map[string][]model.AmazonStock{}
	for _, stock := range amazonStocks {
		stocks, exist := w2bIndex[stock.W2bID]
		if !exist {
			stocks = []model.AmazonStock{}
		}
		stocks = append(stocks, stock)
		w2bIndex[stock.W2bID] = stocks
	}

	timeString := strings.Replace(param.FileName, ".csv", "", -1)
	filename := fmt.Sprintf("%v-store-%v.csv", timeString, storeNum)

	f, err := os.OpenFile(param.FilePath+"/"+filename, os.O_CREATE|os.O_APPEND|os.O_RDWR, os.ModePerm|os.ModeTemporary) //创建文件
	defer f.Close()
	// os.chmod(path + "/" + filename, os.FileMode(0775))
	if err != nil {
		logs.Error("[GenerateStoreStockFile] create output file fail, filename: %v, err: %v", filename, err)
		panic("create output file fail!")
	}
	f.Write([]byte("\xef\xbb\xbf\"sku\",\"price\",\"quantity\"\n"))
	tpl := "\"%v\",\"%v\",\"%v\"\n"

	for _, item := range totalW2bProducts {
		// sku := "x-w2b-" + item.W2bID
		price := item.Price + item.ShipCost + item.Supplierhandling
		if item.Stricted == 1 {
			item.Quantity = 0
		}
		quantity := item.Quantity

		if quantity > 10 {
			// logs.Info("[GenerateStoreStockFile] stock of w2b_id: %v is larger than 10!", item.W2bID)
			continue
		}

		money := 1.5 * float64(price) / 100
		if quantity < 0 {
			money = 2 * float64(price) / 100
		}

		floor := math.Floor(money)
		if money-floor > 0.5 {
			money = floor + 0.99
		} else {
			money = floor + 0.49
		}

		prefixList := []string{"x-w2b-", "c-w2b-", "p-w2b-"}
		for _, prefix := range prefixList {
			sku := prefix + item.W2bID
			stocks, exist := w2bIndex[item.W2bID]
			if !exist {
				continue
			}
			for _, stock := range stocks {
				if stock.SKU != sku {
					continue
				}

				if (stock.Quantity == 0 && item.Quantity > 0 && int64(item.UpdateTime)-stock.UpdateTime < 24*3600) || item.Quantity == 0 {
					if prefix == "c-w2b-" {
						money = money - 10
					}
					str := fmt.Sprintf(tpl, sku, money, quantity)
					f.Write([]byte(fmt.Sprintf(str)))
					logs.Info("[HandleMessage] store: %v, stock.Quantity: %v, item.Quantity: %v, stock.UpdateTime: %v, item.UpdateTime: %v, str: %v", storeNum, stock.Quantity, item.Quantity, stock.UpdateTime, item.UpdateTime, str)
				}
			}
		}
		f.Close()
	}
}

func GeneratePriceFile(param GenerateFileParam) {
	offset := 0
	limit := 2000

	condition := "on_sale = 1"
	if param.IsCanada {
		condition = condition + " and to_canada = 1"
	}
	//conf.DataPlatformDB.Model(&model.W2bProducts{}).Where().Count(&count)
	//logs.Info("[HandleMessage] total item num: %v", count)
	totalW2bProducts := []model.W2bProducts{}
	for {
		w2bProducts := []model.W2bProducts{}
		conf.DataPlatformDB.Where(condition).Offset(offset).Limit(limit).Find(&w2bProducts)
		if len(w2bProducts) == 0 {
			break
		}
		offset += limit
		totalW2bProducts = append(totalW2bProducts, w2bProducts...)
	}

	filename := "america-" + param.FileName
	f, err := os.Create(param.FilePath + "/" + filename) //创建文件
	defer f.Close()
	// os.chmod(path + "/" + filename, os.FileMode(0775))
	if err != nil {
		logs.Error("[MessageHandler] create output file fail, err: %v", err)
		panic("create output file fail!")
	}
	f.Write([]byte("\xef\xbb\xbf\"sku\",\"price\",\"quantity\"\n"))
	tpl := "\"%v\",\"%v\",\"%v\"\n"

	for _, item := range totalW2bProducts {
		sku := "x-w2b-" + item.W2bID
		price := item.Price + item.ShipCost + item.Supplierhandling
		if item.Stricted == 1 {
			// logs.Info("[] ==== w2b_id: %v, stricted: %v", item.W2bID, item.Stricted)
			item.Quantity = 0
		}
		quantity := item.Quantity

		money := 1.5 * float64(price) / 100
		if quantity < 0 {
			money = 2 * float64(price) / 100
		}

		// ad white list
		if item.W2bID == "C507170182" {
			money = 79.99
		}

		if param.IsCanada {
			money = money * 1.4 // 1 USD = 1.4 CAD
		}
		floor := math.Floor(money)
		if money-floor > 0.5 {
			money = floor + 0.99
		} else {
			money = floor + 0.49
		}

		str := fmt.Sprintf(tpl, sku, money, quantity)
		// logs.Info("[HandleMessage] str: %v", str)
		f.Write([]byte(fmt.Sprintf(str)))
		// str = fmt.Sprintf(tpl, "c-w2b-"+item.W2bID, money-10, quantity)
		// f.Write([]byte(fmt.Sprintf(str)))
		// str = fmt.Sprintf(tpl, "p-w2b-"+item.W2bID, money, quantity)
		// f.Write([]byte(fmt.Sprintf(str)))
	}
	f.Close()
	logs.Info("[HandleMessage] ==== end ====")
}
