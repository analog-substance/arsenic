package lead

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/analog-substance/arsenic/lib/util"
	nessus "github.com/reapertechlabs/go_nessus"
)

const (
	metadataFile string = "00-metadata.md"
)

var (
	reconLeadsDir     string = filepath.FromSlash("recon/leads")
	dataLeadFile      string = filepath.FromSlash(".hugo/data/leads.json")
	reportFindingsDir string = filepath.FromSlash("report/findings")
)

type HugoLeadData struct {
	Ignored []string `json:"ignored"`
	Copied  []string `json:"copied"`
}

func (d HugoLeadData) save() error {
	out, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		return err
	}

	err = os.WriteFile(dataLeadFile, out, 0644)
	if err != nil {
		return err
	}

	return nil
}

func readHugoLeadData() (HugoLeadData, error) {
	var data HugoLeadData

	byteValue, err := os.ReadFile(dataLeadFile)
	if err != nil {
		return data, err
	}

	err = json.Unmarshal(byteValue, &data)
	if err != nil {
		return data, err
	}
	return data, nil
}

func IgnoreLead(id string) error {
	md, err := readHugoLeadData()
	if err != nil {
		return err
	}

	if util.IndexOf(md.Ignored, id) < 0 {
		md.Ignored = append(md.Ignored, id)
		err = md.save()
	}

	return err
}

func UnignoreLead(id string) error {
	data, err := readHugoLeadData()
	if err != nil {
		return err
	}

	idx := util.IndexOf(data.Ignored, id)
	if idx >= 0 {
		data.Ignored = util.RemoveIndex(data.Ignored, idx)
		err = data.save()
	}

	return err
}

func CopyLead(id string) error {
	md, err := readHugoLeadData()
	if err != nil {
		return err
	}

	if util.IndexOf(md.Copied, id) >= 0 {
		return errors.New("lead already copied")
	}

	lead, err := LeadFromID(id)
	if err != nil {
		return err
	}

	//Create the finding folder
	findingDir := fmt.Sprintf("%v/%.1f %v %v", reportFindingsDir, lead.Cvss.Score, util.Sanitize(lead.Title), util.Sanitize(id))
	err = os.Mkdir(filepath.FromSlash(findingDir), 0755)
	if err != nil {
		return err
	}

	err = lead.copyFiles(findingDir)
	if err != nil {
		return err
	}

	md.Copied = append(md.Copied, id)
	err = md.save()
	return err
}

func UncopyLead(id string) error {
	// Note: This only clears the ID from the Copied array
	md, err := readHugoLeadData()
	if err != nil {
		return err
	}

	idx := util.IndexOf(md.Copied, id)
	if idx >= 0 {
		md.Copied = util.RemoveIndex(md.Copied, idx)
		err = md.save()
	}

	return err
}

type AffectedAsset struct {
	Name         string `json:"name"`
	Port         int    `xml:"port,attr" json:"port,omitempty"`
	SvcName      string `xml:"svc_name,attr" json:"svc_name,omitempty"`
	Protocol     string `xml:"protocol,attr" json:"protocol,omitempty"`
	PluginOutput string `json:"plugin_output"`

	AffectedHost *nessus.ReportHost `json:"-"`
}

func (a *AffectedAsset) ToAssetString() string {
	return fmt.Sprintf("%s:%d", a.Name, a.Port)
}

type NessusFinding struct {
	ReportItem     *nessus.ReportItem `json:"report_item,omitempty"`
	AffectedAssets []AffectedAsset    `json:"affected_assets,omitempty"`
}

func FromNessusFinding(finding *NessusFinding) *Lead {

	cweRefs := append([]string{}, finding.ReportItem.CWE...)

	pluginInt, err := strconv.Atoi(finding.ReportItem.PluginID)
	if err != nil {
		log.Fatalln(err)
	}
	ID := fmt.Sprintf("deadbeef-cafe-babe-f007-%012d", pluginInt)

	return &Lead{
		Dir:          filepath.Join(reconLeadsDir, util.Sanitize(ID)),
		Title:        finding.ReportItem.PluginName,
		Description:  finding.ReportItem.Synopsis,
		CweRefs:      cweRefs,
		ExternalUUID: ID,
		Cvss: CVSS{
			finding.ReportItem.RiskFactor,
			finding.ReportItem.Cvss3BaseScore,
			finding.ReportItem.Cvss3Vector,
		},
		Exploitation: ExploitationRatings{},
		ExternalData: struct {
			Nessus NessusFinding `json:"nessus"`
		}{
			*finding,
		},
	}
}

