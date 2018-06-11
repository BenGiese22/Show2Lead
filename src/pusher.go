package main

import (
        "encoding/json"
        "fmt"
        "io/ioutil"
        "log"
        "net/http"
        "os"
        "strings"
        "strconv"
        "net/smtp"
        "time"
        "golang.org/x/net/context"
        "golang.org/x/oauth2"
        "github.com/aws/aws-lambda-go/lambda"
)

type ProspectDetailResponse struct {
	Response struct {
		Status string `json:"status"`
		Data []struct {
			CreatedAt            string `json:"created_at"`
			Name                 string `json:"name"`
			Email                string `json:"email"`
			Phone                string `json:"phone"`
			Showtime             string `json:"showtime"`
			Address              string `json:"Address"`
			Unit                 string `json:"Unit"`
			ShowingWasScheduled  string `json:"showing_was_scheduled"`
			ShowingMethod        string `json:"showing_method"`
			NoShow               string `json:"no_show"`
			CurrentStatus        string `json:"current_status"`
			LeadSource           string `json:"lead_source"`
			ReferrerURL          string `json:"referrer_url"`
			SourceEmailAutoreply string `json:"source_email_autoreply"`
			SourceShowmojoPhone  string `json:"source_showmojo_phone"`
			TeamMember           string `json:"team_member"`
			Comments             string `json:"comments"`
			Answer8151           string `json:"answer_8151"`
			Answer8152           string `json:"answer_8152"`
			Answer8153           string `json:"answer_8153"`
			Answer8261           string `json:"answer_8261"`
			Answer8263           string `json:"answer_8263"`
			Answer8275           string `json:"answer_8275"`
		} `json:"data"`
	} `json:"response"`
}

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
        tokFile := "token.json"
        tok, err := tokenFromFile(tokFile)
        if err != nil {
                tok = getTokenFromWeb(config)
                saveToken(tokFile, tok)
        }
        return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
        authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
        fmt.Printf("Go to the following link in your browser then type the "+
                "authorization code: \n%v\n", authURL)

        var authCode string
        if _, err := fmt.Scan(&authCode); err != nil {
                log.Fatalf("Unable to read authorization code: %v", err)
        }

        tok, err := config.Exchange(oauth2.NoContext, authCode)
        if err != nil {
                log.Fatalf("Unable to retrieve token from web: %v", err)
        }
        return tok
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
        f, err := os.Open(file)
        defer f.Close()
        if err != nil {
                return nil, err
        }
        tok := &oauth2.Token{}
        err = json.NewDecoder(f).Decode(tok)
        return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
        fmt.Printf("Saving credential file to: %s\n", path)
        f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
        defer f.Close()
        if err != nil {
                log.Fatalf("Unable to cache oauth token: %v", err)
        }
        json.NewEncoder(f).Encode(token)
}

func main() {

  lambda.Start(Action)

}

func Action() (error) {
  log.Println("Process Startup")
  start, end := getTimes()

  str_start := timeForShowMojo(start)
  str_end := timeForShowMojo(end)

  log.Println("Time Frame: " + start.String(),end.String())

  log.Println("Getting Prospect Info. & Sending Emails")
  //tester_send()
  prospects := GetProspectDetails(str_start,str_end)
    for _, val := range prospects.Response.Data {

      //Created: 7 Jun 2018, 6:58AM  CDT
      v := createdAtToTime(val.CreatedAt)

      if(v.After(start)) {
        if (val.ShowingWasScheduled == "t") {
          log.Println("Lead: " + val.Name + " Email: " + val.Email + "  Agent: " + val.TeamMember + "  Created: " + val.CreatedAt)
          send(val.Address + " " + val.Unit, val.Name, val.Email, val.Phone,val.TeamMember)
        }
      }
    }

  log.Println("Process Complete, Shutting Down")
  return nil
}

//Takes in the parameter of val.CreatedAt
//Converts it to a working time object/Struct
//Returns that time object/Struct
func createdAtToTime(created string) (time.Time) {
  //Created: 7 Jun 2018, 11:58AM  CDT
  str := strings.Replace(created, ",","",-1)
  //Created: 7 Jun 2018 11:58AM  CDT
  split_str := strings.Split(str," ")
  //[7][Jun][2018][11:58AM][CDT]

  //Fix time
  num_time := split_str[3]
  indicator := num_time
  indicator = indicator[len(indicator)-2:]
  num_time = num_time[0:(len(num_time)-2)]
  time_split := strings.Split(num_time,":")
  hr,_ := strconv.Atoi(time_split[0])
  if indicator == "AM" && hr == 12 {
    hr = 0
  } else if indicator == "PM" && hr != 12 {
    hr = hr + 12
  }

  hr = hr + 5

  _,month,_ := time.Now().Date()

  year_str := split_str[2]
  year,_ := strconv.Atoi(year_str)
  day_str := split_str[0]
  day,_ := strconv.Atoi(day_str)
  min_str := time_split[1]
  min,_ := strconv.Atoi(min_str)

  if hr > 23 {
    hr = hr - 23
    day = day + 1
  }

  //create Time Struct
  t := time.Date(year,month,day,hr,min,0,0,time.UTC)
  return t
}

