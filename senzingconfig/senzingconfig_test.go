package senzingconfig

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/senzing/go-common/g2engineconfigurationjson"
	"github.com/senzing/go-logging/logging"
	"github.com/senzing/go-observing/observer"
	"github.com/senzing/init-database/senzingschema"
)

// ----------------------------------------------------------------------------
// Test harness
// ----------------------------------------------------------------------------

func TestMain(m *testing.M) {
	err := setup()
	if err != nil {
		fmt.Print(err)
		os.Exit(1)
	}
	code := m.Run()
	err = teardown()
	if err != nil {
		fmt.Print(err)
	}
	os.Exit(code)
}

func setup() error {
	ctx := context.TODO()
	senzingEngineConfigurationJson, err := g2engineconfigurationjson.BuildSimpleSystemConfigurationJsonUsingEnvVars()
	if err != nil {
		fmt.Print(err)
	}
	senzingSchema := &senzingschema.SenzingSchemaImpl{
		SenzingEngineConfigurationJson: senzingEngineConfigurationJson,
	}
	err = senzingSchema.SetLogLevel(ctx, logging.LevelInfoName)
	if err != nil {
		fmt.Println(err)
	}
	err = senzingSchema.InitializeSenzing(ctx)
	if err != nil {
		fmt.Println(err)
	}
	return err
}

func teardown() error {
	var err error = nil
	return err
}

// ----------------------------------------------------------------------------
// Test interface functions
// ----------------------------------------------------------------------------

func TestSenzingConfigImpl_InitializeSenzing_withDatasources(test *testing.T) {
	ctx := context.TODO()
	senzingEngineConfigurationJson, err := g2engineconfigurationjson.BuildSimpleSystemConfigurationJsonUsingEnvVars()
	if err != nil {
		fmt.Print(err)
	}
	senzingConfig := &SenzingConfigImpl{
		SenzingEngineConfigurationJson: senzingEngineConfigurationJson,
		DataSources:                    []string{"CUSTOMERS", "REFERENCE", "WATCHLIST"},
	}
	err = senzingConfig.SetLogLevel(ctx, logging.LevelInfoName)
	if err != nil {
		fmt.Println(err)
	}
	err = senzingConfig.InitializeSenzing(ctx)
	if err != nil {
		fmt.Println(err)
	}
}

func TestSenzingConfigImpl_InitializeSenzing(test *testing.T) {
	ctx := context.TODO()
	senzingEngineConfigurationJson, err := g2engineconfigurationjson.BuildSimpleSystemConfigurationJsonUsingEnvVars()
	if err != nil {
		fmt.Print(err)
	}
	senzingConfig := &SenzingConfigImpl{
		SenzingEngineConfigurationJson: senzingEngineConfigurationJson,
	}
	err = senzingConfig.SetLogLevel(ctx, logging.LevelInfoName)
	if err != nil {
		fmt.Println(err)
	}
	err = senzingConfig.InitializeSenzing(ctx)
	if err != nil {
		fmt.Println(err)
	}
}

func TestSenzingConfigImpl_RegisterObserver(test *testing.T) {
	ctx := context.TODO()
	anObserver := &observer.ObserverNull{
		Id:       "Observer 1",
		IsSilent: true,
	}
	senzingEngineConfigurationJson, err := g2engineconfigurationjson.BuildSimpleSystemConfigurationJsonUsingEnvVars()
	if err != nil {
		fmt.Print(err)
	}
	senzingConfig := &SenzingConfigImpl{
		SenzingEngineConfigurationJson: senzingEngineConfigurationJson,
	}
	err = senzingConfig.RegisterObserver(ctx, anObserver)
	if err != nil {
		fmt.Print(err)
	}
}

func TestSenzingConfigImpl_SetLogLevel(test *testing.T) {
	ctx := context.TODO()
	senzingEngineConfigurationJson, err := g2engineconfigurationjson.BuildSimpleSystemConfigurationJsonUsingEnvVars()
	if err != nil {
		fmt.Print(err)
	}
	senzingConfig := &SenzingConfigImpl{
		SenzingEngineConfigurationJson: senzingEngineConfigurationJson,
	}
	err = senzingConfig.SetLogLevel(ctx, logging.LevelInfoName)
	if err != nil {
		fmt.Println(err)
	}
}

func TestSenzingConfigImpl_SetObserverOrigin(test *testing.T) {
	ctx := context.TODO()
	senzingEngineConfigurationJson, err := g2engineconfigurationjson.BuildSimpleSystemConfigurationJsonUsingEnvVars()
	if err != nil {
		fmt.Print(err)
	}
	senzingConfig := &SenzingConfigImpl{
		SenzingEngineConfigurationJson: senzingEngineConfigurationJson,
	}
	senzingConfig.SetObserverOrigin(ctx, "TestObserver")
}

func TestSenzingConfigImpl_UnregisterObserver(test *testing.T) {
	ctx := context.TODO()
	anObserver := &observer.ObserverNull{
		Id:       "Observer 1",
		IsSilent: true,
	}
	senzingEngineConfigurationJson, err := g2engineconfigurationjson.BuildSimpleSystemConfigurationJsonUsingEnvVars()
	if err != nil {
		fmt.Print(err)
	}
	senzingConfig := &SenzingConfigImpl{
		SenzingEngineConfigurationJson: senzingEngineConfigurationJson,
	}
	err = senzingConfig.RegisterObserver(ctx, anObserver)
	if err != nil {
		fmt.Print(err)
	}
	err = senzingConfig.UnregisterObserver(ctx, anObserver)
	if err != nil {
		fmt.Print(err)
	}
}
