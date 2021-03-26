package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/medialive"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSMediaLiveChannel_basic(t *testing.T) {
	resourceName := "aws_media_live_channel.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSMediaLive(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsMediaLiveChannelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMediaLiveChannel_basic(acctest.RandString(5)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMediaLiveChannelExists(resourceName),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "medialive", regexp.MustCompile(`channel:[0-9]+`)),
					resource.TestCheckResourceAttr(resourceName, "destinations.0.endpoint", "test"),
					resource.TestCheckResourceAttr(resourceName, "destinations.0.port", "1935"),
					resource.TestCheckResourceAttr(resourceName, "input_class", "STANDARD"),
					resource.TestCheckResourceAttr(resourceName, "input_source_type", "STATIC"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckAwsMediaLiveChannelDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).medialiveconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_media_live_channel" {
			continue
		}

		input := &medialive.DescribeChannelInput{
			ChannelId: aws.String(rs.Primary.ID),
		}

		_, err := conn.DescribeChannel(input)
		if err == nil {
			return fmt.Errorf("MediaLive Channel (%s) not deleted", rs.Primary.ID)
		}

		if !isAWSErr(err, medialive.ErrCodeNotFoundException, "") {
			return err
		}
	}

	return nil
}

func testAccCheckAwsMediaLiveChannelExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := testAccProvider.Meta().(*AWSClient).medialiveconn

		input := &medialive.DescribeChannelInput{
			ChannelId: aws.String(rs.Primary.ID),
		}

		_, err := conn.DescribeChannel(input)

		return err
	}
}

func testAccMediaLiveChannel_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_media_live_input_security_group" "test" {
  ipv4_whitelist = ["192.168.0.1/32"]
  tags = {
    "test": "tf_acceptance"
  }
}

resource "aws_media_live_input" "test" {
  name = "test"
  type = "RTMP_PUSH"
  input_security_groups = [aws_media_live_input_security_group.test.id]

  destinations {
    endpoint = "test"
  }

  tags = {
    "test" = "tf_acceptance"
  }
}

resource "aws_media_package_channel" "test" {
  channel_id  = "test"
  description = "test"

  tags = {
    "test" = "tf_acceptance"
  }
}

resource "aws_media_live_channel" "test" {
  name = "%v"

  class = "SINGLE_PIPELINE"

  destination {
	type = "media_package"
	media_package_channel_ids = [aws_media_package_channel.test.channel_id]
  }

  input_attachment {
    name = "test"
    input_id = aws_media_live_input.test.id
  }

  tags = {
    "test" = "tf_acceptance"
  }
}
`, rName)
}
