package simulator

type SetCommand struct {
	Id      string      `json:"id"`
	Version string      `json:"version"`
	Params  interface{} `json:"params"`
}

type SetCommandResp struct {
	Id   string `json:"id"`
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}
