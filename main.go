package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/360EntSecGroup-Skylar/excelize"
	"io/ioutil"
	"net/http"
	"runtime/debug"
	"strconv"
	"time"
)

var (
	cookieJar = []*http.Cookie{}
	uniHeader = map[string]string{
		"Host":            "www1.nseindia.com",
		"User-Agent":      "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:84.0) Gecko/20100101 Firefox/84.0",
		"Accept":          "*/*",
		"Accept-Language": "en-US,en;q=0.5",
		//"Accept-Encoding": "gzip, deflate, br",
		"X-Requested-With":             "XMLHttpRequest",
		"Referer":                      "https://www1.nseindia.com/products/content/equities/equities/eq_security.htm",
		"Access-Control-Allow-Origin":  "*",
		"Access-Control-Allow-Methods": "GET,POST,PUT,DELETE,OPTIONS",
		"Access-Control-Allow-Headers": "Content-Type, Access-Control-Allow-Headers, Authorization, X-Requested-With",
		"Content-Type":                 "application/json;application/x-www-form-urlencoded; charset=UTF-8",
	}
	botURL = "https://api.telegram.org/bot1835262567:AAG16zksc089ULVg0AYlTFxbG0XjszEhKCY"
)

type Nifty struct {
	Filtered struct {
		Data []OptionData `json:"data"`
	} `json:"filtered"`
}

type OptionData struct {
	StrikePrice int    `json:"strikePrice"`
	ExpDate     string `json:"expiryDate"`
	PE          Option `json:"PE"`
	CE          Option `json:"CE"`
}

type Option struct {
	OptType                string
	StrikePrice            int     `json:"strikePrice"`
	ExpDate                string  `json:"expiryDate"`
	OpenInt                int     `json:"openInterest"`
	ChgOpenInt             int     `json:"changeinOpenInterest"`
	PercentChangeInOpenInt float64 `json:"pchangeinOpenInterest"`
	TotalTradedVolume      int     `json:"totalTradedVolume"`
	ImpliedVolatility      float64 `json:"impliedVolatility"`
	LTP                    float64 `json:"lastPrice"`
	ChangeLTP              float64 `json:"change"`
	PecentChangeLTP        float64 `json:"pChange"`
	TotalBuyQty            int     `json:"totalBuyQuantity"`
	TotalSellQty           int     `json:"totalSellQuantity"`
	BidQty                 int     `json:"bidQty"`
	BidPrice               float64 `json:"bidprice"`
	AskQty                 int     `json:"askQty"`
	AskPrice               float64 `json:"askPrice"`
	UnderLyingValue        float64 `json:"underlyingValue"`
}

func main() {

	/*	c := cron.New()
		err := c.AddFunc("@every 1m", cronJob)
		if err != nil {
			fmt.Println("Cron error : ", err)
		}

		c.Start()
		//cronJob()
		fmt.Println("cron has started..")*/
	//go getCookiesLocally()
	job()
	/*	http.HandleFunc("/", cronJob)

		port := os.Getenv("PORT")
		if port == "" {
			port = "9000" // Default port if not specified
		}
		fmt.Printf("Starting server at port %s\n", port)
		if err := http.ListenAndServe(":"+port, nil); err != nil {
			fmt.Printf("Error caused while starting the server: %v\n", err)
		}*/
	time.Sleep(2 * time.Second)
}

func cronJob(w http.ResponseWriter, r *http.Request) {
	/*flag := job()
	for i := 0; i < 4; i++ {
		if !flag {
			flag = job()
			time.Sleep(1 * time.Minute)
		} else {
			return
		}
	}*/
	fmt.Println("Inside the cronJob function")
	job()
}

