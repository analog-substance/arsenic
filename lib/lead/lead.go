package lead

import (
    "encoding/json"
    "errors"
    "fmt"
    "github.com/analog-substance/arsenic/lib/util"
    "io"
    "io/ioutil"
    "os"
    "path/filepath"
)

const data_lead_file string = ".hugo/data/leads.json"
const recon_leads_dir string = "recon/leads/"
const report_findings_dir string = "report/findings/"
const metadata_file string = "00-metadata.md"

type HugoLeadMetadata struct {
    Ignored  []string `json:"ignored"`
    Copied  []string `json:"copied"`
}

type Cvss struct {
    Severity  string `json:"severity"`
    Score     float32 `json:"score"`
    Vector    string `json:"vector"`
}

type LeadMetadata struct {
    Title  string `json:"title"`
    Cvss   Cvss   `json:"cvss"`
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
        return errors.New("Lead already copied")
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

