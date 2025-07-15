package model

// API response: {"Id":"c6027b37-c08e-4284-b36d-d35122fdf798","Status":"Running","StatusUrl":"http://127.0.0.1:5159/api/actions/c6027b37-c08e-4284-b36d-d35122fdf798"}
type BartenderApIResponse struct {
	Id        string `json:"Id"`
	Status    string `json:"Status"`
	StatusUrl string `json:"StatusUrl"` // API will call to check the status of the job
}
