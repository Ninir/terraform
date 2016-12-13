package aws

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/aws/aws-sdk-go/service/sns"
)

// Built from http://docs.aws.amazon.com/sns/latest/api/API_Operations.html
var snsTopicPermissions := map[string]string {
	"AddPermission",
	"CheckIfPhoneNumberIsOptedOut",
	"ConfirmSubscription",
	"CreatePlatformApplication",
	"CreatePlatformEndpoint",
	"CreateTopic",
	"DeleteEndpoint",
	"DeletePlatformApplication",
	"DeleteTopic",
	"GetEndpointAttributes",
	"GetPlatformApplicationAttributes",
	"GetSMSAttributes",
	"GetSubscriptionAttributes",
	"GetTopicAttributes",
	"ListEndpointsByPlatformApplication",
	"ListPhoneNumbersOptedOut",
	"ListPlatformApplications",
	"ListSubscriptions",
	"ListSubscriptionsByTopic",
	"ListTopics",
	"OptInPhoneNumber",
	"Publish",
	"RemovePermission",
	"SetEndpointAttributes",
	"SetPlatformApplicationAttributes",
	"SetSMSAttributes",
	"SetSubscriptionAttributes",
	"SetTopicAttributes",
	"Subscribe",
	"Unsubscribe",
}

func resourceAwSnsPermission() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsSnsPermissionCreate,
		Read:   resourceAwsSnsPermissionRead,
		Delete: resourceAwsSnsPermissionDelete,
		Schema: map[string]*schema.Schema{
			"action_names": {
				Type:         schema.TypeList,
				Required:     true,
				ForceNew:     true,
			},
			"accounts": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validateAwsAccountId,
			},
			"label": {
				Type:         schema.TypeList,
				Optional:     true,
				ForceNew:     true,
			},
			"topic_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validateArn,
			},
		},
	}
}

func resourceAwsSnsPermissionCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).snsconn

	label := d.Get("label").(string)

	// There is a bug in the API (reported and acknowledged by AWS)
	// which causes some permissions to be ignored when API calls are sent in parallel
	// We work around this bug via mutex
	awsMutexKV.Lock(label)
	defer awsMutexKV.Unlock(label)

	input := sns.AddPermissionInput{
		ActionName:   aws.String(d.Get("action").(string)),
		AWSAccountId: aws.String(d.Get("action").(string)),
		Label:        aws.String(d.Get("label").(string)),
		TopicArn:     aws.String(d.Get("topic_arn").(string)),
	}

	log.Printf("[DEBUG] Adding new SNS permission: %s", input)
	var out *sns.AddPermissionOutput
	err := resource.Retry(1*time.Minute, func() *resource.RetryError {
		var err error
		out, err = conn.AddPermission(&input)

		if err != nil {
			if awsErr, ok := err.(awserr.Error); ok {
				// IAM is eventually consistent :/
				if awsErr.Code() == "ResourceConflictException" {
					return resource.RetryableError(
						fmt.Errorf("[WARN] Error adding new SNS topic permission for %s, retrying: %s",
							*input.TopicArn, err))
				}
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})

	if err != nil {
		return err
	}

	d.SetId(d.Get("label").(string))

	return err
}

func resourceAwsSnsPermissionRead(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceAwsSnsPermissionDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).snsconn

	label := d.Get("label").(string)

	// There is a bug in the API (reported and acknowledged by AWS)
	// which causes some permissions to be ignored when API calls are sent in parallel
	// We work around this bug via mutex
	awsMutexKV.Lock(label)
	defer awsMutexKV.Unlock(label)

	input := sns.RemovePermissionInput{
		Label: aws.String(label),
		TopicArn: aws.String(label),
	}

	log.Printf("[DEBUG] Removing SNS topic permission: %s", input)
	_, err := conn.RemovePermission(&input)
	if err != nil {
		return err
	}

	err = resource.Retry(5*time.Minute, func() *resource.RetryError {
		log.Printf("[DEBUG] Checking if SNS topic permission %q is deleted", d.Id())

		log.Printf("[DEBUG] No error when checking if SNS topic permission %s is deleted", d.Id())
		return nil
	})

	if err != nil {
		return fmt.Errorf("Failed removing SNS permission: %s", err)
	}

	log.Printf("[DEBUG] SNS topic permission with ID %q removed", d.Id())
	d.SetId("")

	return nil
}
