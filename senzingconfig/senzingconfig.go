package senzingconfig

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/senzing/g2-sdk-go/g2api"
	"github.com/senzing/go-logging/logging"
	"github.com/senzing/go-observing/notifier"
	"github.com/senzing/go-observing/observer"
	"github.com/senzing/go-observing/subject"
	"github.com/senzing/go-sdk-abstract-factory/factory"
)

// ----------------------------------------------------------------------------
// Types
// ----------------------------------------------------------------------------

// SenzingConfigImpl is the default implementation of the SenzingConfig interface.
type SenzingConfigImpl struct {
	DataSources                    []string
	g2configmgrSingleton           g2api.G2configmgr
	g2configmgrSyncOnce            sync.Once
	g2configSingleton              g2api.G2config
	g2configSyncOnce               sync.Once
	g2factorySingleton             factory.SdkAbstractFactory
	g2factorySyncOnce              sync.Once
	logger                         logging.LoggingInterface
	logLevel                       string
	observers                      subject.Subject
	SenzingEngineConfigurationJson string
	SenzingModuleName              string
	SenzingVerboseLogging          int
}

// ----------------------------------------------------------------------------
// Variables
// ----------------------------------------------------------------------------

var debugOptions []interface{} = []interface{}{
	&logging.OptionCallerSkip{Value: 5},
}

var traceOptions []interface{} = []interface{}{
	&logging.OptionCallerSkip{Value: 5},
}

var defaultModuleName string = "init-database"

// ----------------------------------------------------------------------------
// Internal methods
// ----------------------------------------------------------------------------

// --- Logging ----------------------------------------------------------------

// Get the Logger singleton.
func (senzingConfig *SenzingConfigImpl) getLogger() logging.LoggingInterface {
	var err error = nil
	if senzingConfig.logger == nil {
		options := []interface{}{
			&logging.OptionCallerSkip{Value: 4},
		}
		senzingConfig.logger, err = logging.NewSenzingToolsLogger(ProductId, IdMessages, options...)
		if err != nil {
			panic(err)
		}
	}
	return senzingConfig.logger
}

// Log message.
func (senzingConfig *SenzingConfigImpl) log(messageNumber int, details ...interface{}) {
	senzingConfig.getLogger().Log(messageNumber, details...)
}

// Debug.
func (senzingConfig *SenzingConfigImpl) debug(messageNumber int, details ...interface{}) {
	details = append(details, debugOptions...)
	senzingConfig.getLogger().Log(messageNumber, details...)
}

// Trace method entry.
func (senzingConfig *SenzingConfigImpl) traceEntry(messageNumber int, details ...interface{}) {
	details = append(details, traceOptions...)
	senzingConfig.getLogger().Log(messageNumber, details...)
}

// Trace method exit.
func (senzingConfig *SenzingConfigImpl) traceExit(messageNumber int, details ...interface{}) {
	details = append(details, traceOptions...)
	senzingConfig.getLogger().Log(messageNumber, details...)
}

// --- Dependent services -----------------------------------------------------

// Create an abstract factory singleton and return it.
func (senzingConfig *SenzingConfigImpl) getG2Factory(ctx context.Context) factory.SdkAbstractFactory {
	senzingConfig.g2factorySyncOnce.Do(func() {
		senzingConfig.g2factorySingleton = &factory.SdkAbstractFactoryImpl{}
	})
	return senzingConfig.g2factorySingleton
}

// Create a G2Config singleton and return it.
func (senzingConfig *SenzingConfigImpl) getG2config(ctx context.Context) (g2api.G2config, error) {
	var err error = nil
	senzingConfig.g2configSyncOnce.Do(func() {
		senzingConfig.g2configSingleton, err = senzingConfig.getG2Factory(ctx).GetG2config(ctx)
		if err != nil {
			return
		}
		if senzingConfig.g2configSingleton.GetSdkId(ctx) == "base" {
			moduleName := senzingConfig.SenzingModuleName
			if len(moduleName) == 0 {
				moduleName = defaultModuleName
			}
			err = senzingConfig.g2configSingleton.Init(ctx, moduleName, senzingConfig.SenzingEngineConfigurationJson, senzingConfig.SenzingVerboseLogging)
		}
	})
	return senzingConfig.g2configSingleton, err
}

