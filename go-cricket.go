package go_cricket

import (
	"encoding/xml"
	"fmt"
	"strconv"
	"io/ioutil"
	"strings"
	"time"
	"net/http"
)

const(
	EVENT_NO_CHANGE = 0
	EVENT_OUT = 1
	EVENT_MATCH_STATUS_CHANGE = 2
	EVENT_OVER_CHANGED = 3
	EVENT_RUN_CHANGE = 4
	CRICBUZZ_URL = "http://synd.cricbuzz.com/j2me/1.0/livematches.xml"
)

type Cricket struct{
	lastFetchedResult MatchStat
}

type Response struct {
	BtTeamName string
	Overs string
	MatchStatus string
	Runs string
}

type ResponseEvent struct {
	EventType int
}

type MatchData struct {
	MatchStats []MatchStat `xml:"match"`
}

type MatchStat struct {
	XMLName xml.Name `xml:"match"`
	Type string `xml:"type,attr"`
	States State `xml:"state"`
	Teams []Team `xml:"Tm"`
	BattingTeam *BattingTeam `xml:"mscr>btTm"`
}

type State struct {
	MatchState string `xml:"mchState,attr"`
	Status string `xml:"status,attr"`
}

type Team struct {
	Name string `xml:"Name,attr"`
}

type InningDetails struct {
	Overs string `xml:"noofovers"`
}

type MatchScore struct {
	BattingTeam *BattingTeam `xml:"btTm"`
	BowlingTeam *BowlingTeam `xml:"blgTm"`
	InningDetails *InningDetails `xml:"inngsdetail"`
}

type BattingTeam struct {
	Name string `xml:"sName,attr"`
	ID string `xml:"id,attr"`
	Inngs []Inning `xml:"Inngs"`
}

type BowlingTeam struct {
	Name string `xml:"sName,attr"`
	ID string `xml:"id,attr"`
	Inngs []Inning `xml:"Inngs"`
}

type Inning struct {
	Description string `xml:"desc,attr"`
	Run string `xml:"r,attr"`
	Overs string `xml:"ovrs,attr"`
	Wickets string `xml:"wkts,attr"`
}

func (m *MatchData ) Print()  {
	for _,v := range m.MatchStats {
		fmt.Println("Type is")
		fmt.Printf("%+v\n", v)
	}
}

func (m *MatchData ) convertToResponse() Response {
	return Response{
	}
}

func (m *MatchStat) TriggerEvent(lastFetchedStat MatchStat,event chan int) {
	var lastBt *BattingTeam
	var newBt *BattingTeam

	if lastFetchedStat.BattingTeam != nil {
		lastBt = lastFetchedStat.BattingTeam
	}

	if(m.BattingTeam != nil){
		newBt = m.BattingTeam
	}else{
		fmt.Println("Match Has not yet Started")
		event <- EVENT_NO_CHANGE
	}

	if(newBt.Inngs != nil && len(newBt.Inngs) > 0){
		in := newBt.Inngs[0]
		run,err := strconv.Atoi(in.Run)
		overs,err := strconv.ParseFloat(in.Overs,32)
		wkts,err := strconv.Atoi(in.Wickets); if err != nil {
			event <- EVENT_NO_CHANGE
		}
		oldRun,_ := strconv.Atoi(lastBt.Inngs[0].Run)
		oldOvers,_ := strconv.ParseFloat(lastBt.Inngs[0].Overs,32)
		oldWkts,_ := strconv.Atoi(lastBt.Inngs[0].Wickets)

		fmt.Println("RUN : ", oldRun, " ", run)
		fmt.Println("OVER : ", int(oldOvers), " ", int(overs))
		fmt.Println("WICKETS : ", oldWkts, " ", wkts)

		if oldRun != run {
			event <- EVENT_RUN_CHANGE
		}
		if int(oldOvers) != int(overs) {
			event <- EVENT_OVER_CHANGED
		}
		if oldWkts != wkts {
			event <- EVENT_OUT
		}
	}
}

func (C *Cricket) Start(event chan int)  {
	var temp MatchData
	var m MatchData
	go func() {
		for {
			resp, _ := http.Get(CRICBUZZ_URL)
			data, _ := ioutil.ReadAll(resp.Body)
			err := xml.Unmarshal(data, &m)
			if err != nil {
				fmt.Print("Error is", err)
			}
			for _,k := range m.MatchStats{
				for _, team := range k.Teams{
					if strings.Compare(team.Name, "NZ") == 0 {
						fmt.Println("Team NZ is playing")
						if len(temp.MatchStats) > 0 {
							k.TriggerEvent(temp.MatchStats[0],event)
						}
					}
				}
			}
			temp = m
			time.Sleep(time.Second*10)
		}
	}()
}
