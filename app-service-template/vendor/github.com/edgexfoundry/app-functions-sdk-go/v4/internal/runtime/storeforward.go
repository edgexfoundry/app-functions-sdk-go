//
// Copyright (c) 2022 Intel Corporation
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
//

package runtime

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/edgexfoundry/app-functions-sdk-go/v4/internal"
	"github.com/edgexfoundry/app-functions-sdk-go/v4/internal/appfunction"
	"github.com/edgexfoundry/app-functions-sdk-go/v4/internal/bootstrap/container"
	"github.com/edgexfoundry/app-functions-sdk-go/v4/pkg/interfaces"
	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
	gometrics "github.com/rcrowley/go-metrics"
)

const (
	defaultMinRetryInterval = 1 * time.Second
)

type storeForwardInfo struct {
	runtime         *FunctionsPipelineRuntime
	dic             *di.Container
	lc              logger.LoggingClient
	serviceKey      string
	dataCount       gometrics.Counter
	retryInProgress atomic.Bool
}

func newStoreAndForward(runtime *FunctionsPipelineRuntime, dic *di.Container, serviceKey string) *storeForwardInfo {
	sf := &storeForwardInfo{
		runtime:    runtime,
		dic:        dic,
		lc:         bootstrapContainer.LoggingClientFrom(dic.Get),
		serviceKey: serviceKey,
		dataCount:  gometrics.NewCounter(),
	}

	metricsManager := bootstrapContainer.MetricsManagerFrom(dic.Get)
	if metricsManager == nil {
		sf.lc.Errorf("Unable to register %s metric: MetricsManager is not available.", internal.StoreForwardQueueSizeName)
		return sf
	}

	err := metricsManager.Register(internal.StoreForwardQueueSizeName, sf.dataCount, nil)
	if err != nil {
		sf.lc.Errorf("Unable to register metric %s. Collection will continue, but metric will not be reported: %v",
			internal.StoreForwardQueueSizeName, err)
	}

	sf.lc.Infof("%s metric has been registered and will be reported (if enabled)", internal.StoreForwardQueueSizeName)

	return sf
}

func (sf *storeForwardInfo) startStoreAndForwardRetryLoop(
	appWg *sync.WaitGroup,
	appCtx context.Context,
	enabledWg *sync.WaitGroup,
	enabledCtx context.Context,
	serviceKey string) {

	appWg.Add(1)
	enabledWg.Add(1)

	config := container.ConfigurationFrom(sf.dic.Get)
	storeClient := container.StoreClientFrom(sf.dic.Get)

	sf.serviceKey = serviceKey

	items, err := storeClient.RetrieveFromStore(serviceKey)
	if err != nil {
		sf.lc.Errorf("Unable to initialize Store and Forward data count: Failed to load items from DB: %v", err)
	} else {
		sf.dataCount.Clear()
		sf.dataCount.Inc(int64(len(items)))
	}

	go func() {
		defer appWg.Done()
		defer enabledWg.Done()

		retryInterval, err := time.ParseDuration(config.GetWritableInfo().StoreAndForward.RetryInterval)
		if err != nil {
			sf.lc.Warn(
				fmt.Sprintf("StoreAndForward RetryInterval failed to parse, defaulting to %s",
					defaultMinRetryInterval.String()))
			retryInterval = defaultMinRetryInterval
		} else if retryInterval < defaultMinRetryInterval {
			sf.lc.Warn(
				fmt.Sprintf("StoreAndForward RetryInterval value %s is less than the allowed minimum value, defaulting to %s",
					retryInterval.String(), defaultMinRetryInterval.String()))
			retryInterval = defaultMinRetryInterval
		}

		if config.GetWritableInfo().StoreAndForward.MaxRetryCount < 0 {
			sf.lc.Warn("StoreAndForward MaxRetryCount can not be less than 0, defaulting to 1")
			if err := config.SetWritableInfo("StoreAndForward.MaxRetryCount", 1); err != nil {
				sf.lc.Errorf("Failed to set StoreAndForward MaxRetryCount to default 1: %v", err)
			}
		}

		sf.lc.Info(
			fmt.Sprintf("Starting StoreAndForward Retry Loop with %s RetryInterval and %d max retries. %d stored items waiting for retry.",
				retryInterval.String(),
				config.GetWritableInfo().StoreAndForward.MaxRetryCount,
				len(items)))

	exit:
		for {
			select {

			case <-appCtx.Done():
				// Exit the loop and function when application service is terminating.
				break exit

			case <-enabledCtx.Done():
				// Exit the loop and function when Store and Forward has been disabled.
				break exit

			case <-time.After(retryInterval):
				sf.retryStoredData(serviceKey)
			}
		}

		sf.lc.Info("Exiting StoreAndForward Retry Loop")
	}()
}

func (sf *storeForwardInfo) storeForLaterRetry(
	payload []byte,
	appContext interfaces.AppFunctionContext,
	pipeline *interfaces.FunctionPipeline,
	pipelinePosition int) {

	item := interfaces.NewStoredObject(sf.runtime.ServiceKey, payload, pipeline.Id, pipelinePosition, pipeline.Hash, appContext.GetAllValues())
	item.CorrelationID = appContext.CorrelationID()

	sf.lc.Tracef("Storing data for later retry for pipeline '%s' (%s=%s)",
		pipeline.Id,
		common.CorrelationHeader,
		appContext.CorrelationID())

	config := container.ConfigurationFrom(sf.dic.Get)
	if !config.GetWritableInfo().StoreAndForward.Enabled {
		sf.lc.Errorf("Failed to store item for later retry for pipeline '%s': StoreAndForward not enabled", pipeline.Id)
		return
	}

	storeClient := container.StoreClientFrom(sf.dic.Get)

	if _, err := storeClient.Store(item); err != nil {
		sf.lc.Errorf("Failed to store item for later retry for pipeline '%s': %s", pipeline.Id, err.Error())
	}

	sf.dataCount.Inc(1)
}

