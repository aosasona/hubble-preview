package api

import (
	"go.trulyao.dev/hubble/web/internal/config"
	"go.trulyao.dev/hubble/web/internal/mail"
	"go.trulyao.dev/hubble/web/internal/objectstore"
	"go.trulyao.dev/hubble/web/internal/otp"
	"go.trulyao.dev/hubble/web/internal/plugin/spec"
	"go.trulyao.dev/hubble/web/internal/queue"
	"go.trulyao.dev/hubble/web/internal/repository"
	"go.trulyao.dev/hubble/web/pkg/llm"
)

type Handler interface {
	Auth() AuthHandler
	User() UserHandler
	Mfa() MfaHandler
	Workspace() WorkspaceHandler
	Collection() CollectionHandler
	Entry() EntryHandler
	Plugin() PluginHandler
}

type api struct {
	// dependencies
	config        *config.Config
	repository    repository.Repository
	mailer        mail.Mailer
	otpManager    otp.Manager
	pluginManager spec.Manager
	objectsStore  *objectstore.Store
	queue         *queue.Queue
	llm           *llm.LLM

	// handlers
	authHandler       AuthHandler
	userHandler       UserHandler
	mfaHandler        MfaHandler
	workspaceHandler  WorkspaceHandler
	collectionHandler CollectionHandler
	entryHandler      EntryHandler
	pluginHandler     PluginHandler
}

type Deps struct {
	Repo          repository.Repository
	Config        *config.Config
	Mailer        mail.Mailer
	OtpManager    otp.Manager
	PluginManager spec.Manager
	ObjectStore   *objectstore.Store
	Queue         *queue.Queue
	LLM           *llm.LLM
}

type baseHandler struct {
	config       *config.Config
	mailer       mail.Mailer
	objectsStore *objectstore.Store
	queue        *queue.Queue
	llm          *llm.LLM

	repos         repository.Repository
	otpManager    otp.Manager
	pluginManager spec.Manager
}

func New(deps *Deps) Handler {
	return &api{
		config:            deps.Config,
		repository:        deps.Repo,
		mailer:            deps.Mailer,
		otpManager:        deps.OtpManager,
		objectsStore:      deps.ObjectStore,
		pluginManager:     deps.PluginManager,
		queue:             deps.Queue,
		llm:               deps.LLM,
		authHandler:       nil,
		userHandler:       nil,
		mfaHandler:        nil,
		workspaceHandler:  nil,
		collectionHandler: nil,
		entryHandler:      nil,
		pluginHandler:     nil,
	}
}

func (a *api) makeBaseHandler() *baseHandler {
	return &baseHandler{
		config:        a.config,
		mailer:        a.mailer,
		repos:         a.repository,
		otpManager:    a.otpManager,
		pluginManager: a.pluginManager,
		objectsStore:  a.objectsStore,
		queue:         a.queue,
		llm:           a.llm,
	}
}

func (a *api) Auth() AuthHandler {
	if a.authHandler == nil {
		a.authHandler = &authHandler{a.makeBaseHandler()}
	}

	return a.authHandler
}

func (a *api) User() UserHandler {
	if a.userHandler == nil {
		a.userHandler = &userHandler{a.makeBaseHandler()}
	}

	return a.userHandler
}

func (a *api) Mfa() MfaHandler {
	if a.mfaHandler == nil {
		a.mfaHandler = &mfaHandler{a.makeBaseHandler()}
	}

	return a.mfaHandler
}

func (a *api) Workspace() WorkspaceHandler {
	if a.workspaceHandler == nil {
		a.workspaceHandler = &workspaceHandler{a.makeBaseHandler()}
	}

	return a.workspaceHandler
}

func (a *api) Collection() CollectionHandler {
	if a.collectionHandler == nil {
		a.collectionHandler = &collectionHandler{a.makeBaseHandler()}
	}

	return a.collectionHandler
}

func (a *api) Entry() EntryHandler {
	if a.entryHandler == nil {
		a.entryHandler = &entryHandler{a.makeBaseHandler()}
	}

	return a.entryHandler
}

func (a *api) Plugin() PluginHandler {
	if a.pluginHandler == nil {
		a.pluginHandler = &pluginHandler{a.makeBaseHandler()}
	}

	return a.pluginHandler
}
