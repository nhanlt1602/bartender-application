package facker

import (
	"bytes"
	"encoding/json"
	"io"
	"kafka-consumer/application/model"
	"net/http"
)

// FakeAPICallWaitingToRun simulates an API call and returns a *http.Response with a BartenderApIResponse as JSON.
func FakeAPICallWaitingToRun() *http.Response {
	resp := model.BartenderApIResponse{
		Id:        "c623be16-da9b-46e1-80f4-fb8f12f7dad5",
		Status:    "WaitingToRun",
		StatusUrl: "http://127.0.0.1:5159/api/actions/c623be16-da9b-46e1-80f4-fb8f12f7dad5",
	}
	body, _ := json.Marshal(resp)
	return &http.Response{
		Status:     "200",
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader(body)),
		Header:     make(http.Header),
	}
}

// FakeAPICallRunning simulates an API call and returns a *http.Response with a BartenderApIResponse as JSON.
func FakeAPICallRunning() *http.Response {
	resp := model.BartenderApIResponse{
		Id:        "c623be16-da9b-46e1-80f4-fb8f12f7dad5",
		Status:    "Running",
		StatusUrl: "http://127.0.0.1:5159/api/actions/c623be16-da9b-46e1-80f4-fb8f12f7dad5",
	}
	body, _ := json.Marshal(resp)
	return &http.Response{
		Status:     "200",
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader(body)),
		Header:     make(http.Header),
	}
}

// FakeAPICallRanToCompletion simulates an API call and returns a *http.Response with a BartenderApIResponse as JSON.
func FakeAPICallRanToCompletion() *http.Response {
	resp := map[string]interface{}{
		"KeepStatusMinutes": 60.0,
		"Id":                "c623be16-da9b-46e1-80f4-fb8f12f7dad5",
		"SubmittedBy":       "DESKTOP-2IG9JTN\\User",
		"SubmittedTime":     "2025-07-16T22:51:55.2797474+07:00",
		"Status":            "RanToCompletion",
	}
	body, _ := json.Marshal(resp)
	return &http.Response{
		Status:     "200",
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader(body)),
		Header:     make(http.Header),
	}
}

func FakeAPICallRanToWaitingCompletion() *http.Response {
	resp := map[string]interface{}{
		"KeepStatusMinutes": 60.0,
		"Id":                "c623be16-da9b-46e1-80f4-fb8f12f7dad5",
		"SubmittedBy":       "DESKTOP-2IG9JTN\\User",
		"SubmittedTime":     "2025-07-16T22:51:55.2797474+07:00",
		"Status":            "RanToWaitingCompletion",
		"StatusUrl":         "http://127.0.0.1:5159/api/actions/c623be16-da9b-46e1-80f4-fb8f12f7dad5",
		"Tracking": []map[string]interface{}{
			{"Step": "Queued", "Timestamp": "2025-07-16T22:51:55.2797474+07:00"},
			{"Step": "Started", "Timestamp": "2025-07-16T22:52:00.0000000+07:00"},
			{"Step": "Running", "Timestamp": "2025-07-16T22:52:10.0000000+07:00"},
			{"Step": "Completed", "Timestamp": "2025-07-16T22:53:00.0000000+07:00"},
		},
	}
	body, _ := json.Marshal(resp)
	return &http.Response{
		Status:     "200",
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader(body)),
		Header:     make(http.Header),
	}
}

// FakeAPICallErrNoConnection Post "http://127.0.0.1:5159/api/actions": dial tcp 127.0.0.1:5159: connectex: No connection could be made because the target machine actively refused it.
func FakeAPICallErrNoConnection() string {
	return "Post \"http://127.0.0.1:5159/api/actions\": dial tcp 127.0.0.1:5159: connectex: No connection could be made because the target machine actively refused it."
}
