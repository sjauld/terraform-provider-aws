package aws

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/medialive"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

// when everything is implemented we can remove this and just use medialive.InputType_Values()
var validAwsMediaLiveInputTypes = []string{
	// medialive.InputTypeInputDevice, // Elemental Link
	// medialive.InputTypeMediaconnect, // MediaConnect
	// medialive.InputTypeMp4File,     // MP4
	// medialive.InputTypeRtmpPull,    // RTMP (pull)
	medialive.InputTypeRtmpPush, // RTMP (push)
	medialive.InputTypeRtpPush,  // RTP
	medialive.InputTypeUdpPush,  // Magic option not listed in console
	// medialive.InputTypeUrlPull,     // HLS
}

func resourceAwsMediaLiveInput() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsMediaLiveInputCreate,
		Read:   resourceAwsMediaLiveInputRead,
		Update: resourceAwsMediaLiveInputUpdate,
		Delete: resourceAwsMediaLiveInputDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"destinations": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"endpoint": {
							Type:     schema.TypeString,
							Required: true,
						},
						"ip": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"port": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"url": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"vpc": {
							Type:     schema.TypeSet,
							Computed: true,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"availability_zone": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"network_interface_id": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
			"input_class": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"input_security_groups": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(validAwsMediaLiveInputTypes, false),
				ForceNew:     true,
			},
			"vpc": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"security_group_ids": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							MaxItems: 5,
						},
						"subnet_ids": {
							Type:     schema.TypeSet,
							Required: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							MinItems: 2,
							MaxItems: 2,
						},
					},
				},
			},
			"tags": tagsSchema(),
		},
	}
}

func resourceAwsMediaLiveInputCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).medialiveconn

	input := &medialive.CreateInputInput{
		Name: aws.String(d.Get("name").(string)),
		Type: aws.String(d.Get("type").(string)),
	}

	if v, ok := d.GetOk("destinations"); ok {
		input.Destinations = expandInputDestinationRequest(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk("input_security_groups"); ok {
		input.InputSecurityGroups = expandStringList(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk("vpc"); ok {
		input.Vpc = expandInputVpcRequest(v.(*schema.Set).List())
	}

	if v := d.Get("tags").(map[string]interface{}); len(v) > 0 {
		input.Tags = keyvaluetags.New(v).IgnoreAws().MedialiveTags()
	}

	resp, err := conn.CreateInput(input)
	if err != nil {
		return fmt.Errorf("error creating MediaLive Input: %s", err)
	}

	d.SetId(aws.StringValue(resp.Input.Id))

	return resourceAwsMediaLiveInputRead(d, meta)
}

func resourceAwsMediaLiveInputRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).medialiveconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	input := &medialive.DescribeInputInput{
		InputId: aws.String(d.Id()),
	}

	resp, err := conn.DescribeInput(input)
	if err != nil {
		return fmt.Errorf("error describing MediaLive Input: %s", err)
	}

	log.Printf("DescribeInput Response: %+v", resp)

	d.Set("arn", resp.Arn)
	if len(resp.Destinations) > 0 {
		d.Set("destinations", flattenInputDestinations(resp.Destinations))
	}

	d.Set("input_class", aws.StringValue(resp.InputClass))
	d.Set("input_source_type", aws.StringValue(resp.InputSourceType))
	d.Set("name", aws.StringValue(resp.Name))

	if err := d.Set("tags", keyvaluetags.MedialiveKeyValueTags(resp.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}

func resourceAwsMediaLiveInputUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).medialiveconn

	if d.HasChange("destinations") {
		_, err := conn.UpdateInput(&medialive.UpdateInputInput{
			InputId:      aws.String(d.Id()),
			Destinations: expandInputDestinationRequest(d.Get("destinations").(*schema.Set).List()),
		})
		if err != nil {
			return fmt.Errorf("error updating MediaLive Input: %s", err)
		}
	}

	if d.HasChange("input_security_groups") {
		_, err := conn.UpdateInput(&medialive.UpdateInputInput{
			InputId:             aws.String(d.Id()),
			InputSecurityGroups: expandStringList(d.Get("input_security_groups").(*schema.Set).List()),
		})
		if err != nil {
			return fmt.Errorf("error updating MediaLive Input: %s", err)
		}
	}

	if d.HasChange("name") {
		_, err := conn.UpdateInput(&medialive.UpdateInputInput{
			InputId: aws.String(d.Id()),
			Name:    aws.String(d.Get("name").(string)),
		})
		if err != nil {
			return fmt.Errorf("error updating MediaLive Input: %s", err)
		}
	}

	arn := d.Get("arn").(string)
	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.MedialiveUpdateTags(conn, arn, o, n); err != nil {
			return fmt.Errorf("error updating MediaPackage Channel (%s) tags: %s", arn, err)
		}
	}

	return resourceAwsMediaLiveInputRead(d, meta)
}

func resourceAwsMediaLiveInputDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).medialiveconn

	input := &medialive.DeleteInputInput{
		InputId: aws.String(d.Id()),
	}

	_, err := conn.DeleteInput(input)
	if err != nil {
		if isAWSErr(err, medialive.ErrCodeNotFoundException, "") {
			return nil
		}
		return fmt.Errorf("error deleting MediaLive Input: %s", err)
	}

	dinput := &medialive.DescribeInputInput{
		InputId: aws.String(d.Id()),
	}
	err = resource.Retry(30*time.Second, func() *resource.RetryError {
		resp, err := conn.DescribeInput(dinput)
		if err != nil {
			if isAWSErr(err, medialive.ErrCodeNotFoundException, "") {
				return nil
			}
			return resource.NonRetryableError(err)
		}
		if aws.StringValue(resp.State) == "DELETED" {
			return nil
		}
		return resource.RetryableError(fmt.Errorf("MediaLive Input (%s) still exists", d.Id()))
	})
	if isResourceTimeoutError(err) {
		_, err = conn.DescribeInput(dinput)
	}
	if err != nil {
		return fmt.Errorf("error waiting for MediaLive Input (%s) deletion: %s", d.Id(), err)
	}

	return nil
}

func expandInputDestinationRequest(v []interface{}) []*medialive.InputDestinationRequest {
	var destinations []*medialive.InputDestinationRequest
	for _, dest := range v {
		destinations = append(destinations, &medialive.InputDestinationRequest{
			StreamName: aws.String(dest.(map[string]interface{})["endpoint"].(string)),
		})
	}
	return destinations
}

func expandInputVpcRequest(v []interface{}) *medialive.InputVpcRequest {
	vpc := v[0].(map[string]interface{})

	req := &medialive.InputVpcRequest{
		SubnetIds: expandStringList(vpc["subnet_ids"].([]interface{})),
	}

	if len(vpc["securitygroup_ids"].([]interface{})) != 0 {
		req.SecurityGroupIds = expandStringList(vpc["security_group_ids"].([]interface{}))
	}

	return req
}

func flattenInputDestinations(is []*medialive.InputDestination) []map[string]interface{} {
	var out []map[string]interface{}
	for _, i := range is {
		var destination = map[string]interface{}{
			"endpoint": endpointFromURL(i.Url),
			"ip":       i.Ip,
			"port":     i.Port,
			"url":      i.Url,
		}

		if i.Vpc != nil {
			destination["vpc"] = map[string]interface{}{
				"availability_zone":    i.Vpc.AvailabilityZone,
				"network_interface_id": i.Vpc.NetworkInterfaceId,
			}
		}
		out = append(out, destination)
	}
	return out
}

func endpointFromURL(u *string) string {
	parts := strings.Split(aws.StringValue(u), "/")
	if len(parts) < 4 {
		return ""
	}
	return parts[3]
}
