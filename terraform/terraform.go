package terraform

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"
	"text/template"

	"github.com/google/uuid"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/hc-install/product"
	"github.com/hashicorp/hc-install/releases"
	"github.com/hashicorp/terraform-exec/tfexec"
	tfjson "github.com/hashicorp/terraform-json"
	"github.com/port-labs/tf-connector/port"
	"github.com/samber/lo"
	"go.uber.org/zap"
)

type Terraform struct {
	exe    string
	logger *zap.SugaredLogger
}

func NewTerraform(logger *zap.SugaredLogger) *Terraform {
	return &Terraform{
		logger: logger,
	}
}

func genUUID(len int) string {
	id := uuid.New()
	return strings.Replace(id.String(), "-", "", -1)[:len]
}

func (t *Terraform) Install(ctx context.Context) error {
	installer := &releases.ExactVersion{
		Product: product.Terraform,
		Version: version.Must(version.NewVersion("1.2.9")),
	}

	execPath, err := installer.Install(ctx)
	if err != nil {
		return err
	}
	t.exe = execPath
	return nil
}

// getStateKey return the key of the state file in the backend (e.g. the object name in the bucket)
// to keep the same state file between actions on the same entity, we use the entity id as the key
// or generate a random key if the entity does not exist yet.
func getStateKey(actionBody *port.ActionBody) string {
	if actionBody.Context.Entity != "" {
		return actionBody.Context.Entity
	}
	return "e_" + genUUID(16)
}

// Apply loads main terraform file with backend configured using go template, loads template file for the wanted resource,
// writes all terraform files to temp working dir and applies using terraform
func (t *Terraform) Apply(actionBody *port.ActionBody, ctx context.Context) error {
	templateFolder := ctx.Value("templateFolder").(string)
	stateKey := getStateKey(actionBody)
	mainTF, err := t.loadMainTF(actionBody, stateKey)
	if err != nil {
		return err
	}
	tmplTF, err := t.loadTemplateTF(actionBody, templateFolder)
	if err != nil {
		return err
	}
	workDir, err := os.MkdirTemp(os.TempDir(), "tf-*")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(workDir)
	t.logger.Infof("Created temp work dir: %s", workDir)
	tf, err := tfexec.NewTerraform(workDir, t.exe)
	if err != nil {
		return fmt.Errorf("failed to create terraform instance: %v", err)
	}
	if err = t.writeTerraformFiles(workDir, map[string][]byte{
		"main.tf":     mainTF,
		"template.tf": tmplTF,
	}); err != nil {
		return err
	}
	t.logger.Info("Running terraform init")
	err = tf.Init(context.Background())
	if err != nil {
		return fmt.Errorf("failed to run terraform init: %v", err)
	}
	props := lo.Assign(map[string]any{}, actionBody.Payload.Entity.Properties, actionBody.Payload.Properties, map[string]any{
		"entity_identifier": stateKey,
		"blueprint":         actionBody.Context.Blueprint,
		"run_id":            actionBody.Context.RunID,
	})
	varsFilepath, err := t.createVarsFile(workDir, props)
	if err != nil {
		return fmt.Errorf("failed to create vars file: %v", err)
	}
	err = tf.Apply(context.Background(), tfexec.VarFile(varsFilepath))
	if err != nil {
		return fmt.Errorf("failed to apply terraform: %v", err)
	}
	return nil
}

func (t *Terraform) createVarsFile(workDir string, properties map[string]any) (string, error) {
	b, err := json.Marshal(properties)
	if err != nil {
		return "", err
	}
	varsFile := path.Join(workDir, "port.tfvars.json")
	if err := os.WriteFile(varsFile, b, 0600); err != nil {
		return "", err
	}
	return varsFile, nil
}

func (t *Terraform) writeTerraformFiles(folder string, files map[string][]byte) error {
	for name, content := range files {
		if err := os.WriteFile(path.Join(folder, name), content, 0600); err != nil {
			return fmt.Errorf("failed to write %s: %v", name, err)
		}
		t.logger.Debugf("Wrote %s: %s", name, string(content))
	}
	return nil
}

func (t *Terraform) loadMainTF(actionBody *port.ActionBody, storageKey string) ([]byte, error) {
	t.logger.Info("Loading default terraform file 'main.tf'")
	tmpl, err := template.ParseFiles("main.tf")
	if err != nil {
		return nil, fmt.Errorf("failed to parse main.tf: %v", err)
	}
	buf := bytes.Buffer{}
	err = tmpl.Execute(&buf, map[string]string{"storage_key": storageKey})
	if err != nil {
		return nil, fmt.Errorf("failed to execute main.tf template: %v", err)
	}
	return buf.Bytes(), nil
}

func (t *Terraform) loadTemplateTF(actionBody *port.ActionBody, templateFolder string) ([]byte, error) {
	templateName := actionBody.Context.Blueprint + ".tf"
	t.logger.Infof("Using template %s", templateName)
	templateFile, err := os.ReadFile(path.Join(templateFolder, templateName))
	if err != nil {
		return nil, fmt.Errorf("failed to read template file %s: %v", templateName, err)
	}
	return templateFile, nil
}

func (t *Terraform) Destroy(actionBody *port.ActionBody, ctx context.Context) error {
	stateKey := actionBody.Context.Entity
	if stateKey == "" {
		return fmt.Errorf("entity id is empty, cannot destroy")
	}
	mainTF, err := t.loadMainTF(actionBody, stateKey)
	if err != nil {
		return err
	}
	workDir, err := os.MkdirTemp(os.TempDir(), "tf-*")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(workDir)
	t.logger.Infof("Created temp work dir: %s", workDir)
	tf, err := tfexec.NewTerraform(workDir, t.exe)
	if err != nil {
		return fmt.Errorf("failed to create terraform instance: %v", err)
	}
	if err = t.writeTerraformFiles(workDir, map[string][]byte{
		"main.tf": mainTF,
	}); err != nil {
		return err
	}
	t.logger.Info("Running terraform init")
	err = tf.Init(context.Background())
	if err != nil {
		return fmt.Errorf("failed to run terraform init: %v", err)
	}
	props := lo.Assign(map[string]any{}, actionBody.Payload.Properties, map[string]any{
		"entity_identifier": stateKey,
		"blueprint":         actionBody.Context.Blueprint,
		"run_id":            actionBody.Context.RunID,
	})
	varsFilepath, err := t.createVarsFile(workDir, props)
	if err != nil {
		return fmt.Errorf("failed to create vars file: %v", err)
	}
	err = tf.Apply(context.Background(), tfexec.VarFile(varsFilepath))
	if err != nil {
		return fmt.Errorf("failed to apply terraform: %v", err)
	}
	return nil
}

// ExtractEntityID extracts the entity ID from the terraform state file
func (t *Terraform) ExtractEntityID(tf *tfexec.Terraform, ctx context.Context) (string, error) {
	state, err := tf.Show(context.Background())
	if err != nil {
		return "", err
	}
	entities := lo.Filter(state.Values.RootModule.Resources, func(r *tfjson.StateResource, _ int) bool {
		return r.Type == "port-labs_entity"
	})
	if len(entities) == 1 {
		return entities[0].AttributeValues["id"].(string), nil
	}
	return state.FormatVersion, nil
}
