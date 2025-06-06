// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT

package xray

import (
	"fmt"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap"
	"go.opentelemetry.io/collector/pipeline"
	"go.opentelemetry.io/collector/processor/batchprocessor"

	"github.com/aws/amazon-cloudwatch-agent/translator/translate/otel/common"
	awsxrayexporter "github.com/aws/amazon-cloudwatch-agent/translator/translate/otel/exporter/awsxray"
	"github.com/aws/amazon-cloudwatch-agent/translator/translate/otel/extension/agenthealth"
	"github.com/aws/amazon-cloudwatch-agent/translator/translate/otel/processor"
	awsxrayreceiver "github.com/aws/amazon-cloudwatch-agent/translator/translate/otel/receiver/awsxray"
	"github.com/aws/amazon-cloudwatch-agent/translator/translate/otel/receiver/otlp"
)

const (
	pipelineName = "xray"
)

var (
	xrayKey = common.ConfigKey(common.TracesKey, common.TracesCollectedKey, common.XrayKey)
	otlpKey = common.ConfigKey(common.TracesKey, common.TracesCollectedKey, common.OtlpKey)
)

type translator struct {
}

var _ common.PipelineTranslator = (*translator)(nil)

func NewTranslator() common.PipelineTranslator {
	return &translator{}
}

func (t *translator) ID() pipeline.ID {
	return pipeline.NewIDWithName(pipeline.SignalTraces, pipelineName)
}

func (t *translator) Translate(conf *confmap.Conf) (*common.ComponentTranslators, error) {
	if conf == nil || !(conf.IsSet(xrayKey) || conf.IsSet(otlpKey)) {
		return nil, &common.MissingKeyError{ID: t.ID(), JsonKey: fmt.Sprint(xrayKey, " or ", otlpKey)}
	}
	translators := &common.ComponentTranslators{
		Receivers:  common.NewTranslatorMap[component.Config, component.ID](),
		Processors: common.NewTranslatorMap(processor.NewDefaultTranslatorWithName(pipelineName, batchprocessor.NewFactory())),
		Exporters:  common.NewTranslatorMap(awsxrayexporter.NewTranslator()),
		Extensions: common.NewTranslatorMap(agenthealth.NewTranslator(agenthealth.TracesName, []string{agenthealth.OperationPutTraceSegments}),
			agenthealth.NewTranslatorWithStatusCode(agenthealth.StatusCodeName, nil, true)),
	}
	if conf.IsSet(xrayKey) {
		translators.Receivers.Set(awsxrayreceiver.NewTranslator())
	}
	if conf.IsSet(otlpKey) {
		translators.Receivers.Set(otlp.NewTranslator(
			otlp.WithSignal(pipeline.SignalTraces),
			otlp.WithConfigKey(otlpKey)),
		)
	}
	return translators, nil
}
