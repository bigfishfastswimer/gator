package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/open-policy-agent/frameworks/constraint/pkg/instrumentation"
	cmdutils "github.com/open-policy-agent/gatekeeper/v3/cmd/gator/util"
	"github.com/open-policy-agent/gatekeeper/v3/pkg/gator/reader"
	"github.com/open-policy-agent/gatekeeper/v3/pkg/gator/test"
	"gopkg.in/yaml.v3"
	"strings"
)

const (
	flagNameFilename = "filename"
	flagNameOutput   = "output"
	flagNameImage    = "image"
	flagNameTempDir  = "tempdir"

	stringJSON          = "json"
	stringYAML          = "yaml"
	stringHumanFriendly = "default"

	fourSpaceTab = "    "
)

func GatorTest(flagFilenames, flagImages []string, flagTempDir string) error {
	unstrucs, err := reader.ReadSources(flagFilenames, flagImages, flagTempDir)
	if err != nil {
		cmdutils.ErrFatalf("reading: %v", err)
	}
	if len(unstrucs) == 0 {
		cmdutils.ErrFatalf("no input data identified")
	}

	responses, err := test.Test(unstrucs, test.Opts{IncludeTrace: false, GatherStats: true, UseK8sCEL: false})
	if err != nil {
		cmdutils.ErrFatalf("auditing objects: %v", err)
	}
	results := responses.Results()

	fmt.Print(formatOutput(stringYAML, results, responses.StatsEntries))

	//constraintObj := &unstructured.Unstructured{}
	//constraintObj.SetKind("kind")

	//c, err := gator.NewOPAClient(false, false)
	//if err != nil {
	//	return err
	//}
	//c.AddData()
}

func formatOutput(flagOutput string, results []*test.GatorResult, stats []*instrumentation.StatsEntry) string {
	switch strings.ToLower(flagOutput) {
	case stringJSON:
		var jsonB []byte
		var err error

		if stats != nil {
			statsAndResults := map[string]interface{}{"results": results, "stats": stats}
			jsonB, err = json.MarshalIndent(statsAndResults, "", fourSpaceTab)
			if err != nil {
				cmdutils.ErrFatalf("marshaling validation json results and stats: %v", err)
			}
		} else {
			jsonB, err = json.MarshalIndent(results, "", fourSpaceTab)
			if err != nil {
				cmdutils.ErrFatalf("marshaling validation json results: %v", err)
			}
		}

		return string(jsonB)
	case stringYAML:
		yamlResults := test.GetYamlFriendlyResults(results)
		var yamlb []byte

		if stats != nil {
			statsAndResults := map[string]interface{}{"results": yamlResults, "stats": stats}

			statsJSONB, err := json.Marshal(statsAndResults)
			if err != nil {
				cmdutils.ErrFatalf("pre-marshaling stats to json: %v", err)
			}

			statsAndResultsUnmarshalled := struct {
				Results []*test.YamlGatorResult
				Stats   []*instrumentation.StatsEntry
			}{}

			err = json.Unmarshal(statsJSONB, &statsAndResultsUnmarshalled)
			if err != nil {
				cmdutils.ErrFatalf("pre-unmarshaling stats from json: %v", err)
			}

			yamlb, err = yaml.Marshal(statsAndResultsUnmarshalled)
			if err != nil {
				cmdutils.ErrFatalf("marshaling validation yaml results and stats: %v", err)
			}
		} else {
			jsonb, err := json.Marshal(yamlResults)
			if err != nil {
				cmdutils.ErrFatalf("pre-marshaling results to json: %v", err)
			}

			unmarshalled := []*test.YamlGatorResult{}
			err = json.Unmarshal(jsonb, &unmarshalled)
			if err != nil {
				cmdutils.ErrFatalf("pre-unmarshaling results from json: %v", err)
			}

			yamlb, err = yaml.Marshal(unmarshalled)
			if err != nil {
				cmdutils.ErrFatalf("marshaling validation yaml results: %v", err)
			}
		}

		return string(yamlb)
	case stringHumanFriendly:
	default:
		var buf bytes.Buffer
		if len(results) > 0 {
			for _, result := range results {
				obj := fmt.Sprintf("%s/%s %s",
					result.ViolatingObject.GetAPIVersion(),
					result.ViolatingObject.GetKind(),
					result.ViolatingObject.GetName(),
				)
				if result.ViolatingObject.GetNamespace() != "" {
					obj = fmt.Sprintf("%s/%s %s/%s",
						result.ViolatingObject.GetAPIVersion(),
						result.ViolatingObject.GetKind(),
						result.ViolatingObject.GetNamespace(),
						result.ViolatingObject.GetName(),
					)
				}
				buf.WriteString(fmt.Sprintf("%s: [%q] Message: %q\n",
					obj,
					result.Constraint.GetName(),
					result.Msg,
				))

				if result.Trace != nil {
					buf.WriteString(fmt.Sprintf("Trace: %v", *result.Trace))
				}
			}
		}
		return buf.String()
	}

	return ""
}
