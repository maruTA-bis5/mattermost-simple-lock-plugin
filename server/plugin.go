package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/mattermost/mattermost-server/model"
	"github.com/mattermost/mattermost-server/plugin"
)

type Plugin struct {
	plugin.MattermostPlugin

	// configurationLock synchronizes access to the configuration.
	resourceLock sync.RWMutex

	// configuration is the active plugin configuration. Consult getConfiguration and
	// setConfiguration for usage.
	configuration *configuration

	router *mux.Router
}

func (p *Plugin) OnActivate() error {
	command := &model.Command{
		Trigger:	"lock",
		DisplayName:	"Lock",
		Description:	"Lock shared resource.",
		AutoComplete:	"true",
		AutoCompleteDesc:	"Lock shared resource.",
		AutoCompleteHint:	"resource [message]",
	}
	err := p.API.RegisterCommand(command)
	if err != nil {
		return err
	}
	p.router = p.initApi()
	return nil
}

func (p *Plugin) ExecuteCommand(c *plugin.Context, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	p.resourceLock.Lock()
	defer p.resourceLock.Unlock()

	targetResource := // TODO
	if p.isAlreadyLocked(targetResource) {
		errResponse := &model.CommandResponse {
			ResponseType:	model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
			Text:			"Resource ["+targetResource+"] is already locked.",
		}
		return errResponse, nil
	}
	p.lockResource(targetResource)

	message := // TODO

	userId := args.UserId
	user, err := p.API.GetUser(userId)
	if err != nil {
		return nil, err
	}
	username := user.Username

	channelId := args.ChannelId

	integrationContext := make(map[string]string)
	integrationContext["resource"] = resourceName
	integrationContext["message"] = message
	integrationContext["username"] = username
	releaseIntegration := &model.PostActionIntegration {
		URL:	fmt.Sprintf("%s/plugins/%s/api/release", siteUrl, pluginId),
		Context: integrationContext,
	}

	releaseAction := &model.PostAction {
		Name:	"Release Lock",
		Integration:	releaseIntegration,
	}

	toReleaseAttachment := &model.SlackAttachment {
		PreText:	targetResource+" "+message+" by "+username,
		Actions:	releaseAction,
	}

	response := &model.CommandResponse{
		ResponseType:	model.COMMAND_RESPONSE_TYPE_IN_CHANNEL,
		Text:			text,
		Username:		username,
		ChannelId:		channelId,
		Attachments:	attachments,
	}
	return response, nil
}

func (p *Plugin) isAlreadyLocked(targetResource string) bool {
	//TODO
	return false
}

func (p *Plugin) lockResource(targetResource string) {
	// TODO
}

func (p *Plugin) ServeHTTP(c *plugin.Context, w http.ResponseWriter, r *http.Request) {
	p.router.ServeHTTP(w, r)
}

func (p *Plugin) initApi() *mux.Router {
	r := mux.NewRouter()
	apiRouter := r.PathPrefix("/api").SubRouter()
	apiRouter.HandleFunc("/release"), p.handleRelease).Methods("POST")
	return r
}

func (p *Plugin) handleRelease(w http.ResponseWriter, r *http.Request) {
	request := model.PostActionIntegrationRequestFromJson(r.Body)
	if request == nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	targetResource := request.Context["resource"]
	if targetResource == nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	p.resourceLock.Lock()
	defer p.resourceLock.Unlock()

	response := &model.PostActionIntegrationResponse{}
	if !p.isAlreadyLocked(targetResource) {
		response.EphemeralText("Resourece ["+targetResource+"] did not locked.")
		return response
	}

	p.unlockResourece(targetResource)

	message := request.Context["message"]
	username := request.Context["username"]
	originalMessage := targetResource+" "+message+" by "+username
	update := &model.Post{}
	update.Message = originalMessage+"\n"+"Lock released at: "+time.Now().String()
	props := make(StringInterface)
	props["attachments"] = make([]model.SlackAttachment)
	update.Props = props

	response.Update = update

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(response.ToJson())
}

func (p *Plugin) unlockResource(targetResource string) {
	// TODO
}

