package aws

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/medialive"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsMediaLiveInputSecurityGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsMediaLiveInputSecurityGroupCreate,
		Read:   resourceAwsMediaLiveInputSecurityGroupRead,
		Update: resourceAwsMediaLiveInputSecurityGroupUpdate,
		Delete: resourceAwsMediaLiveInputSecurityGroupDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ipv4_whitelist": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validateIpv4CIDRNetworkAddress,
				},
			},
			"tags": tagsSchema(),
		},
	}
}

func resourceAwsMediaLiveInputSecurityGroupCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).medialiveconn

	input := &medialive.CreateInputSecurityGroupInput{
		WhitelistRules: expandIPv4Whitelist(d.Get("ipv4_whitelist").(*schema.Set).List()),
	}

	if v := d.Get("tags").(map[string]interface{}); len(v) > 0 {
		input.Tags = keyvaluetags.New(v).IgnoreAws().MedialiveTags()
	}

	resp, err := conn.CreateInputSecurityGroup(input)
	if err != nil {
		return fmt.Errorf("error creating MediaLive Input Security Group: %s", err)
	}

	d.SetId(aws.StringValue(resp.SecurityGroup.Id))

	return resourceAwsMediaLiveInputSecurityGroupRead(d, meta)
}

func resourceAwsMediaLiveInputSecurityGroupRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).medialiveconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	input := &medialive.DescribeInputSecurityGroupInput{
		InputSecurityGroupId: aws.String(d.Id()),
	}

	resp, err := conn.DescribeInputSecurityGroup(input)
	if err != nil {
		return fmt.Errorf("error describing MediaLive Input Security Group: %s", err)
	}

	d.Set("arn", resp.Arn)

	var whitelist []*string

	for _, cidr := range resp.WhitelistRules {
		whitelist = append(whitelist, cidr.Cidr)
	}

	d.Set("ipv4_whitelist", whitelist)

	if err := d.Set("tags", keyvaluetags.MedialiveKeyValueTags(resp.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}

func resourceAwsMediaLiveInputSecurityGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).medialiveconn

	if d.HasChange("ipv4_whitelist") {
		input := &medialive.UpdateInputSecurityGroupInput{
			InputSecurityGroupId: aws.String(d.Id()),
			WhitelistRules:       expandIPv4Whitelist(d.Get("ipv4_whitelist").(*schema.Set).List()),
		}

		_, err := conn.UpdateInputSecurityGroup(input)
		if err != nil {
			return fmt.Errorf("error updating MediaLive Input Security Group: %s", err)
		}
	}

	arn := d.Get("arn").(string)
	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.MedialiveUpdateTags(conn, arn, o, n); err != nil {
			return fmt.Errorf("error updating MediaPackage Channel (%s) tags: %s", arn, err)
		}
	}

	return resourceAwsMediaLiveInputSecurityGroupRead(d, meta)
}

func resourceAwsMediaLiveInputSecurityGroupDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).medialiveconn

	input := &medialive.DeleteInputSecurityGroupInput{
		InputSecurityGroupId: aws.String(d.Id()),
	}

	_, err := conn.DeleteInputSecurityGroup(input)
	if err != nil {
		if isAWSErr(err, medialive.ErrCodeNotFoundException, "") {
			return nil
		}
		return fmt.Errorf("error deleting MediaLive Input Security Group: %s", err)
	}

	dinput := &medialive.DescribeInputSecurityGroupInput{
		InputSecurityGroupId: aws.String(d.Id()),
	}
	err = resource.Retry(5*time.Minute, func() *resource.RetryError {
		resp, err := conn.DescribeInputSecurityGroup(dinput)
		if err != nil {
			if isAWSErr(err, medialive.ErrCodeNotFoundException, "") {
				return nil
			}
			return resource.NonRetryableError(err)
		}
		if aws.StringValue(resp.State) == "DELETED" {
			return nil
		}
		return resource.RetryableError(fmt.Errorf("MediaLive Input Security Group (%s) still exists", d.Id()))
	})
	if isResourceTimeoutError(err) {
		_, err = conn.DescribeInputSecurityGroup(dinput)
	}
	if err != nil {
		return fmt.Errorf("error waiting for MediaLive Input Security Group (%s) deletion: %s", d.Id(), err)
	}
	return nil
}

func expandIPv4Whitelist(cidrs []interface{}) []*medialive.InputWhitelistRuleCidr {
	var rules []*medialive.InputWhitelistRuleCidr
	for _, cidr := range cidrs {
		rules = append(rules, &medialive.InputWhitelistRuleCidr{Cidr: aws.String(cidr.(string))})
	}
	return rules
}