func job() {
	loc, _ := time.LoadLocation("Asia/Kolkata")
	t := time.Now().In(loc)

	h, m, _ := t.Clock()
	y, mth, d := t.Date()
	timeDate := fmt.Sprintf("%v-%v-%v", d, mth, y)
	timeNow := fmt.Sprintf("%v.%v", h, m)
	fmt.Println("above the excel creation function")

	xls, err := excelize.OpenFile("optionchain.xlsm")
	if err != nil {
		fmt.Println("Error while opening excel file : ", err)
	}
	index := xls.NewSheet("Sheet1")
	xls.SetActiveSheet(index)

	/*	a, _, _ := t.Clock()


		weekday := int(t.Weekday())
		if !(a > 8 && a < 19 && weekday > 0 && weekday < 6) {
			fmt.Printf("Cron is running\n")
			return
		}*/
	fmt.Println("Running the job function.")
	t = findThursday(t)
	strikeDate := t.Format("02-Jan-2006")
	niftyData, err := getOptionData()
	if err != nil {
		fmt.Printf("Error at 116 : %v\n", err)
	}
	fmt.Printf("Successfully fetched optiondata:")
	niftyOpt := &Nifty{}
	err = json.Unmarshal(niftyData, niftyOpt)
	if err != nil {
		fmt.Printf("This error in Job function at 122: %v\n", err)
		debug.PrintStack()
	}
	data := niftyOpt.Filtered.Data
	optMap := map[string]OptionData{}
	priceList := make([]int, 0)
	for _, opt := range data {
		priceList = append(priceList, opt.StrikePrice)
		optMap[strconv.Itoa(opt.StrikePrice)+"|"+opt.ExpDate] = opt
	}

	currNifty, marketErr := getMarketStatus()
	if marketErr != nil {
		fmt.Printf("error at line 134 : %v\n", marketErr)
	}

	fmt.Printf("Current Nifty : %.2f\nStrike Date : %s\n\n", currNifty, strikeDate)
	midStrkPrice := int(currNifty/100) * 100

	factor := currNifty - float64(midStrkPrice)

	if factor > 25 && factor <= 75 {
		midStrkPrice = midStrkPrice + 50
	} else if factor > 75 && factor <= 99 {
		midStrkPrice = midStrkPrice + 100
	}
	avgPriceMap := make(map[int]string)

	for j := -2; j < 3; j++ {
		avgPriceMap[midStrkPrice+(50*j)] = "temp"
	}

	excelHeaderList := []string{
		"OI",
		"Chg. In OI",
		"Total Traded Volume",
		"Implied Volatility",
		"LTP",
		"Total Buy Qty",
		"Total Sell Qty",
		"Strike Price",
		"Total Sell Qty",
		"Total Buy Qty",
		"LTP",
		"Implied Volatility",
		"Total Traded Volume",
		"Chg. In OI",
		"OI",
		"Difference",
	}
	xls.SetSheetRow("Sheet1", "A1", &[]string{fmt.Sprintf("Nifty"), fmt.Sprintf("%v", currNifty)})
	xls.SetSheetRow("Sheet1", "A2", &[]string{fmt.Sprintf("Strike Date"), fmt.Sprintf("%v", strikeDate)})
	xls.SetSheetRow("Sheet1", "A3", &[]string{"Date:", timeDate})
	var tempint []interface{}
	tnow, _ := strconv.ParseFloat(timeNow, 64)
	tempint = append(tempint, "Time:", tnow)
	xls.SetSheetRow("Sheet1", "A4", &tempint)
	xls.SetSheetRow("Sheet1", "A5", &excelHeaderList)
	var callOIAvg, putOIAvg, cellToStyle int

	for i, price := range priceList {
		var data []interface{}
		opt := optMap[strconv.Itoa(price)+"|"+strikeDate]
		data = append(data, opt.CE.OpenInt)
		data = append(data, opt.CE.ChgOpenInt)
		data = append(data, opt.CE.TotalTradedVolume)
		data = append(data, opt.CE.ImpliedVolatility)
		data = append(data, opt.CE.LTP)
		data = append(data, opt.CE.TotalBuyQty)
		data = append(data, opt.CE.TotalSellQty)
		data = append(data, opt.CE.StrikePrice)
		data = append(data, opt.PE.TotalSellQty)
		data = append(data, opt.PE.TotalBuyQty)
		data = append(data, opt.PE.LTP)
		data = append(data, opt.PE.ImpliedVolatility)
		data = append(data, opt.PE.TotalTradedVolume)
		data = append(data, opt.PE.ChgOpenInt)
		data = append(data, opt.PE.OpenInt)
		diff := opt.PE.OpenInt - opt.CE.OpenInt
		data = append(data, diff)

		_, ok := avgPriceMap[opt.StrikePrice]
		if ok {
			callOIAvg = callOIAvg + opt.CE.ChgOpenInt
			putOIAvg = putOIAvg + opt.PE.ChgOpenInt
		}

		if opt.StrikePrice == midStrkPrice {
			cellToStyle = i + 6
		}
		xls.SetSheetRow("Sheet1", "A"+strconv.Itoa(i+6), &data)

	}

	/*xls.SetCellFormula("Sheet1", "B19", "=SUM(B9:B13)")
	xls.SetCellFormula("Sheet1", "E19", "=SUM(E9:E13)")
	xls.SetCellFormula("Sheet1", "G19", "=E19-B19")*/
	xls.SetColWidth("Sheet1", "A", "P", 10)
	dataStyle, err := xls.NewStyle(`{"border":[{"type":"left","color":"0000FF","style":3},{"type":"top","color":"00FF00","style":4},{"type":"bottom","color":"FFFF00","style":5},{"type":"right","color":"FF0000","style":6}],
"fill":{"type":"pattern","color":["#E0EBF5"],"pattern":1}}`)
	headerStyle, err := xls.NewStyle(`{"border":[{"type":"left","color":"000000","style":5},{"type":"top","color":"000000","style":5},{"type":"bottom","color":"000000","style":5},{"type":"right","color":"000000","style":5}],
"fill":{"type":"pattern","color":["#FEFFD9"],"pattern":1}}`)
	StrikePriceStyle, err := xls.NewStyle(`{"border":[{"type":"left","color":"000000","style":5},{"type":"top","color":"000000","style":5},{"type":"bottom","color":"000000","style":5},{"type":"right","color":"000000","style":5}],
"fill":{"type":"pattern","color":["#d7ffd7"],"pattern":1}}`)
	if err != nil {
		fmt.Println(err)
	}
	callStyleNo := fmt.Sprintf("G%d", cellToStyle)
	putStyleNo := fmt.Sprintf("I%d", cellToStyle)
	xls.SetCellStyle("Sheet1", "A6", callStyleNo, dataStyle)
	xls.SetCellStyle("Sheet1", putStyleNo, "P112", dataStyle)
	xls.SetCellStyle("Sheet1", "A5", "P5", headerStyle)
	xls.SetCellStyle("Sheet1", "H6", "H112", StrikePriceStyle)
	xls.SetCellValue("Sheet1", "E3", callOIAvg)
	xls.SetCellValue("Sheet1", "F3", putOIAvg)
	xls.SetCellValue("Sheet1", "G3", putOIAvg-callOIAvg)
	xls.UpdateLinkedValue()
	_ = xls.Save()
	/*	for _, price := range strkPrceList {
		opt := optMap[strconv.Itoa(price)+"|"+strikeDate]

		oiData := fmt.Sprintf("Strike Price : %d\n", price)
		oiData = oiData + fmt.Sprintf("Call OI: %d\n", opt.CE.OpenInt)
		oiData = oiData + fmt.Sprintf("Chg. In call OI : %d\n", opt.CE.ChgOpenInt)
		oiData = oiData + fmt.Sprintf("Put OI: %d\n", opt.PE.OpenInt)
		oiData = oiData + fmt.Sprintf("Chg. In Put OI : %d\n", opt.PE.ChgOpenInt)
		diff := opt.PE.OpenInt - opt.CE.OpenInt
		oiData = oiData + fmt.Sprintf("Difference : %d\n", diff)
		oiData = oiData + "Signal: "
		if diff > 0 {
			oiData = oiData + "Buy\n"
		} else if diff == 0 {
			oiData = oiData + "Neutral\n"
		} else {
			oiData = oiData + "Sell\n"
		}
		textMsg = textMsg + "\n" + oiData
	}*/

	/*	err = sendToTelegram(textMsg)
		if err != nil {
			return "Error Occured"
		}*/

}