type CVSS struct {
	Severity string  `json:"severity"`
	Score    float64 `json:"score"`
	Vector   string  `json:"vector"`
}

type ExploitationRatings struct {
	Likelihood string `json:"likelihood"`
	Impact     string `json:"impact"`
	RiskRating string `json:"riskRating"`
}
type Lead struct {
	Dir          string              `json:"-"`
	Title        string              `json:"title"`
	Description  string              `json:"description"`
	CweRefs      []string            `json:"cweRefs"`
	ExternalUUID string              `json:"external_uuid"`
	Cvss         CVSS                `json:"cvss"`
	Exploitation ExploitationRatings `json:"exploitation"`
	ExternalData struct {
		Nessus NessusFinding `json:"nessus"`
	} `json:"externalData"`
}

func (l *Lead) Save() error {
	isNessus := l.ExternalData.Nessus.ReportItem.PluginID != ""

	var summary string
	var recommendations string
	stepsToReproduce := []string{""}
	references := []string{""}
	affectedAssets := []string{""}

	if isNessus {
		summary = l.ExternalData.Nessus.ReportItem.Description
		recommendations = l.ExternalData.Nessus.ReportItem.Solution

		for _, asset := range l.ExternalData.Nessus.AffectedAssets {
			assetStr := asset.ToAssetString()
			affectedAssets = append(affectedAssets, assetStr)
			stepsToReproduce = append(stepsToReproduce, fmt.Sprintf("\n**%s**", assetStr), asset.PluginOutput)
		}

		if len(l.ExternalData.Nessus.ReportItem.CVE) > 0 {
			for _, cve := range l.ExternalData.Nessus.ReportItem.CVE {
				references = append(references, fmt.Sprintf("https://nvd.nist.gov/vuln/detail/%s", cve))
			}
		}

		if l.ExternalData.Nessus.ReportItem.SeeAlso != "" {
			links := strings.Split(l.ExternalData.Nessus.ReportItem.SeeAlso, "\n")
			references = append(references, links...)
		}
	}

	os.MkdirAll(l.Dir, 0755)

	out, err := json.MarshalIndent(l, "", "  ")
	if err != nil {
		return err
	}

	//00-metadata.md  01-summary.md  02-affected_assets.md  03-recommendations.md  04-references.md  05-steps_to_reproduce.md
	err = os.WriteFile(filepath.Join(l.Dir, "00-metadata.md"), out, 0644)
	if err != nil {
		return err
	}

	err = os.WriteFile(filepath.Join(l.Dir, "01-summary.md"), []byte(summary), 0644)
	if err != nil {
		return err
	}

	err = os.WriteFile(filepath.Join(l.Dir, "02-affected_assets.md"), []byte(strings.Join(affectedAssets, "\n* ")), 0644)
	if err != nil {
		return err
	}

	err = os.WriteFile(filepath.Join(l.Dir, "03-recommendations.md"), []byte(recommendations), 0644)
	if err != nil {
		return err
	}

	err = os.WriteFile(filepath.Join(l.Dir, "04-references.md"), []byte(strings.Join(references, "\n* ")), 0644)
	if err != nil {
		return err
	}

	err = os.WriteFile(filepath.Join(l.Dir, "05-steps_to_reproduce.md"), []byte(strings.Join(stepsToReproduce, "\n")), 0644)
	if err != nil {
		return err
	}

	return nil
}

func (l *Lead) files() ([]string, error) {
	var filePaths []string

	files, err := os.ReadDir(l.Dir)
	if err != nil {
		return filePaths, err
	}

	for _, file := range files {
		filePaths = append(filePaths, file.Name())
	}

	return filePaths, nil
}

func (l *Lead) copyFiles(dest string) error {
	// Copy the files
	files, err := l.files()
	if err != nil {
		return err
	}

	for _, file := range files {
		// Open src
		src, err := os.Open(filepath.Join(l.Dir, file))
		if err != nil {
			return err
		}
		defer src.Close()

		// Create dest
		dest, err := os.Create(filepath.Join(dest, file))
		if err != nil {
			return err
		}
		defer dest.Close()

		_, err = io.Copy(dest, src)
		if err != nil {
			return err
		}
	}

	return nil
}

func LeadFromID(id string) (*Lead, error) {
	dir := filepath.Join(reconLeadsDir, util.Sanitize(id))
	metadataFile := filepath.Join(dir, metadataFile)
	bytes, err := os.ReadFile(metadataFile)
	if err != nil {
		return nil, err
	}

	var lead Lead
	err = json.Unmarshal(bytes, &lead)
	if err != nil {
		return nil, err
	}

	lead.ExternalUUID = id
	lead.Dir = dir

	return &lead, nil
}