// Create a G2Configmgr singleton and return it.
func (senzingConfig *SenzingConfigImpl) getG2configmgr(ctx context.Context) (g2api.G2configmgr, error) {
	var err error = nil
	senzingConfig.g2configmgrSyncOnce.Do(func() {
		senzingConfig.g2configmgrSingleton, err = senzingConfig.getG2Factory(ctx).GetG2configmgr(ctx)
		if err != nil {
			return
		}
		if senzingConfig.g2configmgrSingleton.GetSdkId(ctx) == "base" {
			moduleName := senzingConfig.SenzingModuleName
			if len(moduleName) == 0 {
				moduleName = defaultModuleName
			}
			err = senzingConfig.g2configmgrSingleton.Init(ctx, moduleName, senzingConfig.SenzingEngineConfigurationJson, senzingConfig.SenzingVerboseLogging)
		}
	})
	return senzingConfig.g2configmgrSingleton, err
}

// Get dependent services: G2config, G2configmgr
func (senzingConfig *SenzingConfigImpl) getDependentServices(ctx context.Context) (g2api.G2config, g2api.G2configmgr, error) {
	g2Config, err := senzingConfig.getG2config(ctx)
	if err != nil {
		return nil, nil, err
	}
	g2Configmgr, err := senzingConfig.getG2configmgr(ctx)
	if err != nil {
		return g2Config, nil, err
	}
	return g2Config, g2Configmgr, err
}

// --- Misc -------------------------------------------------------------------

// Add datasources to Senzing configuration.
func (senzingConfig *SenzingConfigImpl) addDatasources(ctx context.Context, g2Config g2api.G2config, configHandle uintptr) error {
	var err error = nil
	for _, datasource := range senzingConfig.DataSources {
		inputJson := `{"DSRC_CODE": "` + datasource + `"}`
		_, err = g2Config.AddDataSource(ctx, configHandle, inputJson)
		if err != nil {
			return err
		}
		senzingConfig.log(2001, datasource)
	}
	return err
}

// ----------------------------------------------------------------------------
// Interface methods
// ----------------------------------------------------------------------------

/*
The InitializeSenzing method adds the Senzing default configuration to databases.

Input
  - ctx: A context to control lifecycle.
*/
func (senzingConfig *SenzingConfigImpl) InitializeSenzing(ctx context.Context) error {
	var err error = nil
	var configID int64 = 0

	// Prolog.

	traceExitMessageNumber := 29
	entryTime := time.Now()
	if senzingConfig.isTrace {
		senzingConfig.traceEntry(10)
		defer func() { senzingConfig.traceExit(traceExitMessageNumber, err, configID, time.Since(entryTime)) }()
	}

	// Log entry parameters.

	if senzingConfig.getLogger().IsDebug() {
		asJson, err := json.Marshal(senzingConfig)
		if err != nil {
			traceExitMessageNumber = 11
			return err
		}
		senzingConfig.log(1001, senzingConfig, string(asJson))
	}

	// Create Senzing objects.

	g2Config, g2Configmgr, err := senzingConfig.getDependentServices(ctx)
	if err != nil {
		traceExitMessageNumber = 12
		return err
	}

	// Determine if configuration already exists. If so, return.

	configID, err = g2Configmgr.GetDefaultConfigID(ctx)
	if err != nil {
		traceExitMessageNumber = 13
		return err
	}
	if configID != 0 {
		if senzingConfig.observers != nil {
			go func() {
				details := map[string]string{}
				notifier.Notify(ctx, senzingConfig.observers, ProductId, 8001, err, details)
			}()
		}
		senzingConfig.log(2002, configID)
		traceExitMessageNumber = 14
		return err
	}

	// Create a fresh Senzing configuration.

	configHandle, err := g2Config.Create(ctx)
	if err != nil {
		traceExitMessageNumber = 15
		return err
	}

	// If requested, add DataSources to fresh Senzing configuration.

	if len(senzingConfig.DataSources) > 0 {
		err = senzingConfig.addDatasources(ctx, g2Config, configHandle)
		if err != nil {
			traceExitMessageNumber = 16
			return err
		}
	}

	// Create a JSON string from the in-memory configuration.

	configStr, err := g2Config.Save(ctx, configHandle)
	if err != nil {
		traceExitMessageNumber = 17
		return err
	}

	// Persist the Senzing configuration to the Senzing repository and set as default configuration.

	configComments := fmt.Sprintf("Created by init-database at %s", entryTime.Format(time.RFC3339Nano))
	configID, err = g2Configmgr.AddConfig(ctx, configStr, configComments)
	if err != nil {
		traceExitMessageNumber = 18
		return err
	}
	err = g2Configmgr.SetDefaultConfigID(ctx, configID)
	if err != nil {
		traceExitMessageNumber = 19
		return err
	}

	// Notify observers.

	senzingConfig.log(2003, configID, configComments)
	if senzingConfig.observers != nil {
		go func() {
			details := map[string]string{}
			notifier.Notify(ctx, senzingConfig.observers, ProductId, 8002, err, details)
		}()
	}

	return err
}

