//
// Copyright (c) 2021 One Track Consulting
// Copyright (C) 2024-2025 IOTech Ltd
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package app

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/edgexfoundry/app-functions-sdk-go/v4/internal/bootstrap/container"
	"github.com/edgexfoundry/app-functions-sdk-go/v4/internal/common"
	"github.com/edgexfoundry/app-functions-sdk-go/v4/internal/store/db"
	"github.com/edgexfoundry/app-functions-sdk-go/v4/internal/store/db/postgres"
	"github.com/edgexfoundry/app-functions-sdk-go/v4/internal/store/db/redis"
	"github.com/edgexfoundry/app-functions-sdk-go/v4/pkg/interfaces"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/environment"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/secret"
	bootstrapConfig "github.com/edgexfoundry/go-mod-bootstrap/v4/config"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"
)

// RegisterCustomStoreFactory allows registration of alternative storage implementation to back the Store&Forward loop
func (svc *Service) RegisterCustomStoreFactory(name string, factory func(cfg bootstrapConfig.Database, cred bootstrapConfig.Credentials) (interfaces.StoreClient, error)) error {
	if name == db.RedisDB {
		return fmt.Errorf("cannot register factory for reserved name %q", name)
	}

	if svc.customStoreClientFactories == nil {
		svc.customStoreClientFactories = make(map[string]func(db bootstrapConfig.Database, cred bootstrapConfig.Credentials) (interfaces.StoreClient, error), 1)
	}

	svc.customStoreClientFactories[strings.ToUpper(name)] = factory

	return nil
}

func (svc *Service) createStoreClient(database bootstrapConfig.Database, credentials bootstrapConfig.Credentials) (interfaces.StoreClient, error) {
	switch strings.ToLower(database.Type) {
	case db.RedisDB:
		return redis.NewClient(database, credentials)
	case db.Postgres:
		return postgres.NewClient(svc.ctx.appCtx, database, credentials, svc.lc, svc.serviceKey)
	default:
		if factory, found := svc.customStoreClientFactories[strings.ToUpper(database.Type)]; found {
			return factory(database, credentials)
		}
		return nil, db.ErrUnsupportedDatabase
	}
}

func (svc *Service) startStoreForward() {
	var storeForwardEnabledCtx context.Context
	svc.ctx.storeForwardWg = &sync.WaitGroup{}
	storeForwardEnabledCtx, svc.ctx.storeForwardCancelCtx = context.WithCancel(context.Background())
	svc.runtime.StartStoreAndForward(svc.ctx.appWg, svc.ctx.appCtx, svc.ctx.storeForwardWg, storeForwardEnabledCtx, svc.serviceKey)
}

func (svc *Service) stopStoreForward() {
	svc.lc.Info("Canceling Store and Forward retry loop")
	if svc.ctx.storeForwardCancelCtx != nil {
		svc.ctx.storeForwardCancelCtx()
	}
	svc.ctx.storeForwardWg.Wait()
}

func initializeStoreClient(config *common.ConfigurationStruct, svc *Service) error {
	// Only need the database client if Store and Forward is enabled
	if !config.GetWritableInfo().StoreAndForward.Enabled {
		svc.dic.Update(di.ServiceConstructorMap{
			container.StoreClientName: func(get di.Get) interface{} {
				return nil
			},
		})
		return nil
	}

	logger := bootstrapContainer.LoggingClientFrom(svc.dic.Get)
	secretProvider := bootstrapContainer.SecretProviderFrom(svc.dic.Get)

	var err error

	secrets, err := secretProvider.GetSecret(config.Database.Type)
	if err != nil {
		return fmt.Errorf("unable to get Database Credentials for Store and Forward: %s", err.Error())
	}

	credentials := bootstrapConfig.Credentials{
		Username: secrets[secret.UsernameKey],
		Password: secrets[secret.PasswordKey],
	}

	startup := environment.GetStartupInfo(svc.serviceKey)

	tryUntil := time.Now().Add(time.Duration(startup.Duration) * time.Second)

	var storeClient interfaces.StoreClient
	for time.Now().Before(tryUntil) {
		if storeClient, err = svc.createStoreClient(config.Database, credentials); err != nil {
			logger.Warnf("unable to initialize Database '%s' for Store and Forward: %s", config.Database.Type, err.Error())
			time.Sleep(time.Duration(startup.Interval) * time.Second)
			continue
		}
		break
	}

	if err != nil {
		return fmt.Errorf("initialize Database for Store and Forward failed: %s", err.Error())
	}

	svc.dic.Update(di.ServiceConstructorMap{
		container.StoreClientName: func(get di.Get) interface{} {
			return storeClient
		},
	})
	return nil
}
