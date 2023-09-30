package model

// MyMsg 我的消息
type MyMsg struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	Time    string `json:"time"`
}

type MyBalance struct {
	Title   string  `json:"title"`
	Balance float64 `json:"balance"`
	Time    string  `json:"time"`
}
