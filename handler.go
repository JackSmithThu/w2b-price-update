package main

import (
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"strings"
	"time"

	"github.com/windrainw/bs_common/conf"
	model "github.com/windrainw/bs_common/model/data_platform_model"
	"github.com/windrainw/common-ares/frame/logs"
)

func HandleMessage() {
	logs.Info("[HandleMessage] ==== start ====")
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
			fmt.Printf("[HandleMessage]file exist! ====")
			time.Sleep(1800 * time.Second)
			return
		}
	}

	param := GenerateFileParam{
		IsCanada: false,
		FilePath: path,
		FileName: filename,
	}
	GeneratePriceFile(param)

	param.IsCanada = true
	GeneratePriceFile(param)
}

type GenerateFileParam struct {
	IsCanada bool
	FileName string
	FilePath string
}

func GeneratePriceFile(param GenerateFileParam) {
	offset := 0
	limit := 2000
	conf.InitPlatformDBConnect()
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
	if param.IsCanada {
		filename = "canada-" + param.FileName
	}
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
			item.Quantity = 0
		}
		quantity := item.Quantity

		money := 1.5 * float64(price) / 100
		if quantity < 2 {
			money = 2 * float64(price) / 100
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
		logs.Info("[HandleMessage] str: %v", str)
		f.Write([]byte(fmt.Sprintf(str)))
	}

	logs.Info("[HandleMessage] ==== end ====")
}
