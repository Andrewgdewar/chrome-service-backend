package routes

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/RedHatInsights/chrome-service-backend/rest/models"
	"github.com/RedHatInsights/chrome-service-backend/rest/service"
	"github.com/RedHatInsights/chrome-service-backend/rest/util"
	"github.com/go-chi/chi/v5"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func handleDashboardError(err error, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		resp := util.ErrorResponse{
			Errors: []string{err.Error()},
		}
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(resp)
		return
	} else if err != nil && errors.Is(err, util.ErrNotAuthorized) {
		resp := util.ErrorResponse{
			Errors: []string{"not authorized"},
		}
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(resp)

		return
	} else if err != nil {
		logrus.Errorln(err)
		resp := util.ErrorResponse{
			Errors: []string{err.Error()},
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(resp)
		return
	}

	resp := util.ErrorResponse{
		Errors: []string{"internal server error"},
	}

	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(resp)
}

func handleDashboardResponse[T interface{}, RespType util.ListResponse[T] | util.EntityResponse[T]](rep RespType, err error, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		handleDashboardError(err, w)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(rep)
}

func GetDashboardTemplates(w http.ResponseWriter, r *http.Request) {
	var userDashboardTemplates []models.DashboardTemplate
	var err error
	dashboardParam := r.URL.Query().Get("dashboard")
	user := r.Context().Value(util.USER_CTX_KEY).(models.UserIdentity)
	userID := user.ID

	dashboard := models.AvailableTemplates(dashboardParam)
	err = dashboard.IsValid()
	if dashboard != "" && err != nil {
		handleDashboardError(err, w)
		return

	}
	userDashboardTemplates, err = service.GetDashboardTemplate(userID, dashboard)

	response := util.ListResponse[models.DashboardTemplate]{
		Data: userDashboardTemplates,
	}

	handleDashboardResponse[models.DashboardTemplate, util.ListResponse[models.DashboardTemplate]](response, err, w)
}

func UpdateDashboardTemplate(w http.ResponseWriter, r *http.Request) {
	var dashboardTemplate models.DashboardTemplate
	var err error
	templateID := chi.URLParam(r, "templateId")
	user := r.Context().Value(util.USER_CTX_KEY).(models.UserIdentity)
	userID := user.ID

	templateIdUint, err := strconv.ParseUint(templateID, 10, 64)

	if err != nil {
		handleDashboardError(errors.New("invalid template ID"), w)
		return
	}

	err = json.NewDecoder(r.Body).Decode(&dashboardTemplate)
	if err != nil {
		handleDashboardError(errors.New("unable to parse payload to dashboard template"), w)
		return
	}

	updatedTemplate, err := service.UpdateDashboardTemplate(uint(templateIdUint), userID, dashboardTemplate)
	resp := util.EntityResponse[models.DashboardTemplate]{
		Data: updatedTemplate,
	}
	handleDashboardResponse[models.DashboardTemplate, util.EntityResponse[models.DashboardTemplate]](resp, err, w)
}

func GetBaseDashboardTemplates(w http.ResponseWriter, r *http.Request) {
	dashboardParam := r.URL.Query().Get("dashboard")

	if dashboardParam == "" {
		templates := service.GetAllBaseTemplates()
		resp := util.ListResponse[models.BaseDashboardTemplate]{
			Data: templates,
		}
		handleDashboardResponse[models.BaseDashboardTemplate, util.ListResponse[models.BaseDashboardTemplate]](resp, nil, w)
		return
	}

	template, err := service.GetDashboardTemplateBase(models.AvailableTemplates(dashboardParam))

	resp := util.EntityResponse[models.BaseDashboardTemplate]{
		Data: template,
	}

	handleDashboardResponse[models.BaseDashboardTemplate, util.EntityResponse[models.BaseDashboardTemplate]](resp, err, w)
}

func CopyDashboardTemplate(w http.ResponseWriter, r *http.Request) {
	var err error
	templateID := chi.URLParam(r, "templateId")
	user := r.Context().Value(util.USER_CTX_KEY).(models.UserIdentity)
	userID := user.ID

	templateIdUint, err := strconv.ParseUint(templateID, 10, 64)

	if err != nil {
		handleDashboardError(errors.New("invalid template ID"), w)
		return
	}

	dashboardTemplate, err := service.CopyDashboardTemplate(userID, uint(templateIdUint))

	response := util.EntityResponse[models.DashboardTemplate]{
		Data: dashboardTemplate,
	}

	handleDashboardResponse[models.DashboardTemplate, util.EntityResponse[models.DashboardTemplate]](response, err, w)
}

func DeleteDashboardTemplate(w http.ResponseWriter, r *http.Request) {
	templateID := chi.URLParam(r, "templateId")
	user := r.Context().Value(util.USER_CTX_KEY).(models.UserIdentity)
	userID := user.ID

	templateIdUint, err := strconv.ParseUint(templateID, 10, 64)
	if err != nil {
		handleDashboardError(errors.New("invalid template ID"), w)
		return
	}

	err = service.DeleteTemplate(userID, uint(templateIdUint))
	if err != nil {
		handleDashboardError(err, w)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func ChangeDefaultTemplate(w http.ResponseWriter, r *http.Request) {
	templateID := chi.URLParam(r, "templateId")
	user := r.Context().Value(util.USER_CTX_KEY).(models.UserIdentity)
	userID := user.ID

	templateIdUint, err := strconv.ParseUint(templateID, 10, 64)

	if err != nil {
		handleDashboardError(errors.New("invalid template ID"), w)
		return
	}

	dashboardTemplate, err := service.ChangeDefaultTemplate(userID, uint(templateIdUint))
	resp := util.EntityResponse[models.DashboardTemplate]{
		Data: dashboardTemplate,
	}
	handleDashboardResponse[models.DashboardTemplate, util.EntityResponse[models.DashboardTemplate]](resp, err, w)
}

func MakeDashboardTemplateRoutes(sub chi.Router) {
	sub.Get("/", GetDashboardTemplates)
	sub.Patch("/{templateId}", UpdateDashboardTemplate)
	sub.Delete("/{templateId}", DeleteDashboardTemplate)
	sub.Post("/{templateId}/copy", CopyDashboardTemplate)
	sub.Post("/{templateId}/default", ChangeDefaultTemplate)

	sub.Get("/base-template", GetBaseDashboardTemplates)
}
