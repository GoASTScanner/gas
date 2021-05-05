package output

import (
	"fmt"
	"github.com/google/uuid"
	"runtime/debug"
	"strconv"
	"strings"

	"github.com/securego/gosec/v2"
	"github.com/securego/gosec/v2/cwe"
	"github.com/securego/gosec/v2/sarif"
)

type sarifLevel string

const (
	sarifNone    = sarifLevel("none")
	sarifNote    = sarifLevel("note")
	sarifWarning = sarifLevel("warning")
	sarifError   = sarifLevel("error")
	cweAcronym   = "CWE"
)

func convertToSarifReport(rootPaths []string, data *reportInfo) (*sarif.StaticAnalysisResultsFormatSARIFVersion210JSONSchema, error) {

	type rule struct {
		index int
		rule  *sarif.ReportingDescriptor
	}

	rules := make([]*sarif.ReportingDescriptor, 0)
	rulesIndices := make(map[string]rule)
	lastRuleIndex := -1

	results := []*sarif.Result{}
	taxa := make([]*sarif.ReportingDescriptor, 0)
	weaknesses := make(map[string]cwe.Weakness)

	for _, issue := range data.Issues {
		_, ok := weaknesses[issue.Cwe.ID]
		if !ok {
			weakness := cwe.Get(issue.Cwe.ID)
			weaknesses[issue.Cwe.ID] = weakness
			taxon := parseSarifTaxon(weakness, issue.Cwe.URL)
			taxa = append(taxa, taxon)
		}

		r, ok := rulesIndices[issue.RuleID]
		if !ok {
			lastRuleIndex++
			r = rule{index: lastRuleIndex, rule: parseSarifRule(issue, weaknesses[issue.Cwe.ID])}
			rulesIndices[issue.RuleID] = r
			rules = append(rules, r.rule)
		}

		location, err := parseSarifLocation(issue, rootPaths)
		if err != nil {
			return nil, err
		}

		result := &sarif.Result{
			RuleId:    r.rule.Id,
			RuleIndex: r.index,
			Level:     getSarifLevel(issue.Severity.String()),
			Message: &sarif.Message{
				Text: issue.What,
			},
			Locations: []*sarif.Location{location},
		}

		results = append(results, result)
	}

	tool := buildSarifTool(buildSarifDriver(rules))

	run := buildSarifRun(results, buildSarifTaxonomies(taxa), tool)

	return buildSarifReport(run), nil
}

// buildSarifReport return SARIF report struct
func buildSarifReport(run *sarif.Run) *sarif.StaticAnalysisResultsFormatSARIFVersion210JSONSchema {
	return &sarif.StaticAnalysisResultsFormatSARIFVersion210JSONSchema{
		Version: "2.1.0",
		Schema:  "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/master/Schemata/sarif-schema-2.1.0.json",
		Runs:    []*sarif.Run{run},
	}
}

// parseSarifRule return SARIF rule field struct
func parseSarifRule(issue *gosec.Issue, weakness cwe.Weakness) *sarif.ReportingDescriptor {
	return &sarif.ReportingDescriptor{
		Id:   issue.RuleID,
		Name: issue.What,
		ShortDescription: &sarif.MultiformatMessageString{
			Text: issue.What,
		},
		FullDescription: &sarif.MultiformatMessageString{
			Text: issue.What,
		},
		Help: &sarif.MultiformatMessageString{
			Text: fmt.Sprintf("%s\nSeverity: %s\nConfidence: %s\n", issue.What, issue.Severity.String(), issue.Confidence.String()),
		},
		Properties: &sarif.PropertyBag{
			Tags: []string{fmt.Sprintf("CWE-%s", issue.Cwe.ID), issue.Severity.String()},
		},
		DefaultConfiguration: &sarif.ReportingConfiguration{
			Level: getSarifLevel(issue.Severity.String()),
		},
		Relationships: []*sarif.ReportingDescriptorRelationship{
			buildSarifReportingDescriptorRelationship(weakness),
		},
	}
}

func buildSarifReportingDescriptorRelationship(weakness cwe.Weakness) *sarif.ReportingDescriptorRelationship {
	return &sarif.ReportingDescriptorRelationship{
		Target: &sarif.ReportingDescriptorReference{
			Id:   weakness.ID,
			Guid: uuid3(weakness.Name),
			ToolComponent: &sarif.ToolComponentReference{
				Name: cweAcronym,
			},
		},
		Kinds: []string{"superset"},
	}
}

func buildSarifTool(driver *sarif.ToolComponent) *sarif.Tool {
	return &sarif.Tool{
		Driver: driver,
	}
}

