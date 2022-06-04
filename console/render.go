package console

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func (mc *MinerConsole) render(w http.ResponseWriter, resobj interface{}) {
	restxt, e1 := json.Marshal(resobj)
	if e1 != nil {
		mc.renderError(w, "resobj not json object.")
	} else {
		mc.renderJsonByte(w, restxt)
	}
}

func (mc *MinerConsole) renderError(w http.ResponseWriter, errorstr string) {
	mc.renderJsonString(w, fmt.Sprintf(`{"error":"%s"}`, errorstr))
}

func (mc *MinerConsole) renderJsonString(w http.ResponseWriter, jsonstr string) {
	mc.renderJsonByte(w, []byte(jsonstr))
}

func (mc *MinerConsole) renderJsonByte(w http.ResponseWriter, jsonbyte []byte) {
	w.Header().Set("Content-Type", "text/json;charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Write(jsonbyte) // Customized jsondata data
}
