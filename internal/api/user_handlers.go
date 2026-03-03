package api

import (
	"errors"
	"net/http"

	"github.com/ogiovannyoliveira/go-bid/internal/jsonutils"
	"github.com/ogiovannyoliveira/go-bid/internal/services"
	"github.com/ogiovannyoliveira/go-bid/internal/use_case/user"
)

func (api *Api) handleSignupUser(w http.ResponseWriter, r *http.Request) {
	data, problems, err := jsonutils.DecodeValidJson[user.CreateUserReq](r)
	if err != nil {
		_ = jsonutils.EncodeJson(w, r, http.StatusUnprocessableEntity, problems)
		return
	}

	id, err := api.UserService.CreateUser(r.Context(), data.UserName, data.Email, data.Password, data.Bio)
	if err != nil {
		if errors.Is(err, services.ErrDuplicatedUserNameOrEmail) {
			_ = jsonutils.EncodeJson(w, r, http.StatusUnprocessableEntity, map[string]any{
				"error": "Email or username already exists",
			})
			return
		}
	}

	_ = jsonutils.EncodeJson(w, r, http.StatusUnprocessableEntity, map[string]any{
		"id": id,
	})
}

func (api *Api) handleLoginUser(w http.ResponseWriter, r *http.Request) {
	panic("Not implemented")
}

func (api *Api) handleLogoutUser(w http.ResponseWriter, r *http.Request) {
	panic("Not implemented")
}