func buildSarifTaxonomies(taxa []*sarif.ReportingDescriptor) []*sarif.ToolComponent {
	version := "4.4"
	return []*sarif.ToolComponent{
		{
			Name:           cweAcronym,
			Version:        version,
			ReleaseDateUtc: "2021-03-15",
			InformationUri: fmt.Sprintf("https://cwe.mitre.org/data/published/cwe_v%s.pdf/", version),
			DownloadUri:    fmt.Sprintf("https://cwe.mitre.org/data/xml/cwec_v%s.xml.zip", version),
			Organization:   "MITRE",
			ShortDescription: &sarif.MultiformatMessageString{
				Text: "The MITRE Common Weakness Enumeration",
			},
			Guid:            uuid3(cweAcronym),
			IsComprehensive: true,
			MinimumRequiredLocalizedDataSemanticVersion: version,
			Taxa: taxa,
		},
	}
}

func parseSarifTaxon(weakness cwe.Weakness, url string) *sarif.ReportingDescriptor {
	return buildSarifTaxon(weakness.ID, weakness.Name, url, weakness.Description)
}

func buildSarifTaxon(id string, name string, uri string, description string) *sarif.ReportingDescriptor {
	return &sarif.ReportingDescriptor{
		Id:      id,
		Name:    name,
		Guid:    uuid3(name),
		HelpUri: uri,
		ShortDescription: &sarif.MultiformatMessageString{
			Text: description,
		},
	}
}

func buildSarifDriver(rules []*sarif.ReportingDescriptor) *sarif.ToolComponent {
	buildInfo, ok := debug.ReadBuildInfo()
	var gosecVersion string
	if ok {
		gosecVersion = buildInfo.Main.Version[1:]
	} else {
		gosecVersion = "devel"
	}
	return &sarif.ToolComponent{
		Name:    "gosec",
		Version: gosecVersion,
		SupportedTaxonomies: []*sarif.ToolComponentReference{
			{Name: cweAcronym, Index: 1, Guid: uuid3(cweAcronym)},
		},
		InformationUri: "https://github.com/securego/gosec/",
		Rules:          rules,
	}
}

func uuid3(value string) string {
	return uuid.NewMD5(uuid.Nil, []byte(value)).String()
}

func buildSarifRun(results []*sarif.Result, taxonomies []*sarif.ToolComponent, tool *sarif.Tool) *sarif.Run {
	return &sarif.Run{
		Results:    results,
		Taxonomies: taxonomies,
		Tool:       tool,
	}
}

// parseSarifLocation return SARIF location struct
func parseSarifLocation(issue *gosec.Issue, rootPaths []string) (*sarif.Location, error) {
	var filePath string

	lines := strings.Split(issue.Line, "-")
	startLine, err := strconv.Atoi(lines[0])
	if err != nil {
		return nil, err
	}
	endLine := startLine
	if len(lines) > 1 {
		endLine, err = strconv.Atoi(lines[1])
		if err != nil {
			return nil, err
		}
	}

	col, err := strconv.Atoi(issue.Col)
	if err != nil {
		return nil, err
	}

	for _, rootPath := range rootPaths {
		if strings.HasPrefix(issue.File, rootPath) {
			filePath = strings.Replace(issue.File, rootPath+"/", "", 1)
		}
	}

	return buildSarifLocation(buildSarifPhysicalLocation(parseSarifArtifactLocation(filePath), parseSarifRegion(startLine, endLine, col))), nil

}

func buildSarifLocation(physicalLocation *sarif.PhysicalLocation) *sarif.Location {
	return &sarif.Location{
		PhysicalLocation: physicalLocation,
	}
}

func buildSarifPhysicalLocation(artifactLocation *sarif.ArtifactLocation, region *sarif.Region) *sarif.PhysicalLocation {
	return &sarif.PhysicalLocation{
		ArtifactLocation: artifactLocation,
		Region:           region,
	}
}

func parseSarifArtifactLocation(filePath string) *sarif.ArtifactLocation {
	return &sarif.ArtifactLocation{
		Uri: filePath,
	}
}

func parseSarifRegion(startLine int, endLine int, col int) *sarif.Region {
	return &sarif.Region{
		StartLine:   startLine,
		EndLine:     endLine,
		StartColumn: col,
		EndColumn:   col,
	}
}

// From https://docs.oasis-open.org/sarif/sarif/v2.0/csprd02/sarif-v2.0-csprd02.html#_Toc10127839
// * "warning": The rule specified by ruleId was evaluated and a problem was found.
// * "error": The rule specified by ruleId was evaluated and a serious problem was found.
// * "note": The rule specified by ruleId was evaluated and a minor problem or an opportunity to improve the code was found.
// * "none": The concept of “severity” does not apply to this result because the kind property (§3.27.9) has a value other than "fail".
func getSarifLevel(s string) sarifLevel {
	switch s {
	case "LOW":
		return sarifWarning
	case "MEDIUM":
		return sarifError
	case "HIGH":
		return sarifError
	default:
		return sarifNote
	}
}
