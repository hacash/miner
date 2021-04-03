package minerrelayservice

import (
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

func ResponseCreateData(key string, value interface{}) map[string]interface{} {
	var data = map[string]interface{}{}
	data[key] = value
	return data
}

func ResponseLocation(w http.ResponseWriter, url string) {
	w.Header().Set("Location", url) //跳转地址设置
	w.WriteHeader(302)
}

func ResponseError(w http.ResponseWriter, err error) {
	ResponseErrorString(w, err.Error())
}

func ResponseErrorString(w http.ResponseWriter, errstr string) {
	ResponseErrorStringWithCode(w, 1, errstr)
}

func ResponseErrorStringWithCode(w http.ResponseWriter, errcode int, errstr string) {
	errobj := map[string]interface{}{}
	errobj["ret"] = errcode
	errobj["errmsg"] = errstr
	ResponseJSON(w, errobj)
}

func ResponseList(w http.ResponseWriter, data interface{}) {
	errstr := map[string]interface{}{}
	errstr["ret"] = 0
	errstr["list"] = data
	ResponseData(w, errstr)
}

func ResponseData(w http.ResponseWriter, data map[string]interface{}) {
	if data == nil {
		resdts := map[string]interface{}{}
		resdts["ret"] = 0
		data = resdts
	} else if _, ok := data["ret"]; !ok {
		data["ret"] = 0
	}
	ResponseJSON(w, data)
}

func ResponseJSON(w http.ResponseWriter, resobj interface{}) error {

	// return
	restxt, e1 := json.Marshal(resobj)
	if e1 != nil {
		return e1
	} else {
		e2 := ResponseJSONbytes(w, restxt)
		if e2 != nil {
			return e2
		}
	}
	return nil
}

func ResponseJSONbytes(w http.ResponseWriter, content []byte) error {

	header := w.Header()
	key1 := "Access-Control-Allow-Origin"
	if "" == header.Get(key1) {
		header.Set(key1, "*")
	}
	header.Set("Content", "text/json; utf-8")

	// return
	w.WriteHeader(200)
	_, e2 := w.Write(content)
	if e2 != nil {
		return e2
	}
	return nil
}

///////////////////////////////////////////////////////////////////////////

func CheckParamBool(r *http.Request, key string, defValue bool) bool {
	boolString := strings.Trim(r.FormValue(key), " ")
	valen := len(boolString)
	if valen == 0 || boolString == "0" || strings.ToLower(boolString) == "false" {
		return false
	} else if valen > 0 {
		return true
	}
	return defValue
}

func CheckParamUint64(r *http.Request, key string, defValue uint64) uint64 {
	var value uint64 = defValue
	if v := r.FormValue(key); len(v) > 0 {
		if i, e := strconv.ParseUint(v, 10, 0); e == nil {
			value = i
		}
	}
	return value
}

func CheckParamString(r *http.Request, key string, defValue string) string {
	if v := r.FormValue(key); len(v) > 0 {
		return v
	}
	return defValue
}

func CheckParamHex(r *http.Request, key string, defv []byte) []byte {
	if v := r.FormValue(key); len(v) > 0 {
		v = strings.TrimPrefix(v, "0x")
		hexdts, e := hex.DecodeString(v)
		if e != nil {
			return nil
		}
		return hexdts // ok
	}
	return defv
}

func CheckParamHexMustLen(r *http.Request, key string, mustLen int) []byte {
	if v := r.FormValue(key); len(v) > 0 {
		v = strings.TrimPrefix(v, "0x")
		if len(v) != mustLen*2 {
			return nil // len error
		}
		hexdts, e := hex.DecodeString(v)
		if e != nil {
			return nil
		}
		return hexdts // ok
	}
	return nil
}

func CheckParamUint64Must(r *http.Request, w http.ResponseWriter, key string) (uint64, bool) {
	var value uint64 = 0
	if v := r.FormValue(key); len(v) > 0 {
		if i, e := strconv.ParseUint(v, 10, 0); e == nil {
			value = i
		} else {
			ResponseErrorString(w, "param <"+key+"> format error.")
			return 0, false
		}
	} else {
		ResponseErrorString(w, "param <"+key+"> must give.")
		return 0, false
	}
	return value, true
}
