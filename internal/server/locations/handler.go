package locations

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/alexPavlikov/auth-service/internal/models"
	"github.com/alexPavlikov/auth-service/internal/server/service"
)

type Handler struct {
	Service *service.Service
}

func New(service *service.Service) *Handler {
	return &Handler{
		Service: service,
	}
}

func (h *Handler) Auth(w http.ResponseWriter, r *http.Request) {
	var us models.UserPayLoad

	dec := json.NewDecoder(r.Body)
	defer r.Body.Close()

	if err := dec.Decode(&us); err != nil {
		//...
	}

	//... хешуем пароль

	us.IP = r.RemoteAddr

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)

	defer cancel()

	user, err := h.Service.Auth(ctx, us)
	if err != nil {
		//...
	}
	fmt.Println(user)
	//... выдаем токен
}

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {

}
