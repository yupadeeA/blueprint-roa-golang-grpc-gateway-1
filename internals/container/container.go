package container

import (
	"fmt"
	"git.robodev.co/imp/shared-utility/validator"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/robowealth-mutual-fund/blueprint-roa-golang/internals/config"
	"github.com/robowealth-mutual-fund/blueprint-roa-golang/internals/controller"
	controllerProduct "github.com/robowealth-mutual-fund/blueprint-roa-golang/internals/controller/product"
	"github.com/robowealth-mutual-fund/blueprint-roa-golang/internals/infrastructure/database"
	grpcServer "github.com/robowealth-mutual-fund/blueprint-roa-golang/internals/infrastructure/grpcServer"
	httpServer "github.com/robowealth-mutual-fund/blueprint-roa-golang/internals/infrastructure/httpServer"
	"github.com/robowealth-mutual-fund/blueprint-roa-golang/internals/infrastructure/jaeger"
	"github.com/robowealth-mutual-fund/blueprint-roa-golang/internals/repository/postgres"
	serviceProduct "github.com/robowealth-mutual-fund/blueprint-roa-golang/internals/service/product"
	"github.com/robowealth-mutual-fund/blueprint-roa-golang/internals/service/product/wrapper"
	"github.com/robowealth-mutual-fund/blueprint-roa-golang/internals/utils"
	"github.com/robowealth-mutual-fund/blueprint-roa-golang/internals/utils/logrus"
	log "github.com/sirupsen/logrus"
	"go.uber.org/dig"
	"net/http"
)

type Container struct {
	container *dig.Container
}

func (c *Container) Configure() error {
	if err := c.container.Provide(wrapper.WrapProduct, dig.Name("wrapperProduct")); err != nil {
		return err
	}
	servicesConstructors := []interface{}{
		config.NewConfiguration,
		grpcServer.NewServer,
		database.NewServerBase,
		http.NewServeMux,
		httpServer.NewServer,
		runtime.NewServeMux,
		jaeger.NewJaeger,
		logrus.NewLog,
		controller.NewHealthZController,
		controller.NewPingPongController,
		validator.NewCustomValidator,
		controllerProduct.NewController,
		serviceProduct.NewService,
		postgres.NewRepository,
		utils.NewUtils,
		utils.NewCustomValidator,
	}

	for _, service := range servicesConstructors {
		if err := c.container.Provide(service); err != nil {
			return err
		}
	}
	appConfig := config.NewConfiguration()
	jaeger.NewJaeger(appConfig)
	logrus.NewLog()
	return nil
}

func (c *Container) Start() error {
	log.Info("Start Container")
	if err := c.container.Invoke(func(s *grpcServer.Server, h *httpServer.Server, v *validator.CustomValidator) {
		go func() {
			_ = h.Start()
		}()
		s.Start()

	}); err != nil {
		fmt.Printf("%s", err)

		return err
	}

	return nil
}

//MigrateDB ...
func (c *Container) MigrateDB() error {
	log.Info("Start Container DB")

	if err := c.container.Invoke(func(d *database.DB) {
		d.MigrateDB()
	}); err != nil {
		return err
	}

	return nil
}

func NewContainer() (*Container, error) {
	d := dig.New()

	container := &Container{
		container: d,
	}

	if err := container.Configure(); err != nil {
		return nil, err
	}

	return container, nil
}
