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

// TestEC2InstanceRunningState verifies the running state of EC2 instances.
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

	// Перевірити, чи виведені значення не є пустими
	assert.NotNil(t, dbIP)
	assert.NotNil(t, httpIP)

	// Перевірити, чи стан інстансів EC2 - "running"
	instanceStateDB := aws.GetInstanceState(t, dbIP, "eu-north-1")
	assert.Equal(t, "running", instanceStateDB.Name)

	instanceStateHTTP := aws.GetInstanceState(t, httpIP, "eu-north-1")
	assert.Equal(t, "running", instanceStateHTTP.Name)
}

// Test that http_ip and db_ip are in the VPC and in correct subnets CIDR blocks.
func TestInstanceCIDRBlocks(t *testing.T) {
	terraformOptions := &terraform.Options{
		TerraformDir:    "../terragrunt/infra-module",
		TerraformBinary: "terragrunt",
	}

	vpcCIDR := "192.168.0.0/16"
	networkHTTPCIDR := "192.168.1.0/24"
	networkDBCIDR := "192.168.2.0/24"

	// Defer the destroy until after the test
	defer terraform.Destroy(t, terraformOptions)

	// Apply the Terraform module
	terraform.InitAndApply(t, terraformOptions)

	// Get the outputs
	httpIP := terraform.Output(t, terraformOptions, "http_ip")
	dbIP := terraform.Output(t, terraformOptions, "db_ip")

	// Verify that http_ip is within the VPC CIDR block
	assert.True(t, aws.IsCIDRBlockInCidrBlock(t, httpIP, vpcCIDR))

	// Verify that db_ip is within the VPC CIDR block
	assert.True(t, aws.IsCIDRBlockInCidrBlock(t, dbIP, vpcCIDR))

	// Verify that http_ip is within the networkHTTPCIDR block and not within the networkDBCIDR block
	assert.True(t, aws.IsCIDRBlockInCidrBlock(t, httpIP, networkHTTPCIDR))
	assert.False(t, aws.IsCIDRBlockInCidrBlock(t, httpIP, networkDBCIDR))

	// Verify that db_ip is within the networkDBCIDR block and not within the networkHTTPCIDR block
	assert.True(t, aws.IsCIDRBlockInCidrBlock(t, dbIP, networkDBCIDR))
	assert.False(t, aws.IsCIDRBlockInCidrBlock(t, dbIP, networkHTTPCIDR))
}

// Test that Database Not Accessible From Internet
func TestDatabaseNotAccessibleFromInternet(t *testing.T) {
	terraformOptions := &terraform.Options{
		TerraformDir:    "../terragrunt/infra-module",
		TerraformBinary: "terragrunt",
	}

	// Defer the destroy until after the test
	defer terraform.Destroy(t, terraformOptions)

	// Apply the Terraform module
	terraform.InitAndApply(t, terraformOptions)

	// Get the outputs
	dbIP := terraform.Output(t, terraformOptions, "db_ip")

	// Define a port that you expect your database to be listening on
	dbPort := "5432"

	// Define the timeout for the connection attempt
	timeout := 2 * time.Second

	// Define the expected error message when the connection attempt fails
	expectedErrorMessage := "dial tcp: i/o timeout"

	// Attempt to connect to the database
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%s", dbIP, dbPort), timeout)

	// Check if there was an error connecting
	if err != nil {
		// If the error message matches the expected error message, the database is not accessible from the internet
		if err.Error() == expectedErrorMessage {
			t.Logf("Database is not accessible from the internet as expected")
		} else {
			t.Errorf("Unexpected error: %s", err)
		}
	} else {
		// If the connection was successful, the database is accessible from the internet
		t.Errorf("Unexpected successful connection to the database from the internet")
	}

	// Close the connection if it was opened
	if conn != nil {
		conn.Close()
	}
}