func (sf *storeForwardInfo) retryStoredData(serviceKey string) {
	// Skip if another thread is already doing the retry
	if sf.retryInProgress.Load() {
		return
	}

	sf.retryInProgress.Store(true)
	defer sf.retryInProgress.Store(false)

	storeClient := container.StoreClientFrom(sf.dic.Get)

	items, err := storeClient.RetrieveFromStore(serviceKey)
	if err != nil {
		sf.lc.Errorf("Unable to load store and forward items from DB: %s", err.Error())
		return
	}

	sf.lc.Debugf("%d stored data items found for retrying", len(items))

	if len(items) > 0 {
		itemsToRemove, itemsToUpdate := sf.processRetryItems(items)

		sf.lc.Debugf(" %d stored data items will be removed post retry", len(itemsToRemove))
		sf.lc.Debugf(" %d stored data items will be updated post retry", len(itemsToUpdate))

		for _, item := range itemsToRemove {
			if err := storeClient.RemoveFromStore(item); err != nil {
				sf.lc.Errorf("Unable to remove stored data item for pipeline '%s' from DB, objectID=%s: %s",
					item.PipelineId,
					err.Error(),
					item.ID)
			}
		}

		for _, item := range itemsToUpdate {
			if err := storeClient.Update(item); err != nil {
				sf.lc.Errorf("Unable to update stored data item for pipeline '%s' from DB, objectID=%s: %s",
					item.PipelineId,
					err.Error(),
					item.ID)
			}
		}

		sf.dataCount.Dec(int64(len(itemsToRemove)))
	}
}

func (sf *storeForwardInfo) processRetryItems(items []interfaces.StoredObject) ([]interfaces.StoredObject, []interfaces.StoredObject) {
	config := container.ConfigurationFrom(sf.dic.Get)

	var itemsToRemove []interfaces.StoredObject
	var itemsToUpdate []interfaces.StoredObject

	// Item will be removed from store if:
	//    - successfully retried
	//    - max retries exceeded
	//    - version no longer matches current Pipeline
	// Item will not be removed if retry failed and more retries available (hit 'continue' above)
	for _, item := range items {
		pipeline := sf.runtime.GetPipelineById(item.PipelineId)

		if pipeline == nil {
			sf.lc.Errorf("Stored data item's pipeline '%s' no longer exists. Removing item from DB", item.PipelineId)
			itemsToRemove = append(itemsToRemove, item)
			continue
		}

		if item.Version != pipeline.Hash {
			sf.lc.Error("Stored data item's pipeline Version doesn't match '%s' pipeline's Version. Removing item from DB", item.PipelineId)
			itemsToRemove = append(itemsToRemove, item)
			continue
		}

		if !sf.retryExportFunction(item, pipeline) {
			item.RetryCount++
			if config.GetWritableInfo().StoreAndForward.MaxRetryCount == 0 ||
				item.RetryCount < config.GetWritableInfo().StoreAndForward.MaxRetryCount {
				sf.lc.Tracef("Export retry failed for pipeline '%s'. retries=%d, Incrementing retry count (%s=%s)",
					item.PipelineId,
					item.RetryCount,
					common.CorrelationHeader,
					item.CorrelationID)
				itemsToUpdate = append(itemsToUpdate, item)
				continue
			}

			sf.lc.Tracef("Max retries exceeded for pipeline '%s'. retries=%d, Removing item from DB (%s=%s)",
				item.PipelineId,
				item.RetryCount,
				common.CorrelationHeader,
				item.CorrelationID)
			itemsToRemove = append(itemsToRemove, item)

			// Note that item will be removed for DB below.
		} else {
			sf.lc.Tracef("Retry successful for pipeline '%s'. Removing item from DB (%s=%s)",
				item.PipelineId,
				common.CorrelationHeader,
				item.CorrelationID)
			itemsToRemove = append(itemsToRemove, item)
		}
	}

	return itemsToRemove, itemsToUpdate
}

func (sf *storeForwardInfo) retryExportFunction(item interfaces.StoredObject, pipeline *interfaces.FunctionPipeline) bool {
	appContext := appfunction.NewContext(item.CorrelationID, sf.dic, "")

	for k, v := range item.ContextData {
		appContext.AddValue(strings.ToLower(k), v)
	}

	sf.lc.Tracef("Retrying stored data for pipeline '%s' (%s=%s)",
		item.PipelineId,
		common.CorrelationHeader,
		appContext.CorrelationID())

	return sf.runtime.ExecutePipeline(
		item.Payload,
		appContext,
		pipeline,
		item.PipelinePosition,
		true) == nil
}

func (sf *storeForwardInfo) triggerRetry() {
	if sf.dataCount.Count() > 0 {
		config := container.ConfigurationFrom(sf.dic.Get)
		if !config.GetWritableInfo().StoreAndForward.Enabled {
			sf.lc.Debug("Store and Forward not enabled, skipping triggering retry of failed data")
			return
		}

		sf.lc.Debug("Triggering Store and Forward retry of failed data")
		sf.retryStoredData(sf.serviceKey)
	}
}
