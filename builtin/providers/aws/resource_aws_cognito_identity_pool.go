package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/cognitoidentity"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsCognitoIdentityPool() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsCognitoIdentityPoolCreate,
		Read:   resourceAwsCognitoIdentityPoolRead,
		Delete: resourceAwsCognitoIdentityPoolDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"identity_pool_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
			},

			"cognito_identity_providers": {
				Type:         schema.TypeSet,
				Optional:     true,
			},

			"developer_provider_name": {
				Type:         schema.TypeString,
				Optional:     true,
			},

			"allow_unauthenticated_identities": {
				Type: schema.TypeBool,
				Optional: true,
			},

			"openid_connect_provider_arns": {
				Type: schema.TypeList,
				Optional: true,
			},

			"saml_provider_arns": {
				Type: schema.TypeList,
				Optional: true,
			},

			"supported_login_providers": {
				Type: schema.TypeMap,
				Optional: true,
			},
		},
	}
}

func resourceAwsCognitoIdentityPoolCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cognitoconn
	log.Print("[DEBUG] Creating Cognito Identity Pool")

	params := &cognitoidentity.CreateIdentityPoolInput{
		IdentityPoolName: aws.String(d.Get("identity_pool_name").(string)),
	}

	activity, err := conn.CreateIdentityPool(params)
	if err != nil {
		return fmt.Errorf("Error creating Cognito Identity Pool: %s", err)
	}

	d.SetId(*activity.IdentityPoolId)

	return resourceAwsCognitoIdentityPoolRead(d, meta)
}

func resourceAwsCognitoIdentityPoolRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cognitoconn
	log.Printf("[DEBUG] Reading Cognito Identity Pool: %s", d.Id())

	ip, err := conn.DescribeIdentityPool(&cognitoidentity.DescribeIdentityPoolInput{
		IdentityPoolId: aws.String(d.Id()),
	})
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == "ActivityDoesNotExist" {
			d.SetId("")
			return nil
		}
		return err
	}

	d.Set("identity_pool_name", ip.IdentityPoolName)
	d.Set("allow_unauthenticated_identities", ip.AllowUnauthenticatedIdentities)
	d.Set("cognito_identity_providers", ip.CognitoIdentityProviders)
	d.Set("developer_provider_name", ip.DeveloperProviderName)
	d.Set("openid_connect_provider_arns", ip.OpenIdConnectProviderARNs)
	d.Set("saml_provider_arns", ip.SamlProviderARNs)
	d.Set("supported_login_providers", ip.SupportedLoginProviders)

	return nil
}

func resourceAwsCognitoIdentityPoolDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cognitoconn
	log.Printf("[DEBUG] Deleting Cognito Identity Pool: %s", d.Id())

	return resource.Retry(5*time.Minute, func() *resource.RetryError {
		_, err := conn.DeleteIdentityPool(&cognitoidentity.DeleteIdentityPoolInput{
			IdentityPoolId: aws.String(d.Id()),
		})

		if err == nil {
			return nil
		}

		return resource.NonRetryableError(err)
	})
}
