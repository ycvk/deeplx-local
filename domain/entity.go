package domain

type TranslateRequest struct {
	Text       string `json:"text"`
	SourceLang string `json:"source_lang"`
	TargetLang string `json:"target_lang"`
}

// TranslateResponse
//
//	{
//	 "code": 200,
//	 "id": 4548710002,
//	 "data": "我爱你",
//	 "alternatives": [
//	   "我爱你们",
//	   "我爱您",
//	   "我爱妳"
//	 ]
//	}
type TranslateResponse struct {
	Code int    `json:"code"`
	Data string `json:"data"`
}

type YingTuResponse struct {
	Code int `json:"code"`
	Data struct {
		Total int                 `json:"total"`
		Arr   []YingTuResponseArr `json:"arr"`
	}
}

type YingTuResponseArr struct {
	Url string `json:"url"`
}

type Quake360Response struct {
	Code    int                    `json:"code"`
	Message string                 `json:"message"`
	Data    []Quake360ResponseData `json:"data"`
}

type Quake360ResponseData struct {
	Domain string `json:"domain"`
	Id     string `json:"id"` // hpjx.e.eceping.net_443_tcp 或者 116.204.90.243_2001_tcp
}
