package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"runtime/debug"
	"strconv"
	"time"
)

var (
	cookie string
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

	/*	c := cron.New()
		err := c.AddFunc("@every 1m", cronJob)
		if err != nil {
			fmt.Println("Cron error : ", err)
		}

		c.Start()
		//cronJob()
		fmt.Println("cron has started..")*/
	go getCookiesLocally()
	http.HandleFunc("/", cronJob)

	port := os.Getenv("PORT")
	if port == "" {
		port = "9000" // Default port if not specified
	}
	fmt.Printf("Starting server at port %s\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		fmt.Printf("Error caused while starting the server: %v\n", err)
	}
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
	job()
}

func job() string {
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
	niftyData, err := getOptionData()
	if err != nil {
		return "Error Occured"
	}
	fmt.Printf("Successfully fetched optiondata:")
	niftyOpt := &Nifty{}
	err = json.Unmarshal(niftyData, niftyOpt)
	if err != nil {
		fmt.Printf("This error in Job function: %v\n", err)
		debug.PrintStack()
		return "Error Occured"
	}
	data := niftyOpt.Filtered.Data
	optMap := map[string]OptionData{}
	for _, opt := range data {
		optMap[strconv.Itoa(opt.StrikePrice)+"|"+opt.ExpDate] = opt
	}

	currNifty, marketErr := getMarketStatus()
	if marketErr != nil {
		return "Error Occured"
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

	err = sendToTelegram(textMsg)
	if err != nil {
		return "Error Occured"
	}

	return textMsg
}

func getOptionData() ([]byte, error) {
	var tempData []byte
	httpClient := &http.Client{}
	h := http.Header{}
	h.Add("Cache-Control", "max-age=0")
	h.Add("DNT", "1")
	h.Add("Upgrade-Insecure-Requests", "1")
	//h.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.79 Safari/537.36")
	h.Add("Sec-Fetch-User", "?1")
	h.Add("Sec-Fetch-Site", "none")
	//h.Add("Sec-Fetch-Mode", "navigate")
	//h.Add("Referer", "https://www1.nseindia.com/live_market/dynaContent/live_watch/option_chain/optionKeys.jsp?symbolCode=-9999&symbol=NIFTY&symbol=BANKNIFTY&instrument=OPTIDX&date=-&segmentLink=17&segmentLink=17")
	//h.Add("Accept-Encoding", "gzip, deflate, br")
	h.Add("Accept-Language", "en-US,en;q=0.9,hi;q=0.8")
	//h.Add("authority", "www.nseindia.com")
	//h.Add("cache-control", "no-cache")
	//h.Add("pragma", "no-cache")
	//h.Add("referer", "https://www.nseindia.com/option-chain?symbolCode=-10006&symbol=NIFTY&symbol=NIFTY&instrument=-&date=-&segmentLink=17&symbolCount=2&segmentLink=17")
	//h.Add("sec-fetch-dest", "empty")
	h.Add("sec-fetch-mode", "cors")
	//h.Add("sec-fetch-site", "same-origin")
	h.Add("Host", "www.nseindia.com")
	h.Add("Connection", "keep-alive")
	h.Add("User-Agent", "PostmanRuntime/7.26.8")
	h.Add("Accept", "*/*")
	h.Add("Cookie", cookie)
	b := bytes.NewBuffer([]byte("{}"))
	req, err := http.NewRequest(http.MethodGet, "https://www.nseindia.com/api/option-chain-indices?symbol=NIFTY", b)
	if err != nil {
		fmt.Printf("This is error in getOptionData function: %v\n", err)
		debug.PrintStack()
		return tempData, err
	}
	req.Header = h

	resp, err := httpClient.Do(req)
	if err != nil {
		fmt.Printf("This is error in getOptionData function : %v\n", err)
		debug.PrintStack()
		return tempData, err
	}

	if resp.StatusCode != 200 {
		fmt.Printf("This is resp.Statuscode : %v\n", resp.StatusCode)
		body, _ := ioutil.ReadAll(resp.Body)
		fmt.Println(string(body))
		debug.PrintStack()
		return tempData, fmt.Errorf("Failed to get optiondata with status code : %v\n", resp.StatusCode)
	}
	setCookiesLocally(resp.Cookies())
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("This is error in getOptionData function : %v\n", err)
		debug.PrintStack()
		return tempData, err
	}

	_ = resp.Body.Close()

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

	req, err := http.NewRequest(http.MethodGet, "https://www.nseindia.com/api/marketStatus", b)
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
	fmt.Printf("This is response body : market status : %s\n", string(body))

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

func sendToTelegram(text string) error {
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

func getCookiesLocally() {
	httpClient := &http.Client{}
	h := http.Header{}
	h.Add("User-Agent", "PostmanRuntime/7.26.8")
	h.Add("Accept", "*/*")
	h.Add("Host", "www.nseindia.com")
	h.Add("Accept-Encoding", "gzip, deflate, br")
	h.Add("Connection", "keep-alive")
	b := bytes.NewBuffer([]byte("{}"))
	req, err := http.NewRequest(http.MethodGet, "https://www.nseindia.com/api/option-chain-indices?symbol=NIFTY", b)
	if err != nil {
		fmt.Printf("This is error in getOptionData function: %v\n", err)
		debug.PrintStack()
	}
	req.Header = h

	resp, err := httpClient.Do(req)
	if err != nil {
		fmt.Printf("This is error in getOptionData function : %v\n", err)
		debug.PrintStack()
	}

	if resp.StatusCode != 200 {
		getCookiesLocally()
	}

	setCookiesLocally(resp.Cookies())

}

func setCookiesLocally(cookieJar []*http.Cookie) {

	cookie = ""

	for _, c := range cookieJar {
		cookie = cookie + c.String()
	}

}