func getOptionData() ([]byte, error) {
	url := "https://www.nseindia.com/api/option-chain-indices?symbol=NIFTY"
	var tempData []byte
	client := &http.Client{}
	h := http.Header{}

	for k, v := range uniHeader {
		h.Add(k, v)
	}

	req, _ := http.NewRequest(http.MethodGet, url, nil)

	req.Header = h
	err := setCookies()
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("This is error in getOptionData function : %v\n", err)
		debug.PrintStack()
		return tempData, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("This is error in getOptionData function : %v\n", err)
		debug.PrintStack()
		return tempData, err
	}
	//fmt.Printf("This is body : %+v\n", string(body))

	if resp.StatusCode != 200 {
		fmt.Printf("Status code is not 200: %v\n", resp.StatusCode)
		fmt.Printf("This is body : %s\n", string(body))
		debug.PrintStack()
		return tempData, fmt.Errorf("statuscode is not 200")
	}

	_ = resp.Body.Close()

	return body, nil
}

func getMarketStatus() (float64, error) {
	tempMarket := map[string][]interface{}{}
	httpClient := &http.Client{}
	h := http.Header{}
	for k, v := range uniHeader {
		h.Add(k, v)
	}
	marketVal := 0.0
	req, err := http.NewRequest(http.MethodGet, "https://www.nseindia.com/api/marketStatus", nil)
	if err != nil {
		fmt.Printf("This is Error in getMarketStatus function: %v\n", err)
		debug.PrintStack()
		return marketVal, err
	}

	req.Header = h

	resp, err := httpClient.Do(req)
	if err != nil {
		fmt.Printf("This is Error in getMarketStatus function: %v\n", err)
		debug.PrintStack()
		return marketVal, err
	}

	if resp.StatusCode != 200 {
		fmt.Printf("This is resp status code : %v\n", resp.StatusCode)
		debug.PrintStack()
		return marketVal, fmt.Errorf("Got non-success code while fetching market status data : %v\n", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("This is Error in getMarketStatus function: %v\n", err)
		debug.PrintStack()
		return marketVal, err
	}

	_ = resp.Body.Close()
	//fmt.Printf("This is response body : market status : %s\n", string(body))

	err = json.Unmarshal(body, &tempMarket)
	if err != nil {
		fmt.Printf("This is Error in getMarketStatus function: %v\n", err)
		debug.PrintStack()
		return marketVal, err
	}

	tempMap := tempMarket["marketState"][0].(map[string]interface{})
	marketVal = tempMap["last"].(float64)

	return marketVal, nil
}

func sendMsgToTelegram(text string) error {
	sendMsgURL := botURL + "/sendmessage"
	var msg struct {
		ChatID int    `json:"chat_id"`
		Text   string `json:"text"`
	}
	msg.ChatID = 1371114495
	msg.Text = text
	jsonMsg, err := json.Marshal(msg)
	if err != nil {
		fmt.Printf("This is the error : %v\n", err)
		debug.PrintStack()
		return err
	}

	reqBody := bytes.NewBuffer(jsonMsg)

	resp, err := http.Post(sendMsgURL, "application/json", reqBody)
	if err != nil {
		fmt.Printf("This is the error : %v\n", err)
		debug.PrintStack()
		return err
	}

	if resp.StatusCode != 200 {
		debug.PrintStack()
		return fmt.Errorf("Got non-success code while telegram message : %v\n", resp.StatusCode)
	}

	return nil
}

func findThursday(t time.Time) time.Time {
	const day = 24 * time.Hour
	// get daylight saving time out of the way
	t = time.Date(t.Year(), t.Month(), t.Day(), 12, 0, 0, 0, t.Location())
	//t = time.Date(2021,3 ,12 , 12, 0, 0, 0, t.Location())
	// compute next Friday

	t = t.Add(6 * day)
	t = t.Add(-time.Duration(t.Add(-4*day).Weekday()) * day)
	// check all subsequent Fridays
	return t
}

func setCookies() error {

	url := "https://www.nseindia.com/api/option-chain-indices?symbol=NIFTY"
	method := "GET"
	client := &http.Client{}
	h := http.Header{}

	req, err := http.NewRequest(method, url, nil)
	for k, v := range uniHeader {
		h.Add(k, v)
	}
	req.Header = h
	if err != nil {
		fmt.Println("***************", err)
		return err
	}
	res, err := client.Do(req)
	if err != nil {
		fmt.Println("***************", err)
		return err
	}
	defer res.Body.Close()

	for _, c := range res.Cookies() {
		cookieJar = append(cookieJar, c)
	}

	return nil
}
