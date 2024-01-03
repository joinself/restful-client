package self

import (
	"github.com/joinself/restful-client/internal/connection"
	"github.com/joinself/restful-client/internal/entity"
	"github.com/joinself/restful-client/internal/fact"
	"github.com/joinself/restful-client/internal/message"
	"github.com/joinself/restful-client/internal/request"
	"github.com/joinself/restful-client/pkg/log"
	"github.com/joinself/restful-client/pkg/support"
	"github.com/joinself/restful-client/pkg/webhook"
	selfsdk "github.com/joinself/self-go-sdk"
)

type Runner interface {
	Run(app entity.App) error
	Stop(id string)
	Get(id string) (*selfsdk.Client, bool)
	Poster(id string) (webhook.Poster, bool)
}

type runner struct {
	runners    map[string]Service
	cRepo      connection.Repository
	fRepo      fact.Repository
	mRepo      message.Repository
	rRepo      request.Repository
	logger     log.Logger
	rService   request.Service
	storageKey string
	storageDir string
}

type RunnerConfig struct {
	ConnectionRepo connection.Repository
	FactRepo       fact.Repository
	MessageRepo    message.Repository
	RequestRepo    request.Repository
	Logger         log.Logger
	RequestService request.Service
	StorageKey     string
	StorageDir     string
}

func NewRunner(config RunnerConfig) Runner {
	return &runner{
		runners:    map[string]Service{},
		cRepo:      config.ConnectionRepo,
		fRepo:      config.FactRepo,
		mRepo:      config.MessageRepo,
		rRepo:      config.RequestRepo,
		logger:     config.Logger,
		rService:   config.RequestService,
		storageKey: config.StorageKey,
		storageDir: config.StorageDir,
	}
}

func (r *runner) Get(id string) (*selfsdk.Client, bool) {
	val, ok := r.runners[id]
	return val.Get(), ok
}

func (r *runner) Poster(id string) (webhook.Poster, bool) {
	val, ok := r.runners[id]
	if !ok {
		return nil, false
	}

	return val.Poster(), true
}

func (r *runner) Run(app entity.App) error {
	r.logger.Infof("setting up app %s", app.ID)
	client, err := r.setupSelfClient(app)
	if err != nil {
		r.logger.Errorf("ERROR setting up app %s : %s", app.ID, err.Error())
		return err
	}

	r.runners[app.ID] = NewService(Config{
		ConnectionRepo: r.cRepo,
		FactRepo:       r.fRepo,
		MessageRepo:    r.mRepo,
		RequestRepo:    r.rRepo,
		Logger:         r.logger,
		RequestService: r.rService,
		SelfClient:     support.NewSelfClient(client),
		Poster:         webhook.NewWebhook(app.Callback),
	})
	r.logger.Infof("trying to start %s", app.ID)
	r.runners[app.ID].Run()

	r.logger.Infof("app %s started", app.ID)
	return nil
}

func (r *runner) Stop(id string) {
	r.runners[id].Stop()
}

func (r *runner) setupSelfClient(app entity.App) (*selfsdk.Client, error) {
	selfConfig := selfsdk.Config{
		SelfAppID:           app.ID,
		SelfAppDeviceSecret: app.DeviceSecret,
		StorageKey:          r.storageKey,
		StorageDir:          r.storageDir,
	}

	// TODO: recover this piece if we eventually need to.
	if app.Env != "production" {
		if app.Env == "development" {
			// selfConfig.APIURL = c.SelfAPIURL
			// selfConfig.MessagingURL = c.SelfMessagingURL
		} else {
			selfConfig.Environment = app.Env
		}
	}

	return selfsdk.New(selfConfig)
}

// RunService executes the listeners specified on the Service.
func RunService(service Service, logger log.Logger) {
	service.Run()
}