//Returns the time of method call and 10 minutes before that.
func getTimes() (time.Time, time.Time) {
  time_now := time.Now().UTC()
  subtracted_time := time_now.Add(-10*time.Minute)
  return subtracted_time,time_now
}

//Converts time to string value time for showmojo
func timeForShowMojo(v time.Time) (string) {

	day := strconv.Itoa(v.Day())
	if v.Day() < 10 {
		day = "0" + day
	}
	month := strconv.Itoa(int(v.Month()))
	if int(v.Month()) < 10 {
		month = "0" + month
	}
	minute := strconv.Itoa(int(v.Minute()))
	if v.Minute() < 10 {
		minute = "0" + minute
	}
	hour := strconv.Itoa(int(v.Hour()))
	if v.Hour() < 10 {
		hour = "0" + hour
	}
	return strconv.Itoa(v.Year()) + "-" + month + "-" + day + " " + hour + ":" + minute + ":00"
}

//Sends email with specified Lead Info. to the LeadSimple Email Importer
//LS Email: new-leadd564b4aaf8@newlead.leadsimple.com
//Zapier Email: i5qc1rg9@robot.zapier.com
func send(address string, name string, email string, phone string, agent string) {
  agent_name := getLeadSimpleName(agent)
  notifyUid := getNotifyUid(agent)
  from := "Show2Lead@gmail.com"
  pass := "Miller2179"
  to := "i5qc1rg9@robot.zapier.com"
  body := "Address: " + address + "\r\n" + "Name: " + name + "\r\n" +
  "Phone: " + phone + "\r\n" + "Email: " + email + "\r\n" + "Assign To: " + agent_name + "\r\n"
  
  body = body + "Notify: " + notifyUid + "\r\n"

  msg := "From: " + from + "\n" +
		"To: " + to + "\n" +
		"Subject: New Lead\n\n" +
		body

	err := smtp.SendMail("smtp.gmail.com:587",
		smtp.PlainAuth("", from, pass, "smtp.gmail.com"),
		from, []string{to}, []byte(msg))

	if err != nil {
		log.Printf("smtp error: %s", err)
		return
	}

	log.Print("sent, visit "+ to)
}

//Function for testing the sending capability and duplication testing.
func tester_send() {
  address_0 := "1029 North Jackon St"
  name := "John Smith"
  email := "john.smith@gmail.com"
  phone := "999.999.9999"
  agent := "Matt M"
  send(address_0,name,email,phone,agent)
}

//Returns ProspectDetailResponse Type (See Struct Above) with Prospect info.
func GetProspectDetails(startTime string, endTime string) (ProspectDetailResponse) {
	body2 := readProspects(startTime, endTime)
	//println(body2)
	var response ProspectDetailResponse
	json.Unmarshal(body2, &response)
	return response
}

//Takes in Agent Name from Showmojo (Gino P / Matt M) and converts
//it to LeadSimple format (Gino / Matt)
func getLeadSimpleName(agent string) string {
  //Matt M -> Matt
    switch agent {
    case "Matt M":
      return "5610"
    case "Gino P":
      return "5611"
    case "Shawn J":
      return "5647"
    default:
      return "0"
    }
  }

func getNotifyUid(agent string) string {
  switch agent {
  case "Matt M":
    return "f9b7bd28-ff13-43db-b0a8-5b8fc4e671cf"
  case "Gino P":
    return "632cc4e0-3427-420e-9616-764b03cb8233"
  case "Shawn J":
    return "db627b21-0419-400d-8210-8dc57ec355e3"
  default:
    return "0"
  }
}

//Calls Post to the prospect data and gets returned raw data
func readProspects(startDate string, endDate string) []byte {//todo handle errs
	// Generated by curl-to-Go: https://mholt.github.io/curl-to-go
	//body := strings.NewReader(`start_date=2017-11-20 19:00:00&end_date=2017-11-20 20:00:00`)
	body := strings.NewReader(`start_date=` + startDate + `&end_date=` + endDate)

  //requests the prospect data
	req, err := http.NewRequest("POST", "https://showmojo.com/api/v3/reports/detailed_prospect_data", body)
	if err != nil {
		// handle err
	}
	req.Header.Set("Authorization", "Token token=\"ec6b7cb47c6f1270949f16215227afd9\"") //todo add to configuration
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		// handle err
	}
	defer resp.Body.Close()
	body2, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		//handle err
	}
	return body2
}
