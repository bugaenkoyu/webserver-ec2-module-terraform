package test

import (
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/aws"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
)

// TestInfrastructure verifies the running state of EC2 instances.
func TestInfrastructure(t *testing.T) {
	t.Parallel()

	// The path to Terraform configurations and specify using Terragrunt
	terraformOptions := &terraform.Options{
		TerraformDir:    "../terragrunt/infra-module",
		TerraformBinary: "terragrunt",
	}

	// Remove infrastructure after testing
	defer terraform.Destroy(t, terraformOptions)

	// Start testing by installing the infrastructure
	terraform.InitAndApply(t, terraformOptions)

	// Get output values from Terraform
	dbIP := terraform.Output(t, terraformOptions, "db_ip")
	httpIP := terraform.Output(t, terraformOptions, "http_ip")

	// Verify that the output values are not empty
	assert.NotNil(t, dbIP)
	assert.NotNil(t, httpIP)

	// Verify that the EC2 instances are running
	instanceStateDB := aws.GetInstanceState(t, dbIP, "eu-north-1")
	assert.Equal(t, "running", instanceStateDB.Name)

	instanceStateHTTP := aws.GetInstanceState(t, httpIP, "eu-north-1")
	assert.Equal(t, "running", instanceStateHTTP.Name)
}