/*
The RegisterObserver method adds the observer to the list of observers notified.

Input
  - ctx: A context to control lifecycle.
  - observer: The observer to be added.
*/
func (senzingConfig *SenzingConfigImpl) RegisterObserver(ctx context.Context, observer observer.Observer) error {
	var err error = nil

	// Prolog.

	traceExitMessageNumber := 39
	if senzingConfig.isTrace {
		entryTime := time.Now()
		senzingConfig.traceEntry(30, observer.GetObserverId(ctx))
		defer func() {
			senzingConfig.traceExit(traceExitMessageNumber, observer.GetObserverId(ctx), err, time.Since(entryTime))
		}()
	}

	// Log entry parameters.

	if senzingConfig.getLogger().IsDebug() {
		asJson, err := json.Marshal(senzingConfig)
		if err != nil {
			traceExitMessageNumber = 31
			return err
		}
		senzingConfig.log(1002, senzingConfig, string(asJson))
	}

	// Create empty list of observers.

	if senzingConfig.observers == nil {
		senzingConfig.observers = &subject.SubjectImpl{}
	}

	// Register observer with senzingConfig and dependent services.

	err = senzingConfig.observers.RegisterObserver(ctx, observer)
	if err != nil {
		traceExitMessageNumber = 32
		return err
	}

	// FIXME: Need issue to fix registering observers with g2-sdk-go-*
	// g2Config, g2Configmgr, err := senzingConfig.getDependentServices(ctx)
	// if err != nil {
	// 	traceExitMessageNumber = 33
	// 	return err
	// }
	// err = g2Config.RegisterObserver(ctx, observer)
	// if err != nil {
	// 	traceExitMessageNumber = 34
	// 	return err
	// }
	// err = g2Configmgr.RegisterObserver(ctx, observer)
	// if err != nil {
	// 	traceExitMessageNumber = 35
	// 	return err
	// }

	// Notify observers.

	go func() {
		details := map[string]string{
			"observerID": observer.GetObserverId(ctx),
		}
		notifier.Notify(ctx, senzingConfig.observers, ProductId, 8003, err, details)
	}()

	return err
}

