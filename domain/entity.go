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
