package model

import "time"

type BartenderApIResponse struct {
	Id        string `json:"Id"`
	Status    string `json:"Status"`
	StatusUrl string `json:"StatusUrl"` // API will call to check the status of the job
}

type BartenderTrackingStatusResponse struct {
	KeepStatusMinutes float64   `json:"KeepStatusMinutes"`
	Id                string    `json:"Id"`
	SubmittedBy       string    `json:"SubmittedBy"`
	SubmittedTime     time.Time `json:"SubmittedTime"`
	Status            string    `json:"Status"`
}

//{
//"KeepStatusMinutes": 60.0,
//"Id": "c623be16-da9b-46e1-80f4-fb8f12f7dad5",
//"SubmittedBy": "DESKTOP-2IG9JTN\\User",
//"SubmittedTime": "2025-07-16T22:51:55.2797474+07:00",
//"Status": "RanToCompletion"
//}
