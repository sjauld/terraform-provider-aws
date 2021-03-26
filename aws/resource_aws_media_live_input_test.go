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

func TestAccAWSMediaLiveInput_RTMPPush(t *testing.T) {
	resourceName := "aws_media_live_input.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSMediaLive(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsMediaLiveInputDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMediaLiveInput_RTMPPushConfig(acctest.RandString(5)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMediaLiveInputExists(resourceName),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "medialive", regexp.MustCompile(`input:[0-9]+`)),
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

func testAccCheckAwsMediaLiveInputDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).medialiveconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_media_live_input" {
			continue
		}

		input := &medialive.DescribeInputInput{
			InputId: aws.String(rs.Primary.ID),
		}

		_, err := conn.DescribeInput(input)
		if err == nil {
			return fmt.Errorf("MediaLive Input (%s) not deleted", rs.Primary.ID)
		}

		if !isAWSErr(err, medialive.ErrCodeNotFoundException, "") {
			return err
		}
	}

	return nil
}

func testAccCheckAwsMediaLiveInputExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := testAccProvider.Meta().(*AWSClient).medialiveconn

		input := &medialive.DescribeInputInput{
			InputId: aws.String(rs.Primary.ID),
		}

		_, err := conn.DescribeInput(input)

		return err
	}
}

func testAccPreCheckAWSMediaLive(t *testing.T) {
	conn := testAccProvider.Meta().(*AWSClient).medialiveconn

	input := &medialive.ListInputsInput{}

	_, err := conn.ListInputs(input)

	if testAccPreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccMediaLiveInput_RTMPPushConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_media_live_input_security_group" "test" {
  ipv4_whitelist = ["192.168.0.1/32"]
  tags = {
    "test": "tf_acceptance"
  }
}

resource "aws_media_live_input" "test" {
  name = "%v"
  type = "RTMP_PUSH"
  input_security_groups = [aws_media_live_input_security_group.test.id]

  destinations {
    endpoint = "test"
  }

  tags = {
    "test" = "tf_acceptance"
  }
}
`, rName)
}
