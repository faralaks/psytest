package hendlers

import (
	"context"
	"db"
	"encoding/json"
	"go.mongodb.org/mongo-driver/bson"
	"net/http"
	"time"
	. "tools"
)

type LoginDataResponse struct {
	Status       string `json:"status"`
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

func Login(w http.ResponseWriter, r *http.Request) {
	login := NewB64LowString(TrimStr(r.FormValue("login"), 40))
	pas := Encrypt(TrimStr(r.FormValue("password"), 50))

	var user User
	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)

	err := db.UsersCol.FindOne(ctx, bson.M{"login": login, "pas": pas}).Decode(&user)
	if err != nil {
		JsonMsg{Kind: BadAuthKind}.SendMsg(w)
		return
	}

	at, rt, err := CreateTokens(user.Uid.Hex(), user.Status)
	if err != nil {
		JsonMsg{Kind: FatalKind, Msg: "Не удалось создать токены | " + err.Error()}.SendMsg(w)
		return
	}
	http.SetCookie(w, &http.Cookie{Name: "AccessToken", Value: at, HttpOnly: true, Expires: time.Now().Add(time.Hour * 10)})
	http.SetCookie(w, &http.Cookie{Name: "RefreshToken", Value: rt, HttpOnly: true, Expires: time.Now().Add(time.Hour * 10)})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(LoginDataResponse{user.Status, rt, at})

}