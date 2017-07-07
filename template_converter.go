package roer

import (
	"fmt"

	"github.com/spinnaker/roer/spinnaker"
)

const generatedTemplateHeader = `# GENERATED BY roer
#
# The output generated by this tool should be used as a base for further
# modifications. It does not make assumptions as to what things can be made into
# variables, modules, partials or Jinja templtes. This is your responsibility as
# the owner of the template.
#
# Some recommendations to massage the initial output:
#
# * Rename the pipeline stage IDs, notification names and trigger names to be
#   more meaningful. Enumerated stage IDs is ultimately a detriment for
#   long-term maintainability.
# * Best intentions are made to order most things, but the list of stages 
#   themselves are not ordered: Rearrange the stages so that they're roughly 
#   chronological.
`

func convertPipelineToTemplate(pipelineConfig spinnaker.PipelineConfig) PipelineTemplate {
	t := PipelineTemplate{
		Schema: "1",
		ID:     "generatedTemplate",
		Metadata: PipelineTemplateMetadata{
			Name:        pipelineConfig.Name,
			Description: pipelineConfig.Description,
			Owner:       pipelineConfig.LastModifiedBy,
			Scopes:      []string{pipelineConfig.Application},
		},
		Protect: false,
		Configuration: PipelineTemplateConfig{
			ConcurrentExecutions: map[string]bool{
				"parallel":        pipelineConfig.Parallel,
				"limitConcurrent": pipelineConfig.LimitConcurrent,
			},
			Triggers:      convertTriggers(pipelineConfig.Triggers),
			Parameters:    pipelineConfig.Parameters,
			Notifications: convertNotifications(pipelineConfig.Notifications),
		},
		Variables: make([]interface{}, 0),
		Stages:    convertStages(pipelineConfig.Stages),
	}
	return t
}

func convertTriggers(triggers []map[string]interface{}) (l []map[string]interface{}) {
	for i, t := range triggers {
		t["name"] = fmt.Sprintf("unnamed%d", i)
		l = append(l, t)
	}
	return
}

func convertNotifications(notifications []map[string]interface{}) (l []map[string]interface{}) {
	for i, n := range notifications {
		n["name"] = fmt.Sprintf("%s%d", n["type"], i)
		l = append(l, n)
	}
	return
}

func convertStages(stages []map[string]interface{}) []PipelineTemplateStage {
	convertToStringSlice := func(input interface{}) []string {
		if input == nil {
			return []string{}
		}
		s := make([]string, len(input.([]interface{})))
		for i, v := range input.([]interface{}) {
			s[i] = v.(string)
		}
		return s
	}

	l := []PipelineTemplateStage{}
	for _, s := range stages {
		stage := PipelineTemplateStage{
			ID:        s["type"].(string) + s["refId"].(string),
			Type:      s["type"].(string),
			DependsOn: buildDependsOn(stages, convertToStringSlice(s["requisiteStageRefIds"])),
			Name:      s["name"].(string),
			Config:    getStageConfig(s),
		}
		l = append(l, stage)
	}
	return l
}

func buildDependsOn(stages []map[string]interface{}, reqStageRefIDs []string) []string {
	l := []string{}
	for _, refID := range reqStageRefIDs {
		for _, s := range stages {
			if targetRefID, ok := s["refId"]; ok {
				if targetRefID.(string) == refID {
					l = append(l, s["type"].(string)+targetRefID.(string))
				}
			}
		}
	}
	return l
}

func getStageConfig(s map[string]interface{}) map[string]interface{} {
	config := map[string]interface{}{}
	for k, v := range s {
		config[k] = v
	}
	delete(config, "type")
	delete(config, "name")
	delete(config, "refId")
	delete(config, "requisiteStageRefIds")
	return config
}