/*
The SetLogLevel method sets the level of logging.

Input
  - ctx: A context to control lifecycle.
  - logLevel: The desired log level. TRACE, DEBUG, INFO, WARN, ERROR, FATAL or PANIC.
*/
func (senzingConfig *SenzingConfigImpl) SetLogLevel(ctx context.Context, logLevelName string) error {
	var err error = nil

	// Prolog.

	traceExitMessageNumber := 49
	if senzingConfig.isTrace {
		entryTime := time.Now()
		senzingConfig.traceEntry(40, logLevelName)
		defer func() { senzingConfig.traceExit(traceExitMessageNumber, logLevelName, err, time.Since(entryTime)) }()
	}

	// Log entry parameters.

	if senzingConfig.getLogger().IsDebug() {
		asJson, err := json.Marshal(senzingConfig)
		if err != nil {
			traceExitMessageNumber = 41
			return err
		}
		senzingConfig.log(1003, senzingConfig, string(asJson))
	}

	// Verify value of logLevelName.

	if !logging.IsValidLogLevelName(logLevelName) {
		traceExitMessageNumber = 42
		return fmt.Errorf("invalid error level: %s", logLevelName)
	}

	// Set senzingConfig log level.

	senzingConfig.logLevel = logLevelName
	err = senzingConfig.getLogger().SetLogLevel(logLevelName)
	if err != nil {
		traceExitMessageNumber = 43
		return err
	}
	senzingConfig.isTrace = (logLevelName == logging.LevelTraceName)

	// Set log level for dependent services.

	// TODO: Remove once g2configmgr.SetLogLevel(context.Context, string)
	logLevel := logging.TextToLoggerLevelMap[logLevelName]

	g2Config, g2Configmgr, err := senzingConfig.getDependentServices(ctx)
	if err != nil {
		traceExitMessageNumber = 44
		return err
	}
	err = g2Config.SetLogLevel(ctx, logLevel)
	if err != nil {
		traceExitMessageNumber = 45
		return err
	}
	err = g2Configmgr.SetLogLevel(ctx, logLevel)
	if err != nil {
		traceExitMessageNumber = 46
		return err
	}

	// Notify observers.

	if senzingConfig.observers != nil {
		go func() {
			details := map[string]string{
				"logLevelName": logLevelName,
			}
			notifier.Notify(ctx, senzingConfig.observers, ProductId, 8004, err, details)
		}()
	}

	return err
}

/*
The UnregisterObserver method removes the observer to the list of observers notified.

Input
  - ctx: A context to control lifecycle.
  - observer: The observer to be removed.
*/
func (senzingConfig *SenzingConfigImpl) UnregisterObserver(ctx context.Context, observer observer.Observer) error {
	var err error = nil

	// Prolog.

	traceExitMessageNumber := 59
	if senzingConfig.isTrace {
		entryTime := time.Now()
		senzingConfig.traceEntry(50, observer.GetObserverId(ctx))
		defer func() {
			senzingConfig.traceExit(traceExitMessageNumber, observer.GetObserverId(ctx), err, time.Since(entryTime))
		}()
	}

	// Log entry parameters.

	if senzingConfig.getLogger().IsDebug() {
		asJson, err := json.Marshal(senzingConfig)
		if err != nil {
			traceExitMessageNumber = 51
			return err
		}
		senzingConfig.log(1004, senzingConfig, string(asJson))
	}

	// FIXME: Need issue to fix registering observers with g2-sdk-go-*
	// Unregister observers in dependencies.
	// g2Config, g2Configmgr, err := senzingConfig.getDependentServices(ctx)
	// err = g2Config.UnregisterObserver(ctx, observer)
	// if err != nil {
	// 	traceExitMessageNumber = 52
	// 	return err
	// }
	// err = g2Configmgr.UnregisterObserver(ctx, observer)
	// if err != nil {
	// 	traceExitMessageNumber = 53
	// 	return err
	// }

	// Remove observer from this service.

	if senzingConfig.observers != nil {

		// Tricky code:
		// client.notify is called synchronously before client.observers is set to nil.
		// In client.notify, each observer will get notified in a goroutine.
		// Then client.observers may be set to nil, but observer goroutines will be OK.
		details := map[string]string{
			"observerID": observer.GetObserverId(ctx),
		}
		notifier.Notify(ctx, senzingConfig.observers, ProductId, 8005, err, details)

		err = senzingConfig.observers.UnregisterObserver(ctx, observer)
		if err != nil {
			traceExitMessageNumber = 54
			return err
		}

		if !senzingConfig.observers.HasObservers(ctx) {
			senzingConfig.observers = nil
		}
	}

	return err
}
