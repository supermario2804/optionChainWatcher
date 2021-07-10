package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"runtime/debug"
	"strconv"
	"time"

	"github.com/robfig/cron"
)

var (
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
	OptType     string
	StrikePrice int    `json:"strikePrice"`
	ExpDate     string `json:"expiryDate"`
	OpenInt     int    `json:"openInterest"`
	ChgOpenInt  int    `json:"changeinOpenInterest"`
}

func main() {

	c := cron.New()
	err := c.AddFunc("@every 30s", cronJob)
	if err != nil {
		fmt.Println("Cron error : ", err)
	}

	c.Start()
	fmt.Println("cron has started..")
	/*	port := os.Getenv("PORT")
		if port == "" {
			port = "9000" // Default port if not specified
		}
		if err := http.ListenAndServe(":"+port, nil); err != nil {
			fmt.Printf("Error caused while starting the server")
		}*/
	select {}
}

func cronJob() {
	loc, _ := time.LoadLocation("Asia/Kolkata")
	t := time.Now().In(loc)
	/*	a, _, _ := t.Clock()

		weekday := int(t.Weekday())
		if !(a > 8 && a < 19 && weekday > 0 && weekday < 6) {
			fmt.Printf("Cron is running\n")
			return
		}*/
	fmt.Println("Running the cron job function.")
	t = findThursday(t)
	strikeDate := t.Format("02-Jan-2006")
	fmt.Println("above the get optionData fucntion call")
	niftyData, _ := getOptionData()
	fmt.Println("After the get optiondata function call")
	niftyOpt := &Nifty{}
	jsonErr := json.Unmarshal(niftyData, niftyOpt)
	if jsonErr != nil {
		fmt.Println("This is error1: ", jsonErr)
		debug.PrintStack()
	}
	data := niftyOpt.Filtered.Data
	optMap := map[string]OptionData{}
	for _, opt := range data {
		optMap[strconv.Itoa(opt.StrikePrice)+"|"+opt.ExpDate] = opt
	}

	currNifty, marketErr := getMarketStatus()
	if marketErr != nil {
		fmt.Println("This is error2: ", marketErr)
		debug.PrintStack()
	}

	textMsg := fmt.Sprintf("Current Nifty : %.2f\nStrike Date : %s\n\n", currNifty, strikeDate)
	fmt.Printf("Current Nifty : %.2f\nStrike Date : %s\n\n", currNifty, strikeDate)
	midStrkPrice := int(currNifty/100) * 100
	strkPrceList := make([]int, 0)

	for j := -5; j < 6; j++ {
		strkPrceList = append(strkPrceList, midStrkPrice+(100*j))
	}

	for _, price := range strkPrceList {
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
	}

	telegramErr := sendToTelegram(textMsg)
	if telegramErr != nil {
		fmt.Println("This is error3: ", telegramErr)
		debug.PrintStack()
	}

}

func getOptionData() ([]byte, error) {
	var tempData []byte
	httpClient := &http.Client{}
	h := http.Header{}
	h.Add("Connection", "keep-alive")
	h.Add("Cache-Control", "max-age=0")
	h.Add("DNT", "1")
	h.Add("Upgrade-Insecure-Requests", "1")
	h.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.79 Safari/537.36")
	h.Add("Sec-Fetch-User", "?1")
	//h.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9")
	h.Add("Sec-Fetch-Site", "none")
	h.Add("Sec-Fetch-Mode", "navigate")
	//h.Add("Accept-Encoding", "gzip, deflate, br")
	h.Add("Accept-Language", "en-US,en;q=0.9,hi;q=0.8")
	b := bytes.NewBuffer([]byte("{}"))
	req, httpErr := http.NewRequest(http.MethodGet, "https://www.nseindia.com/api/option-chain-indices?symbol=NIFTY", b)
	if httpErr != nil {
		return tempData, httpErr
	}
	req.Header = h

	resp, httpErr := httpClient.Do(req)
	if httpErr != nil || resp.StatusCode != 200 {
		time.Sleep(15 * time.Second)
		return getOptionData()
	}

	body, httpErr := ioutil.ReadAll(resp.Body)
	if httpErr != nil {
		return tempData, httpErr
	}

	_ = resp.Body.Close()

	//fmt.Println(string(body))

	return body, nil
}

func getMarketStatus() (float64, error) {
	tempMarket := map[string][]interface{}{}
	httpClient := &http.Client{}
	h := http.Header{}
	marketVal := 0.0
	h.Add("Connection", "keep-alive")
	h.Add("Cache-Control", "max-age=0")
	h.Add("DNT", "1")
	h.Add("Upgrade-Insecure-Requests", "1")
	h.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.79 Safari/537.36")
	h.Add("Sec-Fetch-User", "?1")
	//h.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9")
	h.Add("Sec-Fetch-Site", "none")
	h.Add("Sec-Fetch-Mode", "navigate")
	//h.Add("Accept-Encoding", "gzip, deflate, br")
	h.Add("Accept-Language", "en-US,en;q=0.9,hi;q=0.8")
	b := bytes.NewBuffer([]byte("{}"))
	req, httpErr := http.NewRequest(http.MethodGet, "https://www.nseindia.com/api/marketStatus", b)
	if httpErr != nil {
		return 0, httpErr
	}
	req.Header = h

	resp, httpErr := httpClient.Do(req)
	if httpErr != nil || resp.StatusCode != 200 {
		time.Sleep(15 * time.Second)
		fmt.Printf("This is httpErr : %v\n", httpErr)
		return getMarketStatus()
	}

	body, httpErr := ioutil.ReadAll(resp.Body)
	if httpErr != nil {
		fmt.Printf("Error at ioutil.ReadAll : %v\n",httpErr)
		return marketVal, httpErr
	}

	_ = resp.Body.Close()

	//fmt.Println(string(body))
	fmt.Printf("This is status code : %v\n", resp.StatusCode)
	jsonErr := json.Unmarshal(body, &tempMarket)
	if jsonErr != nil {
		time.Sleep(15 * time.Second)
		return getMarketStatus()
	}
	tempMap := tempMarket["marketState"][0].(map[string]interface{})
	marketVal = tempMap["last"].(float64)

	return marketVal, nil
}

func sendToTelegram(text string) error {
	sendMsgURL := botURL + "/sendmessage"
	var msg struct {
		ChatID int    `json:"chat_id"`
		Text   string `json:"text"`
	}
	msg.ChatID = 1371114495
	msg.Text = text
	jsonMsg, jsonErr := json.Marshal(msg)
	if jsonErr != nil {
		return jsonErr
	}

	reqBody := bytes.NewBuffer(jsonMsg)

	resp, httpErr := http.Post(sendMsgURL, "application/json", reqBody)
	if httpErr != nil || resp.StatusCode != 200 {
		return httpErr
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
