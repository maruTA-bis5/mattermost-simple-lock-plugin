package main

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/mattermost/mattermost-server/model"
	"github.com/mattermost/mattermost-server/plugin"
)

type Plugin struct {
	plugin.MattermostPlugin

	// configurationLock synchronizes access to the configuration.
	configurationLock sync.RWMutex
	resourceLock      sync.RWMutex

	// configuration is the active plugin configuration. Consult getConfiguration and
	// setConfiguration for usage.
	configuration *configuration

	router       *mux.Router
	serverConfig *model.Config
}

func (p *Plugin) OnActivate() error {
	command := &model.Command{
		Trigger:          "lock",
		DisplayName:      "Lock",
		Description:      "Lock shared resource.",
		AutoComplete:     true,
		AutoCompleteDesc: "Lock shared resource.",
		AutoCompleteHint: "resource [message]",
	}
	err := p.API.RegisterCommand(command)
	if err != nil {
		return err
	}
	p.router = p.initApi()
	p.OnConfigurationChange()
	return nil
}

func (p *Plugin) ExecuteCommand(c *plugin.Context, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	p.resourceLock.Lock()
	defer p.resourceLock.Unlock()

	parts := strings.Split(args.Command, " ")
	targetResource := parts[1]
	if p.isAlreadyLocked(targetResource) {
		errResponse := &model.CommandResponse{
			ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
			Text:         fmt.Sprintf("Resource [%s] is already locked.", targetResource),
		}
		return errResponse, nil
	}
	p.lockResource(targetResource)

	message := strings.Join(parts[2:], " ")

	userId := args.UserId
	user, err := p.API.GetUser(userId)
	if err != nil {
		return nil, err
	}
	username := user.Username

	postPretext := fmt.Sprintf("Resource[%s] Locked by %s\n%s", targetResource, username, message)
	integrationContext := model.StringInterface{
		"resource":        targetResource,
		"originalMessage": postPretext,
	}
	releaseIntegration := &model.PostActionIntegration{
		URL:     fmt.Sprintf("%s/plugins/%s/api/release", *p.serverConfig.ServiceSettings.SiteURL, manifest.Id),
		Context: integrationContext,
	}

	releaseAction := &model.PostAction{
		Name:        "Release Lock",
		Integration: releaseIntegration,
	}

	toReleaseAttachment := &model.SlackAttachment{
		Pretext: postPretext,
		Actions: []*model.PostAction{releaseAction},
	}

	response := &model.CommandResponse{
		ResponseType: model.COMMAND_RESPONSE_TYPE_IN_CHANNEL,
		Username:     username,
		Attachments:  []*model.SlackAttachment{toReleaseAttachment},
	}
	return response, nil
}

func (p *Plugin) isAlreadyLocked(targetResource string) bool {
	key := "simplelock_locked_" + targetResource
	value, err := p.API.KVGet(key)
	if err != nil {
		// TODO error log
	}
	return value != nil
}

func (p *Plugin) lockResource(targetResource string) {
	key := "simplelock_locked_" + targetResource
	value := []byte("locked")
	err := p.API.KVSet(key, value)
	if err != nil {
		// TODO error log
	}
}

func (p *Plugin) ServeHTTP(c *plugin.Context, w http.ResponseWriter, r *http.Request) {
	p.router.ServeHTTP(w, r)
}

func (p *Plugin) initApi() *mux.Router {
	r := mux.NewRouter()
	apiRouter := r.PathPrefix("/api").Subrouter()
	apiRouter.HandleFunc("/release", p.handleRelease).Methods("POST")
	return r
}

func (p *Plugin) handleRelease(w http.ResponseWriter, r *http.Request) {
	request := model.PostActionIntegrationRequestFromJson(r.Body)
	if request == nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	targetResource := request.Context["resource"].(string)
	if targetResource == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	p.resourceLock.Lock()
	defer p.resourceLock.Unlock()

	response := &model.PostActionIntegrationResponse{}
	if !p.isAlreadyLocked(targetResource) {
		response.EphemeralText = "Resourece [" + targetResource + "] did not locked."
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(response.ToJson())
		return
	}

	p.unlockResource(targetResource)

	originalMessage := request.Context["originalMessage"].(string)
	update := &model.Post{}
	update.Message = originalMessage + "\n" + "Lock released at: " + time.Now().Format(time.RFC3339)
	props := model.StringInterface{}
	props["attachments"] = []*model.SlackAttachment{}
	update.Props = props

	response.Update = update

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(response.ToJson())
}

func (p *Plugin) unlockResource(targetResource string) {
	key := "simplelock_locked_" + targetResource
	err := p.API.KVDelete(key)
	if err != nil {
		// TODO error log
	}
}
