package lead

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/analog-substance/arsenic/lib/util"
	"github.com/reapertechlabs/go_nessus"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

const data_lead_file string = ".hugo/data/leads.json"
const recon_leads_dir string = "recon/leads/"
const report_findings_dir string = "report/findings/"
const metadata_file string = "00-metadata.md"

type HugoLeadMetadata struct {
	Ignored []string `json:"ignored"`
	Copied  []string `json:"copied"`
}

type Cvss struct {
	Severity string  `json:"severity"`
	Score    float32 `json:"score"`
	Vector   string  `json:"vector"`
}

type LeadMetadata struct {
	Title string `json:"title"`
	Cvss  Cvss   `json:"cvss"`
}

func ReadHugoLeadMetadata() (HugoLeadMetadata, error) {
	var metadata HugoLeadMetadata
	jsonFile, err := os.Open(filepath.FromSlash(data_lead_file))
	if err != nil {
		return metadata, err
	}
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)
	err = json.Unmarshal(byteValue, &metadata)
	if err != nil {
		return metadata, err
	}
	return metadata, nil
}

func (md HugoLeadMetadata) save() error {
	out, err := json.MarshalIndent(md, "", "  ")
	if err != nil {
		return err
	}

	fp := filepath.FromSlash(data_lead_file)
	err = ioutil.WriteFile(fp, out, 0644)
	if err != nil {
		return err
	}

	return nil
}

func ReadLeadMetadata(id string) (LeadMetadata, error) {
	var metadata LeadMetadata
	lead_metadata_file := recon_leads_dir + "/" + util.Sanitize(id) + "/" + metadata_file
	jsonFile, err := os.Open(filepath.FromSlash(lead_metadata_file))
	if err != nil {
		return metadata, err
	}
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)
	err = json.Unmarshal(byteValue, &metadata)
	if err != nil {
		return metadata, err
	}
	return metadata, nil
}

func GetLeadDir(id string) string {
	return recon_leads_dir + "/" + id
}

func GetLeadFiles(id string) ([]string, error) {
	filePaths := []string{}
	lead_dir := GetLeadDir(id)
	files, err := ioutil.ReadDir(filepath.FromSlash(lead_dir))
	if err != nil {
		return filePaths, err
	}

	for _, file := range files {
		filePaths = append(filePaths, file.Name())
	}

	return filePaths, nil
}

func IgnoreLead(id string) error {
	fmt.Println("IgnoreLead")
	md, err := ReadHugoLeadMetadata()
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
	fmt.Println("UnignoreLead")
	md, err := ReadHugoLeadMetadata()
	if err != nil {
		return err
	}

	idx := util.IndexOf(md.Ignored, id)
	if idx >= 0 {
		md.Ignored = util.RemoveIndex(md.Ignored, idx)
		err = md.save()
	}

	return err
}

func CopyLead(id string) error {
	fmt.Println("CopyLead")
	md, err := ReadHugoLeadMetadata()
	if err != nil {
		return err
	}

	if util.IndexOf(md.Copied, id) >= 0 {
		return errors.New("lead already copied")
	}

	//Create the finding folder
	lead_md, err := ReadLeadMetadata(id)
	if err != nil {
		return err
	}

	lead_dir := GetLeadDir(id)
	finding_dir := fmt.Sprintf("%v/%.1f %v %v", report_findings_dir, lead_md.Cvss.Score, util.Sanitize(lead_md.Title), util.Sanitize(id))
	err = os.Mkdir(filepath.FromSlash(finding_dir), 0750)
	if err != nil {
		return err
	}

	//Copy the files
	files, err := GetLeadFiles(id)
	if err != nil {
		return err
	}

	for _, file := range files {
		// Open src
		src, err := os.Open(filepath.FromSlash(lead_dir + "/" + file))
		if err != nil {
			return err
		}
		defer src.Close()

		// Create dest
		dest, err := os.Create(filepath.FromSlash(finding_dir + "/" + file))
		if err != nil {
			return err
		}
		defer dest.Close()

		_, err = io.Copy(dest, src)
		if err != nil {
			return err
		}
	}

	md.Copied = append(md.Copied, id)
	err = md.save()
	return err
}

func UncopyLead(id string) error {
	// Note: This only clears the ID from the Copied array
	fmt.Println("UncopyLead")
	md, err := ReadHugoLeadMetadata()
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
	return &Lead{
		Title:        finding.ReportItem.PluginName,
		Description:  finding.ReportItem.Synopsis,
		CweRefs:      finding.ReportItem.CWE,
		ExternalUUID: "",
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

	out, err := json.MarshalIndent(l, "", "  ")
	if err != nil {
		return err
	}

	var ID string
	var summary string
	var recommendations string
	stepsToReproduce := []string{""}
	references := []string{""}
	affectedAssets := []string{""}

	if isNessus {
		pluginInt, err := strconv.Atoi(l.ExternalData.Nessus.ReportItem.PluginID)
		if err != nil {
			log.Fatalln(err)
		}
		ID = fmt.Sprintf("deadbeef-cafe-babe-f007-%012d", pluginInt)
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

	leadDir := path.Join("recon", "leads", ID)
	os.MkdirAll(leadDir, 0755)

	//00-metadata.md  01-summary.md  02-affected_assets.md  03-recommendations.md  04-references.md  05-steps_to_reproduce.md
	err = os.WriteFile(path.Join(leadDir, "00-metadata.md"), out, 0644)
	if err != nil {
		return err
	}

	err = os.WriteFile(path.Join(leadDir, "01-summary.md"), []byte(summary), 0644)
	if err != nil {
		return err
	}

	err = os.WriteFile(path.Join(leadDir, "02-affected_assets.md"), []byte(strings.Join(affectedAssets, "\n* ")), 0644)
	if err != nil {
		return err
	}

	err = os.WriteFile(path.Join(leadDir, "03-recommendations.md"), []byte(recommendations), 0644)
	if err != nil {
		return err
	}

	err = os.WriteFile(path.Join(leadDir, "04-references.md"), []byte(strings.Join(references, "\n* ")), 0644)
	if err != nil {
		return err
	}

	err = os.WriteFile(path.Join(leadDir, "05-steps_to_reproduce.md"), []byte(strings.Join(stepsToReproduce, "\n")), 0644)
	if err != nil {
		return err
	}

	return nil
}
