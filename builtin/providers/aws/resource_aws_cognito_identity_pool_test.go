package aws

import (
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/cognitoidentity"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSCognitoIdentityPool_basic(t *testing.T) {
	name := fmt.Sprintf("identity pool %s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))
	updatedName := fmt.Sprintf("identity pool updated %s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCognitoIdentityPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCognitoIdentityPoolConfig_basic(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoIdentityPoolExists("aws_cognito_identity_pool.main"),
					resource.TestCheckResourceAttr("aws_cognito_identity_pool.main", "identity_pool_name", name),
				),
			},
			{
				Config: testAccAWSCognitoIdentityPoolConfig_basic(updatedName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoIdentityPoolExists("aws_cognito_identity_pool.main"),
					resource.TestCheckResourceAttr("aws_cognito_identity_pool.main", "identity_pool_name", updatedName),
				),
			},
		},
	})
}

func TestAccAWSCognitoIdentityPool_supportedLoginProviders(t *testing.T) {
	name := fmt.Sprintf("identity pool %s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCognitoIdentityPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCognitoIdentityPoolConfig_basic(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoIdentityPoolExists("aws_cognito_identity_pool.main"),
					resource.TestCheckResourceAttr("aws_cognito_identity_pool.main", "identity_pool_name", name),
				),
			},
			{
				Config: testAccAWSCognitoIdentityPoolConfig_supportedLoginProviders(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoIdentityPoolExists("aws_cognito_identity_pool.main"),
					resource.TestCheckResourceAttr("aws_cognito_identity_pool.main", "identity_pool_name", name),
				),
			},
			{
				Config: testAccAWSCognitoIdentityPoolConfig_basic(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoIdentityPoolExists("aws_cognito_identity_pool.main"),
					resource.TestCheckResourceAttr("aws_cognito_identity_pool.main", "identity_pool_name", name),
				),
			},
		},
	})
}

func testAccCheckAWSCognitoIdentityPoolExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("No Cognito Identity Pool ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).cognitoconn

		_, err := conn.DescribeIdentityPool(&cognitoidentity.DescribeIdentityPoolInput{
			IdentityPoolId: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckAWSCognitoIdentityPoolDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).cognitoconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cognito_identity_pool" {
			continue
		}

		_, err := conn.DescribeIdentityPool(&cognitoidentity.DescribeIdentityPoolInput{
			IdentityPoolId: aws.String(rs.Primary.ID),
		})

		if err != nil {
			if wserr, ok := err.(awserr.Error); ok && wserr.Code() == "ResourceNotFoundException" {
				return nil
			}
			return err
		}
	}

	return nil
}

func testAccAWSCognitoIdentityPoolConfig_basic(name string) string {
	return fmt.Sprintf(`
resource "aws_cognito_identity_pool" "main" {
  identity_pool_name               = "%s"
  allow_unauthenticated_identities = false
}
`, name)
}

func testAccAWSCognitoIdentityPoolConfig_supportedLoginProviders(name string) string {
	return fmt.Sprintf(`
resource "aws_cognito_identity_pool" "main" {
  identity_pool_name               = "%s"
  allow_unauthenticated_identities = false

  supported_login_providers {
    "graph.facebook.com" = "7346241598935555"
  }
}
`, name)
}
