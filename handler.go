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
		if strings.Contains(fileItem.Name(), filename) {
			logs.Info("[HandleMessage] current filename is exist, dirFilename: %v, filename %v", fileItem.Name(), filename) //打印当前文件或目录下的文件或目录名
			fmt.Printf("[HandleMessage]file exist! ====")
			time.Sleep(1800 * time.Second)
			return
		}
	}

	count := 0
	offset := 0
	limit := 10
	conf.InitPlatformDBConnect()
	conf.DataPlatformDB.Model(&model.W2bProducts{}).Where("on_sale = 1").Count(&count)
	logs.Info("[HandleMessage] total item num: %v", count)

	f, err := os.Create(path + "/" + filename) //创建文件
	defer f.Close()
	// os.chmod(path + "/" + filename, os.FileMode(0775))
	if err != nil {
		logs.Error("[MessageHandler] create output file fail, err: %v", err)
		panic("create output file fail!")
	}
	f.Write([]byte("\xef\xbb\xbf\"sku\",\"price\",\"quantity\"\n"))
	tpl := "\"%v\",\"%v\",\"%v\"\n"
	for i := 0; i < count/limit; i++ {
		w2bProducts := []model.W2bProducts{}
		conf.DataPlatformDB.Where("on_sale = 1").Offset(offset).Limit(limit).Find(&w2bProducts)
		offset += limit

		for _, item := range w2bProducts {
			sku := "x-w2b-" + item.W2bID
			price := item.Price + item.ShipCost + item.Supplierhandling
			quantity := item.Quantity

			money := 1.5 * float64(price) / 100
			if quantity < 2 {
				money = 2 * float64(price) / 100
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
	}
	logs.Info("[HandleMessage] ==== end ====")
}
